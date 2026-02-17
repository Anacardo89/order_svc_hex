package validator

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAndValidate_FlatStruct(t *testing.T) {
	type Payload struct {
		Name string `json:"name" validate:"required"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name        string
		json        string
		wantMissing []string
	}{
		{
			name:        "valid payload",
			json:        `{"name":"john","age":30}`,
			wantMissing: nil,
		},
		{
			name:        "missing required field",
			json:        `{"age":30}`,
			wantMissing: []string{"name"},
		},
		{
			name:        "zero value allowed",
			json:        `{"name":""}`,
			wantMissing: nil,
		},
		{
			name:        "optional field missing",
			json:        `{"name":"john"}`,
			wantMissing: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Payload
			err := ParseAndValidate(json.RawMessage(tt.json), &p)
			if len(tt.wantMissing) > 0 {
				require.Error(t, err)
				var ve ValidationError
				require.ErrorAs(t, err, &ve)
				assert.ElementsMatch(t, tt.wantMissing, ve.MissingFields)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseAndValidate_FlatStruct_NullAndPointers(t *testing.T) {
	type Payload struct {
		Name *string `json:"name" validate:"required"`
		Age  *int    `json:"age"`
	}

	tests := []struct {
		name        string
		json        string
		wantMissing []string
	}{
		{
			name:        "required pointer present",
			json:        `{"name":"john"}`,
			wantMissing: nil,
		},
		{
			name:        "required pointer null",
			json:        `{"name":null}`,
			wantMissing: []string{"name"},
		},
		{
			name:        "required pointer missing",
			json:        `{}`,
			wantMissing: []string{"name"},
		},
		{
			name:        "optional pointer missing",
			json:        `{"name":"john"}`,
			wantMissing: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Payload
			err := ParseAndValidate(json.RawMessage(tt.json), &p)
			if len(tt.wantMissing) > 0 {
				require.Error(t, err)
				var ve ValidationError
				require.ErrorAs(t, err, &ve)
				assert.ElementsMatch(t, tt.wantMissing, ve.MissingFields)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseAndValidate_NestedStruct(t *testing.T) {
	type User struct {
		Email string `json:"email" validate:"required"`
	}
	type Payload struct {
		User User `json:"user" validate:"required"`
	}

	tests := []struct {
		name        string
		json        string
		wantMissing []string
	}{
		{
			name:        "valid nested struct",
			json:        `{"user":{"email":"a@b.com"}}`,
			wantMissing: nil,
		},
		{
			name:        "missing nested required field",
			json:        `{"user":{}}`,
			wantMissing: []string{"user.email"},
		},
		{
			name:        "missing required parent",
			json:        `{}`,
			wantMissing: []string{"user"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Payload
			err := ParseAndValidate(json.RawMessage(tt.json), &p)
			if len(tt.wantMissing) > 0 {
				require.Error(t, err)
				var ve ValidationError
				require.ErrorAs(t, err, &ve)
				assert.ElementsMatch(t, tt.wantMissing, ve.MissingFields)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseAndValidate_NestedStruct_NullAndPointers(t *testing.T) {
	type Inner struct {
		ID *string `json:"id" validate:"required"`
	}
	type Payload struct {
		Inner *Inner `json:"inner" validate:"required"`
	}

	tests := []struct {
		name        string
		json        string
		wantMissing []string
	}{
		{
			name:        "nested required pointer present",
			json:        `{"inner":{"id":"123"}}`,
			wantMissing: nil,
		},
		{
			name:        "nested required pointer null",
			json:        `{"inner":{"id":null}}`,
			wantMissing: []string{"inner.id"},
		},
		{
			name:        "nested required container null",
			json:        `{"inner":null}`,
			wantMissing: []string{"inner"},
		},
		{
			name:        "nested required container missing",
			json:        `{}`,
			wantMissing: []string{"inner"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Payload
			err := ParseAndValidate(json.RawMessage(tt.json), &p)
			if len(tt.wantMissing) > 0 {
				require.Error(t, err)
				var ve ValidationError
				require.ErrorAs(t, err, &ve)
				assert.ElementsMatch(t, tt.wantMissing, ve.MissingFields)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseAndValidate_OptionalContainerWithRequiredNestedField(t *testing.T) {
	type Inner struct {
		ID string `json:"id" validate:"required"`
	}
	type Payload struct {
		Inner *Inner `json:"inner"`
	}

	tests := []struct {
		name        string
		json        string
		wantMissing []string
	}{
		{
			name:        "optional container missing",
			json:        `{}`,
			wantMissing: nil,
		},
		{
			name:        "container present but missing nested required",
			json:        `{"inner":{}}`,
			wantMissing: []string{"inner.id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Payload
			err := ParseAndValidate(json.RawMessage(tt.json), &p)
			if len(tt.wantMissing) > 0 {
				require.Error(t, err)
				var ve ValidationError
				require.ErrorAs(t, err, &ve)
				assert.ElementsMatch(t, tt.wantMissing, ve.MissingFields)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseAndValidate_SliceOfStructs(t *testing.T) {
	type Item struct {
		ID string `json:"id" validate:"required"`
	}
	type Payload struct {
		Items []Item `json:"items" validate:"required"`
	}

	tests := []struct {
		name        string
		json        string
		wantMissing []string
	}{
		{
			name:        "valid slice",
			json:        `{"items":[{"id":"1"}]}`,
			wantMissing: nil,
		},
		{
			name:        "empty slice allowed",
			json:        `{"items":[]}`,
			wantMissing: nil,
		},
		{
			name:        "missing required slice",
			json:        `{}`,
			wantMissing: []string{"items"},
		},
		{
			name:        "missing nested required in slice",
			json:        `{"items":[{}]}`,
			wantMissing: []string{"items[0].id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Payload
			err := ParseAndValidate(json.RawMessage(tt.json), &p)
			if len(tt.wantMissing) > 0 {
				require.Error(t, err)
				var ve ValidationError
				require.ErrorAs(t, err, &ve)
				assert.ElementsMatch(t, tt.wantMissing, ve.MissingFields)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseAndValidate_SliceOfStructs_NullEntries(t *testing.T) {
	type Item struct {
		ID string `json:"id" validate:"required"`
	}
	type Payload struct {
		Items []Item `json:"items" validate:"required"`
	}

	tests := []struct {
		name        string
		json        string
		wantMissing []string
	}{
		{
			name:        "slice with valid item",
			json:        `{"items":[{"id":"1"}]}`,
			wantMissing: nil,
		},
		{
			name:        "slice with null item",
			json:        `{"items":[null]}`,
			wantMissing: []string{"items[0]"},
		},
		{
			name:        "slice is null",
			json:        `{"items":null}`,
			wantMissing: []string{"items"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Payload
			err := ParseAndValidate(json.RawMessage(tt.json), &p)
			if len(tt.wantMissing) > 0 {
				require.Error(t, err)
				var ve ValidationError
				require.ErrorAs(t, err, &ve)
				assert.ElementsMatch(t, tt.wantMissing, ve.MissingFields)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseAndValidate_SliceOfPointers(t *testing.T) {
	type Item struct {
		ID string `json:"id" validate:"required"`
	}
	type Payload struct {
		Items []*Item `json:"items" validate:"required"`
	}

	tests := []struct {
		name        string
		json        string
		wantMissing []string
	}{
		{
			name:        "slice with valid item",
			json:        `{"items":[{"id":"1"}]}`,
			wantMissing: nil,
		},
		{
			name:        "slice with null item",
			json:        `{"items":[null]}`,
			wantMissing: []string{"items[0]"},
		},
		{
			name:        "slice with missing required field in pointer",
			json:        `{"items":[{}]}`,
			wantMissing: []string{"items[0].id"},
		},
		{
			name:        "slice is null",
			json:        `{"items":null}`,
			wantMissing: []string{"items"},
		},
		{
			name:        "slice is empty",
			json:        `{"items":[]}`,
			wantMissing: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Payload
			err := ParseAndValidate(json.RawMessage(tt.json), &p)
			if len(tt.wantMissing) > 0 {
				require.Error(t, err)
				var ve ValidationError
				require.ErrorAs(t, err, &ve)
				assert.ElementsMatch(t, tt.wantMissing, ve.MissingFields)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseAndValidate_MapOfStructs(t *testing.T) {
	type Item struct {
		Name string `json:"name" validate:"required"`
	}
	type Payload struct {
		Items map[string]Item `json:"items" validate:"required"`
	}

	tests := []struct {
		name        string
		json        string
		wantMissing []string
	}{
		{
			name:        "valid map",
			json:        `{"items":{"a":{"name":"x"}}}`,
			wantMissing: nil,
		},
		{
			name:        "empty map allowed",
			json:        `{"items":{}}`,
			wantMissing: nil,
		},
		{
			name:        "missing nested required in map",
			json:        `{"items":{"a":{}}}`,
			wantMissing: []string{`items["a"].name`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Payload
			err := ParseAndValidate(json.RawMessage(tt.json), &p)
			if len(tt.wantMissing) > 0 {
				require.Error(t, err)
				var ve ValidationError
				require.ErrorAs(t, err, &ve)
				assert.ElementsMatch(t, tt.wantMissing, ve.MissingFields)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseAndValidate_MapOfNullEntries(t *testing.T) {
	type Item struct {
		ID string `json:"id" validate:"required"`
	}
	type Payload struct {
		Items map[string]Item `json:"items" validate:"required"`
	}

	tests := []struct {
		name        string
		json        string
		wantMissing []string
	}{
		{
			name:        "map with valid entry",
			json:        `{"items":{"a":{"id":"1"}}}`,
			wantMissing: nil,
		},
		{
			name:        "map with null entry",
			json:        `{"items":{"a":null}}`,
			wantMissing: []string{`items["a"]`},
		},
		{
			name:        "map missing required field inside entry",
			json:        `{"items":{"a":{}}}`,
			wantMissing: []string{`items["a"].id`},
		},
		{
			name:        "map is null",
			json:        `{"items":null}`,
			wantMissing: []string{"items"},
		},
		{
			name:        "empty map",
			json:        `{"items":{}}`,
			wantMissing: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Payload
			err := ParseAndValidate(json.RawMessage(tt.json), &p)
			if len(tt.wantMissing) > 0 {
				require.Error(t, err)
				var ve ValidationError
				require.ErrorAs(t, err, &ve)
				assert.ElementsMatch(t, tt.wantMissing, ve.MissingFields)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseAndValidate_MapToPointerOfStructs(t *testing.T) {
	type Item struct {
		ID string `json:"id" validate:"required"`
	}
	type Payload struct {
		Items map[string]*Item `json:"items" validate:"required"`
	}

	tests := []struct {
		name        string
		json        string
		wantMissing []string
	}{
		{
			name:        "map with valid pointer entry",
			json:        `{"items":{"a":{"id":"1"}}}`,
			wantMissing: nil,
		},
		{
			name:        "map with null pointer entry",
			json:        `{"items":{"a":null}}`,
			wantMissing: []string{`items["a"]`},
		},
		{
			name:        "map with missing required field in pointer",
			json:        `{"items":{"a":{}}}`,
			wantMissing: []string{`items["a"].id`},
		},
		{
			name:        "map is null",
			json:        `{"items":null}`,
			wantMissing: []string{"items"},
		},
		{
			name:        "empty map",
			json:        `{"items":{}}`,
			wantMissing: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Payload
			err := ParseAndValidate(json.RawMessage(tt.json), &p)
			if len(tt.wantMissing) > 0 {
				require.Error(t, err)
				var ve ValidationError
				require.ErrorAs(t, err, &ve)
				assert.ElementsMatch(t, tt.wantMissing, ve.MissingFields)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
