package server

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"math/big"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/dhaifley/game2d/cache"
	"github.com/dhaifley/game2d/errors"
	"github.com/dhaifley/game2d/logger"
	"github.com/dhaifley/game2d/request"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// Account values represent account data.
type Account struct {
	ID               request.FieldString `bson:"id"                 json:"id"                 yaml:"id"`
	Name             request.FieldString `bson:"name"               json:"name"               yaml:"name"`
	Status           request.FieldString `bson:"status"             json:"status"             yaml:"status"`
	StatusData       request.FieldJSON   `bson:"status_data"        json:"status_data"        yaml:"status_data"`
	Repo             request.FieldString `bson:"repo"               json:"repo"               yaml:"repo"`
	RepoStatus       request.FieldString `bson:"repo_status"        json:"repo_status"        yaml:"repo_status"`
	RepoStatusData   request.FieldJSON   `bson:"repo_status_data"   json:"repo_status_data"   yaml:"repo_status_data"`
	GameCommitHash   request.FieldString `bson:"game_commit_hash"   json:"game_commit_hash"   yaml:"game_commit_hash"`
	GameLimit        request.FieldInt64  `bson:"game_limit"         json:"game_limit"         yaml:"game_limit"`
	Secret           request.FieldString `bson:"secret"             json:"secret"             yaml:"secret"`
	AIAPIKey         request.FieldString `bson:"ai_api_key"         json:"ai_api_key"         yaml:"ai_api_key"`
	AIMaxTokens      request.FieldInt64  `bson:"ai_max_tokens"      json:"ai_max_tokens"      yaml:"ai_max_tokens"`
	AIThinkingBudget request.FieldInt64  `bson:"ai_thinking_budget" json:"ai_thinking_budget" yaml:"ai_thinking_budget"`
	Data             request.FieldJSON   `bson:"data"               json:"data"               yaml:"data"`
	CreatedAt        request.FieldTime   `bson:"created_at"         json:"created_at"         yaml:"created_at"`
	UpdatedAt        request.FieldTime   `bson:"updated_at"         json:"updated_at"         yaml:"updated_at"`
}

// Validate checks that the value contains valid data.
func (a *Account) Validate() error {
	if a.ID.Set {
		if !a.ID.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"id must not be null",
				"account", a)
		}

		if !request.ValidAccountID(a.ID.Value) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid id",
				"account", a)
		}
	}

	if a.Name.Set {
		if !a.Name.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"name must not be null",
				"account", a)
		}

		if !request.ValidAccountName(a.Name.Value) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid name",
				"account", a)
		}
	}

	if a.Status.Set {
		if !a.Status.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"status must not be null",
				"account", a)
		}

		switch a.Status.Value {
		case request.StatusActive, request.StatusInactive:
		default:
			return errors.New(errors.ErrInvalidRequest,
				"invalid status",
				"account", a)
		}
	}

	if a.RepoStatus.Set {
		if !a.RepoStatus.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"repo_status must not be null",
				"account", a)
		}

		switch a.RepoStatus.Value {
		case request.StatusActive, request.StatusInactive,
			request.StatusError, request.StatusImporting:
		default:
			return errors.New(errors.ErrInvalidRequest,
				"invalid repo_status",
				"account", a)
		}
	}

	if a.GameLimit.Set {
		if !a.GameLimit.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"game_limit must not be null",
				"account", a)
		}

		if a.GameLimit.Value < 0 {
			return errors.New(errors.ErrInvalidRequest,
				"invalid game_limit",
				"account", a)
		}
	}

	if a.AIMaxTokens.Set {
		if !a.AIMaxTokens.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"ai_max_tokens must not be null",
				"account", a)
		}

		if a.AIMaxTokens.Value < 0 {
			return errors.New(errors.ErrInvalidRequest,
				"invalid ai_max_tokens",
				"account", a)
		}
	}

	if a.AIThinkingBudget.Set {
		if !a.AIThinkingBudget.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"ai_thinking_budget must not be null",
				"account", a)
		}

		if a.AIThinkingBudget.Value < 0 {
			return errors.New(errors.ErrInvalidRequest,
				"invalid ai_thinking_budget",
				"account", a)
		}
	}

	if a.Secret.Set && !a.Secret.Valid {
		return errors.New(errors.ErrInvalidRequest,
			"secret must not be null",
			"account", a)
	}

	return nil
}

// ValidateCreate checks that the value contains valid data for creation.
func (a *Account) ValidateCreate() error {
	if !a.ID.Set {
		return errors.New(errors.ErrInvalidRequest,
			"missing id",
			"account", a)
	}

	return a.Validate()
}

// Claims values contain token claims information.
type Claims struct {
	AccountID   string `json:"account_id"`
	AccountName string `json:"account_name"`
	UserID      string `json:"user_id"`
	Scopes      string `json:"scopes"`
}

