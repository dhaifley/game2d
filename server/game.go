package server

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"net/url"
	"path/filepath"
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

// CtxKeyMinData is the context key used to indicate that minimal data should be
// returned for a game.
const CtxKeyMinData = "min_data"

// Game is the game type for the game service.
type Game struct {
	AccountID   request.FieldString      `bson:"account_id"  json:"account_id"  yaml:"account_id"`
	W           request.FieldInt64       `bson:"w"           json:"w"           yaml:"w"`
	H           request.FieldInt64       `bson:"h"           json:"h"           yaml:"h"`
	ID          request.FieldString      `bson:"id"          json:"id"          yaml:"id"`
	PreviousID  request.FieldString      `bson:"previous_id" json:"previous_id" yaml:"previous_id"`
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

	if g.ID.Set && g.ID.Valid {
		if !request.ValidGameID(g.ID.Value) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid previous_id",
				"game", g)
		}
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

	return g.Validate()
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

	if _, ok := f["status"]; !ok {
		f["status"] = request.StatusActive
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
			"unable to get games",
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
	ctx = context.WithValue(ctx, CtxKeyMinData, true)

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

	if !req.ID.Set {
		req.ID = request.FieldString{
			Set: true, Valid: true, Value: uuid.NewString(),
		}
	}

	if !req.Status.Set {
		req.Status = request.FieldString{
			Set: true, Valid: true, Value: request.StatusActive,
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

	cDoc := &bson.D{}

	request.SetField(cDoc, "id", req.ID)
	request.SetField(cDoc, "previous_id", req.PreviousID)
	request.SetField(doc, "account_id", req.AccountID)
	request.SetField(cDoc, "created_at", req.CreatedAt)
	request.SetField(cDoc, "created_by", req.CreatedBy)

	doc = &bson.D{{Key: "$set", Value: doc}, {Key: "$setOnInsert", Value: cDoc}}

	if err := s.DB().Collection("games").FindOneAndUpdate(ctx, f, doc,
		options.FindOneAndUpdate().SetProjection(bson.M{"_id": 0}).
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

	if res.PreviousID.Value != "" {
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

			if pgID := pg.PreviousID.Value; pgID != "" {
				if err := s.deleteGame(ctx, pgID); err != nil {
					return nil, errors.Wrap(err, errors.ErrDatabase,
						"unable to delete previous game",
						"previous_id", res.PreviousID.Value)
				}

				pg.PreviousID = request.FieldString{
					Set: true, Valid: true, Value: "",
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
	ctx = context.WithValue(ctx, CtxKeyMinData, true)

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
