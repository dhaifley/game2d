package client_test

import (
	"encoding/json"
	"testing"

	"github.com/dhaifley/empty/client"
	"github.com/stretchr/testify/assert"
)

const (
	TestScript = "function Update(data)\nend"
)

func TestNewScript(t *testing.T) {
	include := []string{"image", "script"}

	script := client.NewScript(TestID, TestName, TestScript, include)
	assert.NotNil(t, script, "Script should not be nil")
}

func TestScriptJSONMarshaling(t *testing.T) {
	include := []string{"image", "script"}

	originalScript := client.NewScript(TestID, TestName, TestScript, include)

	data, err := json.Marshal(originalScript)
	assert.NoError(t, err, "Marshal should not return an error")

	var newScript client.Script

	err = json.Unmarshal(data, &newScript)
	assert.NoError(t, err, "Unmarshal should not return an error")

	originalJSON, err := json.Marshal(originalScript)
	assert.NoError(t, err)

	newJSON, err := json.Marshal(&newScript)
	assert.NoError(t, err)

	assert.JSONEq(t, string(originalJSON), string(newJSON),
		"Original and unmarshaled scripts should be equal")
}
