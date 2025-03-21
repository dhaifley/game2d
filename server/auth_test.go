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

var TestAccount = server.Account{
	ID: request.FieldString{
		Set: true, Valid: true,
		Value: TestID,
	},
	Name: request.FieldString{
		Set: true, Valid: true,
		Value: "testAccount",
	},
	Status: request.FieldString{
		Set: true, Valid: true,
		Value: request.StatusActive,
	},
	StatusData: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"last_error": "test",
		},
	},
	Repo: request.FieldString{
		Set: true, Valid: true,
		Value: "test",
	},
	RepoStatus: request.FieldString{
		Set: true, Valid: true,
		Value: request.StatusActive,
	},
	RepoStatusData: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"last_error": "test",
		},
	},
	Secret: request.FieldString{
		Set: true, Valid: true,
		Value: "test",
	},
	Data: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"test": "test",
		},
	},
}

var TestUser = server.User{
	ID: request.FieldString{
		Set: true, Valid: true,
		Value: TestUUID,
	},
	Email: request.FieldString{
		Set: true, Valid: true,
		Value: "test@game2d.ai",
	},
	LastName: request.FieldString{
		Set: true, Valid: true,
		Value: "testLastName",
	},
	FirstName: request.FieldString{
		Set: true, Valid: true,
		Value: "testFirstName",
	},
	Status: request.FieldString{
		Set: true, Valid: true,
		Value: request.StatusActive,
	},
	Scopes: request.FieldString{
		Set: true, Valid: true,
		Value: request.ScopeSuperuser,
	},
	Data: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"test": "test",
		},
	},
}

func TestAccountServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests")
	}

	t.Parallel()

	data := map[string]any{}

	dataLock := sync.Mutex{}

	tests := []struct {
		name   string
		url    string
		method string
		header map[string]string
		body   map[string]any
		resp   func(t *testing.T, res *http.Response)
	}{{
		name:   "unauthorized",
		url:    "http://localhost:8080/api/v1/account",
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

			dataLock.Unlock()
		},
	}, {
		name:   "get account",
		url:    "http://localhost:8080/api/v1/account",
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

			expB := `"id":"`

			if !strings.Contains(string(b), expB) {
				t.Errorf("Expected body to contain: %v, got: %v",
					expB, string(b))
			}
		},
	}, {
		name:   "post account",
		url:    "http://localhost:8080/api/v1/account",
		method: http.MethodPost,
		body: map[string]any{
			"id":     "test-account",
			"name":   "test-account",
			"status": "active",
			"secret": "test",
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

			expB := `"id":"`

			if !strings.Contains(string(b), expB) {
				t.Errorf("Expected body to contain: %v, got: %v",
					expB, string(b))
			}
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}

			if len(tt.body) > 0 {
				if ct, ok := tt.header["Content-Type"]; ok {
					if !strings.Contains("json", ct) {
						form := url.Values{}

						for k, v := range tt.body {
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

				if buf.Len() == 0 {
					b, err := json.Marshal(tt.body)
					if err != nil {
						t.Error(err)
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

func TestUserServer(t *testing.T) {
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
		body   map[string]any
		resp   func(t *testing.T, res *http.Response)
	}{{
		name:   "unauthorized",
		url:    "http://localhost:8080/api/v1/user",
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

			dataLock.Unlock()
		},
	}, {
		name:   "get user",
		url:    "http://localhost:8080/api/v1/user",
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

			expB := `"id":"`

			if !strings.Contains(string(b), expB) {
				t.Errorf("Expected body to contain: %v, got: %v",
					expB, string(b))
			}
		},
	}, {
		name:   "patch user",
		url:    "http://localhost:8080/api/v1/user",
		method: http.MethodPatch,
		body: map[string]any{
			"data": map[string]any{
				"test": "test",
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

			expB := `"id":"`

			if !strings.Contains(string(b), expB) {
				t.Errorf("Expected body to contain: %v, got: %v",
					expB, string(b))
			}
		},
	}, {
		name:   "put user",
		url:    "http://localhost:8080/api/v1/user",
		method: http.MethodPut,
		body: map[string]any{
			"email":      "test@test.com",
			"first_name": "Test",
			"last_name":  "User",
			"status":     "active",
			"scopes":     request.ScopeSuperuser,
			"data": map[string]any{
				"test": "test",
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

			expB := `"id":"`

			if !strings.Contains(string(b), expB) {
				t.Errorf("Expected body to contain: %v, got: %v",
					expB, string(b))
			}
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}

			if len(tt.body) > 0 {
				if ct, ok := tt.header["Content-Type"]; ok {
					if !strings.Contains("json", ct) {
						form := url.Values{}

						for k, v := range tt.body {
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

				if buf.Len() == 0 {
					b, err := json.Marshal(tt.body)
					if err != nil {
						t.Error(err)
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
