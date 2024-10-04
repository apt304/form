package form

import (
	"encoding"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

var (
	durationType        = reflect.TypeOf(time.Duration(0))
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// ErrorDecode represents an error that occurs during the decoding process.
type ErrorDecode struct {
	fieldName string
	err       error
}

// Error returns the error message for ErrorDecode.
func (e ErrorDecode) Error() string {
	return fmt.Sprintf("Unable to decode tag '%s': %s", e.fieldName, e.err)
}

// Unmarshal iterates over the fields in `dest`, populating them with the appropriate fields from the provided source
// map. `src` is a map containing form values, and `dest` is a pointer to the struct that will be populated.
//
// Example:
//
//	var r *http.Request
//	err := r.ParseForm()
//	if err != nil { ... }
//
//	var submission SampleForm
//	err := form.Unmarshal(r.Form, &submission)
//	if err != nil { ... }
//
// Form data is flat, with key/value pairings: `field = val`, where `field` is matched to a struct tag. Forms allow the
// same key to be reused: `field = val1, field = val2`. Multiple values can be handled by using a slice in the struct.
// Dynamic values are encoded in the keys with a `field[key] = val` syntax. Use `map[string]<type>` as the struct type
// to unmarshal these dynamic pairs.
//
// If multiple form values are provided for a field, parse all values. If the value is not a slice, the first form value
// is set to the struct's field.
func Unmarshal(src map[string][]string, dest any) error {
	return NewDecoder(src).Decode(dest)
}

// Decoder is responsible for decoding form data from the source map to the provided destination struct.
type Decoder struct {
	src map[string][]string
}

// NewDecoder creates a new Decoder instance with the given source form data.
func NewDecoder(src map[string][]string) *Decoder {
	return &Decoder{src: src}
}

// Decode decodes the form data into the provided destination struct by iterating over the fields in `dest`.
// The `dest` must be a pointer to a struct.
func (d *Decoder) Decode(dest any) error {
	// Ensure dest has a value that is a non-nil pointer to a struct
	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Pointer || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("destination (%v) must be a pointer to a struct", reflect.TypeOf(dest))
	}

	// Get value of dest pointer
	val = val.Elem()

	err := d.decodeStruct(val)
	if err != nil {
		return err
	}

	return nil
}

// decodeStruct iterates over the fields of the provided struct and decodes them from form values.
func (d *Decoder) decodeStruct(dest reflect.Value) error {
	// Iterate over the fields in dest
	destType := dest.Type()
	for i := 0; i < dest.NumField(); i++ {
		fieldType := destType.Field(i)
		if !fieldType.IsExported() {
			continue
		}

		// Ignore omitempty flag when decoding
		formTag, _ := parseFieldTag(fieldType)

		if formTag != "" && formTag != "-" {
			// Parse based on field type. All field types but map look up their values from src. Map must iterate over
			// src keys to find all relevant key/value pairs. Map key/value parsing is done once and cached for
			// additional map fields.
			fieldVal := dest.Field(i)

			err := d.decodeFormField(fieldVal, formTag)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// decodeFormField decodes the form value into the provided struct field based on the form tag.
func (d *Decoder) decodeFormField(dest reflect.Value, formTag string) error {
	if dest.Kind() != reflect.Map && len(d.src[formTag]) == 0 {
		return nil
	}

	// Check overridden TextUnmarshaler types first.
	if dest.Type().Implements(textUnmarshalerType) ||
		(dest.CanAddr() && dest.Addr().Type().Implements(textUnmarshalerType)) {
		if !dest.Type().Implements(textUnmarshalerType) {
			return d.decodeFormField(dest.Addr(), formTag)
		}

		ensurePointerIsSet(dest)
		f := dest.MethodByName("UnmarshalText")
		rawValue := []byte(d.src[formTag][0])
		unmarshalArg := []reflect.Value{reflect.ValueOf(rawValue)}
		ret := f.Call(unmarshalArg)
		if !ret[0].IsNil() {
			return ErrorDecode{fieldName: formTag, err: ret[0].Interface().(error)}
		}

		return nil
	}

	if dest.Kind() == reflect.Pointer {
		// Decode the element the pointer references.
		ensurePointerIsSet(dest)
		return d.decodeFormField(dest.Elem(), formTag)
	}

	// Check for structured types
	switch dest.Kind() {
	case reflect.Slice:
		return d.decodeSliceField(dest, formTag)

	case reflect.Map:
		return d.decodeMap(dest, formTag)

	case reflect.Struct:
		return d.decodeStruct(dest)

	default:
		break
	}

	// Decode value. Take the first value from the source slice.
	var strVal string
	if len(d.src[formTag]) > 0 {
		strVal = d.src[formTag][0]
	}

	return d.decodeValue(dest, strVal, formTag)
}

// decodeValue decodes a single value from the form into the provided destination value.
func (d *Decoder) decodeValue(dest reflect.Value, rawValue, formTag string) error {
	switch dest.Type() {
	case durationType:
		duration, err := time.ParseDuration(rawValue)
		if err != nil {
			return ErrorDecode{fieldName: formTag, err: err}
		}
		dest.Set(reflect.ValueOf(duration))
		return nil
	}

	switch dest.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(rawValue, 0, dest.Type().Bits())
		if err != nil {
			return ErrorDecode{fieldName: formTag, err: err}
		}
		dest.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(rawValue, 0, dest.Type().Bits())
		if err != nil {
			return ErrorDecode{fieldName: formTag, err: err}
		}
		dest.SetUint(i)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(rawValue, dest.Type().Bits())
		if err != nil {
			return ErrorDecode{fieldName: formTag, err: err}
		}
		dest.SetFloat(f)

	case reflect.Bool:
		b, err := strconv.ParseBool(rawValue)
		if err != nil {
			return ErrorDecode{fieldName: formTag, err: err}
		}
		dest.SetBool(b)

	case reflect.String:
		dest.SetString(rawValue)
		return nil

	case reflect.Pointer:
		ensurePointerIsSet(dest)
		return d.decodeValue(dest.Elem(), rawValue, formTag)

	default:
		return fmt.Errorf("unsupported type %v", dest.Type())
	}

	return nil
}

// decodeSliceField decodes the form values into the provided slice field.
func (d *Decoder) decodeSliceField(dest reflect.Value, formTag string) error {
	return d.decodeSliceValue(dest, d.src[formTag], formTag)
}

// decodeSliceValue decodes the values from the source slice into the provided destination slice.
func (d *Decoder) decodeSliceValue(dest reflect.Value, rawValues []string, formTag string) error {
	sliceType := dest.Type()

	for _, val := range rawValues {
		elem := reflect.New(sliceType.Elem()).Elem()
		err := d.decodeValue(elem, val, formTag)
		if err != nil {
			return err
		}

		dest.Set(reflect.Append(dest, elem))
	}

	return nil
}

// decodeMap decodes the form values into the provided map field.
func (d *Decoder) decodeMap(dest reflect.Value, formTag string) error {
	regex, err := regexp.Compile(fmt.Sprintf("^%s\\[(.*)]$", formTag))
	if err != nil {
		return ErrorDecode{fieldName: formTag, err: err}
	}

	mapType := dest.Type()
	m := reflect.MakeMap(mapType)

	// Find all src keys that match the form tag.
	for rawKey, val := range d.src {
		captureGroups := regex.FindStringSubmatch(rawKey)
		if len(captureGroups) == 0 || len(val) == 0 {
			continue
		}

		if len(captureGroups) != 2 {
			return ErrorDecode{fieldName: formTag, err: fmt.Errorf("invalid map key: %v", captureGroups)}
		}

		// Handle single values or slices.
		sliceType := mapType.Elem()
		sliceVal := reflect.New(sliceType).Elem()
		if mapType.Elem().Kind() == reflect.Slice {
			err = d.decodeSliceValue(sliceVal, val, formTag)
			if err != nil {
				return ErrorDecode{fieldName: formTag, err: fmt.Errorf("error decoding map slice: %v", err)}
			}

			m.SetMapIndex(reflect.ValueOf(captureGroups[1]), sliceVal)
		} else {
			err = d.decodeValue(sliceVal, val[0], formTag)
			if err != nil {
				return ErrorDecode{fieldName: formTag, err: fmt.Errorf("error decoding map value: %v", err)}
			}

			m.SetMapIndex(reflect.ValueOf(captureGroups[1]), sliceVal)
		}
	}

	if m.Len() > 0 {
		dest.Set(m)
	}

	return nil
}

// ensurePointerIsSet checks if the provided value is a nil pointer, and sets the internal value if the value is nil.
func ensurePointerIsSet(val reflect.Value) {
	if val.Kind() == reflect.Pointer && val.IsNil() {
		val.Set(reflect.New(val.Type().Elem()))
	}
}
