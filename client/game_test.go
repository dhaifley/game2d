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
	game := client.NewGame(nil, client.DefaultGameWidth,
		client.DefaultGameHeight, TestID, TestName, TestDesc)
	assert.NotNil(t, game, "Game should not be nil")
}

func TestUpdate(t *testing.T) {
	game := client.NewGame(nil, client.DefaultGameWidth,
		client.DefaultGameHeight, TestID, TestName, TestDesc)

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
	game := client.NewGame(nil, client.DefaultGameWidth,
		client.DefaultGameHeight, TestID, TestName, TestDesc)

	game.Draw(ebiten.NewImage(client.DefaultGameWidth,
		client.DefaultGameHeight))
}

func TestLayout(t *testing.T) {
	game := client.NewGame(nil, client.DefaultGameWidth,
		client.DefaultGameHeight, TestID, TestName, TestDesc)

	w, h := game.Layout(client.DefaultGameWidth, client.DefaultGameHeight)
	assert.Equal(t, client.DefaultGameWidth, w, "Width should be %d",
		client.DefaultGameWidth)
	assert.Equal(t, client.DefaultGameHeight, h, "Height should be %d",
		client.DefaultGameHeight)
}

func TestGameJSONMarshaling(t *testing.T) {
	originalGame := client.NewGame(nil, client.DefaultGameWidth,
		client.DefaultGameHeight, TestID, TestName, TestDesc)

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
	game := client.NewGame(nil, client.DefaultGameWidth,
		client.DefaultGameHeight, TestID, TestName, TestDesc)

	game.AddScript(client.NewScript(TestID, TestName, TestScript, nil))
	game.AddImage(client.NewImage(TestID, TestName, TestImage))
	game.AddSubject(client.NewSubject(game, TestID, TestName, TestID, TestID,
		nil))
	game.AddObject(client.NewObject(game, TestID, TestName, TestID, TestID,
		nil))

	au := game.APIURL()

	game.SetAPIURL("")
	game.SetAPIToken(game.APIToken())
	game.SetH(game.H())
	game.SetW(game.W())
	game.SetID(game.ID())
	game.SetName(game.Name())

	err := game.Save()
	assert.NoError(t, err)

	t.Cleanup(func() {
		game.SetAPIURL(au)
		os.Remove("test.json")
	})

	err = game.Load()
	assert.NoError(t, err)
}
