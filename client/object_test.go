package client_test

import (
	"encoding/json"
	"testing"

	"github.com/dhaifley/game2d/client"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewObject(t *testing.T) {
	object := client.NewObject(nil, TestID, TestName, "", "", nil)
	assert.NotNil(t, object, "Object should not be nil")
}

func TestObjectUpdate(t *testing.T) {
	object := client.NewObject(client.NewGame(nil, 800, 600,
		TestID, TestName, TestDesc), TestID, TestName, "", "", nil)

	err := object.Update()
	assert.NoError(t, err, "Update should not return an error")
}

func TestObjectDraw(t *testing.T) {
	object := client.NewObject(nil, TestID, TestName, "", "", nil)

	object.Draw(ebiten.NewImage(640, 480))
}

func TestObjectLayout(t *testing.T) {
	object := client.NewObject(nil, TestID, TestName, "", "", nil)

	w, h := object.Layout(800, 600)
	assert.Equal(t, 0, w, "Width should be 0")
	assert.Equal(t, 0, h, "Height should be 0")
}

func TestObjectJSONMarshaling(t *testing.T) {
	originalObject := client.NewObject(nil, TestID, TestName,
		"test script", "", map[string]any{"score": 42})

	data, err := json.Marshal(originalObject)
	assert.NoError(t, err, "Marshal should not return an error")

	var newObject client.Object

	err = json.Unmarshal(data, &newObject)
	assert.NoError(t, err, "Unmarshal should not return an error")

	originalJSON, err := json.Marshal(originalObject)
	assert.NoError(t, err)

	newJSON, err := json.Marshal(&newObject)
	assert.NoError(t, err)

	assert.JSONEq(t, string(originalJSON), string(newJSON),
		"Original and unmarshaled objects should be equal")
}
