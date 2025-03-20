package server

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dhaifley/game2d/cache"
	"github.com/dhaifley/game2d/errors"
	"github.com/dhaifley/game2d/logger"
	"github.com/dhaifley/game2d/request"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gopkg.in/yaml.v3"
)

// Context keys.
const (
	CtxKeyGameNoCount         = "game_no_count"
	CtxKeyGameMinData         = "game_min_data"
	CtxKeyGameAllowPreviousID = "game_allow_previous_id"
	CtxKeyGameAllowTags       = "game_allow_tags"
)

// Game values represent game state data.
type Game struct {
	AccountID   request.FieldString      `bson:"account_id"  json:"account_id"  yaml:"account_id"`
	Public      request.FieldBool        `bson:"public"      json:"public"      yaml:"public"`
	W           request.FieldInt64       `bson:"w"           json:"w"           yaml:"w"`
	H           request.FieldInt64       `bson:"h"           json:"h"           yaml:"h"`
	ID          request.FieldString      `bson:"id"          json:"id"          yaml:"id"`
	PreviousID  request.FieldString      `bson:"previous_id" json:"previous_id" yaml:"previous_id"`
	Name        request.FieldString      `bson:"name"        json:"name"        yaml:"name"`
	Version     request.FieldString      `bson:"version"     json:"version"     yaml:"version"`
	Description request.FieldString      `bson:"description" json:"description" yaml:"description"`
	Icon        request.FieldString      `bson:"icon"        json:"icon"        yaml:"icon"`
	Status      request.FieldString      `bson:"status"      json:"status"      yaml:"status"`
	StatusData  request.FieldJSON        `bson:"status_data" json:"status_data" yaml:"status_data"`
	Subject     request.FieldJSON        `bson:"subject"     json:"subject"     yaml:"subject"`
	Objects     request.FieldJSON        `bson:"objects"     json:"objects"     yaml:"objects"`
	Images      request.FieldJSON        `bson:"images"      json:"images"      yaml:"images"`
	Scripts     request.FieldJSON        `bson:"scripts"     json:"scripts"     yaml:"scripts"`
	Source      request.FieldString      `bson:"source"      json:"source"      yaml:"source"`
	CommitHash  request.FieldString      `bson:"commit_hash" json:"commit_hash" yaml:"commit_hash"`
	Tags        request.FieldStringArray `bson:"tags"        json:"tags"        yaml:"tags"`
	AIData      request.FieldJSON        `bson:"ai_data"     json:"ai_data"     yaml:"ai_data"`
	CreatedAt   request.FieldTime        `bson:"created_at"  json:"created_at"  yaml:"created_at"`
	CreatedBy   request.FieldString      `bson:"created_by"  json:"created_by"  yaml:"created_by"`
	UpdatedAt   request.FieldTime        `bson:"updated_at"  json:"updated_at"  yaml:"updated_at"`
	UpdatedBy   request.FieldString      `bson:"updated_by"  json:"updated_by"  yaml:"updated_by"`
}

// AIData values contain the AI data for a game.
type AIData struct {
	Prompt   request.FieldString `bson:"prompt"   json:"prompt"   yaml:"prompt"`
	Response request.FieldString `bson:"response" json:"response" yaml:"response"`
	Data     request.FieldJSON   `bson:"data"     json:"data"     yaml:"data"`
	Error    request.FieldJSON   `bson:"error"    json:"error"    yaml:"error"`
	GameID   request.FieldString `bson:"game_id"  json:"game_id"  yaml:"game_id"`
}

// Map converts the AIData to a map.
func (a *AIData) Map() map[string]any {
	var res map[string]any

	if a == nil {
		return res
	}

	res = map[string]any{}

	if a.Prompt.Valid {
		res["prompt"] = a.Prompt.Value
	}

	if a.Response.Valid {
		res["response"] = a.Response.Value
	}

	if a.Data.Valid {
		res["data"] = a.Data.Value
	}

	if a.Error.Valid {
		res["error"] = a.Error.Value
	}

	if a.GameID.Valid {
		res["game_id"] = a.GameID.Value
	}

	return res
}

// aiDataFromFieldJSON converts FieldJSON to AIData.
func aiDataFromMap(m map[string]any) *AIData {
	if m == nil {
		return nil
	}

	res := &AIData{}

	if v, ok := m["prompt"]; ok {
		if s, ok := v.(string); ok {
			res.Prompt = request.FieldString{
				Set: true, Valid: true, Value: s,
			}
		}
	}

	if v, ok := m["response"]; ok {
		if s, ok := v.(string); ok {
			res.Response = request.FieldString{
				Set: true, Valid: true, Value: s,
			}
		}
	}

	if v, ok := m["data"]; ok {
		if v == nil {
			res.Data = request.FieldJSON{
				Set: true, Valid: false, Value: nil,
			}
		}

		if s, ok := v.(map[string]any); ok {
			res.Data = request.FieldJSON{
				Set: true, Valid: true, Value: s,
			}
		} else if s, ok := v.(string); ok {
			var m map[string]any

			if err := json.Unmarshal([]byte(s), &m); err == nil {
				res.Data = request.FieldJSON{
					Set: true, Valid: true, Value: m,
				}
			}
		}
	}

	if v, ok := m["error"]; ok {
		if s, ok := v.(map[string]any); ok {
			res.Error = request.FieldJSON{
				Set: true, Valid: true, Value: s,
			}
		} else if s, ok := v.(string); ok {
			var m map[string]any

			if err := json.Unmarshal([]byte(s), &m); err == nil {
				res.Error = request.FieldJSON{
					Set: true, Valid: true, Value: m,
				}
			}
		}
	}

	if v, ok := m["game_id"]; ok {
		if s, ok := v.(string); ok {
			res.GameID = request.FieldString{
				Set: true, Valid: true, Value: s,
			}
		}
	}

	return res
}

