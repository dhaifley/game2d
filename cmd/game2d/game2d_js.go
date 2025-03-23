//go:build js

package main

import (
	"syscall/js"

	"github.com/dhaifley/game2d/client"
)

// initJS initializes the JavaScript API for the game2d client.
func initJS(g *client.Game) {
	setGameID := func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return 1
		}

		g.SetID(args[0].String())

		return 0
	}

	js.Global().Set("setGameID", js.FuncOf(setGameID))

	setGameName := func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return 1
		}

		g.SetName(args[0].String())

		return 0
	}

	js.Global().Set("setGameName", js.FuncOf(setGameName))

	setAPIURL := func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return 1
		}

		g.SetAPIURL(args[0].String())

		return 0
	}

	js.Global().Set("setAPIURL", js.FuncOf(setAPIURL))

	setAPIToken := func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return 1
		}

		g.SetAPIToken(args[0].String())

		return 0
	}

	js.Global().Set("setAPIToken", js.FuncOf(setAPIToken))
}
