package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"io"

	"github.com/dhaifley/game2d/errors"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// Image values represent the images in the game.
type Image struct {
	id, name string
	w, h     int
	data     []byte
	img      *ebiten.Image
}

// NewImage creates and initializes a new image object.
func NewImage(id, name string, data []byte, w, h int) *Image {
	var i *ebiten.Image

	if len(data) > 0 {
		img, err := svgToImage(bytes.NewBuffer(data), w, h)
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

// svgToImage converts an SVG image from an io.Reader into an image.Image.
func svgToImage(r io.Reader, width, height int) (image.Image, error) {
	svgData, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrClient,
			"unable to read SVG data")
	}

	icon, err := oksvg.ReadIconStream(bytes.NewReader(svgData))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrClient,
			"unable to parse SVG data")
	}

	if width <= 0 || height <= 0 {
		w, h := int(icon.ViewBox.W), int(icon.ViewBox.H)

		if w <= 0 || h <= 0 {
			w, h = 256, 256
		}

		if width <= 0 && height > 0 {
			width = height * w / h
		} else if height <= 0 && width > 0 {
			height = width * h / w
		} else if width <= 0 && height <= 0 {
			width, height = w, h
		}
	}

	icon.SetTarget(0, 0, float64(width), float64(height))

	rgba := image.NewRGBA(image.Rect(0, 0, width, height))

	scanner := rasterx.NewScannerGV(width, height, rgba, rgba.Bounds())

	dasher := rasterx.NewDasher(width, height, scanner)

	icon.Draw(dasher, 1.0)

	return rgba, nil
}

// MarshalJSON serializes the image to JSON.
func (i *Image) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		W    int    `json:"w"`
		H    int    `json:"h"`
		Data string `json:"data,omitempty"`
	}{
		ID:   i.id,
		Name: i.name,
		W:    i.w,
		H:    i.h,
		Data: base64.StdEncoding.EncodeToString(i.data),
	})
}

// UnmarshalJSON deserializes the image from JSON.
func (i *Image) UnmarshalJSON(data []byte) error {
	v := &struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		W    int    `json:"w"`
		H    int    `json:"h"`
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
		img, err := svgToImage(bytes.NewBuffer(i.data), v.W, v.H)
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
