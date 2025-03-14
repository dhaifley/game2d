// Package server contains the REST API server, routers, and handlers.
package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dhaifley/game2d/cache"
	"github.com/dhaifley/game2d/config"
	"github.com/dhaifley/game2d/errors"
	"github.com/dhaifley/game2d/logger"
	"github.com/dhaifley/game2d/metric"
	"github.com/dhaifley/game2d/request"
	"github.com/dhaifley/game2d/static"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// The server version.
var Version = ""

// Server values implement HTTP server functionality.
type Server struct {
	http.Server
	sync.RWMutex
	health   uint32
	addr     []string
	cancels  []context.CancelFunc
	cfg      *config.Config
	log      logger.Logger
	metric   metric.Recorder
	tracer   trace.Tracer
	r        chi.Router
	db       *mongo.Client
	cache    cache.Accessor
	dbOnce   sync.Once
	authOnce sync.Once
}

// NewServer creates a new HTTP server.
func NewServer(cfg *config.Config,
	log logger.Logger,
	metric metric.Recorder,
	tracer trace.Tracer,
) (*Server, error) {
	if cfg == nil {
		cfg = &config.Config{}
	}

	if log == nil || (reflect.ValueOf(log).Kind() == reflect.Ptr &&
		reflect.ValueOf(log).IsNil()) {
		log = logger.NullLog
	}

	if metric == nil || (reflect.ValueOf(metric).Kind() == reflect.Ptr &&
		reflect.ValueOf(metric).IsNil()) {
		metric = nil
	}

	if tracer == nil || (reflect.ValueOf(tracer).Kind() == reflect.Ptr &&
		reflect.ValueOf(tracer).IsNil()) {
		tracer = nil
	}

	s := &Server{
		cfg:    cfg,
		addr:   strings.Split(cfg.ServerAddress(), " "),
		health: http.StatusOK,
		log:    log,
		tracer: tracer,
		metric: metric,
	}

	s.Server.IdleTimeout = 30 * time.Second
	s.Server.ReadHeaderTimeout = 30 * time.Second

	if s.cfg.ServerIdleTimeout() != 0 {
		s.Server.IdleTimeout = s.cfg.ServerIdleTimeout()
		s.Server.ReadHeaderTimeout = s.cfg.ServerIdleTimeout()
	}

	if len(s.cfg.CacheServers()) > 0 {
		s.cache = cache.NewClient(s.cfg, s.log, s.metric, s.tracer)

		s.log.Log(context.Background(), logger.LvlDebug,
			"cache connection created",
			"servers", s.cfg.CacheServers())
	}

	s.initRouter()

	s.Server.Handler = s.r

	return s, nil
}

// Health gets the status code for the current server health.
func (s *Server) Health() uint32 {
	s.RLock()
	defer s.RUnlock()

	return s.health
}

// SetHealth sets the status code for the current server health.
func (s *Server) SetHealth(code uint32) {
	s.Lock()
	defer s.Unlock()

	s.health = code
}

// addCancelFunc adds a context cancellation function to the list of cancel
// functions the server needs to call when closing.
func (s *Server) addCancelFunc(cf context.CancelFunc) {
	s.Lock()
	defer s.Unlock()

	s.cancels = append(s.cancels, cf)
}

// Cache gets the server cache for a specific request.
func (s *Server) Cache(ctx context.Context) cache.Accessor {
	s.RLock()
	defer s.RUnlock()

	if s.cache == nil {
		return nil
	}

	if v, err := request.ContextNoCache(ctx); err != nil || v {
		return nil
	}

	return s.cache
}

// DB gets the database used by for the server.
func (s *Server) DB() *mongo.Database {
	s.RLock()
	defer s.RUnlock()

	if s.db == nil {
		return nil
	}

	return s.db.Database(s.cfg.DBDatabase())
}

