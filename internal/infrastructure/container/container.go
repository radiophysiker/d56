package container

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/radiophysiker/d56/internal/config"
	"github.com/radiophysiker/d56/internal/domain/order"
	"github.com/radiophysiker/d56/internal/domain/user"
	"github.com/radiophysiker/d56/internal/domain/withdrawal"
	"github.com/radiophysiker/d56/internal/infrastructure/accrual"
	"github.com/radiophysiker/d56/internal/infrastructure/database"
	"github.com/radiophysiker/d56/internal/infrastructure/http/handler"
	"github.com/radiophysiker/d56/internal/infrastructure/http/middleware"
	"github.com/radiophysiker/d56/internal/infrastructure/http/router"
	"github.com/radiophysiker/d56/internal/infrastructure/jwt"
	"github.com/radiophysiker/d56/internal/infrastructure/password"
	"github.com/radiophysiker/d56/internal/infrastructure/repository/postgres"
	"github.com/radiophysiker/d56/internal/service"
)

// Container содержит все зависимости приложения
type Container struct {
	// Infrastructure
	dbConnection    *database.DatabaseConnection
	passwordService user.PasswordService
	jwtService      *jwt.Service
	accrualClient   *accrual.Client
	authMiddleware  *middleware.AuthMiddleware
	httpRouter      *router.Router

	// Repositories (Infrastructure Layer)
	userRepo       user.Repository
	orderRepo      order.Repository
	withdrawalRepo withdrawal.Repository

	// Services (Application Layer)
	userService             *service.UserService
	orderService            *service.OrderService
	withdrawalService       *service.WithdrawalService
	accrualProcessorService *service.AccrualProcessorService

	// Handlers (Presentation Layer)
	userHandler *handler.UserHandler
}

// NewContainer создает новый контейнер зависимостей
func NewContainer(cfg *config.Config, logger *zap.Logger) (*Container, error) {
	container := &Container{}

	// Initialize Infrastructure Layer
	if err := container.initInfrastructure(cfg, logger); err != nil {
		return nil, err
	}

	// Initialize Repositories
	container.initRepositories()

	// Initialize Services (Application Layer)
	container.initServices(logger)

	// Initialize Handlers (Presentation Layer)
	container.initHandlers()

	// Initialize Router (Infrastructure Layer - HTTP)
	container.initRouter(logger)

	return container, nil
}

// initInfrastructure инициализирует инфраструктурный слой
func (c *Container) initInfrastructure(cfg *config.Config, logger *zap.Logger) error {
	// Database connection
	dbConn, err := database.NewPostgreSQLConnection(cfg.DatabaseURI)
	if err != nil {
		return err
	}
	c.dbConnection = dbConn

	// Password service
	c.passwordService = password.NewBcryptService()

	// JWT service
	jwtService, err := jwt.NewService(cfg.JWTSecretKey)
	if err != nil {
		return fmt.Errorf("failed to initialize JWT service: %w", err)
	}
	c.jwtService = jwtService

	// Accrual client (if configured)
	if cfg.AccrualSystemAddress != "" {
		c.accrualClient = accrual.NewClient(cfg.AccrualSystemAddress)
	}

	// Auth middleware
	c.authMiddleware = middleware.NewAuthMiddleware(c.jwtService)

	return nil
}

// initRepositories инициализирует репозитории
func (c *Container) initRepositories() {
	db := c.dbConnection.GetDB()

	c.userRepo = postgres.NewUserRepository(db)
	c.orderRepo = postgres.NewOrderRepository(db)
	c.withdrawalRepo = postgres.NewWithdrawalRepository(db)
}

// initServices инициализирует сервисы
func (c *Container) initServices(logger *zap.Logger) {
	txManager := c.dbConnection.GetTransactionManager()

	c.userService = service.NewUserService(c.userRepo, c.passwordService)
	c.orderService = service.NewOrderService(c.orderRepo, c.userRepo, txManager)
	c.withdrawalService = service.NewWithdrawalService(c.withdrawalRepo, c.userRepo, txManager)

	// Accrual processor service (if accrual client is available)
	if c.accrualClient != nil {
		c.accrualProcessorService = service.NewAccrualProcessorService(
			c.accrualClient, c.orderService, logger)
	}
}

// initHandlers инициализирует хендлеры
func (c *Container) initHandlers() {
	c.userHandler = handler.NewUserHandler(
		c.userService,
		c.orderService,
		c.withdrawalService,
		c.jwtService,
	)
}

// initRouter инициализирует роутер
func (c *Container) initRouter(logger *zap.Logger) {
	c.httpRouter = router.NewRouter(c.userHandler, c.authMiddleware, logger)
}

// Getters для доступа к зависимостям

func (c *Container) GetUserHandler() *handler.UserHandler {
	return c.userHandler
}

func (c *Container) GetAccrualProcessorService() *service.AccrualProcessorService {
	return c.accrualProcessorService
}

func (c *Container) GetDatabaseConnection() *database.DatabaseConnection {
	return c.dbConnection
}

func (c *Container) GetAuthMiddleware() *middleware.AuthMiddleware {
	return c.authMiddleware
}

func (c *Container) GetRouter() *router.Router {
	return c.httpRouter
}

// Close закрывает все ресурсы контейнера
func (c *Container) Close() error {
	if c.dbConnection != nil {
		return c.dbConnection.Close()
	}
	return nil
}