// fieldJSONFromAIData converts an AIData to a FieldJSON.
func fieldJSONFromAIData(req *AIData) request.FieldJSON {
	if req == nil {
		return request.FieldJSON{}
	}

	return request.FieldJSON{
		Set: true, Valid: true, Value: req.Map(),
	}
}

// aiDataFromFieldJSON converts a FieldJSON to an AIData.
func aiDataFromFieldJSON(req request.FieldJSON) *AIData {
	if !req.Set || !req.Valid {
		return nil
	}

	return aiDataFromMap(req.Value)
}

// Validate checks that the value contains valid data.
func (g *Game) Validate() error {
	if g.AccountID.Set {
		if !g.AccountID.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"account_id must not be null",
				"game", g)
		}

		if !request.ValidAccountID(g.AccountID.Value) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid id",
				"game", g)
		}
	}

	if g.ID.Set {
		if !g.ID.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"id must not be null",
				"game", g)
		}

		if !request.ValidGameID(g.ID.Value) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid id",
				"game", g)
		}
	}

	if g.PreviousID.Set && g.PreviousID.Valid {
		if !request.ValidGameID(g.PreviousID.Value) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid previous_id",
				"game", g)
		}
	}

	if g.Name.Set && !g.Name.Valid {
		return errors.New(errors.ErrInvalidRequest,
			"name must not be null",
			"game", g)
	}

	if g.Status.Set {
		if !g.Status.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"status must not be null",
				"game", g)
		}

		switch g.Status.Value {
		case request.StatusActive, request.StatusInactive:
		default:
			return errors.New(errors.ErrInvalidRequest,
				"invalid status",
				"game", g)
		}
	}

	return nil
}

// ValidateCreate checks that the value contains valid data for creation.
func (g *Game) ValidateCreate() error {
	if !g.AccountID.Set {
		return errors.New(errors.ErrInvalidRequest,
			"missing account_id",
			"game", g)
	}

	if !g.ID.Set {
		return errors.New(errors.ErrInvalidRequest,
			"missing id",
			"game", g)
	}

	if !g.Name.Set {
		return errors.New(errors.ErrInvalidRequest,
			"missing name",
			"game", g)
	}

	return g.Validate()
}

// getGames retrieves games based on a search query.
func (s *Server) getGames(ctx context.Context,
	query *request.Query,
) ([]*Game, int64, error) {
	aID, err := request.ContextAccountID(ctx)
	if err != nil {
		return nil, 0, errors.New(errors.ErrUnauthorized,
			"unable to get account id from context")
	}

	if query == nil {
		query = request.NewQuery()
	}

	res := []*Game{}

	var f, srt bson.M

	if query.Search != "" {
		if err := bson.UnmarshalExtJSON([]byte(query.Search),
			false, &f); err != nil {
			return nil, 0, errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode search query",
				"query", query)
		}
	}

	if f == nil {
		f = bson.M{}
	}

	if _, ok := f["status"]; !ok {
		f["status"] = request.StatusActive
	}

	f["account_id"] = aID

	if query.Sort != "" {
		if err := bson.UnmarshalExtJSON([]byte(query.Sort),
			false, &srt); err != nil {
			return nil, 0, errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode sort query",
				"query", query)
		}
	}

	if srt == nil {
		srt = bson.M{"created_at": -1}
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
		return nil, 0, errors.Wrap(err, errors.ErrDatabase,
			"unable to find games",
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
			return nil, 0, errors.Wrap(err, errors.ErrDatabase,
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
		return nil, 0, errors.Wrap(err, errors.ErrDatabase,
			"unable to get games",
			"query", query)
	}

	n, err := s.DB().Collection("games").CountDocuments(ctx, f,
		options.Count())
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.ErrDatabase,
			"unable to count games",
			"query", query)
	}

	return res, n, nil
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

	var res *Game

	s.getCache(ctx, cache.KeyGame(id), res)

	if res != nil {
		return res, nil
	}

	f := bson.M{"id": id, "account_id": aID}

	pro := bson.M{"_id": 0}

	if v := ctx.Value(CtxKeyGameMinData); v != nil {
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

	if v := ctx.Value(CtxKeyGameMinData); v == nil {
		s.setCache(ctx, cache.KeyGame(res.ID.Value), res)
	}

	return res, nil
}

