package request_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/dhaifley/game2d/request"
	"go.mongodb.org/mongo-driver/v2/bson"
	"gopkg.in/yaml.v3"
)

func TestFieldJSON(t *testing.T) {
	t.Parallel()

	type tests struct {
		Value       request.FieldString      `json:"value"`
		ZeroValue   request.FieldString      `json:"zero_value"`
		Null        request.FieldString      `json:"null"`
		NotSet      request.FieldString      `json:"not_set"`
		Int64       request.FieldInt64       `json:"int64"`
		Float64     request.FieldFloat64     `json:"float64"`
		Bool        request.FieldBool        `json:"bool"`
		Time        request.FieldTime        `json:"time"`
		StringArray request.FieldStringArray `json:"string_array"`
		Int64Array  request.FieldInt64Array  `json:"int64_array"`
		JSON        request.FieldJSON        `json:"json"`
		Duration    request.FieldDuration    `json:"duration"`
	}

	s := `{
		"null":null,
		"value":"test",
		"zero_value":"",
		"int64":1,
		"int64_array": [1, 2, 3],
		"float64":1.1,
		"bool":true,
		"time": "1970-01-01T00:00:01Z",
		"string_array":["test","test2"],
		"json":{"test":"test"},
		"duration":"1s"
	}`

	var v *tests

	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatal(err)
	}

	if !v.Value.Set {
		t.Error("Expected value to be set")
	}

	if !v.Value.Valid {
		t.Error("Expected value to be valid")
	}

	exp := "test"

	if v.Value.Value != exp {
		t.Errorf("Expected value: %v, got: %v", exp, v.Value.Value)
	}

	if !v.ZeroValue.Set {
		t.Error("Expected zero value to be set")
	}

	if !v.ZeroValue.Valid {
		t.Error("Expected zero value to be valid")
	}

	exp = ""

	if v.ZeroValue.Value != exp {
		t.Errorf("Expected zero value: %v, got: %v", exp, v.Value.Value)
	}

	if !v.Null.Set {
		t.Error("Expected null value to be set")
	}

	if v.Null.Valid {
		t.Error("Expected null value not to be valid")
	}

	if v.NotSet.Set {
		t.Error("Expected not set value not to be set")
	}

	if v.NotSet.Valid {
		t.Error("Expected not set value not to be valid")
	}

	if !v.Int64.Set {
		t.Error("Expected int64 value to be set")
	}

	if !v.Int64.Valid {
		t.Error("Expected int64 value to be valid")
	}

	if v.Int64.Value != 1 {
		t.Errorf("Expected int64 value: 1, got: %v", v.Int64.Value)
	}

	if !v.Float64.Set {
		t.Error("Expected float64 value to be set")
	}

	if !v.Float64.Valid {
		t.Error("Expected float64 value to be valid")
	}

	if v.Float64.Value != 1.1 {
		t.Errorf("Expected float64 value: 1.1, got: %v", v.Float64.Value)
	}

	if !v.Bool.Set {
		t.Error("Expected bool value to be set")
	}

	if !v.Bool.Valid {
		t.Error("Expected bool value to be valid")
	}

	if v.Bool.Value != true {
		t.Errorf("Expected bool value: 1, got: %v", v.Bool.Value)
	}

	if !v.Time.Set {
		t.Error("Expected bool value to be set")
	}

	if !v.Time.Valid {
		t.Error("Expected bool value to be valid")
	}

	if v.Time.Value != 1 {
		t.Errorf("Expected time value: 1, got: %v", v.Time.Value)
	}

	if !v.StringArray.Set {
		t.Error("Expected string slice value to be set")
	}

	if !v.StringArray.Valid {
		t.Error("Expected string slice value to be valid")
	}

	if len(v.StringArray.Value) != 2 {
		t.Errorf("Expected string slice length: 2, got: %v",
			len(v.StringArray.Value))
	}

	exp = "test"

	if v.StringArray.Value[0] != exp {
		t.Errorf("Expected string slice value: %v, got: %v",
			exp, v.StringArray.Value[0])
	}

	if !v.Int64Array.Set {
		t.Error("Expected int64 slice value to be set")
	}

	if !v.Int64Array.Valid {
		t.Error("Expected int64 slice value to be valid")
	}

	if len(v.Int64Array.Value) != 3 {
		t.Errorf("Expected int64 slice length: 3, got: %v",
			len(v.Int64Array.Value))
	}

	if v.Int64Array.Value[0] != 1 {
		t.Errorf("Expected int64 slice value: 1, got: %v",
			v.Int64Array.Value[0])
	}

	if !v.JSON.Set {
		t.Error("Expected JSON value to be set")
	}

	if !v.JSON.Valid {
		t.Error("Expected JSON value to be valid")
	}

	if len(v.JSON.Value) != 1 {
		t.Errorf("Expected JSON length: 1, got: %v",
			len(v.JSON.Value))
	}

	if v.JSON.Value[exp] != exp {
		t.Errorf("Expected JSON value: %v, got: %v",
			exp, v.JSON.Value[exp])
	}

	if !v.Duration.Set {
		t.Error("Expected duration value to be set")
	}

	if !v.Duration.Valid {
		t.Error("Expected duration value to be valid")
	}

	if v.Duration.Value != time.Second {
		t.Errorf("Expected duration value: 1, got: %v", v.Duration.Value)
	}

	b, err := json.Marshal(&v)
	if err != nil {
		t.Fatal(err)
	}

	exp = `{"value":"test","zero_value":"","null":null,"not_set":null,` +
		`"int64":1,"float64":1.1,"bool":true,"time":1,` +
		`"string_array":["test","test2"],"int64_array":[1,2,3],` +
		`"json":{"test":"test"},"duration":"1s"}`

	if string(b) != exp {
		t.Errorf("Expected JSON: %v, got: %v", exp, string(b))
	}
}

