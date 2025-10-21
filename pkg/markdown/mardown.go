package markdown

import (
	"fmt"
	"reflect"
	"strings"
)

// Stringify returns a string representation of the given struct value. Each
// field is represented on a new line in the format "- FieldName: FieldValue".
// Fields with empty string values are omitted from the output.
func Stringify[T any](prefix string, v T) string {
	if reflect.TypeOf(v).Kind() != reflect.Struct {
		panic("T must be struct")
	}

	val := reflect.ValueOf(v)
	typ := val.Type()

	var b strings.Builder
	for i := 0; i < val.NumField(); i++ {
		if i > 0 {
			b.WriteByte('\n')
		}
		fieldName := typ.Field(i).Name
		fieldValue := val.Field(i).Interface()
		fmt.Fprintf(&b, "%s%s: %v", prefix, fieldName, fieldValue)
	}
	return b.String()
}

// Ulist returns a markdown-formatted unordered list of the provided items.
// Each item is prefixed with an asterisk (*) and a space, and items are
// separated by newlines.
func Ulist(items ...any) string {
	var b strings.Builder
	for i, v := range items {
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, " * %v", v)
	}
	return b.String()
}
