package client

import (
	"encoding/hex"
	"encoding/json"
	"maps"
	"slices"
	"strconv"

	"github.com/Shopify/go-lua"
	"github.com/dhaifley/empty/errors"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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

	if err := lua.LoadString(o.game.lua, string(src.data)); err != nil {
		return errors.Wrap(err, errors.ErrClient,
			"unable to load script",
			"object_id", o.id,
			"script_id", o.src,
			"script", string(src.data))
	}

	o.game.lua.Call(0, 0)

	o.game.lua.Global("update")

	if !o.game.lua.IsFunction(1) {
		return errors.New(errors.ErrClient,
			"no update function in script",
			"object_id", o.id,
			"script_id", o.src)
	}

	d := map[string]any{
		"x": o.x,
		"y": o.y,
		"z": o.z,
		"w": o.w,
		"h": o.h,
	}

	d["keys"] = []any{}

	if keys := inpututil.AppendPressedKeys(nil); len(keys) > 0 {
		s := make([]any, len(keys))

		for i, k := range keys {
			s[i] = int(k)
		}

		d["keys"] = s
	}

	incImg, incScript := false, false

	for _, inc := range src.include {
		switch inc {
		case "image":
			if o.img == "" {
				continue
			}

			if i, ok := o.game.img[o.img]; ok && i != nil {
				if len(i.data) > 0 {
					d["image"] = hex.EncodeToString(i.data)
				}

				incImg = true
			}
		case "script":
			if o.src == "" {
				continue
			}

			if s, ok := o.game.src[o.src]; ok && s != nil {
				d["script"] = s.data
				incScript = true
			}
		}
	}

	delete(o.data, "keys")

	maps.Copy(d, o.data)

	pushMap(o.game.lua, d)

	o.game.lua.Call(1, 1)

	r, err := tableToMap(o.game.lua, -1)
	if err != nil {
		return errors.Wrap(err, errors.ErrClient,
			"invalid return value from script",
			"object_id", o.id,
			"script_id", o.src)
	}

	o.game.lua.SetTop(0)

	if x, ok := r["x"].(float64); ok {
		o.x = int(x)

		delete(r, "x")
	}

	if y, ok := r["y"].(float64); ok {
		o.y = int(y)

		delete(r, "y")
	}

	if z, ok := r["z"].(float64); ok {
		o.z = int(z)

		delete(r, "z")
	}

	if w, ok := r["w"].(float64); ok {
		o.w = int(w)

		delete(r, "w")
	}

	if h, ok := r["h"].(float64); ok {
		o.h = int(h)

		delete(r, "h")
	}

	if incImg {
		if rImg, ok := r["image"].(string); ok {
			ib, err := hex.DecodeString(rImg)
			if err != nil {
				return errors.Wrap(err, errors.ErrClient,
					"unable to decode image",
					"object_id", o.id,
					"image_id", o.img,
					"image", rImg)
			}

			if i, ok := o.game.img[o.img]; ok && i != nil {
				if !slices.Equal(i.data, ib) && i.img != nil {
					i.img.WritePixels(ib)
					o.w = i.img.Bounds().Size().X
					o.h = i.img.Bounds().Size().Y
				}
			}

			delete(r, "image")
		}
	}

	if incScript {
		if rSrc, ok := r["script"].(string); ok {
			if s, ok := o.game.src[o.src]; ok && s != nil {
				s.data = rSrc
			}

			delete(r, "script")
		}
	}

	if keys, ok := r["keys"].(map[string]any); ok {
		var lk []any

		for _, v := range keys {
			lk = append(lk, v)
		}

		if o.data == nil {
			o.data = make(map[string]any)
		}

		if len(lk) > 0 {
			o.data["keys"] = lk
		} else {
			delete(o.data, "keys")
		}

		delete(r, "keys")
	}

	if len(r) > 0 {
		if o.data == nil {
			o.data = make(map[string]any)
		}

		maps.Copy(o.data, r)
	}

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
	geo.Translate(float64(screen.Bounds().Dx()/2-o.w/2),
		float64(screen.Bounds().Dy()/2-o.h/2))

	if o.game.sub != nil {
		geo.Translate(float64(o.x-o.game.sub.x), float64(o.y-o.game.sub.y))
	}

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

// pushMap adds a map to the lua stack as a table.
func pushMap(l *lua.State, m map[string]any) {
	l.NewTable()

	for k, v := range m {
		l.PushString(k)
		pushValue(l, v)
		l.SetTable(-3)
	}
}

// pushSlice pushes a slice to the lua stack as a table.
func pushSlice(l *lua.State, a []any) {
	l.NewTable()

	for i, v := range a {
		l.PushString(strconv.FormatInt(int64(i+1), 10))
		pushValue(l, v)
		l.SetTable(-3)
	}
}

// pushValue pushes a value to the lua stack.
func pushValue(l *lua.State, v any) {
	switch val := v.(type) {
	case byte:
		l.PushInteger(int(val))
	case int:
		l.PushInteger(val)
	case float64:
		l.PushNumber(val)
	case string:
		l.PushString(val)
	case bool:
		l.PushBoolean(val)
	case map[string]any:
		pushMap(l, val)
	case []any:
		pushSlice(l, val)
	case nil:
		l.PushNil()
	default:
		l.PushNil()
	}
}

// Returns a table, from the lua stack, at index, as a map.
func tableToMap(l *lua.State, index int) (map[string]any, error) {
	if !l.IsTable(index) {
		return nil, errors.New(errors.ErrClient,
			"value at index is not a table",
			"index", index)
	}

	result := make(map[string]any)

	l.PushNil()

	for l.Next(index - 1) {
		if l.IsString(-2) {
			key, _ := l.ToString(-2)

			result[key] = getValue(l, -1)
		}

		l.Pop(1)
	}

	return result, nil
}

// getValue returns the value, at index from the lua stack.
func getValue(l *lua.State, index int) any {
	switch l.TypeOf(index) {
	case lua.TypeNil:
		return nil
	case lua.TypeBoolean:
		return l.ToBoolean(index)
	case lua.TypeNumber:
		v, _ := l.ToNumber(index)

		return v
	case lua.TypeString:
		v, _ := l.ToString(index)

		return v
	case lua.TypeTable:
		v, _ := tableToMap(l, index)

		return v
	default:
		return nil
	}
}
