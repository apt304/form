package form

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
)

func toPtr[T any](val T) *T {
	return &val
}

type ErrorEncode struct {
	fieldName string
	err       error
}

func (e ErrorEncode) Error() string {
	return fmt.Sprintf("unable to encode tag '%s': %s", e.fieldName, e.err)
}

func Marshal(src any) (map[string][]string, error) {
	dest := map[string][]string{}
	err := NewEncoder(dest).Encode(src)
	if err != nil {
		return nil, err
	}

	return dest, nil
}

type Encoder struct {
	dest map[string][]string
}

func NewEncoder(dest map[string][]string) *Encoder {
	return &Encoder{dest: dest}
}

func (e *Encoder) Encode(src any) error {
	// Ensure that src is a struct value
	val := reflect.ValueOf(src)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("source (%v) must be a struct", src)
	}

	return e.encodeStruct(val)
}

func (e *Encoder) encodeStruct(src reflect.Value) error {
	// Iterate over the fields in src
	srcType := src.Type()
	for i := 0; i < src.NumField(); i++ {
		fieldType := srcType.Field(i)
		if !fieldType.IsExported() {
			continue
		}

		formTag, shouldOmitEmpty := parseFieldTag(fieldType)
		if formTag != "" && formTag != "-" {
			// Parse based on field type. All field types but map look up their values from src. Map must iterate over
			// src keys to find all relevant key/value pairs. Map key/value parsing is done once and cached for
			// additional map fields.
			fieldVal := src.Field(i)

			err := e.encodeFormField(fieldVal, formTag, shouldOmitEmpty)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *Encoder) encodeFormField(src reflect.Value, formTag string, shouldOmitEmpty bool) error {
	if src.Type().Implements(textMarshalerType) ||
		(src.CanAddr() && src.Addr().Type().Implements(textMarshalerType)) {
		// If the destination itself doesn't implement TextMarshaler, take the pointer and recursively call
		// encodeFormField.
		if !src.Type().Implements(textMarshalerType) {
			return e.encodeFormField(src.Addr(), formTag, shouldOmitEmpty)
		}

		// Ignore nil pointers
		if src.Kind() == reflect.Pointer && src.IsNil() {
			return nil
		}

		// Don't include zero values with omitempty flags
		if shouldOmitEmpty && isZeroValue(src) {
			return nil
		}

		f := src.MethodByName("MarshalText")
		ret := f.Call(nil)
		if !ret[1].IsNil() {
			return ErrorEncode{fieldName: formTag, err: ret[0].Interface().(error)}
		}

		// Convert returned bytes to string
		retVal := ret[0]
		retStr := string(retVal.Interface().([]byte))

		e.dest[formTag] = append(e.dest[formTag], retStr)

		return nil
	}

	// Check for structured types
	switch src.Kind() {
	case reflect.Slice:
		return e.encodeSliceField(src, formTag, shouldOmitEmpty)

	case reflect.Map:
		return e.encodeMap(src, formTag, shouldOmitEmpty)

	case reflect.Struct:
		return e.encodeStruct(src)

	default:
		break
	}

	encodedVal, err := e.encodeValue(src, formTag, shouldOmitEmpty)
	if err != nil {
		return err
	}
	if encodedVal == nil {
		return nil
	}

	// Don't include zero values with omitempty flags
	if shouldOmitEmpty && isZeroValue(src) {
		return nil
	}

	e.dest[formTag] = append(e.dest[formTag], *encodedVal)

	return nil
}

func (e *Encoder) encodeValue(src reflect.Value, formTag string, shouldOmitEmpty bool) (*string, error) {
	switch src.Type() {
	case durationType:
		return toPtr(fmt.Sprintf("%s", src)), nil
	}

	switch src.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return toPtr(strconv.FormatInt(src.Int(), 10)), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return toPtr(strconv.FormatUint(src.Uint(), 10)), nil

	case reflect.Float32, reflect.Float64:
		return toPtr(strconv.FormatFloat(src.Float(), 'f', -1, 64)), nil

	case reflect.Bool:
		return toPtr(strconv.FormatBool(src.Bool())), nil

	case reflect.String:
		return toPtr(src.String()), nil

	case reflect.Pointer:
		if src.IsNil() {
			return nil, nil
		}

		// Recursively call decode() with the pointer's element
		// This additional pointer dereferencing is needed to handle slice pointer values. Top-level pointers are
		// handled in decodeFormField.
		return e.encodeValue(src.Elem(), formTag, shouldOmitEmpty)

	default:
		return nil, ErrorEncode{fieldName: formTag, err: fmt.Errorf("unsupported kind %v", src.Kind())}
	}
}

func (e *Encoder) encodeSliceField(src reflect.Value, formTag string, shouldOmitEmpty bool) error {
	if src.Len() == 0 && shouldOmitEmpty {
		return nil
	}

	values, err := e.encodeSliceValue(src, formTag, shouldOmitEmpty)
	if err != nil {
		return err
	}

	e.dest[formTag] = values

	return nil
}

func (e *Encoder) encodeSliceValue(src reflect.Value, formTag string, shouldOmitEmpty bool) ([]string, error) {
	if src.Len() == 0 && shouldOmitEmpty {
		return nil, nil
	}

	var values []string
	for i := 0; i < src.Len(); i++ {
		encodedVal, err := e.encodeValue(src.Index(i), formTag, shouldOmitEmpty)
		if err != nil {
			return nil, ErrorEncode{fieldName: formTag, err: fmt.Errorf("unable to encode slice %s: %w", formTag, err)}
		}

		values = append(values, *encodedVal)
	}

	return values, nil
}

func (e *Encoder) encodeMap(src reflect.Value, formTag string, shouldOmitEmpty bool) error {
	if src.Len() == 0 && shouldOmitEmpty {
		return nil
	}

	for _, key := range src.MapKeys() {
		mapKey := fmt.Sprintf("%s[%s]", formTag, key)
		val := src.MapIndex(key)

		// Handle single values or slices
		if val.Kind() == reflect.Slice {
			encodedVal, err := e.encodeSliceValue(val, formTag, shouldOmitEmpty)
			if err != nil {
				return ErrorEncode{fieldName: formTag, err: fmt.Errorf("unable to encode map key %s: %w", mapKey, err)}
			}

			e.dest[mapKey] = encodedVal
		} else {
			encodedVal, err := e.encodeValue(val, formTag, shouldOmitEmpty)
			if err != nil {
				return ErrorEncode{fieldName: formTag, err: fmt.Errorf("unable encode map key %s: %w", mapKey, err)}
			}

			e.dest[mapKey] = append(e.dest[mapKey], *encodedVal)
		}
	}

	return nil
}

// parseFieldTag parses the field's "form" tag.
// Returns the provided tag value and an omitempty flag, if omitempty is present
func parseFieldTag(fieldType reflect.StructField) (string, bool) {
	formTag := fieldType.Tag.Get("form")
	if formTag == "" {
		return "", false
	}

	tagParts := strings.Split(formTag, ",")
	if len(tagParts) == 1 {
		return tagParts[0], false
	}

	tag := tagParts[0]
	for _, part := range tagParts {
		if part == "omitempty" {
			return tag, true
		}
	}

	return tag, false
}

func isZeroValue(val reflect.Value) bool {
	// Maps, slices, structs, and funcs must be checked directly
	switch val.Kind() {
	case reflect.Map, reflect.Slice:
		return val.IsNil()
	case reflect.Struct:
		return reflect.DeepEqual(val.Interface(), reflect.Zero(val.Type()).Interface())
	default:
		break
	}

	// Other types are compared directly to their type's zero value
	return val.Interface() == reflect.Zero(val.Type()).Interface()
}