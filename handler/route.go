package handler

import (
	"github.com/gofiber/fiber/v2"
)

const (
	Register = "Register"
)

// AddRoutes 使用 app 注册路由, h 持有路由对应的处理方法
func AddRoutes(app *fiber.App, h *Handler) {
	v1 := app.Group("/v1")
	{
		v1.Post("/users", h.Register).Name(Register)
	}
}
