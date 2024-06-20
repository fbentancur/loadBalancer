package main

import (
	"github.com/fbentancur/loadBalancer/controllers"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.Any("/*", controllers.ManejarRequest)

	e.Logger.Fatal(e.Start(":8080"))
}
