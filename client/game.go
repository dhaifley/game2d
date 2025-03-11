package client

import (
	"context"
	"encoding/json"
	"os"
	"reflect"

	"github.com/Shopify/go-lua"
	"github.com/dhaifley/empty/errors"
	"github.com/dhaifley/empty/logger"
	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

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

	if keys := inpututil.AppendJustPressedKeys(nil); len(keys) > 0 {
		for _, k := range keys {
			switch k {
			case ebiten.KeyQuote:
				g.debug = !g.debug
			case ebiten.KeyBracketRight:
				if err := g.Save(); err != nil {
					return errors.Wrap(err, errors.ErrClient,
						"unable to save game")
				}
			case ebiten.KeyBracketLeft:
				if err := g.Load(); err != nil {
					return errors.Wrap(err, errors.ErrClient,
						"unable to load game")
				}
			case ebiten.KeyQ:
				ks := inpututil.AppendPressedKeys(nil)

				for _, k := range ks {
					if k == ebiten.KeyControl {
						g.Quit(nil)
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

		if g.debug {
			b, _ := json.MarshalIndent(&g.sub, "", "  ")

			ebitenutil.DebugPrint(screen, string(b))
		}
	}
}

// Layout returns the game object dimensions.
func (g *Game) Layout(w, h int) (int, int) {
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

// Quit exits the game.
func (g *Game) Quit(err error) {
	if err != nil {
		g.log.Log(context.Background(), logger.LvlError,
			"game error",
			"error", err)

		os.Exit(1)
	}

	os.Exit(0)
}