// SetDB sets the database client for the server.
func (s *Server) SetDB(db *mongo.Client) {
	s.Lock()
	defer s.Unlock()

	if db == nil || (reflect.ValueOf(db).Kind() == reflect.Ptr &&
		reflect.ValueOf(db).IsNil()) {
		s.db = nil

		return
	}

	s.db = db
}

// Mux routes and serves a request.
func (s *Server) Mux(w http.ResponseWriter, r *http.Request) {
	s.r.ServeHTTP(w, r)
}

// ConnectDB connects to the NoSQL database.
func (s *Server) ConnectDB() {
	s.dbOnce.Do(func() {
		go func(ctx context.Context) {
			if s.db != nil {
				return
			}

			retry := false

			for {
				if retry {
					time.Sleep(time.Second)
				} else {
					retry = true
				}

				if s.db != nil {
					break
				}

				c, err := mongo.Connect(options.Client().ApplyURI(
					s.cfg.DBConn()),
					options.Client().SetMaxPoolSize(
						uint64(s.cfg.DBMaxPoolSize())),
					options.Client().SetMinPoolSize(
						uint64(s.cfg.DBMinPoolSize())))
				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to connect to NoSQL database",
						"error", err)

					continue
				}

				s.Lock()

				s.db = c

				s.Unlock()

				ctx = context.WithValue(ctx, request.CtxKeyScopes,
					request.ScopeSuperuser)
				ctx = context.WithValue(ctx, request.CtxKeyAccountID,
					request.SystemAccount)

				if _, err := s.createAccount(ctx, &Account{
					ID: request.FieldString{
						Set: true, Valid: true, Value: s.cfg.ServiceName(),
					},
					Name: request.FieldString{
						Set: true, Valid: true, Value: s.cfg.ServiceName(),
					},
				}); err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to create account",
						"error", err)
				}

				if su := os.Getenv("SUPERUSER"); su != "" {
					if sp := os.Getenv("SUPERUSER_PASSWORD"); sp != "" {
						if _, err := s.createUser(ctx, &User{
							ID: request.FieldString{
								Set: true, Valid: true, Value: su,
							},
							Scopes: request.FieldString{
								Set: true, Valid: true,
								Value: request.ScopeSuperuser,
							},
							Password: &sp,
						}); err != nil {
							s.log.Log(ctx, logger.LvlError,
								"unable to create initial superuser",
								"error", err)
						}
					}
				}

				break
			}
		}(context.Background())
	})
}

// initRouter configures the server routing.
func (s *Server) initRouter() {
	base := chi.NewRouter()

	r := chi.NewRouter()

	base.Mount(s.cfg.ServerPathPrefix(), r)

	r.Use(
		s.context,
		s.header,
		s.logger,
	)

	r.NotFound(s.notFound)
	r.MethodNotAllowed(s.methodNotAllowed)

	r.Get("/debug/cmdline", pprof.Cmdline)
	r.Get("/debug/profile", pprof.Profile)
	r.Get("/debug/symbol", pprof.Symbol)
	r.Get("/debug/trace", pprof.Trace)
	r.Get("/debug/goroutine", pprof.Handler("goroutine").ServeHTTP)
	r.Get("/debug/heap", pprof.Handler("heap").ServeHTTP)
	r.Get("/debug/allocs", pprof.Handler("allocs").ServeHTTP)
	r.Get("/debug/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
	r.Get("/debug/block", pprof.Handler("block").ServeHTTP)
	r.Get("/debug/mutex", pprof.Handler("mutex").ServeHTTP)
	r.Get("/debug/pprof", pprof.Index)

	r.Mount("/healthz", s.HealthHandler())
	r.Mount("/health", s.HealthHandler())
	r.Mount("/account", s.accountHandler())
	r.Mount("/user", s.userHandler())
	r.Mount("/login", s.loginHandler())
	r.Mount("/games", s.gamesHandler())

	s.initStaticRoutes(r)

	s.Lock()

	s.r = base

	s.Unlock()
}

// initStaticRoutes initializes routing for embedded static games.
func (s *Server) initStaticRoutes(r chi.Router) {
	r.Get("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		v, err := static.FS.ReadFile("openapi.json")
		if err != nil {
			s.error(err, w, r)

			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		if _, err := w.Write(v); err != nil {
			s.error(err, w, r)

			return
		}
	})

	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		v, err := static.FS.ReadFile("openapi.yaml")
		if err != nil {
			s.error(err, w, r)

			return
		}

		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		if _, err := w.Write(v); err != nil {
			s.error(err, w, r)

			return
		}
	})

	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		v, err := static.FS.ReadFile("index.html")
		if err != nil {
			s.error(err, w, r)

			return
		}

		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		if _, err := w.Write(v); err != nil {
			s.error(err, w, r)

			return
		}
	})
}

