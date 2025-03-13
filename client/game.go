package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/Shopify/go-lua"
	"github.com/dhaifley/game2d/errors"
	"github.com/dhaifley/game2d/logger"
	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// The client version.
var Version = ""

// Game values represent the game state.
type Game struct {
	log       logger.Logger
	debug     bool
	w, h      int
	id        string
	name      string
	ver       string
	desc      string
	lua       *lua.State
	sub       *Object
	obj       map[string]*Object
	img       map[string]*Image
	src       map[string]*Script
	createdBy string
	createdAt int64
	updatedBy string
	updatedAt int64
}

// NewGame creates and initializes a new Game object.
func NewGame(log logger.Logger, w, h int, id, name, desc string) *Game {
	if log == nil || (reflect.ValueOf(log).Kind() == reflect.Ptr &&
		reflect.ValueOf(log).IsNil()) {
		log = logger.NullLog
	}

	l := lua.NewState()
	lua.OpenLibraries(l)

	if id == "" {
		id = uuid.NewString()
	}

	return &Game{
		log:  log,
		w:    w,
		h:    h,
		lua:  l,
		id:   id,
		name: name,
		desc: desc,
		obj:  make(map[string]*Object),
		img:  make(map[string]*Image),
		src:  make(map[string]*Script),
	}
}

// MarshalJSON serializes the game to JSON.
func (g *Game) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID          string             `json:"id"`
		Name        string             `json:"name"`
		Version     string             `json:"version"`
		Description string             `json:"description"`
		Debug       bool               `json:"debug"`
		W           int                `json:"w"`
		H           int                `json:"h"`
		Subject     *Object            `json:"subject,omitempty"`
		Objects     map[string]*Object `json:"objects,omitempty"`
		Images      map[string]*Image  `json:"images,omitempty"`
		Scripts     map[string]*Script `json:"scripts,omitempty"`
		CreatedAt   int64              `json:"created_at,omitempty"`
		CreatedBy   string             `json:"created_by,omitempty"`
		UpdatedAt   int64              `json:"updated_at,omitempty"`
		UpdatedBy   string             `json:"updated_by,omitempty"`
	}{
		ID:          g.id,
		Name:        g.name,
		Version:     g.ver,
		Description: g.desc,
		Debug:       g.debug,
		W:           g.w,
		H:           g.h,
		Subject:     g.sub,
		Objects:     g.obj,
		Images:      g.img,
		Scripts:     g.src,
		CreatedAt:   g.createdAt,
		CreatedBy:   g.createdBy,
		UpdatedAt:   g.updatedAt,
		UpdatedBy:   g.updatedBy,
	})
}

// UnmarshalJSON deserializes the game from JSON.
func (g *Game) UnmarshalJSON(data []byte) error {
	v := &struct {
		ID          string             `json:"id"`
		Name        string             `json:"name"`
		Version     string             `json:"version"`
		Description string             `json:"description"`
		Debug       bool               `json:"debug"`
		W           int                `json:"w"`
		H           int                `json:"h"`
		Subject     *Object            `json:"subject,omitempty"`
		Objects     map[string]*Object `json:"objects,omitempty"`
		Images      map[string]*Image  `json:"images,omitempty"`
		Scripts     map[string]*Script `json:"scripts,omitempty"`
		CreatedAt   int64              `json:"created_at,omitempty"`
		CreatedBy   string             `json:"created_by,omitempty"`
		UpdatedAt   int64              `json:"updated_at,omitempty"`
		UpdatedBy   string             `json:"updated_by,omitempty"`
	}{}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	g.id = v.ID
	g.name = v.Name
	g.desc = v.Description
	g.debug = v.Debug
	g.w = v.W
	g.h = v.H
	g.sub = v.Subject
	g.obj = v.Objects
	g.img = v.Images
	g.src = v.Scripts
	g.createdAt = v.CreatedAt
	g.createdBy = v.CreatedBy
	g.updatedAt = v.UpdatedAt
	g.updatedBy = v.UpdatedBy

	g.lua = lua.NewState()
	lua.OpenLibraries(g.lua)

	return nil
}