// getAllAccounts retrieves a list of all active account ID's.
func (s *Server) getAllAccounts(ctx context.Context) ([]string, error) {
	ctx = context.WithValue(ctx, request.CtxKeyAccountID, request.SystemAccount)

	res := []string{}

	f := bson.M{"status": request.StatusActive}

	pro := bson.M{"id": 1}

	cur, err := s.DB().Collection("accounts").Find(ctx, f,
		options.Find().SetProjection(pro))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to find all accounts",
			"filter", f)
	}

	defer func() {
		if err := cur.Close(ctx); err != nil {
			s.log.Log(ctx, logger.LvlError,
				"unable to close cursor",
				"err", err,
				"filter", f)
		}
	}()

	if err := cur.All(ctx, &res); err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to get all accounts",
			"filter", f)
	}

	if err := cur.Err(); err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to get all accounts",
			"filter", f)
	}

	return res, nil
}

// getAccountSecret retrieves an encryption secret from the database by
// account ID.
func (s *Server) getAccountSecret(ctx context.Context,
	id string,
) ([]byte, error) {
	ctx = context.WithValue(ctx, request.CtxKeyAccountID, id)
	ctx = context.WithValue(ctx, request.CtxKeyScopes, request.ScopeSuperuser)

	a, err := s.getAccount(ctx, id)
	if err != nil {
		return nil, err
	}

	if a == nil || !a.Secret.Valid {
		return nil, errors.New(errors.ErrNotFound,
			"account secret not found",
			"id", id)
	}

	return []byte(a.Secret.Value), nil
}

// getAccount retrieves an account from the database.
func (s *Server) getAccount(ctx context.Context,
	id string,
) (*Account, error) {
	aID, err := request.ContextAccountID(ctx)
	if err != nil {
		return nil, errors.New(errors.ErrUnauthorized,
			"unable to get account id from context")
	}

	if id == "" {
		id = aID
	}

	if !request.ValidAccountID(id) {
		return nil, errors.New(errors.ErrInvalidRequest,
			"invalid account id",
			"id", id)
	}

	if aID != id && aID != request.SystemAccount &&
		!request.ContextHasScope(ctx, request.ScopeSuperuser) {
		return nil, errors.New(errors.ErrUnauthorized,
			"unauthorized request")
	}

	var res *Account

	defer func() {
		if res != nil {
			if err := s.checkScope(ctx, request.ScopeSuperuser); err != nil {
				res.Secret = request.FieldString{}
			}

			if err := s.checkScope(ctx, request.ScopeAccountAdmin); err != nil {
				res.Repo = request.FieldString{}
			}

			if err := s.checkScope(ctx, request.ScopeAccountAdmin); err != nil {
				res.AIAPIKey = request.FieldString{}
			}
		}
	}()

	s.getCache(ctx, cache.KeyAccount(id), res)

	if res != nil {
		return res, nil
	}

	f := bson.M{"id": id}

	if err := s.DB().Collection("accounts").FindOne(ctx, f,
		options.FindOne().SetProjection(bson.M{"_id": 0})).
		Decode(&res); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New(errors.ErrNotFound,
				"account not found",
				"id", id)
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to get account",
			"id", id)
	}

	s.setCache(ctx, cache.KeyAccount(res.ID.Value), res)

	return res, nil
}

