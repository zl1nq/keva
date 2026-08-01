package main

import (
	"embed"

	"keva/internal/app"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	kevaApp := app.New()

	err := wails.Run(&options.App{
		Title:  "KEVA",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 243, G: 248, B: 251, A: 1},
		OnStartup:        kevaApp.Startup,
		OnBeforeClose:    kevaApp.BeforeClose,
		OnShutdown:       kevaApp.Shutdown,
		Bind: []interface{}{
			kevaApp,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
