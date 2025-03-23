package assets_test

import (
	"testing"

	"github.com/dhaifley/game2d/assets"
)

func TestGetImage(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "existing file",
			filename: "avatar.svg",
			wantErr:  false,
		},
		{
			name:     "non-existing file",
			filename: "nonexistent.svg",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := assets.GetImage(tt.filename)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetImage() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && len(got) == 0 {
				t.Errorf("GetImage() returned empty data, expected non-empty")
			}
		})
	}
}

func TestGetScript(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "existing file",
			filename: "avatar.lua",
			wantErr:  false,
		},
		{
			name:     "non-existing file",
			filename: "nonexistent.lua",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := assets.GetScript(tt.filename)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetScript() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && len(got) == 0 {
				t.Errorf("GetScript() returned empty data, expected non-empty")
			}
		})
	}
}
