package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

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

// Game defaults.
const (
	DefaultGameWidth  = 640
	DefaultGameHeight = 480
)

// Game values represent the game state.
type Game struct {
	log      logger.Logger
	debug    bool
	pause    bool
	public   bool
	w, h     int
	id       string
	pid      string
	name     string
	ver      string
	desc     string
	icon     string
	status   string
	source   string
	apiURL   string
	apiToken string
	lua      *lua.State
	sub      *Object
	obj      map[string]*Object
	img      map[string]*Image
	src      string
	err      error
}

// NewGame creates and initializes a new Game object.
func NewGame(log logger.Logger, w, h int, id, name, desc string) *Game {
	if log == nil || (reflect.ValueOf(log).Kind() == reflect.Ptr &&
		reflect.ValueOf(log).IsNil()) {
		log = logger.NullLog
	}

	if w < 0 {
		w = DefaultGameWidth

		if ws := os.Getenv("GAME2D_GAME_WIDTH"); ws != "" {
			if i, err := strconv.Atoi(ws); err == nil {
				w = i
			}
		}
	}

	if h < 0 {
		h = DefaultGameHeight

		if hs := os.Getenv("GAME2D_GAME_HEIGHT"); hs != "" {
			if i, err := strconv.Atoi(hs); err == nil {
				h = i
			}
		}
	}

	l := lua.NewState()
	lua.OpenLibraries(l)

	if _, err := uuid.Parse(id); err != nil {
		id = ""
	}

	if id == "" {
		id = uuid.NewString()
	}

	return &Game{
		pause:  true,
		log:    log,
		w:      w,
		h:      h,
		lua:    l,
		id:     id,
		name:   name,
		source: "app",
		obj:    make(map[string]*Object),
		img:    make(map[string]*Image),
	}
}

// MarshalJSON serializes the game to JSON.
func (g *Game) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Debug   bool               `json:"debug,omitempty"`
		Pause   bool               `json:"pause,omitempty"`
		Public  bool               `json:"public,omitempty"`
		W       int                `json:"w"`
		H       int                `json:"h"`
		ID      string             `json:"id"`
		PID     string             `json:"previous_id,omitempty"`
		Name    string             `json:"name"`
		Ver     string             `json:"version,omitempty"`
		Desc    string             `json:"description,omitempty"`
		Icon    string             `json:"icon,omitempty"`
		Status  string             `json:"status,omitempty"`
		Source  string             `json:"source,omitempty"`
		Subject *Object            `json:"subject,omitempty"`
		Objects map[string]*Object `json:"objects,omitempty"`
		Images  map[string]*Image  `json:"images,omitempty"`
		Script  string             `json:"script"`
	}{
		Debug:   g.debug,
		Pause:   g.pause,
		Public:  g.public,
		W:       g.w,
		H:       g.h,
		ID:      g.id,
		PID:     g.pid,
		Name:    g.name,
		Ver:     g.ver,
		Desc:    g.desc,
		Icon:    g.icon,
		Status:  g.status,
		Source:  g.source,
		Subject: g.sub,
		Objects: g.obj,
		Images:  g.img,
		Script:  base64.StdEncoding.EncodeToString([]byte(g.src)),
	})
}

// UnmarshalJSON deserializes the game from JSON.
func (g *Game) UnmarshalJSON(data []byte) error {
	v := &struct {
		Debug   bool               `json:"debug,omitempty"`
		Pause   bool               `json:"pause,omitempty"`
		Public  bool               `json:"public,omitempty"`
		W       int                `json:"w"`
		H       int                `json:"h"`
		ID      string             `json:"id"`
		PID     string             `json:"previous_id,omitempty"`
		Name    string             `json:"name"`
		Ver     string             `json:"version,omitempty"`
		Desc    string             `json:"description,omitempty"`
		Icon    string             `json:"icon,omitempty"`
		Status  string             `json:"status,omitempty"`
		Source  string             `json:"source,omitempty"`
		Subject *Object            `json:"subject,omitempty"`
		Objects map[string]*Object `json:"objects,omitempty"`
		Images  map[string]*Image  `json:"images,omitempty"`
		Script  string             `json:"script"`
	}{}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	b, err := base64.StdEncoding.DecodeString(v.Script)
	if err != nil {
		return err
	}

	g.debug = v.Debug
	g.pause = v.Pause
	g.public = v.Public
	g.id = v.ID
	g.pid = v.PID
	g.name = v.Name
	g.ver = v.Ver
	g.desc = v.Desc
	g.icon = v.Icon
	g.status = v.Status
	g.source = v.Source
	g.debug = v.Debug
	g.w = v.W
	g.h = v.H
	g.sub = v.Subject
	g.obj = v.Objects
	g.img = v.Images
	g.src = string(b)

	g.lua = lua.NewState()
	lua.OpenLibraries(g.lua)

	return nil
}

