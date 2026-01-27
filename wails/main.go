package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend
var assets embed.FS

// main is the entrypoint for the Wails application
func main() {
	// Create an instance of the app structure
	app, err := NewApp()
	if err != nil {
		println("Failed to create app:", err.Error())
		return
	}

	// Create application with options
	// In Wails v2, exported methods on the app struct are automatically bound to the frontend
	err = wails.Run(&options.App{
		Title:  "Constrictor REST Client",
		Width:  1400,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 19, G: 19, B: 20, A: 1},
		OnStartup:        app.OnStartup,
		// Bind the app instance - exported methods will be available in frontend
		Bind: []interface{}{app},
	})

	if err != nil {
		println("Failed to run Wails app:", err.Error())
	}
}
