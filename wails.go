package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func NewApp() *App {
	return &App{}
}

//go:embed all:frontend/dist
var assets embed.FS

var MainApp *App

func StartWails() {
	go StartMediaServer()
	MainApp = NewApp()

	err := wails.Run(&options.App{
		Title:     "Vid2Vert",
		Width:     1024,
		MaxWidth:  1920 * 4,
		MaxHeight: 1080 * 4,
		Height:    768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        MainApp.startup,
		OnShutdown:       MainApp.shutdown,
		Bind: []interface{}{
			MainApp,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
