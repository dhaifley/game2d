//go:build !js

package main

import (
	_ "image/png"

	"github.com/dhaifley/game2d/client"
)

// initJS initializes the JavaScript API for the game2d client.
func initJS(g *client.Game) {}
