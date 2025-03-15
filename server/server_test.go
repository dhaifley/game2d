package server_test

import (
	"bytes"
	"context"
	"fmt"
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

func TestMain(m *testing.M) {
	for _, arg := range os.Args {
		if arg == "-test.short=true" {
			// Skipping integration tests.
			os.Exit(0)
		}
	}

	su := os.Getenv("SUPERUSER")
	if su == "" {
		su = "admin"
	}

	sp := os.Getenv("SUPERUSER_PASSWORD")
	if sp == "" {
		sp = "admin"
	}

	os.Setenv("SUPERUSER", su)
	os.Setenv("SUPERUSER_PASSWORD", sp)

	cfg := config.NewDefault()

	svr, err := server.NewServer(cfg,
		logger.New(cfg.LogOut(), cfg.LogFormat(), cfg.LogLevel()), nil, nil)
	if err != nil {
		fmt.Println("server init error", err)
		os.Exit(1)
	}

	svr.ConnectDB()

	for svr.DB() == nil {
		time.Sleep(time.Millisecond * 100)
	}

	svr.UpdateGameImports()

	ctx := context.Background()

	go func() {
		if err := svr.Serve(); err != nil {
			fmt.Println("server error", err)

			os.Exit(1)
		}
	}()

	time.Sleep(time.Second)

	code := m.Run()

	svr.Shutdown(ctx)

	os.Exit(code)
}

func BenchmarkServerPostGame(b *testing.B) {
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

	u := "https://localhost:8080/api/v1/games"

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

func BenchmarkServerGetGame(b *testing.B) {
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

	u := "https://localhost:8080/v1/api/games?size=1"

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
