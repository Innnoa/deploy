package main

import (
	"context"
	"fmt"
	"recovery-unit-deploy/service"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
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
// GetComputerInfo 返回计算机信息，导出给前端调用
func (a *App) GetComputerInfo() service.ComputerInfo {
	var computer service.ComputerInfo
	return computer.GetComputerInfo()
}

func (a *App) GetAllPackages() service.PackageInfo {
	var packages service.PackageInfo
    packages.GetAllPackages("") // Pass an empty string or appropriate value
    return packages
}