// createAccount inserts a new account in the database.
func (s *Server) createAccount(ctx context.Context,
	req *Account,
) (*Account, error) {
	aID, err := request.ContextAccountID(ctx)
	if err != nil {
		return nil, errors.New(errors.ErrUnauthorized,
			"unable to get account id from context")
	}

	if req == nil {
		return nil, errors.New(errors.ErrInvalidRequest,
			"missing account")
	}

	if req.ID.Value == "" {
		req.ID = request.FieldString{
			Set: true, Valid: true, Value: aID,
		}
	}

	if req.ID.Value != aID && aID != request.SystemAccount &&
		!request.ContextHasScope(ctx, request.ScopeSuperuser) {
		return nil, errors.New(errors.ErrUnauthorized,
			"unauthorized request")
	}

	if err := req.ValidateCreate(); err != nil {
		return nil, err
	}

	var res *Account

	defer func() {
		if res != nil {
			if err := s.checkScope(ctx, request.ScopeSuperuser); err != nil {
				res.Secret = request.FieldString{}
			}

			if err := s.checkScope(ctx, request.ScopeAccountAdmin); err != nil {
				res.Repo = request.FieldString{}
			}

			if err := s.checkScope(ctx, request.ScopeAccountAdmin); err != nil {
				res.AIAPIKey = request.FieldString{}
			}
		}
	}()

	req.CreatedAt = request.FieldTime{
		Set: true, Valid: true, Value: time.Now().Unix(),
	}

	req.UpdatedAt = request.FieldTime{
		Set: true, Valid: true, Value: req.CreatedAt.Value,
	}

	req.GameLimit = request.FieldInt64{
		Set: true, Valid: true, Value: s.cfg.GameLimitDefault(),
	}

	f := bson.M{"id": req.ID.Value}

	doc := &bson.D{}

	request.SetField(doc, "name", req.Name)
	request.SetField(doc, "status", req.Status)
	request.SetField(doc, "status_data", req.StatusData)
	request.SetField(doc, "repo", req.Repo)
	request.SetField(doc, "repo_status", req.RepoStatus)
	request.SetField(doc, "repo_status_data", req.RepoStatusData)
	request.SetField(doc, "ai_api_key", req.AIAPIKey)
	request.SetField(doc, "ai_max_tokens", req.AIMaxTokens)
	request.SetField(doc, "ai_thinking_budget", req.AIThinkingBudget)
	request.SetField(doc, "data", req.Data)
	request.SetField(doc, "updated_at", req.UpdatedAt)

	cDoc := &bson.D{}

	request.SetField(cDoc, "id", req.ID)
	request.SetField(cDoc, "created_at", req.CreatedAt)
	request.SetField(cDoc, "game_limit", req.GameLimit)
	request.SetField(cDoc, "secret", req.Secret)

	doc = &bson.D{{Key: "$set", Value: doc}, {Key: "$setOnInsert", Value: cDoc}}

	if err := s.DB().Collection("accounts").FindOneAndUpdate(ctx, f, doc,
		options.FindOneAndUpdate().SetProjection(bson.M{"_id": 0}).
			SetReturnDocument(options.After).SetUpsert(true)).
		Decode(&res); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New(errors.ErrNotFound,
				"account not found",
				"req", req)
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to create account",
			"req", req)
	}

	s.setCache(ctx, cache.KeyAccount(res.ID.Value), res)

	return res, nil
}

// accountHandler performs routing for account requests.
func (s *Server) accountHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(s.dbAvail)

	r.With(s.stat, s.trace, s.auth).Get("/", s.getAccountHandler)
	r.With(s.stat, s.trace, s.auth).Post("/", s.postAccountHandler)

	return r
}

// getAccountHandler is the get handler function for accounts.
func (s *Server) getAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeAccountRead); err != nil {
		s.error(err, w, r)

		return
	}

	res, err := s.getAccount(ctx, "")
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// postAccountHandler is the post handler function for accounts.
func (s *Server) postAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeAccountAdmin); err != nil {
		s.error(err, w, r)

		return
	}

	req := &Account{}

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

	res, err := s.createAccount(ctx, req)
	if err != nil {
		s.error(err, w, r)

		return
	}

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

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// User values represent user data.
type User struct {
	AccountID request.FieldString `bson:"account_id"         json:"account_id"         yaml:"account_id"`
	ID        request.FieldString `bson:"id"                 json:"id"                 yaml:"id"`
	Email     request.FieldString `bson:"email"              json:"email"              yaml:"email"`
	LastName  request.FieldString `bson:"last_name"          json:"last_name"          yaml:"last_name"`
	FirstName request.FieldString `bson:"first_name"         json:"first_name"         yaml:"first_name"`
	Status    request.FieldString `bson:"status"             json:"status"             yaml:"status"`
	Scopes    request.FieldString `bson:"scopes"             json:"scopes"             yaml:"scopes"`
	Data      request.FieldJSON   `bson:"data"               json:"data"               yaml:"data"`
	Password  *string             `bson:"password,omitempty" json:"password,omitempty" yaml:"password,omitempty"`
	CreatedAt request.FieldTime   `bson:"created_at"         json:"created_at"         yaml:"created_at"`
	CreatedBy request.FieldString `bson:"created_by"         json:"created_by"         yaml:"created_by"`
	UpdatedAt request.FieldTime   `bson:"updated_at"         json:"updated_at"         yaml:"updated_at"`
	UpdatedBy request.FieldString `bson:"updated_by"         json:"updated_by"         yaml:"updated_by"`
}

