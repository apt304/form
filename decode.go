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

type ErrorDecode struct {
	fieldName string
	err       error
}

func (e ErrorDecode) Error() string {
	return fmt.Sprintf("Unable to decode tag '%s': %s", e.fieldName, e.err)
}

// Rules:
//		1. Form data is flat -- the key/value mapping is always at the top level. Keys must contain dynamic encoding.
//		2. Nested data structures are represented with dot-notation keys. e.g. nestedStruct.Param1, nestedStruct.Param2
//		3. Dynamic form data is modeled with param[key] = val syntax. `param` matches with the form struct tag. Keys are
//			parsed and used to populate the map.

// Check if the value slice len > 1. If it is and the value is a slice, parse all values. If the value is not a slice,
// parse the first value. Add a config option to error on unexpected slice values.

// Map syntax: brackets -- paramName[key] = value

// TODO: Nested struct: Dot syntax. e.g.  nestedStruct.param1, nestedStruct.param2

// Unmarshal iterates over the fields in `dest`, populating them with the appropriate fields from values.
func Unmarshal(src map[string][]string, dest any) error {
	return NewDecoder(src).Decode(dest)
}

type Decoder struct {
	src map[string][]string
}

func NewDecoder(src map[string][]string) *Decoder {
	return &Decoder{src: src}
}

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

// dest must be a non-pointer value
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

func (d *Decoder) decodeFormField(dest reflect.Value, formTag string) error {
	if dest.Kind() != reflect.Map && len(d.src[formTag]) == 0 {
		return nil
	}

	// Check overridden TextUnmarshaler types first, either the value itself, or a pointer to it.
	// e.g. *time.Time implements TextUnmarshaler. If a time.Time value is provided, the reference
	// must be processed.
	if dest.Type().Implements(textUnmarshalerType) ||
		(dest.CanAddr() && dest.Addr().Type().Implements(textUnmarshalerType)) {
		// If the destination itself doesn't implement TextUnmarshaler, take the pointer and recursively call
		// decodeFormField.
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
		// Recursively call decode() with the pointer's element
		// Decode the element the pointer references
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

	// Decode value. Take the first value from the source slice
	var strVal string
	if len(d.src[formTag]) > 0 {
		strVal = d.src[formTag][0]
	}

	return d.decodeValue(dest, strVal, formTag)
}

// TODO setup decoding for time.time. The primary usecase is for html <input type="time" /> tags
func (d *Decoder) decodeValue(dest reflect.Value, rawValue, formTag string) error {
	switch dest.Type() {
	// Custom types (duration)
	case durationType:
		duration, err := time.ParseDuration(rawValue)
		if err != nil {
			return ErrorDecode{fieldName: formTag, err: err}
		}
		dest.Set(reflect.ValueOf(duration))
		return nil
	}

	switch dest.Kind() {
	// literals
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
		// Recursively call decode() with the pointer's element
		// This additional pointer dereferencing is needed to handle slice pointer values. Top-level pointers are
		// handled in decodeFormField.
		ensurePointerIsSet(dest)
		return d.decodeValue(dest.Elem(), rawValue, formTag)

	default:
		return fmt.Errorf("unsupported type %v", dest.Type())
	}

	return nil
}

func (d *Decoder) decodeSliceField(dest reflect.Value, formTag string) error {
	return d.decodeSliceValue(dest, d.src[formTag], formTag)
}

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

func (d *Decoder) decodeMap(dest reflect.Value, formTag string) error {
	regex, err := regexp.Compile(fmt.Sprintf("^%s\\[(.*)]$", formTag))
	if err != nil {
		return ErrorDecode{fieldName: formTag, err: err}
	}

	mapType := dest.Type()
	m := reflect.MakeMap(mapType)

	// Find all src keys that match the form tag
	for rawKey, val := range d.src {
		captureGroups := regex.FindStringSubmatch(rawKey)
		if len(captureGroups) == 0 || len(val) == 0 {
			// Ignore non-matching src keys and empty values
			continue
		}

		if len(captureGroups) != 2 {
			return ErrorDecode{fieldName: formTag, err: fmt.Errorf("invalid map key: %v", captureGroups)}
		}

		// Handle single values or slices
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

// ensurePointer checks if the provided value is a nil pointer, and sets the internal value, if the value is nil.
func ensurePointerIsSet(val reflect.Value) {
	if val.Kind() == reflect.Pointer && val.IsNil() {
		val.Set(reflect.New(val.Type().Elem()))
	}
}
