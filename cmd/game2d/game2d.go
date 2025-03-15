package main

import (
	"context"
	_ "image/png"
	"os"
	"strconv"

	"github.com/dhaifley/game2d/assets"
	"github.com/dhaifley/game2d/client"
	"github.com/dhaifley/game2d/logger"
)

// Main entry point for the game.
func main() {
	ctx := context.Background()

	log := logger.New(logger.OutStderr, logger.FmtJSON,
		logger.LvlDebug)

	gameID := os.Getenv("GAME2D_GAME_ID")

	g := client.NewGame(log, -1, -1, gameID, "game2d", "2D gaming framework")

	ib, err := assets.GetImage("avatar.png")
	if err != nil {
		log.Log(ctx, logger.LvlError,
			"unable to read image",
			"error", err,
			"file", "avatar.png")

		os.Exit(1)
	}

	g.AddImage(client.NewImage("p1", "avatar.png", ib))

	script, err := assets.GetScript("avatar.lua")
	if err != nil {
		log.Log(ctx, logger.LvlError,
			"unable to read script",
			"error", err,
			"file", "avatar.lua")

		os.Exit(1)
	}

	g.AddScript(client.NewScript("p1", "avatar.lua", script, []string{"image"}))

	g.AddSubject(client.NewSubject(g, "p1", "Hello Aaron!", "p1", "p1", nil))

	ibb, err := assets.GetImage("bg.png")
	if err != nil {
		log.Log(ctx, logger.LvlError,
			"unable to read image",
			"error", err,
			"file", "bg.png")

		os.Exit(1)
	}

	g.AddImage(client.NewImage("bg", "bg.png", ibb))

	for i := -6; i < 7; i++ {
		for j := -4; j < 6; j++ {
			ids := "bg_" + strconv.Itoa(i) + "_" + strconv.Itoa(j)

			obj := client.NewObject(g, ids, ids, "", "bg", nil)
			obj.SetX(i * 64)
			obj.SetY(j * 64)

			g.AddObject(obj)
		}
	}

	if err := g.Run(ctx); err != nil {
		log.Log(ctx, logger.LvlError,
			"game error",
			"error", err)

		os.Exit(1)
	}
}
