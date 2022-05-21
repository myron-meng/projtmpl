package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"projtmpl/env"
	"projtmpl/handler"
	"projtmpl/internal/database"
	"projtmpl/internal/repository"
	"projtmpl/internal/service"
	"projtmpl/pkg/log"
)

func main() {
	// 加载环境变量
	if err := env.Load(); err != nil {
		panic(err)
	}
	// 初始化 logger
	log.Setup()

	// 初始化数据库连接池
	db, err := database.NewConnectionPool()
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("create database connection failed")
	}
	// 初始化 repos
	repos := repository.New(db)
	// 初始化 service
	services := service.New(repos)
	// 初始化 fiber App
	h := handler.Handler{Services: services, Repositories: repos}
	app := handler.NewApp(&h)

	// 接收退出信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-quit
		log.Logger.Info().Str("signal", sig.String()).Msg("exiting...")
		if err = app.Shutdown(); err != nil {
			log.Logger.Fatal().Err(err).Msg("shutdown app failed")
		}
	}()

	// 启动 HTTP 服务
	log.Logger.Info().Msg("starting...")
	if err = app.Listen(fmt.Sprintf(":%d", env.Envs.Port)); err != nil {
		log.Logger.Fatal().Err(err).Msg("start app failed")
	}
	log.Logger.Info().Msg("exited")
}
