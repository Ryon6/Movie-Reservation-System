package testutils

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// StructToQuery converts a struct to a URL query string.
// It uses reflection to iterate over struct fields and their `json` or `form` tags.
func StructToQuery(s interface{}) (string, error) {
	if s == nil {
		return "", nil
	}

	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return "", fmt.Errorf("StructToQuery only accepts structs; got %T", s)
	}

	params := url.Values{}
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Handle embedded structs
		if fieldType.Anonymous && field.Kind() == reflect.Struct {
			embeddedQuery, err := StructToQuery(field.Interface())
			if err != nil {
				return "", err
			}
			embeddedParams, err := url.ParseQuery(embeddedQuery)
			if err != nil {
				return "", err
			}
			for key, values := range embeddedParams {
				for _, value := range values {
					params.Add(key, value)
				}
			}
			continue
		}

		// Get tag, preferring `form` over `json`
		tag := fieldType.Tag.Get("form")
		if tag == "" {
			tag = fieldType.Tag.Get("json")
		}

		// Skip fields without a tag
		if tag == "" {
			continue
		}

		// Skip `omitempty` fields that are zero-valued
		parts := strings.Split(tag, ",")
		tagName := parts[0]
		isOmitempty := false
		if len(parts) > 1 && parts[1] == "omitempty" {
			isOmitempty = true
		}

		if isOmitempty && field.IsZero() {
			continue
		}

		// Convert field value to string and add to params
		var value string
		switch field.Kind() {
		case reflect.String:
			value = field.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			value = strconv.FormatInt(field.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			value = strconv.FormatUint(field.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			value = strconv.FormatFloat(field.Float(), 'f', -1, 64)
		case reflect.Bool:
			value = strconv.FormatBool(field.Bool())
		default:
			if field.Type() == reflect.TypeOf(time.Time{}) {
				t := field.Interface().(time.Time)
				if !t.IsZero() {
					value = t.Format(time.RFC3339)
				}
			}
		}

		if value != "" {
			params.Add(tagName, value)
		}
	}

	return params.Encode(), nil
}
