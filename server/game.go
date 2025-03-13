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

// getGames retrieves games based on a search query.
func (s *Server) getGames(ctx context.Context,
	query url.Values,
) ([]*Game, error) {
	_, _ = ctx, query

	return nil, nil
}

// getGame retrieves a game by ID.
func (s *Server) getGame(ctx context.Context,
	id string,
) (*Game, error) {
	_, _ = ctx, id

	return nil, nil
}

// createGame creates a new game.
func (s *Server) createGame(ctx context.Context,
	req *Game,
) (*Game, error) {
	_, _ = ctx, req

	return nil, nil
}

// updateGame updates an existing game.
func (s *Server) updateGame(ctx context.Context,
	req *Game,
) (*Game, error) {
	_, _ = ctx, req

	return nil, nil
}

// deleteGame deletes a game by ID.
func (s *Server) deleteGame(ctx context.Context,
	id string,
) error {
	_, _ = ctx, id

	return nil
}

// importGames imports games from a source.
func (s *Server) importGames(ctx context.Context,
	force bool,
) error {
	_, _ = ctx, force

	return nil
}

// importGame imports a single game by ID.
func (s *Server) importGame(ctx context.Context,
	id string,
) error {
	_, _ = ctx, id

	return nil
}

// getTags retrieves all game tags.
func (s *Server) getTags(ctx context.Context,
) ([]string, error) {
	_ = ctx

	return nil, nil
}

// getGameTags retrieves tags for a specific game by ID.
func (s *Server) getGameTags(ctx context.Context,
	id string,
) ([]string, error) {
	_, _ = ctx, id

	return nil, nil
}

// addGameTags adds tags to a game by ID.
func (s *Server) addGameTags(ctx context.Context,
	id string,
	tags []string,
) ([]string, error) {
	_, _ = ctx, id
	_ = tags

	return nil, nil
}

// deleteGameTags deletes tags from a game by ID.
func (s *Server) deleteGameTags(ctx context.Context,
	id string,
	tags []string,
) error {
	_, _ = ctx, id
	_ = tags

	return nil
}

// gamesHandler performs routing for event type requests.
func (s *Server) gamesHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(s.dbAvail)

	r.With(s.stat, s.trace, s.auth).Post("/{id}/import", s.postImportGameHandler)
	r.With(s.stat, s.trace, s.auth).Post("/import", s.postImportGamesHandler)

	r.With(s.stat, s.trace, s.auth).Get("/tags", s.getAllGameTagsHandler)
	r.With(s.stat, s.trace, s.auth).Get("/{id}/tags",
		s.getGameTagsHandler)
	r.With(s.stat, s.trace, s.auth).Post("/{id}/tags",
		s.postGameTagsHandler)
	r.With(s.stat, s.trace, s.auth).Delete("/{id}/tags",
		s.deleteGameTagsHandler)

	r.With(s.stat, s.trace, s.auth).Get("/", s.getGamesHandler)
	r.With(s.stat, s.trace, s.auth).Get("/{id}", s.getGameHandler)
	r.With(s.stat, s.trace, s.auth).Post("/", s.postGameHandler)
	r.With(s.stat, s.trace, s.auth).Patch("/{id}", s.putGameHandler)
	r.With(s.stat, s.trace, s.auth).Put("/{id}", s.putGameHandler)
	r.With(s.stat, s.trace, s.auth).Delete("/{id}", s.deleteGameHandler)

	return r
}

// getGamesHandler is the search handler function for game types.
func (s *Server) getGamesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesRead); err != nil {
		s.error(err, w, r)

		return
	}

	res, err := s.getGames(ctx, r.URL.Query())
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// getGameHandler is the get handler function for game types.
func (s *Server) getGameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesRead); err != nil {
		s.error(err, w, r)

		return
	}

	id := chi.URLParam(r, "id")

	res, err := s.getGame(ctx, id)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// postGameHandler is the post handler function for game types.
func (s *Server) postGameHandler(w http.ResponseWriter, r *http.Request) {
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

	res, err := s.createGame(ctx, req)
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

// putGameHandler is the put handler function for game types.
func (s *Server) putGameHandler(w http.ResponseWriter, r *http.Request) {
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

	res, err := s.updateGame(ctx, req)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// deleteGameHandler is the delete handler function for game types.
func (s *Server) deleteGameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesWrite); err != nil {
		s.error(err, w, r)

		return
	}

	id := chi.URLParam(r, "id")

	if err := s.deleteGame(ctx, id); err != nil {
		s.error(err, w, r)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// postImportGamesHandler is the post handler used to import games.
func (s *Server) postImportGamesHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := s.importGames(ctx, force); err != nil {
		s.error(err, w, r)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// postImportGameHandler is the post handler used to import a single game.
func (s *Server) postImportGameHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesAdmin); err != nil {
		s.error(err, w, r)

		return
	}

	id := chi.URLParam(r, "id")

	if err := s.importGame(ctx, id); err != nil {
		s.error(err, w, r)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getAllGameTagsHandler is the get handler function for all game tags.
func (s *Server) getAllGameTagsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesRead); err != nil {
		s.error(err, w, r)

		return
	}

	res, err := s.getTags(ctx)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// getGameTagsHandler is the get handler function for game tags.
func (s *Server) getGameTagsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesRead); err != nil {
		s.error(err, w, r)

		return
	}

	gameID := chi.URLParam(r, "id")

	res, err := s.getGameTags(ctx, gameID)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// postGameTagsHandler is the post handler function for game tags.
func (s *Server) postGameTagsHandler(w http.ResponseWriter, r *http.Request) {
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

	res, err := s.addGameTags(ctx, gameID, tags)
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

// deleteGameTagsHandler is the delete handler function for game tags.
func (s *Server) deleteGameTagsHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := s.deleteGameTags(ctx, gameID, tags); err != nil {
		s.error(err, w, r)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