// UpdateAuthConfig retrieves and begins periodic update of authentication
// configuration data, if configured to do so.
func (s *Server) UpdateAuthConfig() {
	s.authOnce.Do(func() {
		go func() {
			if s.cfg.AuthTokenWellKnown() == "" {
				return
			}

			for s.db == nil {
				time.Sleep(100 * time.Millisecond)
			}

			s.addCancelFunc(s.updateAuth(context.Background()))
		}()
	})
}

// Serve listens for and processes HTTP requests.
func (s *Server) Serve() error {
	ctx := context.Background()

	s.RLock()

	addr := s.addr

	s.RUnlock()

	s.log.Log(ctx, logger.LvlDebug, "starting server",
		"address", addr)

	if len(addr) == 0 {
		return errors.New(errors.ErrConfiguration,
			"no servers configured")
	}

	ech := make(chan error, len(addr))

	var wg sync.WaitGroup

	for _, a := range addr {
		wg.Add(1)

		go func(addr string) {
			defer wg.Done()

			lis, err := net.Listen("tcp", addr)
			if err != nil {
				ech <- errors.Wrap(err, errors.ErrServer,
					"server unable to start listening on "+addr)

				return
			}

			s.log.Log(ctx, logger.LvlInfo, "server listening",
				"address", addr)

			if err = s.Server.Serve(lis); err != nil {
				if err != http.ErrServerClosed {
					ech <- errors.Wrap(err, errors.ErrServer,
						"server error")

					return
				}
			}

			ech <- nil
		}(a)
	}

	go func() {
		wg.Wait()
		close(ech)
	}()

	for err := range ech {
		if err != nil {
			s.log.Log(ctx, logger.LvlError, "server error",
				"error", err)

			return err
		}
	}

	return nil
}

// Close releases all server games immediately.
func (s *Server) Close() {
	ctx := context.Background()

	s.Lock()

	s.log.Log(ctx, logger.LvlInfo, "server closing")

	s.health = http.StatusServiceUnavailable

	s.Unlock()

	s.RLock()

	defer s.RUnlock()

	if err := s.Server.Close(); err != nil {
		s.log.Log(ctx, logger.LvlError,
			"error during server close",
			"error", err)
	}

	for _, canc := range s.cancels {
		if canc != nil {
			canc()
		}
	}

	if s.db != nil {
		if err := s.db.Disconnect(ctx); err != nil {
			s.log.Log(ctx, logger.LvlError,
				"error during database disconnect",
				"error", err)
		}
	}
}

// Shutdown releases all server games gracefully.
func (s *Server) Shutdown(ctx context.Context) {
	s.Lock()

	s.log.Log(ctx, logger.LvlInfo, "server shutting down")

	s.health = http.StatusServiceUnavailable

	s.Unlock()

	s.RLock()

	defer s.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, s.cfg.ServerTimeout())

	defer cancel()

	if err := s.Server.Shutdown(ctx); err != nil {
		s.log.Log(ctx, logger.LvlError, "error during server shutdown",
			"error", err)

		if err := s.Server.Close(); err != nil {
			s.log.Log(ctx, logger.LvlError, "error during server close",
				"error", err)
		}

		return
	}

	for _, canc := range s.cancels {
		if canc != nil {
			canc()
		}
	}

	if s.db != nil {
		if err := s.db.Disconnect(ctx); err != nil {
			s.log.Log(ctx, logger.LvlError,
				"error during database disconnect",
				"error", err)
		}
	}
}

