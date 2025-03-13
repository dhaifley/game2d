package server

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"math/big"
	"net/http"
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

// Account is the account object returned by the API.
type Account struct {
	ID             request.FieldString `bson:"id"               json:"id"               yaml:"id"`
	Name           request.FieldString `bson:"name"             json:"name"             yaml:"name"`
	Status         request.FieldString `bson:"status"           json:"status"           yaml:"status"`
	StatusData     request.FieldJSON   `bson:"status_data"      json:"status_data"      yaml:"status_data"`
	Repo           request.FieldString `bson:"repo"             json:"repo"             yaml:"repo"`
	RepoStatus     request.FieldString `bson:"repo_status"      json:"repo_status"      yaml:"repo_status"`
	RepoStatusData request.FieldJSON   `bson:"repo_status_data" json:"repo_status_data" yaml:"repo_status_data"`
	Secret         request.FieldString `bson:"secret"           json:"secret"           yaml:"secret"`
	Data           request.FieldJSON   `bson:"data"             json:"data"             yaml:"data"`
}

// Claims values contain token claims information.
type Claims struct {
	AccountID   string `json:"account_id"`
	AccountName string `json:"account_name"`
	UserID      string `json:"user_id"`
	Scopes      string `json:"scopes"`
}

// getAccountSecret retrieves an encryption secret from the database by
// account ID.
func (s *Server) getAccountSecret(ctx context.Context,
	id string,
) ([]byte, error) {
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

	if !request.ContextHasScope(ctx, request.ScopeSuperuser) && aID != id {
		return nil, errors.New(errors.ErrUnauthorized,
			"unauthorized request")
	}

	var r *Account

	if s.Cache(ctx) != nil {
		ck := cache.KeyAccount(id)

		ci, err := s.Cache(ctx).Get(ctx, ck)
		if err != nil && !errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to get account cache key",
				"error", err,
				"cache_key", ck,
				"id", id)
		} else if ci != nil {
			buf := bytes.NewBuffer(ci.Value)

			if err := json.NewDecoder(buf).Decode(&r); err != nil {
				s.log.Log(ctx, logger.LvlError,
					"unable to decode account cache value",
					"error", err,
					"cache_key", ck,
					"cache_value", string(ci.Value),
					"id", id)
			}
		}
	}

	if r == nil {
		f := bson.M{"id": id}

		if err := s.DB().Collection("accounts").FindOne(ctx, f,
			options.FindOne().SetProjection(bson.M{"_id": 0})).
			Decode(&r); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, errors.New(errors.ErrNotFound,
					"account not found",
					"id", id)
			}

			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to get account",
				"id", id)
		}

		if s.cache != nil {
			ck := cache.KeyAccount(r.ID.Value)

			buf, err := json.Marshal(r)
			if err != nil {
				s.log.Log(ctx, logger.LvlError,
					"unable to encode account cache value",
					"error", err,
					"cache_key", ck,
					"cache_value", r,
					"id", id)
			} else if len(buf) < s.cfg.CacheMaxBytes() {
				if err := s.Cache(ctx).Set(ctx, &cache.Item{
					Key:        ck,
					Value:      buf,
					Expiration: s.cfg.CacheExpiration(),
				}); err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to set account cache value",
						"error", err,
						"cache_key", ck,
						"cache_value", string(buf),
						"expiration", s.cfg.CacheExpiration(),
						"id", id)
				}
			}
		}
	}

	return r, nil
}

// createAccount inserts a new account in the database.
func (s *Server) createAccount(ctx context.Context,
	req *Account,
) (*Account, error) {
	_, _ = ctx, req

	return nil, nil
}

// AccountRepo values represent an account import repository.
type AccountRepo struct {
	Repo           request.FieldString `json:"repo"`
	RepoStatus     request.FieldString `json:"repo_status"`
	RepoStatusData request.FieldJSON   `json:"repo_status_data"`
}

// getAccountRepo retrieves the account repository from the database.
func (s *Server) getAccountRepo(ctx context.Context) (*AccountRepo, error) {
	admin := true

	if !request.ContextHasScope(ctx, request.ScopeSuperuser) &&
		!request.ContextHasScope(ctx, request.ScopeAccountAdmin) {
		admin = false
	}

	_ = admin

	return nil, nil
}

