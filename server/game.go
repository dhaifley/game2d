package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dhaifley/game2d/cache"
	"github.com/dhaifley/game2d/errors"
	"github.com/dhaifley/game2d/logger"
	"github.com/dhaifley/game2d/request"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CtxKeyMinData is the context key used to indicate that minimal data should be
// returned for a game.
const CtxKeyMinData = "min_data"

// Game is the game type for the game service.
type Game struct {
	AccountID   request.FieldString      `bson:"account_id"  json:"account_id"  yaml:"account_id"`
	W           request.FieldInt64       `bson:"w"           json:"w"           yaml:"w"`
	H           request.FieldInt64       `bson:"h"           json:"h"           yaml:"h"`
	ID          request.FieldString      `bson:"id"          json:"id"          yaml:"id"`
	Name        request.FieldString      `bson:"name"        json:"name"        yaml:"name"`
	Version     request.FieldString      `bson:"version"     json:"version"     yaml:"version"`
	Description request.FieldString      `bson:"description" json:"description" yaml:"description"`
	Status      request.FieldString      `bson:"status"      json:"status"      yaml:"status"`
	StatusData  request.FieldJSON        `bson:"status_data" json:"status_data" yaml:"status_data"`
	Subject     request.FieldJSON        `bson:"subject"     json:"subject"     yaml:"subject"`
	Objects     request.FieldJSON        `bson:"objects"     json:"objects"     yaml:"objects"`
	Images      request.FieldJSON        `bson:"images"      json:"images"      yaml:"images"`
	Scripts     request.FieldJSON        `bson:"scripts"     json:"scripts"     yaml:"scripts"`
	Source      request.FieldString      `bson:"source"      json:"source"      yaml:"source"`
	CommitHash  request.FieldString      `bson:"commit_hash" json:"commit_hash" yaml:"commit_hash"`
	Tags        request.FieldStringArray `bson:"tags"        json:"tags"        yaml:"tags"`
	CreatedAt   request.FieldTime        `bson:"created_at"  json:"created_at"  yaml:"created_at"`
	CreatedBy   request.FieldString      `bson:"created_by"  json:"created_by"  yaml:"created_by"`
	UpdatedAt   request.FieldTime        `bson:"updated_at"  json:"updated_at"  yaml:"updated_at"`
	UpdatedBy   request.FieldString      `bson:"updated_by"  json:"updated_by"  yaml:"updated_by"`
}

// Validate checks that the value contains valid data.
func (g *Game) Validate() error {
	if g.AccountID.Set {
		if !g.AccountID.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"account_id must not be null",
				"user", g)
		}

		if !request.ValidAccountID(g.AccountID.Value) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid id",
				"user", g)
		}
	}

	if g.ID.Set {
		if !g.ID.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"id must not be null",
				"user", g)
		}

		if !request.ValidGameID(g.ID.Value) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid id",
				"user", g)
		}
	}

	if g.Status.Set {
		if !g.Status.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"status must not be null",
				"user", g)
		}

		switch g.Status.Value {
		case request.StatusActive, request.StatusInactive:
		default:
			return errors.New(errors.ErrInvalidRequest,
				"invalid status",
				"user", g)
		}
	}

	return nil
}

// ValidateCreate checks that the value contains valid data for creation.
func (u *Game) ValidateCreate() error {
	if !u.AccountID.Set {
		return errors.New(errors.ErrInvalidRequest,
			"missing account_id",
			"user", u)
	}

	if !u.ID.Set {
		return errors.New(errors.ErrInvalidRequest,
			"missing user_id",
			"user", u)
	}

	return u.Validate()
}

