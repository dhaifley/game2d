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
)

func TestStatsServer(t *testing.T) {
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
		name:   "search games empty",
		url:    "http://localhost:8080/api/v1/health",
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
				t.Errorf("Unexpected error decoding response: %v", err)
			}

			health, ok := m["health"].(float64)
			if !ok {
				t.Errorf("Expected health in response: %v", m)
			}

			dataLock.Lock()
			data["health"] = health
			dataLock.Unlock()
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