// Validate checks that the value contains valid data.
func (u *User) Validate() error {
	if u.AccountID.Set {
		if !u.AccountID.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"account_id must not be null",
				"user", u)
		}

		if !request.ValidAccountID(u.AccountID.Value) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid account_id",
				"user", u)
		}
	}

	if u.ID.Set {
		if !u.ID.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"id must not be null",
				"user", u)
		}

		if !request.ValidUserID(u.ID.Value) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid id",
				"user", u)
		}
	}

	if u.Status.Set {
		if !u.Status.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"status must not be null",
				"user", u)
		}

		switch u.Status.Value {
		case request.StatusActive, request.StatusInactive:
		default:
			return errors.New(errors.ErrInvalidRequest,
				"invalid status",
				"user", u)
		}
	}

	if u.Email.Set && u.Email.Valid {
		if _, err := mail.ParseAddress(u.Email.Value); err != nil {
			return errors.New(errors.ErrInvalidRequest,
				"invalid email",
				"user", u)
		}
	}

	if u.Scopes.Set {
		if !u.Scopes.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"scopes must not be null",
				"user", u)
		}

		if !request.ValidScopes(u.Scopes.Value) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid scope",
				"user", u)
		}
	}

	return nil
}

// ValidateCreate checks that the value contains valid data for creation.
func (u *User) ValidateCreate() error {
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

// getUser retrieves a user from the database.
func (s *Server) getUser(ctx context.Context,
	id string,
) (*User, error) {
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

	if id == "" {
		id = uID
	}

	if !request.ValidUserID(id) {
		return nil, errors.New(errors.ErrInvalidRequest,
			"invalid user id",
			"id", id)
	}

	if uID != id && aID != request.SystemAccount {
		if err := s.checkScope(ctx, request.ScopeUserAdmin); err != nil {
			return nil, err
		}
	}

	var res *User

	defer func() {
		if res != nil {
			if err := s.checkScope(ctx, request.ScopeSuperuser); err != nil {
				res.Password = nil
			}
		}
	}()

	s.getCache(ctx, cache.KeyUser(id), res)

	if res != nil {
		return res, nil
	}

	f := bson.M{"id": id, "account_id": aID}

	if err := s.DB().Collection("users").FindOne(ctx, f,
		options.FindOne().SetProjection(bson.M{"_id": 0})).
		Decode(&res); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New(errors.ErrNotFound,
				"user not found",
				"id", id)
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to get user",
			"id", id)
	}

	s.setCache(ctx, cache.KeyUser(res.ID.Value), res)

	return res, nil
}

// createUser inserts a new user in the database.
func (s *Server) createUser(ctx context.Context,
	req *User,
) (*User, error) {
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

	var res *User

	defer func() {
		if res != nil {
			if err := s.checkScope(ctx, request.ScopeSuperuser); err != nil {
				res.Password = nil
			}
		}
	}()

	f := bson.M{"id": req.ID.Value, "account_id": aID}

	doc := &bson.D{}

	request.SetField(doc, "email", req.Email)
	request.SetField(doc, "last_name", req.LastName)
	request.SetField(doc, "first_name", req.FirstName)
	request.SetField(doc, "status", req.Status)
	request.SetField(doc, "scopes", req.Scopes)
	request.SetField(doc, "data", req.Data)
	request.SetField(doc, "updated_at", req.UpdatedAt)
	request.SetField(doc, "updated_by", req.UpdatedBy)

	if req.Password != nil {
		hp, err := hashPassword(*req.Password)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to hash password")
		}

		request.SetField(doc, "password", request.FieldString{
			Set: true, Valid: true, Value: hp,
		})
	}

	cDoc := &bson.D{}

	request.SetField(cDoc, "account_id", req.AccountID)
	request.SetField(cDoc, "id", req.ID)
	request.SetField(cDoc, "created_at", req.CreatedAt)
	request.SetField(doc, "created_by", req.CreatedBy)

	doc = &bson.D{{Key: "$set", Value: doc}, {Key: "$setOnInsert", Value: cDoc}}

	if err := s.DB().Collection("users").FindOneAndUpdate(ctx, f, doc,
		options.FindOneAndUpdate().SetProjection(bson.M{"_id": 0}).
			SetReturnDocument(options.After).SetUpsert(true)).
		Decode(&res); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New(errors.ErrNotFound,
				"user not found",
				"req", req)
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to create user",
			"req", req)
	}

	s.setCache(ctx, cache.KeyUser(res.ID.Value), res)

	return res, nil
}