// getGames retrieves games based on a search query.
func (s *Server) getGames(ctx context.Context,
	query *request.Query,
) ([]*Game, error) {
	aID, err := request.ContextAccountID(ctx)
	if err != nil {
		return nil, errors.New(errors.ErrUnauthorized,
			"unable to get account id from context")
	}

	if err := s.checkScope(ctx, request.ScopeGamesRead); err != nil {
		return nil, err
	}

	if query == nil {
		query = request.NewQuery()
	}

	res := []*Game{}

	var f, srt bson.M

	if query.Search != "" {
		if err := bson.Unmarshal([]byte(query.Search), &f); err != nil {
			return nil, errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode search query",
				"query", query)
		}
	}

	if f == nil {
		f = bson.M{}
	}

	f["account_id"] = aID

	if query.Sort != "" {
		if err := bson.Unmarshal([]byte(query.Sort), &srt); err != nil {
			return nil, errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode sort query",
				"query", query)
		}
	}

	pro := bson.M{
		"_id":     0,
		"subject": 0,
		"objects": 0,
		"images":  0,
		"scripts": 0,
	}

	cur, err := s.DB().Collection("games").Find(ctx, f, options.Find().
		SetLimit(query.Size).SetSkip(query.Skip).
		SetSort(srt).SetProjection(pro))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to get games",
			"query", query)
	}

	defer func() {
		if err := cur.Close(ctx); err != nil {
			s.log.Log(ctx, logger.LvlError,
				"unable to close cursor",
				"err", err,
				"query", query)
		}
	}()

	for cur.Next(ctx) {
		var g *Game

		if err := cur.Decode(&g); err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to decode game",
				"query", query)
		}

		if g == nil {
			continue
		}

		res = append(res, g)

		s.setCache(ctx, cache.KeyGame(g.ID.Value), g)
	}

	if err := cur.Err(); err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to iterate games",
			"query", query)
	}

	return res, nil
}

// getGame retrieves a game by ID.
func (s *Server) getGame(ctx context.Context,
	id string,
) (*Game, error) {
	aID, err := request.ContextAccountID(ctx)
	if err != nil {
		return nil, errors.New(errors.ErrUnauthorized,
			"unable to get account id from context")
	}

	if id == "" {
		return nil, errors.New(errors.ErrInvalidRequest,
			"missing game id",
			"id", id)
	}

	if !request.ValidGameID(id) {
		return nil, errors.New(errors.ErrInvalidRequest,
			"invalid game id",
			"id", id)
	}

	if err := s.checkScope(ctx, request.ScopeGamesRead); err != nil {
		return nil, err
	}

	var res *Game

	s.getCache(ctx, cache.KeyGame(id), res)

	if res != nil {
		return res, nil
	}

	f := bson.M{"id": id, "account_id": aID}

	pro := bson.M{"_id": 0}

	if v := ctx.Value(CtxKeyMinData); v != nil {
		pro = bson.M{
			"_id":     0,
			"subject": 0,
			"objects": 0,
			"images":  0,
			"scripts": 0,
		}
	}

	if err := s.DB().Collection("games").FindOne(ctx, f,
		options.FindOne().SetProjection(pro)).Decode(&res); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New(errors.ErrNotFound,
				"game not found",
				"id", id)
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to get game",
			"id", id)
	}

	if v := ctx.Value(CtxKeyMinData); v == nil {
		s.setCache(ctx, cache.KeyGame(res.ID.Value), res)
	}

	return res, nil
}

