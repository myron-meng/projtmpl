package handler

import (
	"projtmpl/env"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/helmet/v2"
)

// NewApp 新建并初始化 HTTP App
func NewApp(h *Handler) *fiber.App {
	// 初始化并配置 fiber app
	// 配置参数说明参考 https://docs.gofiber.io/api/fiber
	// [Read|Write|Idle]Timeout 参数说明参考：
	// - [So you want to expose Go on the Internet](https://blog.cloudflare.com/exposing-go-on-the-internet/)
	// - [The complete guide to Go net/http timeouts](https://colobu.com/2016/07/01/the-complete-guide-to-golang-net-http-timeouts/) 注：此文章的大多中文翻译有严重的错误
	app := fiber.New(fiber.Config{
		CaseSensitive:         true,
		StrictRouting:         true,
		ReadTimeout:           env.Envs.HTTPReadTimeout,
		WriteTimeout:          env.Envs.HTTPWriteTimeout,
		IdleTimeout:           env.Envs.HTTPIdleTimeout,
		DisableStartupMessage: env.Envs.Tier == env.Prod,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return InternalServerError(c, err)
		},
	})

	// 添加中间件
	app.Use(Recover())
	app.Use(CORS())
	app.Use(helmet.New())
	app.Use(RequestID())
	app.Use(Logging())

	// 注册 API 路由
	AddRoutes(app, h)

	return app
}