// updateUser updates a user in the database.
func (s *Server) updateUser(ctx context.Context,
	req *User,
) (*User, error) {
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

	if req.ID.Value == "" {
		req.ID = request.FieldString{
			Set: true, Valid: true, Value: uID,
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

	var res *User

	defer func() {
		if res != nil {
			if err := s.checkScope(ctx, request.ScopeSuperuser); err != nil {
				res.Password = nil
			}
		}
	}()

	f := bson.M{"id": req.ID.Value, "account_id": aID}

	doc := &bson.D{}

	request.SetField(doc, "email", req.Email)
	request.SetField(doc, "last_name", req.LastName)
	request.SetField(doc, "first_name", req.FirstName)
	request.SetField(doc, "status", req.Status)
	request.SetField(doc, "scopes", req.Scopes)
	request.SetField(doc, "data", req.Data)
	request.SetField(doc, "password", req.Password)
	request.SetField(doc, "updated_at", req.UpdatedAt)
	request.SetField(doc, "updated_by", req.UpdatedBy)

	if req.Password != nil {
		hp, err := hashPassword(*req.Password)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to hash password")
		}

		request.SetField(doc, "password", request.FieldString{
			Set: true, Valid: true, Value: hp,
		})
	}

	if err := s.DB().Collection("users").FindOneAndUpdate(ctx, f,
		&bson.D{{Key: "$set", Value: doc}},
		options.FindOneAndUpdate().SetProjection(bson.M{"_id": 0}).
			SetReturnDocument(options.After).SetUpsert(false)).
		Decode(&res); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New(errors.ErrNotFound,
				"user not found",
				"req", req)
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to update user",
			"req", req)
	}

	s.setCache(ctx, cache.KeyUser(res.ID.Value), res)

	return res, nil
}

// deleteUser deletes a user from the database.
func (s *Server) deleteUser(ctx context.Context,
	id string,
) error {
	aID, err := request.ContextAccountID(ctx)
	if err != nil {
		return errors.New(errors.ErrUnauthorized,
			"unable to get account id from context")
	}

	if !request.ValidUserID(id) {
		return errors.New(errors.ErrInvalidRequest,
			"invalid user id",
			"id", id)
	}

	f := bson.M{"id": id, "account_id": aID}

	if res, err := s.DB().Collection("users").
		DeleteOne(ctx, f, options.DeleteOne()); err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to delete user",
			"id", id)
	} else if res.DeletedCount == 0 {
		return errors.New(errors.ErrNotFound,
			"user not found",
			"id", id)
	}

	s.deleteCache(ctx, cache.KeyUser(id))

	return nil
}

// userHandler performs routing for user requests.
func (s *Server) userHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(s.dbAvail)

	r.With(s.stat, s.trace, s.auth).Get("/", s.getUserHandler)
	r.With(s.stat, s.trace, s.auth).Patch("/", s.putUserHandler)
	r.With(s.stat, s.trace, s.auth).Put("/", s.putUserHandler)
	r.With(s.stat, s.trace, s.auth).Delete("/{id}", s.deleteUserHandler)

	return r
}

// getUserHandler is the get handler function for users.
func (s *Server) getUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeUserRead); err != nil {
		s.error(err, w, r)

		return
	}

	res, err := s.getUser(ctx, "")
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// putUserHandler is the put handler function for users.
func (s *Server) putUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeUserWrite); err != nil {
		s.error(err, w, r)

		return
	}

	req := &User{}

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
		Set: true, Valid: true, Value: "",
	}

	res, err := s.updateUser(ctx, req)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// deleteUserHandler is the delete handler function for game types.
