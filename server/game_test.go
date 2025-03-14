package server_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dhaifley/game2d/request"
	"github.com/dhaifley/game2d/server"
)

var TestGame = server.Game{
	ID: request.FieldString{
		Set: true, Valid: true,
		Value: TestUUID,
	},
	Name: request.FieldString{
		Set: true, Valid: true,
		Value: "testName",
	},
	Version: request.FieldString{
		Set: true, Valid: true,
		Value: "1",
	},
	Description: request.FieldString{
		Set: true, Valid: true,
		Value: "testDescription",
	},
	Status: request.FieldString{
		Set: true, Valid: true,
		Value: request.StatusNew,
	},
	StatusData: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"last_error": "testError",
		},
	},
	Source: request.FieldString{
		Set: true, Valid: true,
		Value: "testSource",
	},
	CommitHash: request.FieldString{
		Set: true, Valid: true,
		Value: "testHash",
	},
	CreatedBy: request.FieldString{
		Set: true, Valid: true,
		Value: TestID,
	},
	CreatedAt: request.FieldTime{
		Set: true, Valid: true,
		Value: 1,
	},
	UpdatedBy: request.FieldString{
		Set: true, Valid: true,
		Value: TestID,
	},
	UpdatedAt: request.FieldTime{
		Set: true, Valid: true,
		Value: 1,
	},
}

func TestSearchGame(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		query  string
		header map[string]string
		code   int
		resp   string
	}{{
		name:   "success",
		w:      httptest.NewRecorder(),
		url:    basePath + "/games",
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp:   `"game_id":"` + TestGame.ID.Value + `"`,
	}, {
		name:   "summary",
		w:      httptest.NewRecorder(),
		url:    basePath + "/games",
		query:  `?search=and(name:*)&summary=status`,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp:   `"count":1`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u := tt.url + tt.query

			r, err := http.NewRequest(http.MethodGet, u, nil)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			svr.Mux(tt.w, r)

			if tt.w.Code != tt.code {
				t.Errorf("Code expected: %v, got: %v", tt.code, tt.w.Code)
			}

			res := tt.w.Body.String()
			if !strings.Contains(res, tt.resp) {
				t.Errorf("Expected body to contain: %v, got: %v", tt.resp, res)
			}
		})
	}
}

func TestGetGame(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		header map[string]string
		code   int
		resp   string
	}{{
		name:   "success",
		w:      httptest.NewRecorder(),
		url:    basePath + "/games/" + TestGame.ID.Value,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp:   `"game_id":"` + TestGame.ID.Value + `"`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := http.NewRequest(http.MethodGet, tt.url, nil)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			svr.Mux(tt.w, r)

			if tt.w.Code != tt.code {
				t.Errorf("Code expected: %v, got: %v", tt.code, tt.w.Code)
			}

			res := tt.w.Body.String()
			if !strings.Contains(res, tt.resp) {
				t.Errorf("Expected body to contain: %v, got: %v", tt.resp, res)
			}
		})
	}
}

func TestPostGame(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		body   string
		header map[string]string
		code   int
		resp   string
	}{{
		name: "success",
		w:    httptest.NewRecorder(),
		url:  basePath + "/games",
		body: `{
			"event_id": "` + TestGame.ID.Value + `",
			"name":"test",
			"status":"` + request.StatusActive + `"
		}`,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusCreated,
		resp:   `"game_id":"` + TestGame.ID.Value + `"`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewBufferString(tt.body)

			r, err := http.NewRequest(http.MethodPost, tt.url, buf)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			svr.Mux(tt.w, r)

			if tt.w.Code != tt.code {
				t.Errorf("Code expected: %v, got: %v", tt.code, tt.w.Code)
			}

			res := tt.w.Body.String()
			if !strings.Contains(res, tt.resp) {
				t.Errorf("Expected body to contain: %v, got: %v", tt.resp, res)
			}
		})
	}
}

func TestPutGame(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		body   string
		header map[string]string
		code   int
		resp   string
	}{{
		name: "success",
		w:    httptest.NewRecorder(),
		url:  basePath + "/games/" + TestGame.ID.Value,
		body: `{
			"name": "changed",
			"status":"` + request.StatusActive + `"
		}`,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp:   `"game_id":"` + TestGame.ID.Value + `"`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewBufferString(tt.body)

			r, err := http.NewRequest(http.MethodPut, tt.url, buf)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			svr.Mux(tt.w, r)

			if tt.w.Code != tt.code {
				t.Errorf("Code expected: %v, got: %v", tt.code, tt.w.Code)
			}

			res := tt.w.Body.String()

			if !strings.Contains(res, tt.resp) {
				t.Errorf("Expected body to contain: %v, got: %v", tt.resp, res)
			}
		})
	}
}

func TestDeleteGame(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		header map[string]string
		code   int
	}{{
		name:   "success",
		w:      httptest.NewRecorder(),
		url:    basePath + "/games/" + TestGame.ID.Value,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusNoContent,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := http.NewRequest(http.MethodDelete, tt.url, nil)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			svr.Mux(tt.w, r)

			if tt.w.Code != tt.code {
				t.Errorf("Code expected: %v, got: %v", tt.code, tt.w.Code)
			}
		})
	}
}

func TestPostUpdateGames(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		body   string
		header map[string]string
		code   int
		resp   string
	}{{
		name: "success",
		w:    httptest.NewRecorder(),
		url:  basePath + "/games/update/" + TestID + "/" + TestUUID,
		body: `{
			"games": [
				{
					"game_id": "` + TestUUID + `",
					"account_id": "` + TestID + `",
					"cleared_on": 1
				}
			]
		}`,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp:   `"game_id":"` + TestUUID + `"`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewBufferString(tt.body)

			r, err := http.NewRequest(http.MethodPost, tt.url, buf)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			svr.Mux(tt.w, r)

			if tt.w.Code != tt.code {
				t.Errorf("Code expected: %v, got: %v", tt.code, tt.w.Code)
			}

			res := tt.w.Body.String()
			if !strings.Contains(res, tt.resp) {
				t.Errorf("Expected body to contain: %v, got: %v", tt.resp, res)
			}
		})
	}
}

func TestPostImportGames(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		header map[string]string
		code   int
	}{{
		name:   "success",
		w:      httptest.NewRecorder(),
		url:    basePath + "/games/import",
		header: map[string]string{"Authorization": "admin"},
		code:   http.StatusNoContent,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := http.NewRequest(http.MethodPost, tt.url, nil)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			svr.Mux(tt.w, r)

			if tt.w.Code != tt.code {
				t.Errorf("Code expected: %v, got: %v", tt.code, tt.w.Code)
			}
		})
	}
}
