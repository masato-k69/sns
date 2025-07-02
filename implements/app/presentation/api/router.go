package presentation

import (
	api "app/gen/api"
	apiv1 "app/gen/api/v1"
	impl "app/presentation/api/implement"
	implv1 "app/presentation/api/implement/v1"
	"os"

	"github.com/labstack/echo/v4"
)

func NewRouter() (*echo.Echo, error) {
	router := echo.New()
	router.HideBanner = true
	router.Use(NewMiddlewaresCommon()...)

	routerGroup := router.Group(os.Getenv("ROUTER_GROUP"))
	routerGroup.Use(NewMiddlewares()...)
	handler := impl.NewHandler()
	api.RegisterHandlers(routerGroup, &handler)

	routerGroupV1 := router.Group(os.Getenv("ROUTER_GROUP_V1"))
	routerGroupV1.Use(NewMiddlewaresForV1()...)
	handlerv1 := implv1.NewHandler()
	apiv1.RegisterHandlers(routerGroupV1, &handlerv1)

	return router, nil
}