func (s *Server) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeUserAdmin); err != nil {
		s.error(err, w, r)

		return
	}

	id := chi.URLParam(r, "id")

	if err := s.deleteUser(ctx, id); err != nil {
		s.error(err, w, r)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// authJWT authenticates using a JWT token.
func (s *Server) authJWT(ctx context.Context,
	token, accountID string,
) (*Claims, error) {
	res := &Claims{}

	aID := ""

	if accountID != "" {
		aCtx := context.WithValue(ctx, request.CtxKeyAccountID, "sys")

		a, err := s.getAccount(aCtx, accountID)
		if err != nil {
			return nil, errors.New(errors.ErrUnauthorized,
				"invalid account",
				"token", token,
				"account_id", accountID)
		}

		aID = a.ID.Value
	}

	tok, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		switch token.Method.(type) {
		case *jwt.SigningMethodHMAC:
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, errors.New(errors.ErrServer,
					"unable to find kid in token headers",
					"token", token)
			}

			return s.getAccountSecret(ctx, kid)
		case *jwt.SigningMethodECDSA:
			key, err := jwt.ParseECPublicKeyFromPEM(
				s.cfg.AuthTokenPublicKey())
			if err != nil {
				return nil, errors.New(errors.ErrServer,
					"unable to parse server token key",
					"token", token)
			}

			return key, nil
		case *jwt.SigningMethodRSA:
			key, err := jwt.ParseRSAPublicKeyFromPEM(
				s.cfg.AuthTokenPublicKey())
			if err != nil {
				return nil, errors.New(errors.ErrServer,
					"unable to parse server token key",
					"token", token)
			}

			if key == nil {
				kid, ok := token.Header["kid"].(string)
				if !ok {
					return nil, errors.New(errors.ErrServer,
						"unable to find kid in token headers",
						"token", token)
				}

				key = s.cfg.AuthTokenJWKSPublicKey(kid)
			}

			if key == nil {
				return nil, errors.New(errors.ErrServer,
					"unable to find public key for token",
					"token", token)
			}

			return key, nil
		default:
			return nil, errors.New(errors.ErrUnauthorized,
				"invalid authentication token signing method",
				"token", token)
		}
	})
	if err != nil {
		s.log.Log(ctx, logger.LvlDebug,
			"unable to parse authentication token",
			"error", err,
			"token", token)

		return nil, errors.New(errors.ErrUnauthorized,
			"invalid authentication token",
			"token", token)
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok || !tok.Valid {
		s.log.Log(ctx, logger.LvlDebug,
			"invalid authentication token used",
			"error", err,
			"token", token,
			"account_id", aID,
			"claims", claims)

		return nil, errors.New(errors.ErrUnauthorized,
			"invalid authentication token",
			"token", token)
	}

	res.AccountID = s.cfg.AccountID()
	res.AccountName = s.cfg.AccountName()

	ca, err := request.ContextAccountID(ctx)
	if err != nil || ca != request.SystemAccount {
		ctx = context.WithValue(ctx, request.CtxKeyAccountID, res.AccountID)
		ctx = context.WithValue(ctx, request.CtxKeyScopes,
			request.ScopeSuperuser)

		oa, err := s.getAccount(ctx, res.AccountID)
		if err != nil && !errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlDebug,
				"unable to retrieve account",
				"error", err,
				"token", token,
				"account_id", aID,
				"claims", claims)

			return nil, err
		}

		if oa == nil {
			s.log.Log(ctx, logger.LvlDebug,
				"valid authentication token used with invalid "+
					"account",
				"error", err,
				"token", token,
				"claims", claims)

			return nil, errors.New(errors.ErrUnauthorized,
				"invalid authentication token",
				"token", token)
		}
	}

	res.Scopes, _ = claims["scopes"].(string)

	sysAdmin := false

	if strings.Contains(res.Scopes, request.ScopeSuperuser) {
		sysAdmin = true

		if aID, err := request.ContextAccountID(ctx); err == nil {
			ctx = context.WithValue(ctx, request.CtxKeyAccountID, aID)

			res.AccountID = aID
		}
	}

	if aID != "" && res.AccountID != aID && sysAdmin {
		// Cross-tenant requests currently only permitted for system admin.
		res.AccountID = aID
	}

	ctx = context.WithValue(ctx, request.CtxKeyAccountID, res.AccountID)

	uID, ok := claims["sub"].(string)
	if !ok {
		s.log.Log(ctx, logger.LvlDebug,
			"unable to get subject from claims",
			"error", err,
			"token", token,
			"account_id", accountID,
			"claims", claims)

		return nil, errors.New(errors.ErrUnauthorized,
			"invalid authentication token",
			"token", token)
	}

	if !request.ValidUserID(uID) {
		s.log.Log(ctx, logger.LvlDebug,
			"invalid subject found in claims",
			"error", err,
			"token", token,
			"account_id", accountID,
			"claims", claims)

		return nil, errors.New(errors.ErrUnauthorized,
			"invalid authentication token",
			"token", token)
	}

	res.UserID = uID

	return res, nil
}

// authPassword authenticates using a user password.
func (s *Server) authPassword(ctx context.Context,
	userID, password, accountID string,
) (*Claims, error) {
	var err error

	if !request.ValidUserID(userID) {
		return nil, errors.New(errors.ErrInvalidParameter,
			"invalid user_id",
			"user_id", userID)
	}

	aID, aName := s.cfg.AccountID(), s.cfg.AccountName()

	if accountID != "" {
		aCtx := context.WithValue(ctx, request.CtxKeyAccountID, "sys")

		a, err := s.getAccount(aCtx, accountID)
		if err != nil {
			return nil, errors.New(errors.ErrUnauthorized,
				"invalid account",
				"account_id", accountID)
		}

		aID = a.ID.Value
		aName = a.Name.Value
	}

	ctx = context.WithValue(ctx, request.CtxKeyAccountID, aID)

	ctx = context.WithValue(ctx, request.CtxKeyUserID, userID)

	ctx = context.WithValue(ctx, request.CtxKeyScopes,
		request.ScopeSuperuser)

	u, err := s.getUser(ctx, userID)
	if err != nil {
		return nil, errors.New(errors.ErrUnauthorized,
			"invalid user id or password",
			"user_id", userID)
	}

	if err := verifyPassword(*u.Password, password); err != nil {
		return nil, errors.New(errors.ErrUnauthorized,
			"invalid user id or password",
			"user_id", userID)
	}

	return &Claims{
		AccountID:   aID,
		AccountName: aName,
		UserID:      userID,
		Scopes:      u.Scopes.Value,
	}, nil
}

