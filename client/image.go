package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image/png"

	"github.com/dhaifley/game2d/errors"
	"github.com/hajimehoshi/ebiten/v2"
)

// Image values represent the images in the game.
type Image struct {
	id, name string
	data     []byte
	img      *ebiten.Image
}

// MarshalJSON serializes the image to JSON.
func (i *Image) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Data string `json:"data,omitempty"`
	}{
		ID:   i.id,
		Name: i.name,
		Data: base64.StdEncoding.EncodeToString(i.data),
	})
}

// UnmarshalJSON deserializes the image from JSON.
func (i *Image) UnmarshalJSON(data []byte) error {
	v := &struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Data string `json:"data,omitempty"`
	}{}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	i.id = v.ID
	i.name = v.Name

	b, err := base64.StdEncoding.DecodeString(v.Data)
	if err != nil {
		return errors.Wrap(err, errors.ErrClient,
			"unable to base64 decode image",
			"id", i.id,
			"name", i.name)
	}

	i.data = b

	if len(i.data) > 0 {
		img, err := png.Decode(bytes.NewBuffer(i.data))
		if err != nil {
			return errors.Wrap(err, errors.ErrClient,
				"unable to decode image",
				"id", i.id,
				"name", i.name)
		}

		i.img = ebiten.NewImageFromImage(img)
	} else {
		i.img = nil
	}

	return nil
}

// NewImage creates and initializes a new image object.
func NewImage(id, name string, data []byte) *Image {
	var i *ebiten.Image

	if len(data) > 0 {
		img, err := png.Decode(bytes.NewBuffer(data))
		if err == nil {
			i = ebiten.NewImageFromImage(img)
		}
	}

	return &Image{
		id:   id,
		name: name,
		data: data,
		img:  i,
	}
}
