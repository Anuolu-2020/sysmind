package main

import (
	"embed"
	"flag"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"sysmind/internal/version"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Handle command line flags
	versionFlag := flag.Bool("version", false, "Print version information and exit")
	verboseFlag := flag.Bool("verbose", false, "Print detailed build information")
	flag.Parse()

	if *versionFlag {
		if *verboseFlag {
			version.PrintBuildInfo()
		} else {
			fmt.Println(version.String())
		}
		os.Exit(0)
	}

	// Print version on startup
	fmt.Printf("Starting %s\n", version.String())

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  fmt.Sprintf("SysMind %s", version.Short()),
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			About: &mac.AboutInfo{
				Title: fmt.Sprintf("SysMind %s", version.Short()),
				Message: fmt.Sprintf("AI-powered system monitoring assistant\nVersion: %s\nBuild: %s", version.Short(), func() string {
					commit := version.Get().GitCommit
					if len(commit) > 8 {
						return commit[:8]
					}
					return commit
				}()),
			},
		},
		Linux: &linux.Options{
			WindowIsTranslucent: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