// updateAuthConfig periodically updates authentication configuration data.
func (s *Server) updateAuthConfig(ctx context.Context) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)

	if tu, err := uuid.NewRandom(); err == nil {
		ctx = context.WithValue(ctx, request.CtxKeyTraceID, tu.String())
	}

	go func(ctx context.Context) {
		tick := time.NewTimer(0)

		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				if s.db == nil {
					break
				}

				ctx, cancel := request.ContextReplaceTimeout(ctx,
					s.cfg.AuthUpdateInterval())

				if tu, err := uuid.NewRandom(); err == nil {
					ctx = context.WithValue(ctx, request.CtxKeyTraceID,
						tu.String())
				}

				aid := s.cfg.AuthIdentityDomain()
				wkp := s.cfg.AuthTokenWellKnown()

				if aid == "" || wkp == "" {
					cancel()

					break
				}

				wkURL := url.URL{
					Scheme: "https",
					Host:   aid,
					Path:   wkp,
				}

				r, err := http.NewRequestWithContext(ctx, http.MethodGet,
					wkURL.String(), nil)
				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to create auth well known info request",
						"error", err,
						"url", wkURL.String())

					cancel()

					break
				}

				cli := &http.Client{Timeout: time.Second * 10}

				resp, err := cli.Do(r)
				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to retrieve auth well known info",
						"error", err)

					cancel()

					break
				}

				wk := map[string]any{}

				err = json.NewDecoder(resp.Body).Decode(&wk)

				if err := resp.Body.Close(); err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to close well known info response body",
						"error", err)
				}

				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to read well known info response body",
						"error", err)

					cancel()

					break
				}

				jwksURI, ok := wk["jwks_uri"].(string)
				if !ok || jwksURI == "" {
					s.log.Log(ctx, logger.LvlError,
						"JWKS URI not found in well known info",
						"error", err)

					cancel()

					break
				}

				rk, err := http.NewRequestWithContext(ctx, http.MethodGet,
					jwksURI, nil)
				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to create auth well known info request",
						"error", err,
						"url", wkURL.String())

					cancel()

					break
				}

				resp, err = cli.Do(rk)
				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to retrieve auth JWKS",
						"error", err)

					cancel()

					break
				}

				jwksRes := map[string]any{}

				err = json.NewDecoder(resp.Body).Decode(&jwksRes)
				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to read JWKS response body",
						"error", err)

					cancel()

					break
				}

				if err := resp.Body.Close(); err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to close JWKS response body",
						"error", err)
				}

				jwksList, ok := jwksRes["keys"].([]any)
				if !ok || len(jwksList) == 0 {
					s.log.Log(ctx, logger.LvlError,
						"keys not found in JWKS data",
						"response", jwksRes)

					cancel()

					break
				}

				jwks := map[string]*rsa.PublicKey{}

				for _, j := range jwksList {
					jm, ok := j.(map[string]any)
					if !ok {
						continue
					}

					alg, ok := jm["alg"].(string)
					if !ok || alg != "RS256" {
						continue
					}

					kid, ok := jm["kid"].(string)
					if !ok || kid == "" {
						continue
					}

					n, ok := jm["n"].(string)
					if !ok || n == "" {
						continue
					}

					e, ok := jm["e"].(string)
					if !ok && e == "" {
						continue
					}

					nb, err := base64.RawURLEncoding.DecodeString(n)
					if err != nil {
						s.log.Log(ctx, logger.LvlError,
							"unable to decode n value in JWKS data",
							"error", err,
							"jwks", jm,
							"n", n)

						continue
					}

					ev := 0

					if e == "AQAB" || e == "AAEAAQ" {
						ev = 65537
					} else {
						eb, err := base64.RawURLEncoding.DecodeString(e)
						if err != nil {
							s.log.Log(ctx, logger.LvlError,
								"unable to decode e value in JWKS data",
								"error", err,
								"jwks", jm,
								"e", e)
						}

						ebi := new(big.Int).SetBytes(eb)

						ev = int(ebi.Int64())
					}

					jwks[kid] = &rsa.PublicKey{
						N: new(big.Int).SetBytes(nb),
						E: ev,
					}
				}

				s.cfg.SetAuthTokenJWKS(jwks)

				cancel()
			}

			tick = time.NewTimer(s.cfg.AuthUpdateInterval())
		}
	}(ctx)

	return cancel
}