// createGame creates a new game.
func (s *Server) createGame(ctx context.Context,
	req *Game,
) (*Game, error) {
	ctx = context.WithValue(ctx, CtxKeyGameMinData, true)

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

	if req.PreviousID.Value != "" {
		if k := ctx.Value(CtxKeyGameAllowPreviousID); k == nil {
			return nil, errors.New(errors.ErrInvalidRequest,
				"previous_id not allowed",
				"req", req)
		}
	}

	if !req.ID.Set {
		req.ID = request.FieldString{
			Set: true, Valid: true, Value: uuid.NewString(),
		}
	}

	if req.Status.Value == "" {
		req.Status = request.FieldString{
			Set: true, Valid: true, Value: request.StatusActive,
		}
	}

	if req.Source.Value == "" {
		req.Source = request.FieldString{
			Set: true, Valid: true, Value: "app",
		}
	}

	if err := req.ValidateCreate(); err != nil {
		return nil, err
	}

	a, err := s.getAccount(ctx, aID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to get account game limits",
			"account_id", aID,
			"req", req)
	}

	if a == nil {
		return nil, errors.New(errors.ErrNotFound,
			"account game limits not found",
			"account_id", aID,
			"req", req)
	}

	if a.GameLimit.Value > 0 {
		f := bson.M{
			"account_id": aID,
			"status":     request.StatusActive,
		}

		n, err := s.DB().Collection("games").CountDocuments(ctx, f,
			options.Count())
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to count games",
				"account_id", aID,
				"req", req)
		}

		if n >= a.GameLimit.Value {
			return nil, errors.New(errors.ErrorRateLimit,
				"account game limit reached",
				"account_id", aID,
				"game_limit", a.GameLimit.Value,
				"game_count", n)
		}
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

	request.SetField(doc, "public", req.Public)
	request.SetField(doc, "w", req.W)
	request.SetField(doc, "h", req.H)
	request.SetField(doc, "previous_id", req.PreviousID)
	request.SetField(doc, "name", req.Name)
	request.SetField(doc, "version", req.Version)
	request.SetField(doc, "description", req.Description)
	request.SetField(doc, "icon", req.Icon)
	request.SetField(doc, "status", req.Status)
	request.SetField(doc, "status_data", req.StatusData)
	request.SetField(doc, "subject", req.Subject)
	request.SetField(doc, "objects", req.Objects)
	request.SetField(doc, "images", req.Images)
	request.SetField(doc, "scripts", req.Scripts)
	request.SetField(doc, "commit_hash", req.CommitHash)
	request.SetField(doc, "ai_data", req.AIData)
	request.SetField(doc, "updated_at", req.UpdatedAt)
	request.SetField(doc, "updated_by", req.UpdatedBy)

	cDoc := &bson.D{}

	request.SetField(cDoc, "account_id", req.AccountID)
	request.SetField(cDoc, "id", req.ID)
	request.SetField(cDoc, "source", req.Source)
	request.SetField(cDoc, "created_at", req.CreatedAt)
	request.SetField(cDoc, "created_by", req.CreatedBy)

	if v := ctx.Value(CtxKeyGameAllowTags); v != nil {
		request.SetField(doc, "tags", req.Tags)
	}

	doc = &bson.D{{Key: "$set", Value: doc}, {Key: "$setOnInsert", Value: cDoc}}

	pro := bson.M{"_id": 0}

	if v := ctx.Value(CtxKeyGameMinData); v != nil {
		pro = bson.M{
			"_id":     0,
			"subject": 0,
			"objects": 0,
			"images":  0,
			"scripts": 0,
		}
	}

	if err := s.DB().Collection("games").FindOneAndUpdate(ctx, f, doc,
		options.FindOneAndUpdate().SetProjection(pro).
			SetReturnDocument(options.After).SetUpsert(true)).
		Decode(&res); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New(errors.ErrNotFound,
				"game not found",
				"req", req)
		}

		var wex mongo.WriteException

		if errors.As(err, &wex) {
			for _, writeErr := range wex.WriteErrors {
				if writeErr.Code == 10334 || writeErr.Code == 16793 {
					return nil, errors.New(errors.ErrInvalidRequest,
						"game data exceeds 16MB size limit",
						"req", req)
				}
			}
		} else if errors.ErrorHas(err, "exceeded maximum BSON document size") {
			return nil, errors.New(errors.ErrInvalidRequest,
				"game data exceeds 16MB size limit",
				"req", req)
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to create game",
			"req", req)
	}

	s.setCache(ctx, cache.KeyGame(res.ID.Value), res)

	if req.PreviousID.Value != "" {
		pg, err := s.getGame(ctx, res.PreviousID.Value)
		if err != nil && !errors.Has(err, errors.ErrNotFound) {
			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to get previous game",
				"req", req,
				"previous_id", res.PreviousID.Value)
		}

		if pg.Status.Value != request.StatusInactive {
			pg.Status = request.FieldString{
				Set: true, Valid: true, Value: request.StatusInactive,
			}

			if pg.PreviousID.Value != res.ID.Value &&
				pg.PreviousID.Value != "" {
				if err := s.deleteGame(ctx, pg.PreviousID.Value); err != nil {
					return nil, errors.Wrap(err, errors.ErrDatabase,
						"unable to delete previous game",
						"previous_id", res.PreviousID.Value)
				}

				pg.PreviousID = request.FieldString{
					Set: true, Valid: false, Value: "",
				}
			}

			if _, err := s.updateGame(ctx, pg); err != nil {
				return nil, errors.Wrap(err, errors.ErrDatabase,
					"unable to update previous game status",
					"req", req,
					"previous_id", res.PreviousID.Value)
			}
		}
	}

	return res, nil
}

