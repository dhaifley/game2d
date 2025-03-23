package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
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

func TestGamesServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests")
	}

	data := map[string]any{}

	dataLock := sync.Mutex{}

	tests := []struct {
		name   string
		url    string
		method string
		header map[string]string
		body   any
		resp   func(t *testing.T, res *http.Response)
	}{{
		name:   "unauthorized",
		url:    "http://localhost:8080/api/v1/games",
		method: http.MethodGet,
		header: map[string]string{"Authorization": "test"},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusUnauthorized

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			expB := `"Unauthorized"`

			if !strings.Contains(string(b), expB) {
				t.Errorf("Expected body to contain: %v, got: %v",
					expB, string(b))
			}
		},
	}, {
		name:   "password login",
		url:    "http://localhost:8080/api/v1/login/token",
		method: http.MethodPost,
		header: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		body: map[string]any{
			"username": "admin",
			"password": "admin",
			"scope":    request.ScopeSuperuser,
		},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusOK

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			m := map[string]any{}

			json.Unmarshal(b, &m)
			if err != nil {
				t.Errorf("Unexpected error decoding response: %v", err)
			}

			at, ok := m["access_token"].(string)
			if !ok {
				t.Errorf("Unexpected response: %v", m)
			}

			if len(at) < 8 {
				t.Errorf("Expected access token, got: %v", at)
			}

			dataLock.Lock()
			data["access_token"] = at
			data["id"] = ""
			dataLock.Unlock()
		},
	}, {
		name:   "search games empty",
		url:    `http://localhost:8080/api/v1/games`,
		method: http.MethodGet,
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusOK

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			var games []any
			if err := json.Unmarshal(b, &games); err != nil {
				t.Errorf("Unexpected error decoding response: %v", err)
			}
		},
	}, {
		name:   "create game",
		url:    "http://localhost:8080/api/v1/games",
		method: http.MethodPost,
		body: map[string]any{
			"id":          TestUUID,
			"name":        "Test Game",
			"version":     "1",
			"description": "A test game",
			"status":      "active",
			"data": map[string]any{
				"test": "test",
			},
		},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusCreated

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			m := map[string]any{}

			if err := json.Unmarshal(b, &m); err != nil {
				t.Errorf("Unexpected error decoding response: %v", err)
			}

			gameID, ok := m["id"].(string)
			if !ok {
				t.Errorf("Expected id in response: %v", m)
			}

			dataLock.Lock()
			data["id"] = gameID
			dataLock.Unlock()
		},
	}, {
		name:   "get game",
		url:    "http://localhost:8080/api/v1/games/{{id}}",
		method: http.MethodGet,
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusOK

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			m := map[string]any{}

			if err := json.Unmarshal(b, &m); err != nil {
				t.Errorf("Unexpected error decoding response: %v ", err)
			}

			if _, ok := m["id"].(string); !ok {
				t.Errorf("Expected id in response: %v", m)
			}
		},
	}, {
		name:   "patch game",
		url:    "http://localhost:8080/api/v1/games/{{id}}",
		method: http.MethodPatch,
		body: map[string]any{
			"description": "Updated test game",
			"data": map[string]any{
				"updated": true,
			},
		},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusOK

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			m := map[string]any{}

			if err := json.Unmarshal(b, &m); err != nil {
				t.Errorf("Unexpected error decoding response: %v", err)
			}

			desc, ok := m["description"].(string)
			if !ok || desc != "Updated test game" {
				t.Errorf("Expected updated description in response: %v", m)
			}
		},
	}, {
		name:   "put game",
		url:    "http://localhost:8080/api/v1/games/{{id}}",
		method: http.MethodPut,
		body: map[string]any{
			"name":        "Test Game Updated",
			"version":     "2",
			"description": "A fully updated test game",
			"status":      "active",
			"data": map[string]any{
				"test": "updated",
			},
		},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusOK

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			m := map[string]any{}

			if err := json.Unmarshal(b, &m); err != nil {
				t.Errorf("Unexpected error decoding response: %v", err)
			}

			version, ok := m["version"].(string)
			if !ok || version != "2" {
				t.Errorf("Expected updated version in response: %v", m)
			}
		},
	}, {
		name:   "get game tags",
		url:    "http://localhost:8080/api/v1/games/{{id}}/tags",
		method: http.MethodGet,
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusOK

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			var tags []any
			if err := json.Unmarshal(b, &tags); err != nil {
				t.Errorf("Unexpected error decoding response: %v %v", err, tags)
			}
		},
	}, {
		name:   "create game tags",
		url:    "http://localhost:8080/api/v1/games/{{id}}/tags",
		method: http.MethodPost,
		body:   []string{"test:tag1", "test:tag2"},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusCreated

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			var tags []string
			if err := json.Unmarshal(b, &tags); err != nil {
				t.Errorf("Unexpected error decoding response: %v", err)
			}

			if len(tags) != 2 {
				t.Errorf("Expected 2 tags, got: %v", len(tags))
			}
		},
	}, {
		name:   "delete game tags",
		url:    "http://localhost:8080/api/v1/games/{{id}}/tags",
		method: http.MethodDelete,
		body:   []string{"test:tag1", "test:tag2"},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusNoContent

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}
		},
	}, {
		name:   "prompt game",
		url:    "http://localhost:8080/api/v1/games/prompt",
		method: http.MethodPost,
		body:   map[string]any{"game_id": TestUUID, "prompt": "test"},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusCreated

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			m := map[string]any{}

			if err := json.Unmarshal(b, &m); err != nil {
				t.Errorf("Unexpected error decoding response: %v ", err)
			}

			id, ok := m["game_id"].(string)
			if !ok {
				t.Errorf("Expected game_id in response: %v", m)
			}

			dataLock.Lock()
			data["game_id"] = id
			dataLock.Unlock()
		},
	}, {
		name:   "undo prompt",
		url:    "http://localhost:8080/api/v1/games/undo",
		method: http.MethodPost,
		body:   map[string]any{"game_id": TestUUID},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusCreated

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			m := map[string]any{}

			if err := json.Unmarshal(b, &m); err != nil {
				t.Errorf("Unexpected error decoding response: %v ", err)
			}

			if _, ok := m["game_id"].(string); !ok {
				t.Errorf("Expected id in response: %v", m)
			}
		},
	}, {
		name:   "copy game",
		url:    "http://localhost:8080/api/v1/games/copy",
		method: http.MethodPost,
		body:   map[string]any{"id": TestUUID},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusCreated

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			m := map[string]any{}

			if err := json.Unmarshal(b, &m); err != nil {
				t.Errorf("Unexpected error decoding response: %v ", err)
			}

			id, ok := m["id"].(string)
			if !ok {
				t.Errorf("Expected id in response: %v", m)
			}

			dataLock.Lock()
			data["copy_id"] = id
			dataLock.Unlock()
		},
	}, {
		name:   "delete game copy",
		url:    "http://localhost:8080/api/v1/games/{{copy_id}}",
		method: http.MethodDelete,
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusNoContent

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}
		},
	}, {
		name:   "delete game",
		url:    "http://localhost:8080/api/v1/games/{{id}}",
		method: http.MethodDelete,
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusNoContent

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if strings.Contains(tt.url, "{{id}}") {
				dataLock.Lock()
				gameID, _ := data["id"].(string)
				dataLock.Unlock()

				tt.url = strings.ReplaceAll(tt.url, "{{id}}",
					gameID)
			}

			if strings.Contains(tt.url, "{{copy_id}}") {
				dataLock.Lock()
				gameID, _ := data["copy_id"].(string)
				dataLock.Unlock()

				tt.url = strings.ReplaceAll(tt.url, "{{copy_id}}",
					gameID)
			}

			buf := &bytes.Buffer{}

			if tt.body != nil {
				if ct, ok := tt.header["Content-Type"]; ok {
					if !strings.Contains("json", ct) {
						if bm, ok := tt.body.(map[string]any); ok {
							form := url.Values{}

							for k, v := range bm {
								switch vv := v.(type) {
								case string:
									form.Add(k, vv)
								default:
									b, err := json.Marshal(vv)
									if err != nil {
										t.Error(err)
									}

									form.Add(k, string(b))
								}
							}

							buf = bytes.NewBufferString(form.Encode())
						}
					}
				}

				if buf.Len() == 0 {
					var (
						b   []byte
						err error
					)

					if bm, ok := tt.body.(map[string]any); ok {
						dataLock.Lock()
						gID, _ := data["game_id"].(string)
						dataLock.Unlock()

						if gID != "" {
							bm["game_id"] = gID
						}

						b, err = json.Marshal(bm)
						if err != nil {
							t.Error(err)
						}
					} else {
						b, err = json.Marshal(tt.body)
						if err != nil {
							t.Error(err)
						}
					}

					buf = bytes.NewBuffer(b)
				}
			}

			var br io.Reader

			if buf.Len() > 0 {
				br = buf
			}

			r, err := http.NewRequest(tt.method, tt.url, br)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			dataLock.Lock()

			at, _ := data["access_token"].(string)

			dataLock.Unlock()

			if at != "" {
				if tt.header == nil {
					tt.header = map[string]string{}
				}

				tt.header["Authorization"] = "Bearer " + at
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			res, err := http.DefaultClient.Do(r)
			if err != nil {
				t.Errorf("Unexpected client error: %v", err)
			}

			defer res.Body.Close()

			tt.resp(t, res)
		})
	}
}
