package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"gobang"
	"html/template"
	"io"
	"github.com/labstack/echo/middleware"
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
	t := &Template{
		templates:template.Must(template.New("").Funcs(template.FuncMap{
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
	e.GET("/status", standard.WrapHandler(gobang.Status()))
	e.GET("/game", gobang.Game)
	e.Run(standard.New(":8011"))
}