// updateGame updates an existing game.
func (s *Server) updateGame(ctx context.Context,
	req *Game,
) (*Game, error) {
	ctx = context.WithValue(ctx, CtxKeyGameMinData, true)

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

	if req.PreviousID.Set {
		if k := ctx.Value(CtxKeyGameAllowPreviousID); k == nil {
			return nil, errors.New(errors.ErrInvalidRequest,
				"previous_id not allowed",
				"req", req)
		}
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	req.UpdatedAt = request.FieldTime{
		Set: true, Valid: true, Value: time.Now().Unix(),
	}

	req.UpdatedBy = request.FieldString{
		Set: true, Valid: true, Value: uID,
	}

	var res *Game

	f := bson.M{"id": req.ID.Value, "account_id": aID}

	doc := &bson.D{}

	request.SetField(doc, "public", req.Public)
	request.SetField(doc, "w", req.W)
	request.SetField(doc, "h", req.H)
	request.SetField(doc, "previous_id", req.PreviousID)
	request.SetField(doc, "name", req.Name)
	request.SetField(doc, "version", req.Version)
	request.SetField(doc, "description", req.Description)
	request.SetField(doc, "icon", req.Icon)
	request.SetField(doc, "status", req.Status)
	request.SetField(doc, "status_data", req.StatusData)
	request.SetField(doc, "subject", req.Subject)
	request.SetField(doc, "objects", req.Objects)
	request.SetField(doc, "images", req.Images)
	request.SetField(doc, "scripts", req.Scripts)
	request.SetField(doc, "commit_hash", req.CommitHash)
	request.SetField(doc, "ai_data", req.AIData)
	request.SetField(doc, "updated_at", req.UpdatedAt)
	request.SetField(doc, "updated_by", req.UpdatedBy)

	if v := ctx.Value(CtxKeyGameAllowTags); v != nil {
		request.SetField(doc, "tags", req.Tags)
	}

	pro := bson.M{"_id": 0}

	if v := ctx.Value(CtxKeyGameMinData); v != nil {
		pro = bson.M{
			"_id":     0,
			"subject": 0,
			"objects": 0,
			"images":  0,
			"scripts": 0,
		}
	}

	if err := s.DB().Collection("games").FindOneAndUpdate(ctx, f,
		&bson.D{{Key: "$set", Value: doc}},
		options.FindOneAndUpdate().SetProjection(pro).
			SetReturnDocument(options.After).SetUpsert(false)).
		Decode(&res); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New(errors.ErrNotFound,
				"game not found",
				"req", req)
		}

		var wex mongo.WriteException

		if errors.As(err, &wex) {
			for _, writeErr := range wex.WriteErrors {
				if writeErr.Code == 10334 || writeErr.Code == 16793 {
					return nil, errors.New(errors.ErrInvalidRequest,
						"game data exceeds 16MB size limit",
						"req", req)
				}
			}
		} else if errors.ErrorHas(err, "exceeded maximum BSON document size") {
			return nil, errors.New(errors.ErrInvalidRequest,
				"game data exceeds 16MB size limit",
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
	ctx = context.WithValue(ctx, request.CtxKeyUserID, request.SystemUser)
	ctx = context.WithValue(ctx, request.CtxKeyScopes, request.ScopeSuperuser)

	ar, err := s.getAccountRepo(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to get account repository")
	}

	if !force && ar.RepoStatus.Value == request.StatusImporting {
		if pli, ok := ar.RepoStatusData.Value["games_last_imported"]; ok {
			if i, ok := pli.(int64); ok && i > time.Now().Unix()-120 {
				return errors.Wrap(err, errors.ErrImport,
					"unable to import games, another import in progress")
			}
		}
	}

	ar.RepoStatus = request.FieldString{
		Set: true, Valid: true, Value: request.StatusImporting,
	}

	dm := ar.RepoStatusData.Value

	if dm == nil {
		dm = map[string]any{}
	}

	dm["games_last_imported"] = time.Now().Unix()

	ar.RepoStatusData = request.FieldJSON{
		Set: true, Valid: true, Value: dm,
	}

	if err := s.setAccountRepo(ctx, ar); err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to set account repository status")
	}

	updated, deleted, iErr := s.importRepoGames(ctx, ar, force)

	ar, err = s.getAccountRepo(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to get account repository")
	}

	ar.RepoStatus = request.FieldString{
		Set: true, Valid: true, Value: request.StatusActive,
	}

	dm = ar.RepoStatusData.Value

	if dm == nil {
		dm = map[string]any{}
	}

	dm["games_updated"] = updated

	dm["games_deleted"] = deleted

	if iErr != nil {
		ar.RepoStatus.Value = request.StatusError

		dm["games_last_error"] = iErr.Error()
	} else {
		delete(dm, "games_last_error")
	}

	ar.RepoStatusData = request.FieldJSON{
		Set: true, Valid: true, Value: dm,
	}

	if err := s.setAccountRepo(ctx, ar); err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to set account repository status")
	}

	if iErr != nil {
		return iErr
	}

	return nil
}

