package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/dhaifley/game2d/errors"
	"github.com/dhaifley/game2d/request"
	"github.com/go-chi/chi/v5"
)

// Game is the game type for the game service.
type Game struct {
	W           request.FieldInt64  `bson:"w"           json:"w"           yaml:"w"`
	H           request.FieldInt64  `bson:"h"           json:"h"           yaml:"h"`
	ID          request.FieldString `bson:"id"          json:"id"          yaml:"id"`
	Name        request.FieldString `bson:"name"        json:"name"        yaml:"name"`
	Version     request.FieldString `bson:"version"     json:"version"     yaml:"version"`
	Description request.FieldString `bson:"description" json:"description" yaml:"description"`
	Status      request.FieldString `bson:"status"      json:"status"      yaml:"status"`
	StatusData  request.FieldJSON   `bson:"status_data" json:"status_data" yaml:"status_data"`
	Subject     request.FieldJSON   `bson:"subject"     json:"subject"     yaml:"subject"`
	Objects     request.FieldJSON   `bson:"objects"     json:"objects"     yaml:"objects"`
	Images      request.FieldJSON   `bson:"images"      json:"images"      yaml:"images"`
	Scripts     request.FieldJSON   `bson:"scripts"     json:"scripts"     yaml:"scripts"`
	Source      request.FieldString `bson:"source"      json:"source"      yaml:"source"`
	CommitHash  request.FieldString `bson:"commit_hash" json:"commit_hash" yaml:"commit_hash"`
	CreatedBy   request.FieldString `bson:"created_by"  json:"created_by"  yaml:"created_by"`
	CreatedAt   request.FieldTime   `bson:"created_at"  json:"created_at"  yaml:"created_at"`
	UpdatedBy   request.FieldString `bson:"updated_by"  json:"updated_by"  yaml:"updated_by"`
	UpdatedAt   request.FieldTime   `bson:"updated_at"  json:"updated_at"  yaml:"updated_at"`
}

// GetGames retrieves games based on a search query.
func (s *Server) GetGames(ctx context.Context,
	query url.Values,
) ([]*Game, error) {
	return nil, nil
}

// GetGame retrieves a game by ID.
func (s *Server) GetGame(ctx context.Context,
	id string,
) (*Game, error) {
	return nil, nil
}

// CreateGame creates a new game.
func (s *Server) CreateGame(ctx context.Context,
	req *Game,
) (*Game, error) {
	return nil, nil
}

// UpdateGame updates an existing game.
func (s *Server) UpdateGame(ctx context.Context,
	req *Game,
) (*Game, error) {
	return nil, nil
}

// DeleteGame deletes a game by ID.
func (s *Server) DeleteGame(ctx context.Context,
	id string,
) error {
	return nil
}

// ImportGames imports games from a source.
func (s *Server) ImportGames(ctx context.Context,
	force bool,
) error {
	return nil
}

// ImportGame imports a single game by ID.
func (s *Server) ImportGame(ctx context.Context,
	id string,
) error {
	return nil
}

// GetTags retrieves all game tags.
func (s *Server) GetTags(ctx context.Context,
) ([]string, error) {
	return nil, nil
}

// GetGameTags retrieves tags for a specific game by ID.
func (s *Server) GetGameTags(ctx context.Context,
	id string,
) ([]string, error) {
	return nil, nil
}

// AddGameTags adds tags to a game by ID.
func (s *Server) AddGameTags(ctx context.Context,
	id string,
	tags []string,
) ([]string, error) {
	return nil, nil
}

// DeleteGameTags deletes tags from a game by ID.
func (s *Server) DeleteGameTags(ctx context.Context,
	id string,
	tags []string,
) error {
	return nil
}

// GamesHandler performs routing for event type requests.
func (s *Server) GamesHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(s.dbAvail)

	r.With(s.Stat, s.Trace, s.Auth).Post("/{id}/import", s.PostImportGameHandler)
	r.With(s.Stat, s.Trace, s.Auth).Post("/import", s.PostImportGamesHandler)

	r.With(s.Stat, s.Trace, s.Auth).Get("/tags", s.GetAllGameTagsHandler)
	r.With(s.Stat, s.Trace, s.Auth).Get("/{id}/tags",
		s.GetGameTagsHandler)
	r.With(s.Stat, s.Trace, s.Auth).Post("/{id}/tags",
		s.PostGameTagsHandler)
	r.With(s.Stat, s.Trace, s.Auth).Delete("/{id}/tags",
		s.DeleteGameTagsHandler)

	r.With(s.Stat, s.Trace, s.Auth).Get("/", s.SearchGameHandler)
	r.With(s.Stat, s.Trace, s.Auth).Get("/{id}", s.GetGameHandler)
	r.With(s.Stat, s.Trace, s.Auth).Post("/", s.PostGameHandler)
	r.With(s.Stat, s.Trace, s.Auth).Patch("/{id}", s.PutGameHandler)
	r.With(s.Stat, s.Trace, s.Auth).Put("/{id}", s.PutGameHandler)
	r.With(s.Stat, s.Trace, s.Auth).Delete("/{id}", s.DeleteGameHandler)

	return r
}

