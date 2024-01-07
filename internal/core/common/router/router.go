package router

import (
	"context"
	_ "fernandoglatz/url-management/docs"
	"fernandoglatz/url-management/internal/controller"
	"fernandoglatz/url-management/internal/core/common/utils/log"
	"fernandoglatz/url-management/internal/core/service"
	"fernandoglatz/url-management/internal/infrastructure/config"
	"fernandoglatz/url-management/internal/infrastructure/repository"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup(ctx context.Context, engine *gin.Engine) {
	log.Info(ctx).Msg("Configuring routes")

	contextPath := config.ApplicationConfig.Server.ContextPath
	router := engine.Group(contextPath)

	redirectRepository := repository.NewRedirectCacheRepository(repository.NewRedirectRepository())
	redirectService := service.NewRedirectService(redirectRepository)
	redirectController := controller.NewRedirectController(redirectService)

	healthController := controller.NewHealthController()

	engine.GET("", redirectController.Execute)
	router.GET("", redirectController.Execute)
	router.GET("/", redirectController.Execute) //swagger
	routerRedirect := router.Group("/redirect")
	routerRedirect.GET("", redirectController.Get)
	routerRedirect.GET(":id", redirectController.GetId)
	routerRedirect.PUT("", redirectController.Put)
	routerRedirect.PUT(":id", redirectController.PutId)
	routerRedirect.POST(":id", redirectController.Post)
	routerRedirect.DELETE(":id", redirectController.DeleteId)

	router.GET("/health", healthController.Health)
	router.GET("/swagger-ui/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	engine.NoRoute(redirectController.NoRoute)

	log.Info(ctx).Msg("Routes configured")
}