// getAccountGameCommitHash retrieves the current account commit hash.
func (s *Server) getAccountGameCommitHash(ctx context.Context,
) (string, error) {
	aID, err := request.ContextAccountID(ctx)
	if err != nil {
		return "", errors.New(errors.ErrUnauthorized,
			"unable to get account id from context")
	}

	a, err := s.getAccount(ctx, "")
	if err != nil {
		return "", errors.Wrap(err, errors.ErrDatabase,
			"unable to get account game commit hash",
			"account_id", aID)
	}

	return a.GameCommitHash.Value, nil
}

// setAccountGameCommitHash sets the current account commit hash.
func (s *Server) setAccountGameCommitHash(ctx context.Context,
	commit string,
) error {
	aID, err := request.ContextAccountID(ctx)
	if err != nil {
		return errors.New(errors.ErrUnauthorized,
			"unable to get account id from context")
	}

	a, err := s.getAccount(ctx, aID)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to get account game commit hash",
			"account_id", aID)
	}

	if a == nil {
		return errors.New(errors.ErrNotFound,
			"account not found",
			"account_id", aID)
	}

	a.GameCommitHash = request.FieldString{
		Set: true, Valid: true, Value: commit,
	}

	if _, err := s.createAccount(ctx, a); err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to update account game commit hash",
			"account_id", aID)
	}

	return nil
}

// deleteRepoGames deletes all imported games that do not have the specified
// commit hash.
func (s *Server) deleteRepoGames(ctx context.Context,
	commit string,
) (int, error) {
	aID, err := request.ContextAccountID(ctx)
	if err != nil {
		return 0, errors.New(errors.ErrUnauthorized,
			"unable to get account id from context")
	}

	f := bson.M{
		"account_id":  aID,
		"source":      "git",
		"commit_hash": bson.M{"$ne": commit},
	}

	pro := bson.M{"id": 1}

	cur, err := s.DB().Collection("games").Find(ctx, f,
		options.Find().SetProjection(pro))
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrDatabase,
			"unable to get games to delete",
			"filter", f)
	}

	defer func() {
		if err := cur.Close(ctx); err != nil {
			s.log.Log(ctx, logger.LvlError,
				"unable to close cursor",
				"err", err)
		}
	}()

	n := 0

	for cur.Next(ctx) {
		var g *Game

		if err := cur.Decode(&g); err != nil {
			return n, errors.Wrap(err, errors.ErrDatabase,
				"unable to decode game")
		}

		if g == nil {
			continue
		}

		df := bson.M{
			"account_id": aID,
			"id":         g.ID.Value,
			"source":     "git",
		}

		if _, err := s.DB().Collection("games").
			DeleteOne(ctx, df, options.DeleteOne()); err != nil {
			return n, errors.Wrap(err, errors.ErrDatabase,
				"unable to delete imported game",
				"filter", df)
		}

		s.deleteCache(ctx, cache.KeyGame(g.ID.Value))

		n++
	}

	if err := cur.Err(); err != nil {
		return n, errors.Wrap(err, errors.ErrDatabase,
			"unable to delete imported games",
			"filter", f)
	}

	return n, nil
}

