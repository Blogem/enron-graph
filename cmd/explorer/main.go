package main

import (
	"database/sql"
	"embed"
	"log"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/pkg/utils"
	_ "github.com/lib/pq"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Load configuration
	cfg, err := utils.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	client, err := ent.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer client.Close()

	// Also open a direct sql.DB connection for raw queries
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open SQL connection: %v", err)
	}
	defer db.Close()

	// Create an instance of the app structure
	app := NewApp(client, db)

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "Enron Graph Explorer",
		Width:  1400,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
