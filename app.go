package main

import (
	"context"
	"fmt"
	"recovery-unit-deploy/service/common"
)

// App struct
type App struct {
	ctx       context.Context
	startPage string
}

// NewApp creates a new App application struct
func NewApp(page string) *App {
	return &App{startPage: page}
}

// 供前端调用的方法
func (a *App) GetStartPage() string {
	return a.startPage
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) LogFromFrontend(message string) {
	common.AppLogger.Info(fmt.Sprintln("[Frontend]", message))
}