// importRepoGames updates the games based on the contents of the account
// import repository.
func (s *Server) importRepoGames(ctx context.Context,
	ar *AccountRepo,
	force bool,
) (int, int, error) {
	ctx, cancel := request.ContextReplaceTimeout(ctx, s.cfg.ServerTimeout())

	defer cancel()

	cli, err := s.getRepoClient(ar.Repo.Value)
	if err != nil {
		return 0, 0, errors.Wrap(err, errors.ErrImport,
			"unable to create repository client")
	}

	newHash, err := cli.Commit(ctx)
	if err != nil {
		return 0, 0, errors.Wrap(err, errors.ErrImport,
			"unable to get repository commit hash")
	}

	ch, err := s.getAccountGameCommitHash(ctx)
	if err != nil {
		return 0, 0, errors.Wrap(err, errors.ErrImport,
			"unable to get account commit_hash")
	}

	if !force && ch == newHash {
		s.log.Log(ctx, logger.LvlDebug,
			"game import completed, commit unchanged",
			"updated", 0,
			"deleted", 0)

		return 0, 0, nil
	}

	res, err := cli.ListAll(ctx, "games/")
	if err != nil {
		return 0, 0, errors.Wrap(err, errors.ErrImport,
			"unable to list repository path",
			"path", "games/")
	}

	updated := 0

	errs := errors.New(errors.ErrImport,
		"unable to import games")

	for _, i := range res {
		if i.Type == "file" || i.Type == "commit_file" {
			ctx, cancel := request.ContextReplaceTimeout(ctx,
				s.cfg.ServerTimeout())

			defer cancel()

			gID := strings.TrimPrefix(strings.TrimPrefix(i.Path, "/"), "games/")

			ext := filepath.Ext(gID)

			gID = strings.TrimSuffix(gID, ext)

			g, err := s.getGame(ctx, gID)
			if err != nil && !errors.Has(err, errors.ErrNotFound) {
				errs.Errors = append(errs.Errors, errors.Wrap(err,
					errors.ErrDatabase,
					"unable to get current game",
					"game_id", gID))

				continue
			}

			if g != nil && (!force && g.Version.Value == i.Commit) {
				if g.CommitHash.Value != newHash {
					g.CommitHash = request.FieldString{
						Set: true, Valid: true, Value: newHash,
					}

					if _, err := s.updateGame(ctx, g); err != nil {
						errs.Errors = append(errs.Errors, errors.Wrap(err,
							errors.ErrDatabase,
							"unable to update repository game",
							"game", g))

						continue
					}

					updated++
				}

				continue
			}

			vb, err := cli.Get(ctx, "games/"+gID+ext)
			if err != nil {
				errs.Errors = append(errs.Errors, errors.Wrap(err,
					errors.ErrImport,
					"unable to get game repository file",
					"game_id", gID))

				continue
			}

			if err := yaml.Unmarshal(vb, &g); err != nil {
				errs.Errors = append(errs.Errors, errors.Wrap(err,
					errors.ErrImport,
					"unable to parse game repository file",
					"game_id", gID))

				continue
			}

			g.ID = request.FieldString{
				Set: true, Valid: true, Value: gID,
			}

			g.Version = request.FieldString{
				Set: true, Valid: true, Value: newHash,
			}

			g.Status = request.FieldString{
				Set: true, Valid: true, Value: request.StatusActive,
			}

			g.Source = request.FieldString{
				Set: true, Valid: true, Value: "git",
			}

			g.CommitHash = request.FieldString{
				Set: true, Valid: true, Value: newHash,
			}

			if _, err := s.createGame(ctx, g); err != nil {
				errs.Errors = append(errs.Errors, errors.Wrap(err,
					errors.ErrDatabase,
					"unable to create imported game",
					"game", g))

				continue
			}

			updated++
		}
	}

	if len(errs.Errors) > 0 {
		s.log.Log(ctx, logger.LvlWarn,
			"unable to complete game import",
			"updated", updated,
			"errors", errs.Errors)

		return updated, 0, errs
	}

	ctx, cancel = request.ContextReplaceTimeout(ctx, s.cfg.ServerTimeout())

	defer cancel()

	deleted := 0

	if newHash != "" {
		err := s.setAccountGameCommitHash(ctx, newHash)
		if err != nil {
			errs.Errors = append(errs.Errors, errors.Wrap(err,
				errors.ErrDatabase,
				"unable to set account game_commit_hash"))
		} else {
			deleted, err = s.deleteRepoGames(ctx, newHash)
			if err != nil {
				errs.Errors = append(errs.Errors, errors.Wrap(err,
					errors.ErrDatabase,
					"unable to delete removed repository games",
					"commit_hash", newHash))
			}
		}
	}

	if len(errs.Errors) > 0 {
		s.log.Log(ctx, logger.LvlWarn,
			"unable to complete game import",
			"updated", updated,
			"deleted", deleted,
			"errors", errs.Errors)

		return updated, deleted, errs
	}

	s.log.Log(ctx, logger.LvlInfo,
		"game import completed",
		"updated", updated,
		"deleted", deleted)

	return updated, deleted, nil
}

// updateGameImports periodically imports game data.
func (s *Server) updateGameImports(ctx context.Context,
) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)

	go func(ctx context.Context) {
		tick := time.NewTimer(0)

		adj := time.Duration(0)

		retries := 0

		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				accounts, err := s.getAllAccounts(ctx)
				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to get accounts to import games",
						"error", err)

					break
				}

				var wg sync.WaitGroup

				for _, aID := range accounts {
					wg.Add(1)

					go func(ctx context.Context, accountID string) {
						ctx = context.WithValue(ctx, request.CtxKeyAccountID,
							accountID)
						ctx = context.WithValue(ctx, request.CtxKeyUserID,
							request.SystemUser)
						ctx = context.WithValue(ctx, request.CtxKeyScopes,
							request.ScopeSuperuser)

						if tu, err := uuid.NewRandom(); err == nil {
							ctx = context.WithValue(ctx, request.CtxKeyTraceID,
								tu.String())
						}

						if err := s.importGames(ctx, false); err != nil {
							lvl := logger.LvlError
							if errors.ErrorHas(err,
								"another import in progress") {
								lvl = logger.LvlDebug
							}

							s.log.Log(ctx, lvl,
								"unable to import resources",
								"error", err)

							adj = s.cfg.ImportInterval()*
								time.Duration(retries) +
								time.Duration(float64(
									s.cfg.ImportInterval())*rand.Float64())

							retries++

							if retries > 10 {
								retries = 10
							}
						} else {
							retries = 0
						}

						wg.Done()
					}(ctx, aID)
				}

				wg.Wait()
			}

			tick = time.NewTimer(s.cfg.ImportInterval() + adj)

			adj = 0
		}
	}(ctx)

	return cancel
}

