package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	"gobang"
	"html/template"
	"io"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	// debug
	e.SetDebug(true)
	e.Use(middleware.Static("public"))
	e.Use(middleware.Logger())
	t := &Template{
		templates: template.Must(template.New("").Funcs(template.FuncMap{
			"loop": func(n int) []int {
				slice := make([]int, n)
				for i := 0; i < n; i++ {
					slice[i] = i + 1
				}
				return slice
			},
		}).ParseGlob("template/*.html")),
	}
	e.SetRenderer(t)
	e.GET("/status", standard.WrapHandler(gobang.HandleStatusSocket()))
	e.GET("/socket", standard.WrapHandler(gobang.HandleGameSocket()))
	e.GET("/game", gobang.Game)
	e.Run(standard.New(":8011"))
}
