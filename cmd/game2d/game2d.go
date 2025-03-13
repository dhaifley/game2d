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

// main initializes and starts the game.
func main() {
	ctx := context.Background()

	log := logger.New(logger.OutStderr, logger.FmtJSON,
		logger.LvlDebug)

	g := client.NewGame(log, 800, 600, "game2d", "game2d",
		"2D gaming framework")

	ib, err := assets.GetImage("kefka.png")
	if err != nil {
		log.Log(ctx, logger.LvlError,
			"unable to read image",
			"error", err,
			"file", "kefka.png")

		os.Exit(1)
	}

	g.AddImage(client.NewImage("p1", "kefka.png", ib))

	script, err := assets.GetScript("kefka.lua")
	if err != nil {
		log.Log(ctx, logger.LvlError,
			"unable to read script",
			"error", err,
			"file", "kefka.lua")

		os.Exit(1)
	}

	g.AddScript(client.NewScript("p1", "kefka.lua", script, []string{"image"}))

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