// createToken is used to create a JWT token that can be used for tokens.
func (s *Server) createToken(ctx context.Context,
	userID string,
	expiration int64,
	scopes, accountID string,
) (string, error) {
	aID := s.cfg.AccountID()

	if accountID != "" {
		aCtx := context.WithValue(ctx, request.CtxKeyAccountID, "sys")

		a, err := s.getAccount(aCtx, accountID)
		if err != nil {
			return "", errors.New(errors.ErrUnauthorized,
				"invalid account",
				"account_id", accountID)
		}

		aID = a.ID.Value
	}

	if !request.ValidUserID(userID) {
		return "", errors.New(errors.ErrInvalidParameter,
			"invalid user_id",
			"user_id", userID)
	}

	if !request.ValidScopes(scopes) {
		return "", errors.New(errors.ErrInvalidParameter,
			"invalid scopes",
			"scopes", scopes)
	}

	now := time.Now()

	if now.Unix() >= expiration {
		return "", errors.New(errors.ErrInvalidParameter,
			"invalid expiration",
			"expiration", expiration)
	}

	claims := jwt.MapClaims{
		"exp":    expiration,
		"iat":    now.Unix(),
		"nbf":    now.Unix(),
		"iss":    s.cfg.AuthTokenIssuer(),
		"sub":    userID,
		"aud":    []string{s.cfg.ServiceName()},
		"scopes": scopes,
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tok.Header = map[string]any{
		"alg": "HS512",
		"typ": "JWT",
		"kid": aID,
	}

	secret, err := s.getAccountSecret(ctx, aID)
	if err != nil {
		return "", err
	}

	authToken, err := tok.SignedString(secret)
	if err != nil {
		return "", errors.New(errors.ErrServer,
			"unable to create token secret")
	}

	return authToken, nil
}

// auth wraps an http handler with authentication verification.
func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

		if token == "" {
			cookie, err := r.Cookie("x-api-key")
			if err != nil && !errors.Is(err, http.ErrNoCookie) {
				s.log.Log(ctx, slog.LevelWarn,
					"invalid authentication cookie received",
					"error", err,
					"cookies", r.Cookies(),
					"request", r)
			} else if cookie != nil {
				token = strings.TrimPrefix(cookie.Value, "Bearer ")
			}
		}

		if token == "" {
			if _, pw, ok := r.BasicAuth(); ok {
				token = pw
			}
		}

		tenant := r.Header.Get("securitytenant")

		claims, err := s.authJWT(ctx, token, tenant)
		if err != nil {
			if e, ok := err.(*errors.Error); ok {
				s.error(e, w, r)

				return
			}

			s.error(errors.New(errors.ErrForbidden,
				"unauthenticated request"), w, r)

			return
		}

		if tenant != "" {
			s.log.Log(ctx, logger.LvlInfo,
				"cross-tenant request authorized",
				"error", err,
				"token", token,
				"tenant", tenant,
				"claims", claims,
				"request_method", r.Method,
				"request_url", r.URL.String(),
				"request_headers", r.Header,
				"request_remote", r.RemoteAddr)
		}

		ctx = context.WithValue(ctx, request.CtxKeyJWT, token)
		ctx = context.WithValue(ctx, request.CtxKeyAccountID, claims.AccountID)
		ctx = context.WithValue(ctx, request.CtxKeyScopes, claims.Scopes)

		if claims.UserID != "" {
			ctx = context.WithValue(ctx, request.CtxKeyUserID, claims.UserID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// checkScope verifies the request has the specified scope. It returns false
// following an error response if the required scope is missing.
func (s *Server) checkScope(ctx context.Context, scope string) error {
	if !request.ContextHasScope(ctx, scope) &&
		!request.ContextHasScope(ctx, request.ScopeSuperuser) {
		return errors.New(errors.ErrForbidden,
			"request not authorized")
	}

	return nil
}

// loginHandler performs routing for login requests.
func (s *Server) loginHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(s.dbAvail)

	r.With(s.stat, s.trace).Post("/token", s.postLoginTokenHandler)

	return r
}

// postLoginTokenHandler is the post handler for password authentication to
// obtain an API access token.
func (s *Server) postLoginTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenant := r.Header.Get("securitytenant")

	claims, err := s.authPassword(ctx, r.FormValue("username"),
		r.FormValue("password"), tenant)
	if err != nil {
		s.error(err, w, r)

		return
	}

	tok, err := s.createToken(ctx, claims.UserID,
		time.Now().Add(s.cfg.AuthTokenExpiresIn()).Unix(),
		claims.Scopes, tenant)
	if err != nil {
		s.error(err, w, r)

		return
	}

	res := map[string]any{
		"access_token": tok,
		"token_type":   "bearer",
		"account_id":   claims.AccountID,
		"account_name": claims.AccountName,
		"id":           claims.UserID,
		"scopes":       claims.Scopes,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// hashPassword creates a hashed password.
func hashPassword(password string) (string, error) {
	hp, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrServer,
			"unable to hash password")
	}

	return string(hp), nil
}

// verifyPassword verifies if a password matches a hashed password.
func verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword),
		[]byte(password))
}
