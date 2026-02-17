package validator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type ValidationError struct {
	MissingFields []string
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("missing required fields: %s", strings.Join(v.MissingFields, ", "))
}

func ParseAndValidate(raw json.RawMessage, payload any) error {
	var (
		ve     ValidationError
		rawMap map[string]json.RawMessage
	)
	rv := reflect.ValueOf(payload)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("payload must be a non-nil pointer")
	}
	rt := rv.Type().Elem()
	if rt.Kind() != reflect.Struct {
		return errors.New("payload must point to a struct")
	}
	if err := json.Unmarshal(raw, &rawMap); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if err := json.Unmarshal(raw, payload); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	checkMissingFields(rawMap, rt, rv.Elem(), "", &ve)
	if len(ve.MissingFields) > 0 {
		return ve
	}
	return nil
}

func checkMissingFields(rawMap map[string]json.RawMessage, t reflect.Type, v reflect.Value, path string, ve *ValidationError) {
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		fv := v.Field(i)
		// dereference top-level pointer
		fieldType, fieldValue := derefTypeValue(sf.Type, fv)
		// set field path
		jsonTag := sf.Tag.Get("json")
		fieldPath := jsonFieldPath(sf, path)
		if fieldPath == "" {
			continue
		}
		// validate required
		raw, exists := rawMap[jsonTag]
		isNull := exists && bytes.Equal(raw, []byte("null"))
		isRequired := strings.Contains(sf.Tag.Get("validate"), "required")
		if isRequired && (!exists || isNull) {
			ve.MissingFields = append(ve.MissingFields, fieldPath)
			continue
		}
		if !exists || isNull {
			continue
		}
		// recurse
		switch fieldType.Kind() {
		case reflect.Struct:
			var nested map[string]json.RawMessage
			if err := json.Unmarshal(raw, &nested); err == nil {
				checkMissingFields(nested, fieldType, fieldValue, fieldPath, ve)
			}

		case reflect.Slice:
			checkMissingSlice(raw, fieldType, fieldValue, fieldPath, ve)

		case reflect.Map:
			checkMissingMap(raw, fieldType, fieldValue, fieldPath, ve)
		}
	}
}