// createGame creates a new game.
func (s *Server) createGame(ctx context.Context,
	req *Game,
) (*Game, error) {
	aID, err := request.ContextAccountID(ctx)
	if err != nil {
		return nil, errors.New(errors.ErrUnauthorized,
			"unable to get account id from context")
	}

	uID, err := request.ContextUserID(ctx)
	if err != nil {
		return nil, errors.New(errors.ErrUnauthorized,
			"unable to get user id from context")
	}

	if err := s.checkScope(ctx, request.ScopeGamesWrite); err != nil {
		return nil, err
	}

	if req == nil {
		return nil, errors.New(errors.ErrInvalidRequest,
			"missing user")
	}

	if aID != request.SystemAccount && aID != req.AccountID.Value {
		if !req.AccountID.Set {
			req.AccountID = request.FieldString{
				Set: true, Valid: true, Value: aID,
			}
		} else if err := s.checkScope(ctx, request.ScopeSuperuser); err != nil {
			return nil, errors.New(errors.ErrUnauthorized,
				"unauthorized request",
				"account_id", aID,
				"user_id", uID)
		}
	}

	if err := req.ValidateCreate(); err != nil {
		return nil, err
	}

	req.CreatedAt = request.FieldTime{
		Set: true, Valid: true, Value: time.Now().Unix(),
	}

	req.CreatedBy = request.FieldString{
		Set: true, Valid: true, Value: uID,
	}

	req.UpdatedAt = request.FieldTime{
		Set: true, Valid: true, Value: req.CreatedAt.Value,
	}

	req.UpdatedBy = request.FieldString{
		Set: true, Valid: true, Value: uID,
	}

	var res *Game

	f := bson.M{"id": req.ID.Value, "account_id": aID}

	doc := &bson.D{}

	request.SetField(doc, "account_id", req.AccountID)
	request.SetField(doc, "w", req.W)
	request.SetField(doc, "h", req.H)
	request.SetField(doc, "id", req.ID)
	request.SetField(doc, "name", req.Name)
	request.SetField(doc, "version", req.Version)
	request.SetField(doc, "description", req.Description)
	request.SetField(doc, "status", req.Status)
	request.SetField(doc, "status_data", req.StatusData)
	request.SetField(doc, "subject", req.Subject)
	request.SetField(doc, "objects", req.Objects)
	request.SetField(doc, "images", req.Images)
	request.SetField(doc, "scripts", req.Scripts)
	request.SetField(doc, "source", req.Source)
	request.SetField(doc, "commit_hash", req.CommitHash)
	request.SetField(doc, "created_at", req.CreatedAt)
	request.SetField(doc, "created_by", req.CreatedBy)
	request.SetField(doc, "updated_at", req.UpdatedAt)
	request.SetField(doc, "updated_by", req.UpdatedBy)

	if err := s.DB().Collection("games").FindOneAndReplace(ctx, f, req,
		options.FindOneAndReplace().SetProjection(bson.M{"_id": 0}).
			SetReturnDocument(options.After).SetUpsert(true)).
		Decode(&res); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New(errors.ErrNotFound,
				"game not found",
				"req", req)
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to create game",
			"req", req)
	}

	s.setCache(ctx, cache.KeyGame(res.ID.Value), res)

	return res, nil
}

// updateGame updates an existing game.
func (s *Server) updateGame(ctx context.Context,
	req *Game,
) (*Game, error) {
	aID, err := request.ContextAccountID(ctx)
	if err != nil {
		return nil, errors.New(errors.ErrUnauthorized,
			"unable to get account id from context")
	}

	uID, err := request.ContextUserID(ctx)
	if err != nil {
		return nil, errors.New(errors.ErrUnauthorized,
			"unable to get user id from context")
	}

	if err := s.checkScope(ctx, request.ScopeGamesWrite); err != nil {
		return nil, err
	}

	if req == nil {
		return nil, errors.New(errors.ErrInvalidRequest,
			"missing game")
	}

	if aID != request.SystemAccount && aID != req.AccountID.Value {
		if !req.AccountID.Set {
			req.AccountID = request.FieldString{
				Set: true, Valid: true, Value: aID,
			}
		} else if err := s.checkScope(ctx, request.ScopeSuperuser); err != nil {
			return nil, errors.New(errors.ErrUnauthorized,
				"unauthorized request",
				"account_id", aID,
				"user_id", uID)
		}
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	req.UpdatedAt = request.FieldTime{
		Set: true, Valid: true, Value: req.CreatedAt.Value,
	}

	req.UpdatedBy = request.FieldString{
		Set: true, Valid: true, Value: uID,
	}

	var res *Game

	f := bson.M{"id": req.ID.Value, "account_id": aID}

	doc := &bson.D{}

	request.SetField(doc, "w", req.W)
	request.SetField(doc, "h", req.H)
	request.SetField(doc, "name", req.Name)
	request.SetField(doc, "version", req.Version)
	request.SetField(doc, "description", req.Description)
	request.SetField(doc, "status", req.Status)
	request.SetField(doc, "status_data", req.StatusData)
	request.SetField(doc, "subject", req.Subject)
	request.SetField(doc, "objects", req.Objects)
	request.SetField(doc, "images", req.Images)
	request.SetField(doc, "scripts", req.Scripts)
	request.SetField(doc, "source", req.Source)
	request.SetField(doc, "commit_hash", req.CommitHash)
	request.SetField(doc, "updated_at", req.UpdatedAt)
	request.SetField(doc, "updated_by", req.UpdatedBy)

	if err := s.DB().Collection("games").FindOneAndUpdate(ctx, f,
		&bson.D{{Key: "$set", Value: doc}},
		options.FindOneAndUpdate().SetProjection(bson.M{"_id": 0}).
			SetReturnDocument(options.After).SetUpsert(false)).
		Decode(&res); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New(errors.ErrNotFound,
				"game not found",
				"req", req)
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to update game",
			"req", req)
	}

	s.setCache(ctx, cache.KeyGame(res.ID.Value), res)

	return res, nil
}

