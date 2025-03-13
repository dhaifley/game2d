package client_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/dhaifley/game2d/client"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
)

const (
	TestID   = "test"
	TestName = "test"
	TestDesc = "test"
)

func TestNewGame(t *testing.T) {
	game := client.NewGame(nil, 800, 600, TestID, TestName, TestDesc)
	assert.NotNil(t, game, "Game should not be nil")
}

func TestUpdate(t *testing.T) {
	game := client.NewGame(nil, 800, 600, TestID, TestName, TestDesc)

	game.AddScript(client.NewScript(TestID, TestName, TestScript, nil))
	game.AddImage(client.NewImage(TestID, TestName, TestImage))
	game.AddSubject(client.NewSubject(game, TestID, TestName, TestID, TestID,
		nil))
	game.AddObject(client.NewObject(game, TestID, TestName, TestID, TestID,
		nil))

	err := game.Update()
	assert.NoError(t, err, "Update should not return an error")
}

func TestDraw(t *testing.T) {
	game := client.NewGame(nil, 800, 600, TestID, TestName, TestDesc)

	game.Draw(ebiten.NewImage(640, 480))
}

func TestLayout(t *testing.T) {
	game := client.NewGame(nil, 800, 600, TestID, TestName, TestDesc)

	w, h := game.Layout(800, 600)
	assert.Equal(t, 800, w, "Width should be 800")
	assert.Equal(t, 600, h, "Height should be 600")
}

func TestGameJSONMarshaling(t *testing.T) {
	originalGame := client.NewGame(nil, 800, 600, TestID, TestName, TestDesc)

	data, err := json.Marshal(originalGame)
	assert.NoError(t, err, "Marshal should not return an error")

	var newGame client.Game

	err = json.Unmarshal(data, &newGame)
	assert.NoError(t, err, "Unmarshal should not return an error")

	originalJSON, err := json.Marshal(originalGame)
	assert.NoError(t, err)

	newJSON, err := json.Marshal(&newGame)
	assert.NoError(t, err)

	assert.JSONEq(t, string(originalJSON), string(newJSON),
		"Original and unmarshaled values should be equal")
}

func TestGameSaveLoad(t *testing.T) {
	game := client.NewGame(nil, 800, 600, TestID, TestName, TestDesc)

	game.AddScript(client.NewScript(TestID, TestName, TestScript, nil))
	game.AddImage(client.NewImage(TestID, TestName, TestImage))
	game.AddSubject(client.NewSubject(game, TestID, TestName, TestID, TestID,
		nil))
	game.AddObject(client.NewObject(game, TestID, TestName, TestID, TestID,
		nil))

	err := game.Save()
	assert.NoError(t, err)

	t.Cleanup(func() { os.Remove("test.json") })

	err = game.Load()
	assert.NoError(t, err)
}
