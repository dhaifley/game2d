package server_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/dhaifley/game2d/config"
	"github.com/dhaifley/game2d/logger"
	"github.com/dhaifley/game2d/server"
)

const (
	TestKey  = int64(1)
	TestID   = "1"
	TestUUID = "11223344-5566-7788-9900-aabbccddeeff"
	TestName = "test"
	basePath = config.DefaultServerPathPrefix
)

var servicesLock sync.Mutex

func TestNewClose(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svr.Close()
}

func TestShutdown(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svr.Shutdown(context.Background())
}

func TestServe(t *testing.T) {
	t.Parallel()

	cfg := config.NewDefault()

	sCfg := &config.ServerConfig{Address: ":18086"}

	sCfg.Load()

	cfg.SetServer(sCfg)

	svr, err := server.NewServer(cfg, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		err = svr.Serve()

		wg.Done()
	}()

	time.Sleep(time.Millisecond * 100)

	svr.Close()

	wg.Wait()

	if err != nil {
		t.Fatal(err)
	}
}

func TestHeader(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}

	tests := []struct {
		name      string
		w         *httptest.ResponseRecorder
		headers   map[string]string
		expCORS   string
		expAllow  string
		expServer string
		expCode   int
	}{{
		name: "options CORS",
		w:    httptest.NewRecorder(),
		headers: map[string]string{
			"Origin": "https://game2d.ai",
		},
		expCORS:   "GET, PUT, POST, OPTIONS",
		expAllow:  "https://game2d.ai",
		expServer: host,
		expCode:   http.StatusNoContent,
	}, {
		name: "options invalid origin",
		w:    httptest.NewRecorder(),
		headers: map[string]string{
			"Origin": "test.api.foo",
		},
		expCORS:   "",
		expAllow:  "",
		expServer: host,
		expCode:   http.StatusNoContent,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := http.NewRequest(http.MethodOptions, basePath+"/", nil)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			for hk, hv := range tt.headers {
				r.Header.Set(hk, hv)
			}

			svr.Mux(tt.w, r)

			if tt.w.Code != tt.expCode {
				t.Errorf("Status code expected: %v, got: %v",
					tt.expCode, tt.w.Code)
			}

			if tt.expServer != tt.w.Header().Get("X-Server") {
				t.Errorf("X-Server expected: %v, got: %v",
					tt.expServer, tt.w.Header().Get("X-Server"))
			}

			if tt.expCORS != tt.w.Header().Get("Access-Control-Allow-Methods") {
				t.Errorf("Access-Control-Allow-Methods expected: %v, got: %v",
					tt.expCORS, tt.w.Header().Get("Access-Control-Allow-Methods"))
			}

			if tt.expAllow != tt.w.Header().Get("Access-Control-Allow-Origin") {
				t.Errorf("Access-Control-Allow-Origin expected: %v, got: %v",
					tt.expAllow, tt.w.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func BenchmarkServerPostResource(b *testing.B) {
	l := logger.New(logger.OutStderr, logger.FmtJSON, logger.LvlInfo)

	os.Setenv("AUTH_TOKEN_PUBLIC_KEY_FILE", "../../certs/tls.crt")

	os.Setenv("AUTH_TOKEN_PRIVATE_KEY_FILE", "../../certs/tls.key")

	c := config.NewDefault()

	svr, err := server.NewServer(c, l, nil, nil)
	if err != nil {
		b.Fatal(err)
	}

	svr.ConnectDB()

	for svr.DB() == nil {
		time.Sleep(time.Millisecond * 100)
	}

	authToken := ""

	if v := os.Getenv("USER_AUTH_TOKEN"); v != "" {
		authToken = v
	}

	w := httptest.NewRecorder()

	u := "https://localhost:8080/api/v1/resources"

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		buf := bytes.NewBufferString(`{"name":"test","key_field":"test"}`)

		r, err := http.NewRequest(http.MethodPost, u, buf)
		if err != nil {
			b.Fatal("Failed to initialize request", err)
		}

		if authToken != "" {
			r.Header.Set("Authorization", "Bearer "+authToken)
		} else {
			r.Header.Set("Authorization", "test")
		}

		b.StartTimer()

		svr.Mux(w, r)
	}
}

func BenchmarkServerGetResource(b *testing.B) {
	l := logger.New(logger.OutStderr, logger.FmtJSON, logger.LvlInfo)

	os.Setenv("AUTH_TOKEN_PUBLIC_KEY_FILE", "../../certs/tls.crt")

	os.Setenv("AUTH_TOKEN_PUBLIC_KEY_FILE", "../../certs/tls.crt")

	os.Setenv("AUTH_TOKEN_PRIVATE_KEY_FILE", "../../certs/tls.key")

	c := config.NewDefault()

	svr, err := server.NewServer(c, l, nil, nil)
	if err != nil {
		b.Fatal(err)
	}

	svr.ConnectDB()

	for svr.DB() == nil {
		time.Sleep(time.Millisecond * 100)
	}

	authToken := ""

	if v := os.Getenv("USER_AUTH_TOKEN"); v != "" {
		authToken = v
	}

	w := httptest.NewRecorder()

	u := "https://localhost:8080/v1/api/resources?size=1"

	r, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		b.Fatal("Failed to initialize request", err)
	}

	if authToken != "" {
		r.Header.Set("Authorization", "Bearer "+authToken)
	} else {
		r.Header.Set("Authorization", "test")
	}

	for i := 0; i < b.N; i++ {
		svr.Mux(w, r)
	}
}
