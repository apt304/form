package form

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"math"
	"net/url"
	"testing"
)

func TestMarshal(t *testing.T) {
	type In struct {
		name     string
		FormData any
	}
	type Out struct {
		Err  error
		Resp map[string][]string
	}

	tests := []struct {
		in  In
		out Out
	}{
		{
			in: In{
				name: "success -- literals and slice",
				FormData: FormStruct{
					StringParam:     "string param",
					StringPtrParam:  toPtr("two"),
					IntParam:        2,
					IntPtrParam:     toPtr(3),
					UintParam:       4,
					UintPtrParam:    toPtr(uint(5)),
					Float32Param:    6.0,
					Float32PtrParam: toPtr(float32(7)),
					Float64Param:    8.0,
					Float64PtrParam: toPtr(float64(9)),

					BoolParam:    true,
					BoolPtrParam: toPtr(false),

					TimeParam:        MustParseTime("2024-08-19T05:09:29-01:00"),
					TimePtrParam:     toPtr(MustParseTime("2024-08-19T05:09:29Z")),
					DurationParam:    MustParseDuration("30s"),
					DurationPtrParam: toPtr(MustParseDuration("30m")),

					SliceParam:       []string{"4", "5"},
					SliceIntPtrParam: []*int{toPtr(6), toPtr(7)},
				},
			},
			out: Out{
				Resp: url.Values{
					"stringParam":       []string{"string param"},
					"stringPtrParam":    []string{"two"},
					"int_param":         []string{"2"},
					"int_ptr_param":     []string{"3"},
					"uint_param":        []string{"4"},
					"uint_ptr_param":    []string{"5"},
					"float32_param":     []string{"6"},
					"float32_ptr_param": []string{"7"},
					"float64_param":     []string{"8"},
					"float64_ptr_param": []string{"9"},

					"bool_param":     []string{"true"},
					"bool_ptr_param": []string{"false"},

					"timeParam":          []string{"2024-08-19T05:09:29-01:00"},
					"time_ptr_param":     []string{"2024-08-19T05:09:29Z"},
					"durationParam":      []string{"30s"},
					"duration_ptr_param": []string{"30m0s"},

					"slice_param":         []string{"4", "5"},
					"slice_int_ptr_param": []string{"6", "7"},
				},
			},
		},
		{
			in: In{
				name: "success -- maps",
				FormData: FormStruct{
					MapStringStringSlice: map[string][]string{
						"keyOne": {"valueOne"},
						"keyTwo": {"valueTwo"},
					},
					MapStringString: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					MapStringIntSlice: map[string][]int{
						"key": {1, 2},
					},
					MapStringInt: map[string]int{
						"key": 2,
					},
				},
			},
			out: Out{
				Resp: url.Values{
					"map_string_slice[keyOne]":  []string{"valueOne"},
					"map_string[key1]":          []string{"value1"},
					"map_string_slice[keyTwo]":  []string{"valueTwo"},
					"map_string[key2]":          []string{"value2"},
					"map_string_int_slice[key]": []string{"1", "2"},
					"map_string_int[key]":       []string{"2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.in.name, func(t *testing.T) {
			formValues, err := Marshal(tt.in.FormData)
			assert.Equal(t, tt.out.Err, err, "expected equal errors")
			if err == nil {
				assert.Equal(t, tt.out.Resp, formValues, "expected equal form struct")
			} else {
				assert.Nil(t, tt.out.Resp, "expected nil form struct")
			}
		})
	}
}

// FuzzStruct does not include time.Time or time.Duration values, as they are unlikely to round-trip marshal and
// unmarshal.
type FuzzStruct struct {
	StringParam     string   `form:"stringParam,omitempty"`
	StringPtrParam  *string  `form:"stringPtrParam,omitempty"`
	IntParam        int      `form:"int_param,omitempty"`
	IntPtrParam     *int     `form:"int_ptr_param,omitempty"`
	UintParam       uint     `form:"uint_param,omitempty"`
	UintPtrParam    *uint    `form:"uint_ptr_param,omitempty"`
	Float32Param    float32  `form:"float32_param,omitempty"`
	Float32PtrParam *float32 `form:"float32_ptr_param,omitempty"`
	Float64Param    float64  `form:"float64_param,omitempty"`
	Float64PtrParam *float64 `form:"float64_ptr_param,omitempty"`
	BoolParam       bool     `form:"bool_param,omitempty"`
	BoolPtrParam    *bool    `form:"bool_ptr_param,omitempty"`

	SliceParam       []string `form:"slice_param,omitempty"`
	SliceIntPtrParam []*int   `form:"slice_int_ptr_param,omitempty"`

	MapStringStringSlice map[string][]string `form:"map_string_slice,omitempty"`
	MapStringString      map[string]string   `form:"map_string,omitempty"`
	MapStringIntSlice    map[string][]int    `form:"map_string_int_slice,omitempty"`
	MapStringInt         map[string]int      `form:"map_string_int,omitempty"`

	NestedStruct FormStructNested `form:"nestedStruct,omitempty"`
}

func FuzzEncode(f *testing.F) {
	type Test struct {
		encodedForm string
	}

	tests := []Test{
		{
			encodedForm: "durationParam=30s",
		},
		{
			encodedForm: "durationParam=30s&duration_ptr_param=30m0s&float32_param=6&float32_ptr_param=7&float64_param=8&float64_ptr_param=9&int_param=2&int_ptr_param=3&slice_int_ptr_param=6&slice_int_ptr_param=7&slice_param=4&slice_param=5&stringParam=one&stringPtrParam=two&timeParam=2024-08-19T05%3A09%3A29-01%3A00&time_ptr_param=2024-08-19T05%3A09%3A29Z&uint_param=4&uint_ptr_param=5",
		},
		{
			encodedForm: "map_string%5Bkey1%5D=value1&map_string%5Bkey2%5D=value2&map_string_int%5Bkey%5D=2&map_string_int_slice%5Bkey%5D=1&map_string_int_slice%5Bkey%5D=2&map_string_slice%5BkeyOne%5D=valueOne&map_string_slice%5BkeyTwo%5D=valueTwo",
		},
		{
			encodedForm: "0",
		},
		{
			encodedForm: "durationParam=0&duration_ptr_param=0&float32_param=1&float64_param=1&float64_ptr_param=0&int_param=1&int_ptr_param=0&slice_int_ptr_param=0&slice_param&&stringParam=0&stringPtrParam&&timeParam=0000-01-01T0%3A00%3A00-00%3A00",
		},
		{
			encodedForm: "timeParam",
		},
		{
			encodedForm: "int64_param=4",
		},
	}

	for _, tc := range tests {
		f.Add(tc.encodedForm) // Use f.Add to provide a seed corpus
	}

	f.Fuzz(func(t *testing.T, encodedForm string) {
		parsedForm, err := url.ParseQuery(encodedForm)
		if err != nil {
			t.Skip()
		}

		parsedMap := map[string][]string(parsedForm)

		var formInput FuzzStruct
		err = Unmarshal(parsedMap, &formInput)
		if err != nil && !errors.As(err, &ErrorDecode{}) {
			t.Fatal(err)
		}

		// Skip cases where floating values are NaN. IEEE defines NaN != NaN
		if math.IsNaN(float64(formInput.Float32Param)) ||
			(formInput.Float32PtrParam != nil && math.IsNaN(float64(*formInput.Float32PtrParam))) ||
			math.IsNaN(formInput.Float64Param) ||
			(formInput.Float64PtrParam != nil && math.IsNaN(*formInput.Float64PtrParam)) {
			t.Skip()
		}

		result, err := Marshal(formInput)
		if err != nil {
			t.Fatal(err)
		}

		var formInputRoundTrip FuzzStruct
		err = Unmarshal(result, &formInputRoundTrip)
		if err != nil && !errors.As(err, &ErrorDecode{}) {
			t.Fatal(err)
		}

		assert.Equal(t, formInput, formInputRoundTrip, "expected equal form inputs")
	})
}