// deleteGame deletes a game by ID.
func (s *Server) deleteGame(ctx context.Context,
	id string,
) error {
	aID, err := request.ContextAccountID(ctx)
	if err != nil {
		return errors.New(errors.ErrUnauthorized,
			"unable to get account id from context")
	}

	if err := s.checkScope(ctx, request.ScopeGamesWrite); err != nil {
		return err
	}

	if id == "" {
		return errors.New(errors.ErrInvalidRequest,
			"missing game id",
			"id", id)
	}

	if !request.ValidGameID(id) {
		return errors.New(errors.ErrInvalidRequest,
			"invalid game id",
			"id", id)
	}

	f := bson.M{"id": id, "account_id": aID}

	if res, err := s.DB().Collection("games").
		DeleteOne(ctx, f, options.DeleteOne()); err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to delete game",
			"id", id)
	} else if res.DeletedCount == 0 {
		return errors.New(errors.ErrNotFound,
			"game not found",
			"id", id)
	}

	s.deleteCache(ctx, cache.KeyGame(id))

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

// getAllGameTags retrieves all game tags.
func (s *Server) getAllGameTags(ctx context.Context,
) ([]string, error) {
	gs, err := s.getGames(ctx, nil)
	if err != nil {
		return nil, err
	}

	tm := make(map[string]struct{}, len(gs))

	for _, g := range gs {
		for _, t := range g.Tags.Value {
			tm[t] = struct{}{}
		}
	}

	tags := make([]string, 0, len(tm))
	for t := range tm {
		tags = append(tags, t)
	}

	return tags, nil
}

// getGameTags retrieves tags for a specific game by ID.
func (s *Server) getGameTags(ctx context.Context,
	id string,
) ([]string, error) {
	ctx = context.WithValue(ctx, CtxKeyMinData, true)

	g, err := s.getGame(ctx, id)
	if err != nil {
		return nil, err
	}

	return g.Tags.Value, nil
}

// addGameTags adds tags to a game by ID.
func (s *Server) addGameTags(ctx context.Context,
	id string,
	tags []string,
) ([]string, error) {
	ctx = context.WithValue(ctx, CtxKeyMinData, true)

	g, err := s.getGame(ctx, id)
	if err != nil {
		return nil, err
	}

	tags = append(tags, g.Tags.Value...)

	tm := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		tm[t] = struct{}{}
	}

	tags = tags[:0]
	for t := range tm {
		tags = append(tags, t)
	}

	g.Tags = request.FieldStringArray{
		Set: true, Valid: true, Value: tags,
	}

	if len(g.Tags.Value) == 0 {
		g.Tags.Valid = false
	}

	if _, err := s.updateGame(ctx, &Game{
		ID:   g.ID,
		Tags: g.Tags,
	}); err != nil {
		return nil, err
	}

	return tags, nil
}

// deleteGameTags deletes tags from a game by ID.
func (s *Server) deleteGameTags(ctx context.Context,
	id string,
	tags []string,
) error {
	ctx = context.WithValue(ctx, CtxKeyMinData, true)

	g, err := s.getGame(ctx, id)
	if err != nil {
		return err
	}

	tm := make(map[string]struct{}, len(g.Tags.Value))
	for _, t := range g.Tags.Value {
		tm[t] = struct{}{}
	}

	for _, t := range tags {
		delete(tm, t)
	}

	tags = tags[:0]
	for t := range tm {
		tags = append(tags, t)
	}

	g.Tags = request.FieldStringArray{
		Set: true, Valid: true, Value: tags,
	}

	if len(g.Tags.Value) == 0 {
		g.Tags.Valid = false
	}

	if _, err := s.updateGame(ctx, &Game{
		ID:   g.ID,
		Tags: g.Tags,
	}); err != nil {
		return err
	}

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

	query, err := request.ParseQuery(r.URL.Query())
	if err != nil {
		s.error(err, w, r)

		return
	}

	res, err := s.getGames(ctx, query)
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

	res, err := s.getAllGameTags(ctx)
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
