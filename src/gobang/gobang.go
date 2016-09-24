package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
)

func main() {
	e := echo.New()
	e.Static("/", "public")
	e.Run(standard.New(":8011"))
}