// context wraps request handlers to setup the request context.
func (s *Server) context(next http.Handler) http.Handler {
	timeout := s.cfg.ServerTimeout()
	if timeout == 0 {
		timeout = 30 * time.Second // Default 30 second timeout.
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)

		defer cancel()

		ctx = context.WithValue(ctx, request.CtxKeyService,
			s.cfg.ServiceName())

		if tID, err := request.ContextTraceID(ctx); err != nil || tID == "" {
			if tu, err := uuid.NewRandom(); err != nil {
				s.log.Log(ctx, logger.LvlError,
					"unable to create UUID for trace_id",
					"error", err,
					"request", r)
			} else {
				tID = tu.String()

				ctx = context.WithValue(ctx, request.CtxKeyTraceID, tID)
			}
		}

		if aID := r.Header.Get("X-Account-ID"); aID != "" {
			ctx = context.WithValue(ctx, request.CtxKeyAccountID, aID)
		}

		if v := r.Header.Get("X-No-Cache"); v != "" && v != "0" &&
			!strings.EqualFold(v, "f") && !strings.EqualFold(v, "false") {
			ctx = context.WithValue(ctx, request.CtxKeyNoCache, true)
		}

		if v := r.URL.Query().Get("no_cache"); v != "" && v != "0" &&
			!strings.EqualFold(v, "f") && !strings.EqualFold(v, "false") {
			ctx = context.WithValue(ctx, request.CtxKeyNoCache, true)
		}

		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body,
				s.cfg.ServerMaxRequestSize())
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// header wraps request handlers with default header values.
func (s *Server) header(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wd := s.cfg.ServerHost()

		if strings.HasSuffix(r.Header.Get("Origin"), "."+wd) ||
			r.Header.Get("Origin") == wd ||
			r.Header.Get("Origin") == "https://"+wd ||
			r.Header.Get("Origin") == "http://"+wd {
			originStr := r.Header.Get("Origin")

			w.Header().Set("Access-Control-Allow-Origin", originStr)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers",
				"Origin, X-Requested-With, X-HTTP-Method-Override, "+
					"Content-Type, Accept, Referer, User-Agent")
			w.Header().Set("Access-Control-Allow-Methods",
				"GET, PUT, POST, OPTIONS")
		}

		host, err := os.Hostname()
		if err != nil {
			host = "unknown"
		}

		w.Header().Set("X-Server", host)
		w.Header().Set("X-Version", Version)
		w.Header().Set("Vary", "Accept-Encoding, Origin")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if s.cfg.ServiceMaintenance() {
			s.error(errors.New(errors.ErrMaintenance,
				"The service is currently undergoing maintenance, "+
					"please try back later"), w, r)

			return
		}

		if r.Method == http.MethodOptions {
			s.noContent(w, r)

			return
		}

		next.ServeHTTP(w, r)
	})
}

// logger wraps request handlers with logging functionality.
func (s *Server) logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		r.Header.Set("X-Status-Code", "200")

		remote := r.RemoteAddr
		if r.Header.Get("X-Forwarded-For") != "" {
			remote = r.Header.Get("X-Forwarded-For")
		}

		ctx := context.WithValue(r.Context(), request.CtxKeyRemote, remote)

		logData := []any{
			"kind", r.Method,
			"uri", r.RequestURI,
			"remote", remote,
			"headers", r.Header,
		}

		s.log.Log(ctx, logger.LvlDebug, "request received", logData...)

		next.ServeHTTP(w, r.WithContext(ctx))

		sc, err := strconv.ParseInt(r.Header.Get("X-Status-Code"),
			10, 64)
		if err != nil {
			s.log.Log(ctx, logger.LvlWarn,
				"unable to retrieve status code from header",
				append([]any{"error", err}, logData...)...)
		}

		lvl := logger.LvlError
		if sc < http.StatusInternalServerError {
			lvl = logger.LvlWarn
		}

		if sc < http.StatusMultipleChoices {
			lvl = logger.LvlInfo
		}

		logData = append(logData,
			"latency", time.Since(start).String(),
			"status", sc,
			"route", chi.RouteContext(ctx).RoutePattern())

		if err != nil {
			s.log.Log(ctx, logger.LvlError,
				"unable to format audit event log data",
				"error", err,
				"log_data", logData)
		}

		s.log.Log(ctx, lvl, "request processed", logData...)
	})
}

