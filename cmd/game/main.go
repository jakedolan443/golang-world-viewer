package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"shipping/internal/config"
	"shipping/internal/game"
)

func main() {
	cfg, err := config.LoadEnv(".env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting %s (quality: %s, window: %dx%d, render: %s)...",
		cfg.Title, cfg.Quality,
		cfg.Display.WindowWidth, cfg.Display.WindowHeight,
		cfg.Display.Resolution.Name)

	g, err := game.New(cfg)
	if err != nil {
		log.Fatalf("Error: %v\n\nEnsure %s exists", err, cfg.Map.ShapefilePath)
	}

	ebiten.SetWindowTitle(cfg.Title)
	g.ApplyDisplay()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