func TestFieldBSON(t *testing.T) {
	t.Parallel()

	type tests struct {
		Null        request.FieldString      `bson:"null,omitempty"`
		NotSet      request.FieldString      `bson:"not_set,omitempty"`
		Value       request.FieldString      `bson:"value,omitempty"`
		ZeroValue   request.FieldString      `bson:"zero_value"`
		Int64       request.FieldInt64       `bson:"int64"`
		Float64     request.FieldFloat64     `bson:"float64"`
		Bool        request.FieldBool        `bson:"bool"`
		Time        request.FieldTime        `bson:"time"`
		StringArray request.FieldStringArray `bson:"string_array"`
		Int64Array  request.FieldInt64Array  `bson:"int64_array"`
		JSON        request.FieldJSON        `bson:"json"`
		Duration    request.FieldDuration    `bson:"duration"`
	}

	doc := map[string]any{
		"null":         nil,
		"value":        "test",
		"zero_value":   "",
		"int64":        int64(1),
		"float64":      1.1,
		"bool":         true,
		"time":         int64(1),
		"duration":     "1s",
		"string_array": []any{"test", "test2"},
		"int64_array":  []any{int64(1), int64(2), int64(3)},
		"json":         map[string]any{"test": "test"},
	}

	b, err := bson.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}

	var v *tests

	if err := bson.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}

	if !v.Value.Set {
		t.Error("Expected value to be set")
	}

	if !v.Value.Valid {
		t.Error("Expected value to be valid")
	}

	exp := "test"

	if v.Value.Value != exp {
		t.Errorf("Expected value: %v, got: %v", exp, v.Value.Value)
	}

	if !v.ZeroValue.Set {
		t.Error("Expected zero value to be set")
	}

	if !v.ZeroValue.Valid {
		t.Error("Expected zero value to be valid")
	}

	exp = ""

	if v.ZeroValue.Value != exp {
		t.Errorf("Expected zero value: %v, got: %v", exp, v.Value.Value)
	}

	if !v.Int64.Set {
		t.Error("Expected int64 value to be set")
	}

	if !v.Int64.Valid {
		t.Error("Expected int64 value to be valid")
	}

	if v.Int64.Value != 1 {
		t.Errorf("Expected int64 value: 1, got: %v", v.Int64.Value)
	}

	if !v.Float64.Set {
		t.Error("Expected float64 value to be set")
	}

	if !v.Float64.Valid {
		t.Error("Expected float64 value to be valid")
	}

	if v.Float64.Value != 1.1 {
		t.Errorf("Expected float64 value: 1.1, got: %v", v.Float64.Value)
	}

	if !v.Bool.Set {
		t.Error("Expected bool value to be set")
	}

	if !v.Bool.Valid {
		t.Error("Expected bool value to be valid")
	}

	if v.Bool.Value != true {
		t.Errorf("Expected bool value: true, got: %v", v.Bool.Value)
	}

	if !v.Time.Set {
		t.Error("Expected time value to be set")
	}

	if !v.Time.Valid {
		t.Error("Expected time value to be valid")
	}

	if v.Time.Value != 1 {
		t.Errorf("Expected time value: 1, got: %v", v.Time.Value)
	}

	if !v.StringArray.Set {
		t.Error("Expected string slice value to be set")
	}

	if !v.StringArray.Valid {
		t.Error("Expected string slice value to be valid")
	}

	if len(v.StringArray.Value) != 2 {
		t.Errorf("Expected string slice length: 2, got: %v",
			len(v.StringArray.Value))
	}

	exp = "test"

	if v.StringArray.Value[0] != exp {
		t.Errorf("Expected string slice value: %v, got: %v",
			exp, v.StringArray.Value[0])
	}

	if !v.Int64Array.Set {
		t.Error("Expected int64 slice value to be set")
	}

	if !v.Int64Array.Valid {
		t.Error("Expected int64 slice value to be valid")
	}

	if len(v.Int64Array.Value) != 3 {
		t.Errorf("Expected int64 slice length: 3, got: %v",
			len(v.Int64Array.Value))
	}

	if v.Int64Array.Value[0] != 1 {
		t.Errorf("Expected int64 slice value: 1, got: %v",
			v.Int64Array.Value[0])
	}

	if !v.JSON.Set {
		t.Error("Expected JSON value to be set")
	}

	if !v.JSON.Valid {
		t.Error("Expected JSON value to be valid")
	}

	if v.JSON.Value == nil || len(v.JSON.Value) != 1 {
		t.Errorf("Expected JSON length: 1, got: %v",
			len(v.JSON.Value))
	} else if v.JSON.Value[exp] != exp {
		t.Errorf("Expected JSON value: %v, got: %v",
			exp, v.JSON.Value[exp])
	}

	if !v.Duration.Set {
		t.Error("Expected duration value to be set")
	}

	if !v.Duration.Valid {
		t.Error("Expected duration value to be valid")
	}

	if v.Duration.Value != time.Second {
		t.Errorf("Expected duration value: 1s, got: %v", v.Duration.Value)
	}
}

