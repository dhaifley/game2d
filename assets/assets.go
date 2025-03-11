package assets

import "embed"

//go:embed *.png *.lua
var Assets embed.FS

// GetImage retrieves an image from the embedded assets.
func GetImage(name string) ([]byte, error) {
	data, err := Assets.ReadFile(name)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// GetScript retrieves a script from the embedded assets.
func GetScript(name string) (string, error) {
	data, err := Assets.ReadFile(name)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
