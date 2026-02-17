package validator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func derefTypeValue(t reflect.Type, v reflect.Value) (reflect.Type, reflect.Value) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return t, reflect.Value{}
			}
			v = v.Elem()
		}
	}
	return t, v
}

func jsonFieldPath(sf reflect.StructField, parent string) string {
	jsonTag := sf.Tag.Get("json")
	if idx := strings.Index(jsonTag, ","); idx != -1 {
		jsonTag = jsonTag[:idx]
	}
	if jsonTag == "-" {
		return ""
	}
	if jsonTag == "" {
		jsonTag = sf.Name
	}
	if parent != "" {
		return parent + "." + jsonTag
	}
	return jsonTag
}

func checkMissingSlice(raw []byte, fieldType reflect.Type, fieldValue reflect.Value, path string, ve *ValidationError) {
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil {
		return
	}
	for i, elem := range arr {
		elemType, elemValue := derefTypeValue(fieldType.Elem(), reflect.Value{})
		if fieldValue.IsValid() && i < fieldValue.Len() {
			elemValue = fieldValue.Index(i)
			elemType, elemValue = derefTypeValue(fieldType.Elem(), elemValue)
		}
		isNull := bytes.Equal(elem, []byte("null"))
		if elemType.Kind() == reflect.Struct && !isNull {
			var nested map[string]json.RawMessage
			if err := json.Unmarshal(elem, &nested); err == nil {
				checkMissingFields(nested, elemType, elemValue, fmt.Sprintf("%s[%d]", path, i), ve)
			}
		} else if elemType.Kind() == reflect.Struct && isNull {
			ve.MissingFields = append(ve.MissingFields, fmt.Sprintf("%s[%d]", path, i))
		}
	}
}

func checkMissingMap(raw []byte, fieldType reflect.Type, fieldValue reflect.Value, path string, ve *ValidationError) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return
	}
	for key, elem := range m {
		elemType, mv := derefTypeValue(fieldType.Elem(), reflect.Value{})
		if fieldValue.IsValid() {
			mv = fieldValue.MapIndex(reflect.ValueOf(key))
			elemType, mv = derefTypeValue(fieldType.Elem(), mv)
		}
		isNull := bytes.Equal(elem, []byte("null"))
		if elemType.Kind() == reflect.Struct && !isNull {
			var nested map[string]json.RawMessage
			if err := json.Unmarshal(elem, &nested); err == nil {
				checkMissingFields(nested, elemType, mv, fmt.Sprintf("%s[%q]", path, key), ve)
			}
		} else if elemType.Kind() == reflect.Struct && isNull {
			ve.MissingFields = append(ve.MissingFields, fmt.Sprintf("%s[%q]", path, key))
		}
	}
}