// ID returns the game ID.
func (g *Game) ID() string {
	return g.id
}

// SetID sets the game ID.
func (g *Game) SetID(id string) {
	g.id = id
}

// Name returns the game name.
func (g *Game) Name() string {
	return g.name
}

// SetName sets the game name.
func (g *Game) SetName(name string) {
	g.name = name
}

// W returns the game width.
func (g *Game) W() int {
	return g.w
}

// SetW sets the game width.
func (g *Game) SetW(w int) {
	g.w = w
}

// H returns the game height.
func (g *Game) H() int {
	return g.h
}

// SetH sets the game height.
func (g *Game) SetH(h int) {
	g.h = h
}

// APIURL returns the API URL.
func (g *Game) APIURL() string {
	return g.apiURL
}

// SetAPIURL sets the API URL.
func (g *Game) SetAPIURL(apiURL string) {
	g.apiURL = apiURL
}

// APIToken returns the API token.
func (g *Game) APIToken() string {
	return g.apiToken
}

// SetAPIToken sets the API token.
func (g *Game) SetAPIToken(apiToken string) {
	g.apiToken = apiToken
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

// SetScript sets the game script.
func (g *Game) SetScript(src string) {
	g.src = src
}

// Update updates the game state each frame.
func (g *Game) Update() error {
	keyMap := map[string]any{}

	debug, save, load, pause, reset := false, false, false, false, false

	if keys := inpututil.AppendPressedKeys(nil); len(keys) > 0 {
		if slices.Contains(keys, ebiten.KeyControl) {
			if jpk := inpututil.AppendJustPressedKeys(nil); len(jpk) > 0 {
				for _, jk := range jpk {
					switch jk {
					case ebiten.KeyQuote:
						debug = true
					case ebiten.KeyS:
						save = true
					case ebiten.KeyL:
						load = true
					case ebiten.KeyP:
						pause = true
					case ebiten.KeyQ:
						reset = true
					}
				}
			}
		} else {
			for i, k := range keys {
				keyMap[strconv.Itoa(i)] = int(k)
			}

			if g.pause && len(keyMap) > 0 {
				pause = true
			}
		}
	}

	if !g.pause && g.src != "" {
		objects := make(map[string]any, len(g.obj))

		for k, obj := range g.obj {
			objects[k] = obj.Map()
		}

		if g.sub == nil {
			return errors.New(errors.ErrClient,
				"game subject object not found",
				"game", g)
		}

		d := map[string]any{
			"id":      g.id,
			"name":    g.name,
			"debug":   g.debug,
			"w":       g.w,
			"h":       g.h,
			"subject": g.sub.Map(),
			"objects": objects,
			"keys":    keyMap,
		}

		buf := bytes.NewBufferString(g.src)

		if err := g.lua.Load(buf, "Update", "text"); err != nil {
			return errors.Wrap(err, errors.ErrClient,
				"unable to load script",
				"game", g,
				"script", g.src)
		}

		g.lua.Call(0, 0)

		g.lua.Global("Update")

		if !g.lua.IsFunction(-1) {
			return errors.New(errors.ErrClient,
				"no Update function in script",
				"game", g,
				"script", g.src)
		}

		pushMap(g.lua, d)

		g.lua.Call(1, 1)

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
	}

	if debug {
		g.debug = !g.debug
	}

	if save {
		if err := g.Save(); err != nil {
			g.log.Log(context.Background(), logger.LvlError,
				"unable to save game",
				"error", err)
		}
	}

	if reset || load {
		if err := g.Load(); err != nil {
			g.log.Log(context.Background(), logger.LvlError,
				"unable to load game",
				"error", err)
		}

		pause = true
	}

	if pause {
		g.pause = !g.pause
	}

	return nil
}

// Draw renders the game state and all objects each frame.
func (g *Game) Draw(screen *ebiten.Image) {
	zi := map[int][]*Object{}

	for _, obj := range g.obj {
		if obj == nil || obj.hidden {
			continue
		}

		if _, ok := zi[obj.z]; !ok {
			zi[obj.z] = []*Object{}
		}

		zi[obj.z] = append(zi[obj.z], obj)
	}

	indexes := make([]int, 0, len(zi))
	for i := range zi {
		indexes = append(indexes, i)
	}

	slices.Sort(indexes)

	for i := range indexes {
		for _, obj := range zi[i] {
			obj.Draw(screen)
		}
	}

	if g.sub != nil {
		g.sub.Draw(screen)
	}

	if g.debug {
		ebitenutil.DebugPrint(screen,
			strings.ReplaceAll(
				fmt.Sprintf("ID: "+g.id+"\nFPS: %f\nTPS: %f\nErr: %+v",
					ebiten.ActualFPS(), ebiten.ActualTPS(), g.err),
				`,"`, "\n,\""))
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
func (g *Game) Save() (rErr error) {
	ebiten.SetWindowTitle(g.name + " (saving...)")

	g.source = "app"

	defer func() {
		if rErr != nil {
			g.err = rErr
		}

		ebiten.SetWindowTitle(g.name)
	}()

	b, err := json.MarshalIndent(&g, "", "  ")
	if err != nil {
		return errors.Wrap(err, errors.ErrClient,
			"unable to encode game save")
	}

	if g.apiURL != "" {
		u, err := url.Parse(g.apiURL)
		if err != nil {
			return errors.Wrap(err, errors.ErrClient,
				"unable to parse game2d API URL",
				"api_url", g.apiURL)
		}

		u = u.JoinPath("games")

		apiURL := u.String()

		req, err := http.NewRequest(http.MethodPost, apiURL,
			bytes.NewBuffer(b))
		if err != nil {
			return errors.Wrap(err, errors.ErrClient,
				"unable to create save request",
				"api_url", apiURL)
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "game2d")
		req.Header.Set("X-Game-ID", g.id)

		if g.apiToken != "" {
			req.Header.Set("Authorization", "Bearer "+g.apiToken)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errors.Wrap(err, errors.ErrClient,
				"unable to save game",
				"api_url", apiURL)
		}

		defer resp.Body.Close()

		rb, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, errors.ErrClient,
				"unable to read save game response",
				"api_url", apiURL)
		}

		if resp.StatusCode != http.StatusCreated &&
			resp.StatusCode != http.StatusOK {
			return errors.New(errors.ErrClient,
				"unable to save game",
				"api_url", apiURL,
				"status_code", resp.StatusCode,
				"response", string(rb))
		}
	} else {
		if err := os.WriteFile("game2d.json", b, 0o644); err != nil {
			return errors.Wrap(err, errors.ErrClient,
				"unable to write game save",
				"file", g.name+".json")
		}
	}

	return nil
}

// Load retrieves a persisted game state.
func (g *Game) Load() (rErr error) {
	var b []byte

	ebiten.SetWindowTitle(g.name + " (loading...)")

	defer func() {
		ebiten.SetWindowTitle(g.name)
		g.err = rErr
	}()

	if g.apiURL != "" {
		u, err := url.Parse(g.apiURL)
		if err != nil {
			return errors.Wrap(err, errors.ErrClient,
				"unable to parse game2d API URL",
				"api_url", g.apiURL)
		}

		u = u.JoinPath("games", g.id)

		apiURL := u.String()

		req, err := http.NewRequest(http.MethodGet, apiURL, nil)
		if err != nil {
			return errors.Wrap(err, errors.ErrClient,
				"unable to create load request",
				"api_url", apiURL)
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "game2d")
		req.Header.Set("X-Game-ID", g.id)

		if g.apiToken != "" {
			req.Header.Set("Authorization", "Bearer "+g.apiToken)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errors.Wrap(err, errors.ErrClient,
				"unable to load game",
				"api_url", apiURL)
		}

		defer resp.Body.Close()

		rb, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, errors.ErrClient,
				"unable to read load game response",
				"api_url", apiURL)
		}

		if resp.StatusCode != http.StatusOK {
			return errors.New(errors.ErrClient,
				"unable to load game",
				"api_url", apiURL,
				"status_code", resp.StatusCode,
				"response", string(rb))
		}

		b = rb
	} else {
		if fb, err := os.ReadFile("game2d.json"); err != nil {
			return errors.Wrap(err, errors.ErrClient,
				"unable to load game",
				"file", g.name+".json")
		} else {
			b = fb
		}
	}

	var g2 Game

	if err := json.Unmarshal(b, &g2); err != nil {
		return errors.Wrap(err, errors.ErrClient,
			"unable to decode game save")
	}

	if g2.w <= 0 || g2.h <= 0 {
		return errors.New(errors.ErrClient,
			"game save data not found",
			"game", g2)
	}

	g.debug = g2.debug
	g.pause = g2.pause
	g.public = g2.public
	g.w = g2.w
	g.h = g2.h
	g.id = g2.id
	g.pid = g2.pid
	g.name = g2.name
	g.ver = g2.ver
	g.desc = g2.desc
	g.icon = g2.icon
	g.status = g2.status
	g.source = g2.source
	g.img = g2.img
	g.src = g2.src

	if g2.sub == nil {
		return errors.New(errors.ErrClient,
			"game subject object not found",
			"game", g2)
	}

	g.sub = g2.sub
	g.sub.game = g

	if len(g2.obj) == 0 {
		return errors.New(errors.ErrClient,
			"game objects not found",
			"game", g2)
	}

	g.obj = g2.obj

	for i, s := range g.obj {
		if s == nil {
			continue
		}

		g.obj[i].game = g
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

	go func() {
		time.Sleep(50 * time.Millisecond)

		if err := g.Load(); err != nil {
			g.log.Log(ctx, logger.LvlError,
				"unable to initialize game",
				"error", err)

			g.err = err
		}
	}()

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
func tableToMap(l *lua.State, index int) (any, error) {
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

	resA := make([]any, 0)

	for l.Next(-2) {
		if l.IsString(-2) {
			key, _ := l.ToString(-2)
			result[key] = getValue(l, -1)
		} else if l.IsNumber(-2) {
			resA = append(resA, getValue(l, -1))
		} else {
			break
		}

		l.Pop(1)
	}

	l.Pop(1)

	if len(resA) > 0 {
		return resA, nil
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

// pullMap retrieves a map from the lua global table.
func pullMap(l *lua.State) (map[string]any, error) {
	if !l.IsTable(-1) {
		l.Pop(1)

		return nil, errors.New(errors.ErrClient,
			"game table not found")
	}

	val, err := tableToMap(l, -1)
	if err != nil {
		return nil, err
	}

	result, ok := val.(map[string]any)
	if !ok {
		return nil, errors.New(errors.ErrClient,
			"invalid game definition received")
	}

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

	if v, ok := m["pause"].(bool); ok {
		g.pause = v
	}

	if v, ok := m["id"].(string); ok {
		g.id = v
	}

	if v, ok := m["name"].(string); ok {
		g.name = v
	}

	if v, ok := m["w"].(int); ok {
		g.w = v
	}

	if v, ok := m["h"].(int); ok {
		g.h = v
	}

	if v, ok := m["subject"].(map[string]any); ok {
		obj := NewObjectFromMap(v)
		if obj == nil {
			return errors.New(errors.ErrClient,
				"game subject object not found",
				"game", g)
		}

		g.sub = obj
		g.sub.game = g
	}

	if v, ok := m["objects"].(map[string]any); ok {
		g.obj = make(map[string]*Object, len(v))

		for id, v := range v {
			if vv, ok := v.(map[string]any); ok {
				obj := NewObjectFromMap(vv)
				if obj == nil {
					continue
				}

				obj.game = g

				g.obj[id] = obj
			}
		}
	}

	return nil
}