// AddSubject adds a subject to the game.
func (g *Game) AddSubject(sub *Object) {
	g.sub = sub
}

// AddObject adds an object to the game.
func (g *Game) AddObject(obj *Object) {
	if obj == nil {
		return
	}

	if g.obj == nil {
		g.obj = make(map[string]*Object)
	}

	g.obj[obj.id] = obj
}

// AddImage adds an image to the game.
func (g *Game) AddImage(img *Image) {
	if img == nil {
		return
	}

	if g.img == nil {
		g.img = make(map[string]*Image)
	}

	g.img[img.id] = img
}

// AddScript adds a script to the game.
func (g *Game) AddScript(src *Script) {
	if src == nil {
		return
	}

	if g.src == nil {
		g.src = make(map[string]*Script)
	}

	g.src[src.id] = src
}

// Update updates the game state each frame.
func (g *Game) Update() error {
	objects := make(map[string]any, len(g.obj))

	for k, obj := range g.obj {
		objects[k] = obj.Map()
	}

	d := map[string]any{
		"id":          g.id,
		"version":     g.ver,
		"name":        g.name,
		"description": g.desc,
		"debug":       g.debug,
		"w":           g.w,
		"h":           g.h,
		"subject":     g.sub.Map(),
		"objects":     objects,
	}

	if keys := inpututil.AppendPressedKeys(nil); len(keys) > 0 {
		keyMap := map[string]any{}

		for i, k := range keys {
			keyMap[strconv.Itoa(i)] = int(k)
		}

		d["keys"] = keyMap
	}

	pushMap(g.lua, d)
	g.lua.SetGlobal("game")

	for _, obj := range g.obj {
		if err := obj.Update(); err != nil {
			return err
		}
	}

	if g.sub != nil {
		if err := g.sub.Update(); err != nil {
			return err
		}
	}

	luaState, err := pullMap(g.lua)
	if err != nil {
		return errors.Wrap(err, errors.ErrClient,
			"unable to retrieve game state from lua")
	}

	delete(luaState, "keys")

	if err := g.updateFromMap(luaState); err != nil {
		return errors.Wrap(err, errors.ErrClient,
			"unable to update game state from lua")
	}

	if keys := inpututil.AppendJustPressedKeys(nil); len(keys) > 0 {
		for _, k := range keys {
			switch k {
			case ebiten.KeyQuote:
				g.debug = !g.debug
			case ebiten.KeyRightBracket:
				ks := inpututil.AppendPressedKeys(nil)

				for _, k := range ks {
					if k == ebiten.KeyControl {
						if err := g.Save(); err != nil {
							return errors.Wrap(err, errors.ErrClient,
								"unable to save game")
						}
					}
				}
			case ebiten.KeyLeftBracket:
				ks := inpututil.AppendPressedKeys(nil)

				for _, k := range ks {
					if k == ebiten.KeyControl {
						if err := g.Load(); err != nil {
							return errors.Wrap(err, errors.ErrClient,
								"unable to load game")
						}
					}
				}
			}
		}
	}

	return nil
}

// Draw renders the game state and all objects each frame.
func (g *Game) Draw(screen *ebiten.Image) {
	for _, obj := range g.obj {
		obj.Draw(screen)
	}

	if g.sub != nil {
		g.sub.Draw(screen)
	}

	if g.debug {
		ebitenutil.DebugPrint(screen,
			fmt.Sprintf("FPS: %f\nTPS: %f",
				ebiten.ActualFPS(), ebiten.ActualTPS()))
	}
}

// Layout returns the game object dimensions.
func (g *Game) Layout(w, h int) (int, int) {
	if g.w == 0 || g.h == 0 {
		g.w = w
		g.h = h
	}

	if g.w != w || g.h != h {
		g.w = w
		g.h = h
	}

	if g.w < 1 {
		g.w = 1
	}

	if g.h < 1 {
		g.h = 1
	}

	return g.w, g.h
}

