package main

import (
	"context"
	"log"

	drawhandler "backend/internal/adapter/http/handler"
	"backend/internal/app"
)

// main.go で使用する依存の差し替えポイントを集約したファイル

type containerFactory func(ctx context.Context) (*app.Container, error)

type routerFactory func(drawHandler *drawhandler.DrawHandler, postHandler *drawhandler.PostHandler) routerRunner

type routerRunner interface {
	Run(...string) error
}

type containerCloser func(container *app.Container) error

var (
	newContainer containerFactory = app.NewContainer
	newRouter    routerFactory    = func(drawHandler *drawhandler.DrawHandler, postHandler *drawhandler.PostHandler) routerRunner {
		return drawhandler.NewRouter(drawHandler, postHandler)
	}
	closeContainer containerCloser = func(container *app.Container) error {
		return container.Close()
	}
	runFunc = run
	fatalf  = log.Fatalf
)