// setAccountRepo sets the account repository in the database.
func (s *Server) setAccountRepo(ctx context.Context,
	v *AccountRepo,
) error {
	if !request.ContextHasScope(ctx, request.ScopeSuperuser) &&
		!request.ContextHasScope(ctx, request.ScopeAccountAdmin) {
		return errors.New(errors.ErrForbidden,
			"unable to set account repo",
			"repo", v)
	}

	return nil
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
			"error", err)

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

	res.AccountID = s.cfg.ServiceName()
	res.AccountName = s.cfg.ServiceName()

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
) error {
	var err error

	if !request.ValidUserID(userID) {
		return errors.New(errors.ErrInvalidParameter, "invalid user_id",
			"user_id", userID)
	}

	aID := ""

	if accountID != "" {
		aCtx := context.WithValue(ctx, request.CtxKeyAccountID, "sys")

		a, err := s.getAccount(aCtx, accountID)
		if err != nil {
			return errors.New(errors.ErrUnauthorized,
				"invalid account",
				"account_id", accountID)
		}

		aID = a.ID.Value
	} else {
		aID = s.cfg.ServiceName()
	}

	hp := new(string)

	*hp, err = hashPassword(aID)
	if err != nil {
		return errors.New(errors.ErrServer,
			"unable to hash password",
			"error", err,
			"user_id", userID)
	}

	if /*hp == nil || */ *hp == "" {
		return errors.New(errors.ErrUnauthorized,
			"user cannot login",
			"user_id", userID)
	}

	if err := verifyPassword(*hp, password); err != nil {
		return errors.New(errors.ErrUnauthorized,
			"invalid user id or password",
			"user_id", userID)
	}

	return nil
}

// updateAuth periodically updates authentication data.
func (s *Server) updateAuth(ctx context.Context) context.CancelFunc {
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
	aID := ""

	if accountID != "" {
		aCtx := context.WithValue(ctx, request.CtxKeyAccountID, "sys")

		a, err := s.getAccount(aCtx, accountID)
		if err != nil {
			return "", errors.New(errors.ErrUnauthorized,
				"invalid account",
				"account_id", accountID)
		}

		aID = a.ID.Value
	} else {
		aID = s.cfg.ServiceName()
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

// accountHandler performs routing for account requests.
func (s *Server) accountHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(s.dbAvail)

	r.With(s.stat, s.trace, s.auth).Get("/repo", s.getAccountRepoHandler)
	r.With(s.stat, s.trace, s.auth).Post("/repo", s.postAccountRepoHandler)

	r.With(s.stat, s.trace, s.auth).Get("/", s.getAccountHandler)
	r.With(s.stat, s.trace, s.auth).Post("/", s.postAccountHandler)

	return r
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

// getAccountRepoHandler is the get handler function for account repos.
func (s *Server) getAccountRepoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeAccountRead); err != nil {
		s.error(err, w, r)

		return
	}

	res, err := s.getAccountRepo(ctx)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// postAccountRepoHandler is the post handler function for account repos.
func (s *Server) postAccountRepoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeAccountWrite); err != nil {
		s.error(err, w, r)

		return
	}

	req := &AccountRepo{}

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

	if err := s.setAccountRepo(ctx, req); err != nil {
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

	if err := json.NewEncoder(w).Encode(req); err != nil {
		s.error(err, w, r)
	}
}

// User is the user object returned by the API.
type User struct {
	ID        request.FieldString `bson:"id"                 json:"id"                 yaml:"id"`
	Email     request.FieldString `bson:"email"              json:"email"              yaml:"email"`
	LastName  request.FieldString `bson:"last_name"          json:"last_name"          yaml:"last_name"`
	FirstName request.FieldString `bson:"first_name"         json:"first_name"         yaml:"first_name"`
	Status    request.FieldString `bson:"status"             json:"status"             yaml:"status"`
	Scopes    request.FieldString `bson:"scopes"             json:"scopes"             yaml:"scopes"`
	Data      request.FieldJSON   `bson:"data"               json:"data"               yaml:"data"`
	Password  *string             `bson:"password,omitempty" json:"password,omitempty" yaml:"password,omitempty"`
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

// getUser retrieves a user from the database.
func (s *Server) getUser(ctx context.Context,
	id string,
) (*User, error) {
	_, _ = ctx, id

	return nil, nil
}

// createUser inserts a new user in the database.
func (s *Server) createUser(ctx context.Context,
	req *User,
) (*User, error) {
	_, _ = ctx, req

	return nil, nil
}

// updateUser updates a user in the database.
func (s *Server) updateUser(ctx context.Context,
	req *User,
) (*User, error) {
	_, _ = ctx, req

	return nil, nil
}

// deleteUser deletes a user from the database.
func (s *Server) deleteUser(ctx context.Context,
	id string,
) error {
	_, _ = ctx, id

	return nil
}

// userHandler performs routing for user requests.
func (s *Server) userHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(s.dbAvail)

	r.With(s.stat, s.trace, s.auth).Get("/", s.getUserHandler)
	r.With(s.stat, s.trace, s.auth).Patch("/", s.putUserHandler)
	r.With(s.stat, s.trace, s.auth).Put("/", s.putUserHandler)

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

	if err := s.authPassword(ctx,
		r.FormValue("username"),
		r.FormValue("password"),
		tenant); err != nil {
		s.error(err, w, r)

		return
	}

	tok, err := s.createToken(ctx, r.FormValue("username"),
		time.Now().Add(s.cfg.AuthTokenExpiresIn()).Unix(),
		r.FormValue("scope"), tenant)
	if err != nil {
		s.error(err, w, r)

		return
	}

	res := map[string]any{
		"access_token": tok,
		"token_type":   "bearer",
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}