// Save persists a game state.
func (g *Game) Save() error {
	b, err := json.MarshalIndent(&g, "", "  ")
	if err != nil {
		return errors.Wrap(err, errors.ErrClient,
			"unable to encode game save")
	}

	if err := os.WriteFile(g.name+".json", b, 0o644); err != nil {
		return errors.Wrap(err, errors.ErrClient,
			"unable to write game save",
			"file", g.name+".json")
	}

	return nil
}

// Load retrieves a persisted game state.
func (g *Game) Load() error {
	b, err := os.ReadFile(g.name + ".json")
	if err != nil {
		return errors.Wrap(err, errors.ErrClient,
			"unable to read game save",
			"file", g.name+".json")
	}

	var g2 Game

	if err := json.Unmarshal(b, &g2); err != nil {
		return errors.Wrap(err, errors.ErrClient,
			"unable to decode game save")
	}

	g.debug = g2.debug
	g.w = g2.w
	g.h = g2.h
	g.id = g2.id
	g.name = g2.name
	g.desc = g2.desc
	g.img = g2.img
	g.src = g2.src
	g.sub = g2.sub
	g.sub.game = g
	g.obj = g2.obj

	for i, s := range g.obj {
		s.game = g
		g.obj[i] = s
	}

	g.lua = lua.NewState()
	lua.OpenLibraries(g.lua)

	return nil
}

// Run starts the game processing.
func (g *Game) Run(ctx context.Context) error {
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(g.w, g.h)
	ebiten.SetWindowTitle(g.name)

	if err := ebiten.RunGame(g); err != nil {
		return err
	}

	return nil
}

// pushMap adds a map to the lua stack as a table and sets it as the lua global
// table.
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
		l.PushInteger(i + 1)
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

// tableToMap retrieves a table from the lua stack, at index, as a map.
func tableToMap(l *lua.State, index int) (map[string]any, error) {
	if !l.IsTable(index) {
		return nil, errors.New(errors.ErrClient,
			"value at index is not a table",
			"index", index)
	}

	if l.IsNil(index) {
		return nil, nil
	}

	l.PushValue(index)
	l.PushNil()

	result := make(map[string]any)

	for l.Next(-2) {
		if l.IsString(-2) {
			key, _ := l.ToString(-2)
			result[key] = getValue(l, -1)
		}

		l.Pop(1)

		if _, ok := result["1"]; ok {
			break
		}

		if !l.IsTable(-2) {
			break
		}
	}

	l.Pop(1)

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

// pullMap retrieves a map from the lua global table.
func pullMap(l *lua.State) (map[string]any, error) {
	l.Global("game")

	if !l.IsTable(-1) {
		l.Pop(1)

		return nil, errors.New(errors.ErrClient,
			"global game table not found")
	}

	result, err := tableToMap(l, -1)

	l.Pop(1)

	return result, err
}

// updateFromMap updates the game state from a map retrieved from lua.
func (g *Game) updateFromMap(m map[string]any) error {
	if m == nil {
		return nil
	}

	if v, ok := m["debug"].(bool); ok {
		g.debug = v
	}

	if v, ok := m["id"].(string); ok {
		g.id = v
	}

	if v, ok := m["version"].(string); ok {
		g.ver = v
	}

	if v, ok := m["name"].(string); ok {
		g.name = v
	}

	if v, ok := m["description"].(string); ok {
		g.desc = v
	}

	if v, ok := m["w"].(int); ok {
		g.w = v
	}

	if v, ok := m["h"].(int); ok {
		g.h = v
	}

	if v, ok := m["subject"].(map[string]any); ok {
		obj := NewObjectFromMap(v)

		g.sub = obj
		g.sub.game = g
	}

	if v, ok := m["objects"].(map[string]any); ok {
		for id, v := range v {
			if vv, ok := v.(map[string]any); ok {
				obj := NewObjectFromMap(vv)
				obj.game = g

				g.obj[id] = obj
			}
		}
	}

	return nil
}