func TestFieldYAML(t *testing.T) {
	t.Parallel()

	type tests struct {
		Value       request.FieldString      `yaml:"value"`
		ZeroValue   request.FieldString      `yaml:"zero_value"`
		Null        request.FieldString      `yaml:"null_value"`
		NotSet      request.FieldString      `yaml:"not_set"`
		Int64       request.FieldInt64       `yaml:"int64"`
		Float64     request.FieldFloat64     `yaml:"float64"`
		Bool        request.FieldBool        `yaml:"bool"`
		Time        request.FieldTime        `yaml:"time"`
		StringArray request.FieldStringArray `yaml:"string_array"`
		Int64Array  request.FieldInt64Array  `yaml:"int64_array"`
		JSON        request.FieldJSON        `yaml:"json"`
		Duration    request.FieldDuration    `yaml:"duration"`
	}

	s := `value: test
zero_value: ""
null_value: null
int64: 1
int64_array: [1, 2, 3]
float64: 1.1
bool: true
time: 1
string_array: ["test", "test2"]
json: {"test": "test"}
duration: 1s
`

	var v *tests

	if err := yaml.Unmarshal([]byte(s), &v); err != nil {
		t.Fatal(err)
	}

	if !v.Value.Set {
		t.Error("Expected value to be set")
	}

	if !v.Value.Valid {
		t.Error("Expected value to be valid")
	}

	exp := "test"

	if v.Value.Value != exp {
		t.Errorf("Expected value: %v, got: %v", exp, v.Value.Value)
	}

	if !v.ZeroValue.Set {
		t.Error("Expected zero value to be set")
	}

	if !v.ZeroValue.Valid {
		t.Error("Expected zero value to be valid")
	}

	exp = ""

	if v.ZeroValue.Value != exp {
		t.Errorf("Expected zero value: %v, got: %v", exp, v.Value.Value)
	}

	if v.Null.Set {
		t.Error("Expected null value not to be set")
	}

	if v.Null.Valid {
		t.Error("Expected null value not to be valid")
	}

	if v.NotSet.Set {
		t.Error("Expected not set value not to be set")
	}

	if v.NotSet.Valid {
		t.Error("Expected not set value not to be valid")
	}

	if !v.Int64.Set {
		t.Error("Expected int64 value to be set")
	}

	if !v.Int64.Valid {
		t.Error("Expected int64 value to be valid")
	}

	if v.Int64.Value != 1 {
		t.Errorf("Expected int64 value: 1, got: %v", v.Int64.Value)
	}

	if !v.Float64.Set {
		t.Error("Expected float64 value to be set")
	}

	if !v.Float64.Valid {
		t.Error("Expected float64 value to be valid")
	}

	if v.Float64.Value != 1.1 {
		t.Errorf("Expected float64 value: 1.1, got: %v", v.Float64.Value)
	}

	if !v.Bool.Set {
		t.Error("Expected bool value to be set")
	}

	if !v.Bool.Valid {
		t.Error("Expected bool value to be valid")
	}

	if v.Bool.Value != true {
		t.Errorf("Expected bool value: 1, got: %v", v.Bool.Value)
	}

	if !v.Time.Set {
		t.Error("Expected bool value to be set")
	}

	if !v.Time.Valid {
		t.Error("Expected bool value to be valid")
	}

	if v.Time.Value != 1 {
		t.Errorf("Expected time value: 1, got: %v", v.Time.Value)
	}

	if !v.StringArray.Set {
		t.Error("Expected string slice value to be set")
	}

	if !v.StringArray.Valid {
		t.Error("Expected string slice value to be valid")
	}

	if len(v.StringArray.Value) != 2 {
		t.Errorf("Expected string slice length: 2, got: %v",
			len(v.StringArray.Value))
	}

	exp = "test"

	if v.StringArray.Value[0] != exp {
		t.Errorf("Expected string slice value: %v, got: %v",
			exp, v.StringArray.Value[0])
	}

	if !v.Int64Array.Set {
		t.Error("Expected int64 slice value to be set")
	}

	if !v.Int64Array.Valid {
		t.Error("Expected int64 slice value to be valid")
	}

	if len(v.Int64Array.Value) != 3 {
		t.Errorf("Expected int64 slice length: 3, got: %v",
			len(v.Int64Array.Value))
	}

	if v.Int64Array.Value[0] != 1 {
		t.Errorf("Expected int64 slice value: 1, got: %v",
			v.Int64Array.Value[0])
	}

	if !v.JSON.Set {
		t.Error("Expected JSON value to be set")
	}

	if !v.JSON.Valid {
		t.Error("Expected JSON value to be valid")
	}

	if len(v.JSON.Value) != 1 {
		t.Errorf("Expected JSON length: 1, got: %v",
			len(v.JSON.Value))
	}

	if v.JSON.Value[exp] != exp {
		t.Errorf("Expected JSON value: %v, got: %v",
			exp, v.JSON.Value[exp])
	}

	if !v.Duration.Set {
		t.Error("Expected duration value to be set")
	}

	if !v.Duration.Valid {
		t.Error("Expected duration value to be valid")
	}

	if v.Duration.Value != time.Second {
		t.Errorf("Expected duration value: 1, got: %v", v.Duration.Value)
	}

	b, err := yaml.Marshal(&v)
	if err != nil {
		t.Fatal(err)
	}

	exp = `value: test
zero_value: ""
null_value: null
not_set: null
int64: 1
float64: 1.1
bool: true
time: 1
string_array:
    - test
    - test2
int64_array:
    - 1
    - 2
    - 3
json:
    test: test
duration: 1s
`

	if string(b) != exp {
		t.Errorf("Expected YAML: %v, got: %v", exp, string(b))
	}
}

