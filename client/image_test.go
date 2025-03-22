package client_test

import (
	"encoding/json"
	"testing"

	"github.com/dhaifley/game2d/client"
	"github.com/stretchr/testify/assert"
)

var TestImage = []byte("")

func TestNewImage(t *testing.T) {
	image := client.NewImage(TestID, TestName, TestImage, 0, 0)
	assert.NotNil(t, image, "Image should not be nil")
}

func TestImageJSONMarshaling(t *testing.T) {
	originalImage := client.NewImage(TestID, TestName, TestImage, 0, 0)

	data, err := json.Marshal(originalImage)
	assert.NoError(t, err, "Marshal should not return an error")

	var newImage client.Image

	err = json.Unmarshal(data, &newImage)
	assert.NoError(t, err, "Unmarshal should not return an error")

	originalJSON, err := json.Marshal(originalImage)
	assert.NoError(t, err)

	newJSON, err := json.Marshal(&newImage)
	assert.NoError(t, err)

	assert.JSONEq(t, string(originalJSON), string(newJSON),
		"Original and unmarshaled images should be equal")
}
