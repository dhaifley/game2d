package client

import (
	"encoding/base64"
	"encoding/json"
)

// Script values represent the scripts in the game.
type Script struct {
	id, name string
	data     string
}

// MarshalJSON serializes the script to JSON.
func (s *Script) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Data string `json:"data,omitempty"`
	}{
		ID:   s.id,
		Name: s.name,
		Data: base64.StdEncoding.EncodeToString([]byte(s.data)),
	})
}

// UnmarshalJSON deserializes the image from JSON.
func (s *Script) UnmarshalJSON(data []byte) error {
	v := &struct {
		ID      string   `json:"id"`
		Name    string   `json:"name"`
		Data    string   `json:"data,omitempty"`
		Include []string `json:"include,omitempty"`
	}{}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	s.id = v.ID
	s.name = v.Name

	b, err := base64.StdEncoding.DecodeString(v.Data)
	if err != nil {
		return err
	}

	s.data = string(b)

	return nil
}

// NewScript creates and initializes a new script object.
func NewScript(id, name, data string, include []string) *Script {
	return &Script{
		id:   id,
		name: name,
		data: data,
	}
}