func TestSetField(t *testing.T) {
	t.Parallel()

	doc := &bson.D{}

	request.SetField(doc, "string",
		request.FieldString{Set: true, Valid: true})
	request.SetField(doc, "string_not",
		request.FieldString{Set: false, Valid: true})
	request.SetField(doc, "string_null",
		request.FieldString{Set: true, Valid: false})
	request.SetField(doc, "int64",
		request.FieldInt64{Set: true, Valid: true})
	request.SetField(doc, "int64_not",
		request.FieldInt64{Set: false, Valid: true})
	request.SetField(doc, "int64_null",
		request.FieldInt64{Set: true, Valid: false})
	request.SetField(doc, "float64",
		request.FieldFloat64{Set: true, Valid: true})
	request.SetField(doc, "float64_not",
		request.FieldFloat64{Set: false, Valid: true})
	request.SetField(doc, "float64_null",
		request.FieldFloat64{Set: true, Valid: false})
	request.SetField(doc, "bool",
		request.FieldBool{Set: true, Valid: true})
	request.SetField(doc, "bool_not",
		request.FieldBool{Set: false, Valid: true})
	request.SetField(doc, "bool_null",
		request.FieldBool{Set: true, Valid: false})
	request.SetField(doc, "time",
		request.FieldTime{Set: true, Valid: true})
	request.SetField(doc, "time_not",
		request.FieldTime{Set: false, Valid: true})
	request.SetField(doc, "time_null",
		request.FieldTime{Set: true, Valid: false})
	request.SetField(doc, "string_array",
		request.FieldStringArray{Set: true, Valid: true})
	request.SetField(doc, "string_array_not",
		request.FieldStringArray{Set: false, Valid: true})
	request.SetField(doc, "string_array_null",
		request.FieldStringArray{Set: true, Valid: false})
	request.SetField(doc, "int64_array",
		request.FieldInt64Array{Set: true, Valid: true})
	request.SetField(doc, "int64_array_not",
		request.FieldInt64Array{Set: false, Valid: true})
	request.SetField(doc, "int64_array_null",
		request.FieldInt64Array{Set: true, Valid: false})
	request.SetField(doc, "json",
		request.FieldJSON{Set: true, Valid: true})
	request.SetField(doc, "json_not",
		request.FieldJSON{Set: false, Valid: true})
	request.SetField(doc, "json_null",
		request.FieldJSON{Set: true, Valid: false})
	request.SetField(doc, "duration",
		request.FieldDuration{Set: true, Valid: true})
	request.SetField(doc, "duration_not",
		request.FieldDuration{Set: false, Valid: true})
	request.SetField(doc, "duration_null",
		request.FieldDuration{Set: true, Valid: false})

	exp := 18

	if len(*doc) != exp {
		t.Errorf("Expected sets length: %v, got: %v", exp, len(*doc))
	}
}
