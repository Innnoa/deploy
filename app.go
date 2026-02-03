package main

import (
	"context"
	"fmt"
	"recovery-unit-deploy/service/common"
)

// App struct
type App struct {
	ctx        context.Context
	startPage  string
	shouldQuit bool
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

func (a *App) ForceQuit() {
	a.shouldQuit = true
	runtime.Quit(a.ctx)
}

func (a *App) BeforeClose(ctx context.Context) (prevent bool) {
	if a.shouldQuit {
		return false
	}
	// 发送事件给前端，而不是直接弹系统对话框
	runtime.EventsEmit(a.ctx, "onBeforeClose")
	// 返回 true 以阻止窗口立即关闭，等待前端反馈
	return true
}