// getAllGameTags retrieves all game tags.
func (s *Server) getAllGameTags(ctx context.Context,
) ([]string, error) {
	ctx = context.WithValue(ctx, CtxKeyGameNoCount, true)

	gs, _, err := s.getGames(ctx, nil)
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
	ctx = context.WithValue(ctx, CtxKeyGameMinData, true)

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
	ctx = context.WithValue(ctx, CtxKeyGameMinData, true)

	g, err := s.getGame(ctx, id)
	if err != nil {
		return nil, err
	}

	newTags := append(tags, g.Tags.Value...)

	tm := make(map[string]struct{}, len(newTags))
	for _, t := range newTags {
		tm[t] = struct{}{}
	}

	nt := make([]string, 0, len(tm))
	for t := range tm {
		nt = append(nt, t)
	}

	g.Tags = request.FieldStringArray{
		Set: true, Valid: true, Value: nt,
	}

	if len(g.Tags.Value) == 0 {
		g.Tags.Valid = false
	}

	ctx = context.WithValue(ctx, CtxKeyGameAllowTags, true)

	if _, err := s.updateGame(ctx, g); err != nil {
		return nil, err
	}

	return tags, nil
}

// deleteGameTags deletes tags from a game by ID.
func (s *Server) deleteGameTags(ctx context.Context,
	id string,
	tags []string,
) error {
	ctx = context.WithValue(ctx, CtxKeyGameMinData, true)

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

	nt := make([]string, 0, len(tm))
	for t := range tm {
		nt = append(nt, t)
	}

	g.Tags = request.FieldStringArray{
		Set: true, Valid: true, Value: nt,
	}

	if len(g.Tags.Value) == 0 {
		g.Tags.Valid = false
	}

	ctx = context.WithValue(ctx, CtxKeyGameAllowTags, true)

	if _, err := s.updateGame(ctx, g); err != nil {
		return err
	}

	return nil
}

// gamesHandler performs routing for event type requests.
func (s *Server) gamesHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(s.dbAvail)

	r.With(s.stat, s.trace, s.auth).Post("/import", s.postImportGamesHandler)
	r.With(s.stat, s.trace, s.auth).Post("/copy", s.postGamesCopyHandler)
	r.With(s.stat, s.trace, s.auth).Post("/prompt", s.postGamesPromptHandler)
	r.With(s.stat, s.trace, s.auth).Post("/undo", s.postGamesUndoHandler)

	r.With(s.stat, s.trace, s.auth).Get("/tags", s.getAllGamesTagsHandler)
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

	res, n, err := s.getGames(ctx, query)
	if err != nil {
		s.error(err, w, r)

		return
	}

	w.Header().Add("X-Total-Count", strconv.FormatInt(n, 10))

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