// dbAvail wraps request handlers with a check to ensure the database is up.
func (s *Server) dbAvail(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.DB() == nil {
			s.error(errors.New(errors.ErrUnavailable,
				"The service database is currently unavailable, "+
					"please try back later"), w, r)

			return
		}

		next.ServeHTTP(w, r)
	})
}

// error responds to the current request with a standard error response.
func (s *Server) error(err error, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Ensure any error is wrapped and formatted.
	e, ok := err.(*errors.Error)
	if !ok {
		if errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) {
			e = errors.Context(ctx)
		} else {
			e = errors.Wrap(err, errors.ErrServer, err.Error())
		}
	}

	// Store the status code in context
	r.Header.Set("X-Status-Code", strconv.FormatInt(int64(e.Code.Status), 10))

	// Send information to the user if the service is under maintenance.
	if e.Code.Name == "Maintenance" {
		w.WriteHeader(e.Code.Status)

		if err := json.NewEncoder(w).Encode(map[string]string{
			"status": "The service is currently undergoing maintenance",
		}); err != nil {
			s.log.Log(ctx, logger.LvlError,
				"unable to encode error into JSON",
				"error", err)
		}

		return
	}

	logData := []any{
		"error", e,
		"kind", r.Method,
		"uri", r.RequestURI,
	}

	if str, err := request.ContextRequestBody(ctx); err == nil {
		logData = append(logData, "request_body", str)
	}

	lvl := logger.LvlError
	if e.Code.Status < http.StatusInternalServerError {
		lvl = logger.LvlWarn
	}

	remote, err := request.ContextRemote(ctx)
	if err != nil {
		remote = r.RemoteAddr
	}

	if remote != "" {
		logData = append(logData, "remote", remote)
	}

	route := "not found"

	rc := chi.RouteContext(ctx)

	if rc != nil {
		route = rc.RoutePattern()
	}

	s.log.Log(ctx, lvl, e.Msg, logData...)

	const (
		routeTag = "route:"
		codeTag  = "code:"
	)

	if mr := s.metric; mr != nil {
		mr.RecordValue(ctx, "status_code", float64(e.Code.Status),
			routeTag+route)
		mr.Increment(ctx, "status_code",
			codeTag+strconv.Itoa(e.Code.Status), routeTag+route)
	}

	if s.tracer != nil {
		span := trace.SpanFromContext(ctx)

		if span != nil {
			span.SetStatus(codes.Error, e.Msg)
			span.RecordError(e)
			span.SetAttributes(attribute.Int("http.status_code",
				e.Code.Status))
		}
	}

	w.WriteHeader(e.Code.Status)

	if err := json.NewEncoder(w).Encode(e); err != nil {
		s.log.Log(ctx, logger.LvlError,
			"unable to encode error into JSON",
			"error", err)
	}
}

// noContent is the handler function for empty responses.
func (s *Server) noContent(w http.ResponseWriter, _ *http.Request) {
	w.Header().Del("Content-Type")
	w.WriteHeader(http.StatusNoContent)
}

// notFound is the handler function for 404 errors.
func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	s.error(errors.New(errors.ErrNotFound,
		"game not found"), w, r)
}

// methodNotAllowed is the handler function for 405 errors.
func (s *Server) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	s.error(errors.New(errors.ErrNotAllowed,
		"method not allowed"), w, r)
}