// SearchGameHandler is the search handler function for game types.
func (s *Server) SearchGameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesRead); err != nil {
		s.error(err, w, r)

		return
	}

	res, err := s.GetGames(ctx, r.URL.Query())
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// GetGameHandler is the get handler function for game types.
func (s *Server) GetGameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesRead); err != nil {
		s.error(err, w, r)

		return
	}

	id := chi.URLParam(r, "id")

	res, err := s.GetGame(ctx, id)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// PostGameHandler is the post handler function for game types.
func (s *Server) PostGameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesWrite); err != nil {
		s.error(err, w, r)

		return
	}

	req := &Game{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		switch e := err.(type) {
		case *errors.Error:
			s.error(e, w, r)
		default:
			s.error(errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode request"), w, r)
		}

		return
	}

	res, err := s.CreateGame(ctx, req)
	if err != nil {
		s.error(err, w, r)

		return
	}

	w.WriteHeader(http.StatusCreated)

	scheme := "https"
	if strings.Contains(r.Host, "localhost") {
		scheme = "http"
	}

	loc := &url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   r.URL.Path + "/" + res.ID.Value,
	}

	w.Header().Set("Location", loc.String())

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// PutGameHandler is the put handler function for game types.
func (s *Server) PutGameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesWrite); err != nil {
		s.error(err, w, r)

		return
	}

	id := chi.URLParam(r, "id")

	req := &Game{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		switch e := err.(type) {
		case *errors.Error:
			s.error(e, w, r)
		default:
			s.error(errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode request"), w, r)
		}

		return
	}

	req.ID = request.FieldString{
		Set: true, Valid: true,
		Value: id,
	}

	res, err := s.UpdateGame(ctx, req)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// DeleteGameHandler is the delete handler function for game types.
func (s *Server) DeleteGameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesWrite); err != nil {
		s.error(err, w, r)

		return
	}

	id := chi.URLParam(r, "id")

	if err := s.DeleteGame(ctx, id); err != nil {
		s.error(err, w, r)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PostImportGamesHandler is the post handler used to import games.
func (s *Server) PostImportGamesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesAdmin); err != nil {
		s.error(err, w, r)

		return
	}

	force := false

	fs := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("force")))
	if fs != "" && fs != "0" && fs != "f" && fs != "false" {
		force = true
	}

	if err := s.ImportGames(ctx, force); err != nil {
		s.error(err, w, r)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PostImportGameHandler is the post handler used to import a single game.
func (s *Server) PostImportGameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesAdmin); err != nil {
		s.error(err, w, r)

		return
	}

	id := chi.URLParam(r, "id")

	if err := s.ImportGame(ctx, id); err != nil {
		s.error(err, w, r)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetAllGameTagsHandler is the get handler function for all game tags.
func (s *Server) GetAllGameTagsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesRead); err != nil {
		s.error(err, w, r)

		return
	}

	res, err := s.GetTags(ctx)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// GetGameTagsHandler is the get handler function for game tags.
func (s *Server) GetGameTagsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesRead); err != nil {
		s.error(err, w, r)

		return
	}

	gameID := chi.URLParam(r, "id")

	res, err := s.GetGameTags(ctx, gameID)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// PostGameTagsHandler is the post handler function for game tags.
func (s *Server) PostGameTagsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesWrite); err != nil {
		s.error(err, w, r)

		return
	}

	gameID := chi.URLParam(r, "id")

	tags := []string{}

	if err := json.NewDecoder(r.Body).Decode(&tags); err != nil {
		switch e := err.(type) {
		case *errors.Error:
			s.error(e, w, r)
		default:
			s.error(errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode request"), w, r)
		}

		return
	}

	res, err := s.AddGameTags(ctx, gameID, tags)
	if err != nil {
		s.error(err, w, r)

		return
	}

	w.Header().Set("Location", r.URL.String())

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// DeleteGameTagsHandler is the delete handler function for game tags.
func (s *Server) DeleteGameTagsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesWrite); err != nil {
		s.error(err, w, r)

		return
	}

	gameID := chi.URLParam(r, "id")

	tags := []string{}

	if err := json.NewDecoder(r.Body).Decode(&tags); err != nil {
		switch e := err.(type) {
		case *errors.Error:
			s.error(e, w, r)
		default:
			s.error(errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode request"), w, r)
		}

		return
	}

	if err := s.DeleteGameTags(ctx, gameID, tags); err != nil {
		s.error(err, w, r)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