// postGamesCopyHandler is the post handler used to copy a game definition.
func (s *Server) postGamesCopyHandler(w http.ResponseWriter,
	r *http.Request,
) {
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

	doCopy := func(req *Game) (*Game, error) {
		if req == nil {
			return nil, errors.New(errors.ErrInvalidRequest,
				"missing request")
		}

		if req.ID.Value == "" {
			return nil, errors.New(errors.ErrInvalidRequest,
				"missing game id",
				"req", req)
		}

		ctx = context.WithValue(ctx, CtxKeyGameAllowPreviousID, true)

		g, err := s.getGame(ctx, req.ID.Value)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to get game for copy",
				"req", req)
		}

		if g == nil {
			return nil, errors.New(errors.ErrNotFound,
				"game not found for copy",
				"req", req)
		}

		res := &Game{
			AccountID: g.AccountID,
			W:         g.W,
			H:         g.H,
			PreviousID: request.FieldString{
				Set: true, Valid: false,
			},
			Name:        req.Name,
			Version:     g.Version,
			Description: g.Description,
			Icon:        g.Icon,
			Status: request.FieldString{
				Set: true, Valid: true, Value: request.StatusActive,
			},
			StatusData: g.StatusData,
			Subject:    g.Subject,
			Objects:    g.Objects,
			Images:     g.Images,
			Scripts:    g.Scripts,
			Source: request.FieldString{
				Set: true, Valid: true, Value: "app",
			},
			Tags:   g.Tags,
			AIData: g.AIData,
		}

		res, err = s.createGame(ctx, res)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to create new game from prompt",
				"req", req,
				"new_game", res)
		}

		return res, nil
	}

	res, err := doCopy(req)
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
		Path:   r.URL.Path,
	}

	w.Header().Set("Location", loc.String())

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// postGamesPromptHandler is the post handler used to send a prompt about a game
// to an AI service.
func (s *Server) postGamesPromptHandler(w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesWrite); err != nil {
		s.error(err, w, r)

		return
	}

	req := &AIData{}

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

	doPrompt := func(req *AIData) (*AIData, error) {
		if req == nil {
			return nil, errors.New(errors.ErrInvalidRequest,
				"missing request")
		}

		if req.GameID.Value == "" {
			return nil, errors.New(errors.ErrInvalidRequest,
				"missing game id",
				"req", req)
		}

		ctx = context.WithValue(ctx, CtxKeyGameAllowPreviousID, true)

		g, err := s.getGame(ctx, req.GameID.Value)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to get game for prompt",
				"req", req)
		}

		if g == nil {
			return nil, errors.New(errors.ErrNotFound,
				"game not found for prompt",
				"req", req)
		}

		if g.Source.Value == "git" {
			return nil, errors.New(errors.ErrInvalidRequest,
				"unable to create prompts for games with source git",
				"req", req)
		}

		if g.Status.Value == request.StatusInactive {
			return nil, errors.New(errors.ErrInvalidRequest,
				"unable to create prompts for inactive games",
				"req", req)
		}

		aid := aiDataFromFieldJSON(g.AIData)
		if aid == nil {
			aid = req
		}

		if aid.Response.Value != "" {
			aid.Response.Value += "\n\n"
		}

		req.Response = request.FieldString{
			Set: true, Valid: true, Value: aid.Response.Value +
				"Prompt:\n" + req.Prompt.Value +
				"\n\nResponse:\nThe AI has responded.",
		}

		aid.Prompt = request.FieldString{
			Set: true, Valid: true, Value: req.Prompt.Value,
		}

		aid.Response = request.FieldString{
			Set: true, Valid: true, Value: req.Response.Value,
		}

		ng := &Game{
			AccountID: g.AccountID,
			W:         g.W,
			H:         g.H,
			PreviousID: request.FieldString{
				Set: true, Valid: true, Value: g.ID.Value,
			},
			Name:        g.Name,
			Version:     g.Version,
			Description: g.Description,
			Icon:        g.Icon,
			Status:      g.Status,
			StatusData:  g.StatusData,
			Subject:     g.Subject,
			Objects:     g.Objects,
			Images:      g.Images,
			Scripts:     g.Scripts,
			Source: request.FieldString{
				Set: true, Valid: true, Value: "app",
			},
			Tags:   g.Tags,
			AIData: fieldJSONFromAIData(aid),
		}

		ng, err = s.createGame(ctx, ng)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to create new game from prompt",
				"req", req,
				"new_game", ng)
		}

		req.GameID = request.FieldString{
			Set: true, Valid: true, Value: ng.ID.Value,
		}

		return req, nil
	}

	res, err := doPrompt(req)
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
		Path:   r.URL.Path,
	}

	w.Header().Set("Location", loc.String())

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// postGamesUndoHandler is the post handler used to undo to the point prior to
// the last AI prompt.
func (s *Server) postGamesUndoHandler(w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeGamesWrite); err != nil {
		s.error(err, w, r)

		return
	}

	req := &AIData{}

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

	doUndo := func(req *AIData) (*AIData, error) {
		if req == nil {
			return nil, errors.New(errors.ErrInvalidRequest,
				"missing request")
		}

		if req.GameID.Value == "" {
			return nil, errors.New(errors.ErrInvalidRequest,
				"missing game id",
				"req", req)
		}

		ctx = context.WithValue(ctx, CtxKeyGameMinData, true)
		ctx = context.WithValue(ctx, CtxKeyGameAllowPreviousID, true)

		g, err := s.getGame(ctx, req.GameID.Value)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to get game to undo",
				"req", req)
		}

		if g == nil {
			return nil, errors.New(errors.ErrNotFound,
				"game not found to undo",
				"req", req)
		}

		if g.PreviousID.Value == "" {
			return nil, errors.New(errors.ErrInvalidRequest,
				"unable to undo game, no previous game",
				"req", req)
		}

		pg, err := s.getGame(ctx, g.PreviousID.Value)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to get previous game to undo",
				"req", req)
		}

		if pg == nil {
			return nil, errors.New(errors.ErrNotFound,
				"previous game not found",
				"req", req)
		}

		if pg.PreviousID.Value != g.ID.Value &&
			pg.PreviousID.Value != "" {
			if err := s.deleteGame(ctx, pg.PreviousID.Value); err != nil {
				return nil, errors.Wrap(err, errors.ErrDatabase,
					"unable to delete previous game",
					"previous_id", pg.PreviousID.Value,
					"req", req)
			}
		}

		g.PreviousID = request.FieldString{
			Set: true, Valid: false, Value: "",
		}

		g.Status = request.FieldString{
			Set: true, Valid: true, Value: request.StatusInactive,
		}

		pg.PreviousID = request.FieldString{
			Set: true, Valid: true, Value: g.ID.Value,
		}

		pg.Status = request.FieldString{
			Set: true, Valid: true, Value: request.StatusActive,
		}

		pg, err = s.updateGame(ctx, pg)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to update previous game to undo",
				"req", req,
				"previous_game", pg)
		}

		g, err = s.updateGame(ctx, g)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to update game to undo",
				"req", req,
				"game", g)
		}

		req.Response = request.FieldString{
			Set: true, Valid: true, Value: req.Response.Value +
				"\n\nUndo." + req.Prompt.Value +
				"\n\nResponse:\nThe previous prompt has been undone.",
		}

		req.GameID = request.FieldString{
			Set: true, Valid: true, Value: pg.ID.Value,
		}

		return req, nil
	}

	res, err := doUndo(req)
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
		Path:   r.URL.Path,
	}

	w.Header().Set("Location", loc.String())

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// getAllGamesTagsHandler is the get handler function for all game tags.
func (s *Server) getAllGamesTagsHandler(w http.ResponseWriter,
	r *http.Request,
) {
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
