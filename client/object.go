package client

import (
	"bytes"
	"encoding/json"

	"github.com/dhaifley/game2d/errors"
	"github.com/hajimehoshi/ebiten/v2"
)

// Object values represent the objects in the game.
type Object struct {
	game          *Game
	sub, hidden   bool
	w, h, x, y, z int
	id, name      string
	src, img      string
	data          map[string]any
}

// NewObject creates and initializes a new object.
func NewObject(
	game *Game,
	id, name string,
	src, img string,
	data map[string]any,
) *Object {
	w, h := 0, 0

	if img != "" && game != nil {
		if i, ok := game.img[img]; ok && i != nil && i.img != nil {
			w = i.img.Bounds().Size().X
			h = i.img.Bounds().Size().Y
		}
	}

	return &Object{
		game: game,
		w:    w,
		h:    h,
		id:   id,
		name: name,
		src:  src,
		img:  img,
		data: data,
	}
}

// NewSubject creates and initializes a new subject object.
func NewSubject(
	game *Game,
	id, name string,
	src, img string,
	data map[string]any,
) *Object {
	sub := NewObject(game, id, name, src, img, data)

	sub.sub = true

	return sub
}

// SetHidden sets the object hidden state.
func (o *Object) SetHidden(hidden bool) {
	o.hidden = hidden
}

// SetName sets the object name.
func (o *Object) SetName(name string) {
	o.name = name
}

// SetX sets the object x-coordinate.
func (o *Object) SetX(x int) {
	o.x = x
}

// SetY sets the object y-coordinate.
func (o *Object) SetY(y int) {
	o.y = y
}

// SetZ sets the object z-index.
func (o *Object) SetZ(z int) {
	o.z = z
}

// SetW sets the object width.
func (o *Object) SetW(w int) {
	o.w = w
}

// SetH sets the object height.
func (o *Object) SetH(h int) {
	o.h = h
}

// SetImage sets the object image.
func (o *Object) SetImage(img string) {
	o.img = img

	if i, ok := o.game.img[img]; ok && i != nil && i.img != nil {
		o.w = i.img.Bounds().Size().X
		o.h = i.img.Bounds().Size().Y
	}
}

// SetScript sets the object script.
func (o *Object) SetScript(src string) {
	o.src = src
}

// SetData sets the object data.
func (o *Object) SetData(data map[string]any) {
	o.data = data
}

func (o *Object) Map() map[string]any {
	return map[string]any{
		"id":      o.id,
		"name":    o.name,
		"hidden":  o.hidden,
		"subject": o.sub,
		"x":       o.x,
		"y":       o.y,
		"z":       o.z,
		"w":       o.w,
		"h":       o.h,
		"image":   o.img,
		"script":  o.src,
		"data":    o.data,
	}
}

// NewObjectFromMap creates a new object from a map.
func NewObjectFromMap(m map[string]any) *Object {
	hidden, _ := m["hidden"].(bool)
	id, _ := m["id"].(string)
	name, _ := m["name"].(string)
	src, _ := m["script"].(string)
	img, _ := m["image"].(string)
	data, _ := m["data"].(map[string]any)
	sub, _ := m["subject"].(bool)
	x, _ := m["x"].(float64)
	y, _ := m["y"].(float64)
	z, _ := m["z"].(float64)
	w, _ := m["w"].(float64)
	h, _ := m["h"].(float64)

	if id == "" {
		return nil
	}

	return &Object{
		id:     id,
		name:   name,
		hidden: hidden,
		src:    src,
		img:    img,
		data:   data,
		sub:    sub,
		x:      int(x),
		y:      int(y),
		z:      int(z),
		w:      int(w),
		h:      int(h),
	}
}

// MarshalJSON serializes the object to JSON.
func (o *Object) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID     string         `json:"id"`
		Name   string         `json:"name"`
		Hidden bool           `json:"hidden"`
		X      int            `json:"x"`
		Y      int            `json:"y"`
		Z      int            `json:"z"`
		W      int            `json:"w"`
		H      int            `json:"h"`
		Image  string         `json:"image,omitempty"`
		Script string         `json:"script,omitempty"`
		Data   map[string]any `json:"data,omitempty"`
	}{
		ID:     o.id,
		Name:   o.name,
		Hidden: o.hidden,
		X:      o.x,
		Y:      o.y,
		Z:      o.z,
		W:      o.w,
		H:      o.h,
		Image:  o.img,
		Script: o.src,
		Data:   o.data,
	})
}

// UnmarshalJSON deserializes the object from JSON.
func (o *Object) UnmarshalJSON(data []byte) error {
	v := &struct {
		ID     string         `json:"id"`
		Name   string         `json:"name"`
		Hidden bool           `json:"hidden"`
		X      int            `json:"x"`
		Y      int            `json:"y"`
		Z      int            `json:"z"`
		W      int            `json:"w"`
		H      int            `json:"h"`
		Image  string         `json:"image,omitempty"`
		Script string         `json:"script,omitempty"`
		Data   map[string]any `json:"data,omitempty"`
	}{}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	o.id = v.ID
	o.name = v.Name
	o.hidden = v.Hidden
	o.x = v.X
	o.y = v.Y
	o.z = v.Z
	o.w = v.W
	o.h = v.H
	o.src = v.Script
	o.img = v.Image
	o.data = v.Data

	return nil
}

// Update updates the object state each frame.
func (o *Object) Update() error {
	if o.src == "" || o.game == nil || o.game.lua == nil || o.game.src == nil {
		return nil
	}

	src, ok := o.game.src[o.src]
	if !ok || src == nil {
		return errors.New(errors.ErrClient,
			"script not found",
			"object_id", o.id,
			"script_id", o.src)
	}

	l := o.game.lua

	buf := bytes.NewBufferString(src.data)

	if err := l.Load(buf, "Update", "text"); err != nil {
		return errors.Wrap(err, errors.ErrClient,
			"unable to load script",
			"object_id", o.id,
			"script_id", o.src,
			"script", string(src.data))
	}

	l.Call(0, 0)

	l.Global("Update")

	if !l.IsFunction(-1) {
		return errors.New(errors.ErrClient,
			"no Update function in script",
			"object_id", o.id,
			"script_id", o.src,
			"script", string(src.data))
	}

	d := map[string]any{"id": o.id}

	pushMap(l, d)

	l.Call(1, 0)

	return nil
}

// Draw renders the object each frame.
func (o *Object) Draw(screen *ebiten.Image) {
	if o.hidden || o.img == "" || o.game == nil || o.game.img == nil {
		return
	}

	if !o.sub && o.game.sub != nil {
		if o.x-o.game.sub.x > screen.Bounds().Dx() ||
			o.x-o.game.sub.x < -screen.Bounds().Dx() ||
			o.y-o.game.sub.y < -screen.Bounds().Dy() ||
			o.y-o.game.sub.y > screen.Bounds().Dy() {
			return
		}
	}

	geo := ebiten.GeoM{}
	geo.Translate(float64(o.x), float64(o.y))

	op := &ebiten.DrawImageOptions{GeoM: geo}

	img := o.game.img[o.img]
	if img == nil || img.img == nil {
		return
	}

	screen.DrawImage(img.img, op)
}

// Layout returns the object dimensions.
func (o *Object) Layout(w, h int) (int, int) {
	return o.w, o.h
}
