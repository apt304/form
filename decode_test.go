package form

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func MustParseTime(timestamp string) time.Time {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		panic(err)
	}
	return t
}

func MustParseDuration(duration string) time.Duration {
	d, err := time.ParseDuration(duration)
	if err != nil {
		panic(err)
	}
	return d
}

type FormStruct struct {
	StringParam     string   `form:"stringParam,omitempty"`
	StringPtrParam  *string  `form:"stringPtrParam,omitempty"`
	IntParam        int      `form:"int_param,omitempty"`
	IntPtrParam     *int     `form:"int_ptr_param,omitempty"`
	Int64Param      int64    `form:"int64_param,omitempty"`
	UintParam       uint     `form:"uint_param,omitempty"`
	UintPtrParam    *uint    `form:"uint_ptr_param,omitempty"`
	Float32Param    float32  `form:"float32_param,omitempty"`
	Float32PtrParam *float32 `form:"float32_ptr_param,omitempty"`
	Float64Param    float64  `form:"float64_param,omitempty"`
	Float64PtrParam *float64 `form:"float64_ptr_param,omitempty"`
	BoolParam       bool     `form:"bool_param,omitempty"`
	BoolPtrParam    *bool    `form:"bool_ptr_param,omitempty"`

	TimeParam        time.Time      `form:"timeParam,omitempty"`
	TimePtrParam     *time.Time     `form:"time_ptr_param,omitempty"`
	DurationParam    time.Duration  `form:"durationParam,omitempty"`
	DurationPtrParam *time.Duration `form:"duration_ptr_param,omitempty"`

	SliceParam       []string `form:"slice_param,omitempty"`
	SliceIntPtrParam []*int   `form:"slice_int_ptr_param,omitempty"`

	MapStringStringSlice map[string][]string `form:"map_string_slice,omitempty"`
	MapStringString      map[string]string   `form:"map_string,omitempty"`
	MapStringIntSlice    map[string][]int    `form:"map_string_int_slice,omitempty"`
	MapStringInt         map[string]int      `form:"map_string_int,omitempty"`

	NestedStruct FormStructNested `form:"nestedStruct,omitempty"`

	IgnoreParam     string `form:"-"`
	MissingParam    string
	unexportedParam string `form:"unexportedParam"` //nolint:all
}

type FormStructNested struct {
	NestedString string `form:"nestedString,omitempty"`
	NestedInt    int    `form:"nestedInt,omitempty"`
}

type TruckSettingsInput struct {
	Location         *string `form:"location"` // Eventually will become structured, once we have location services
	OpenTime         *string `form:"openTime"`
	CloseTime        *string `form:"closeTime"`
	AutoAcceptOrders *bool   `form:"autoAcceptOrders"`
	AutoClose        *bool   `form:"autoClose"`
}

func TestUnmarshal(t *testing.T) {
	type In struct {
		name     string
		FormData url.Values
	}
	type Out struct {
		Err  string
		Resp any
	}

	tests := []struct {
		in  In
		out Out
	}{
		{
			in: In{
				name: "success -- literals and slices",
				FormData: url.Values{
					"stringParam":       []string{"one"},
					"stringPtrParam":    []string{"two"},
					"int_param":         []string{"2"},
					"int_ptr_param":     []string{"3"},
					"int64_param":       []string{"4"},
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
					"duration_ptr_param": []string{"30m"},

					"slice_param":         []string{"4", "5"},
					"slice_int_ptr_param": []string{"6", "7"},

					"unexportedParam": []string{"should", "not", "decode"},
				},
			},
			out: Out{
				Resp: FormStruct{
					StringParam:     "one",
					StringPtrParam:  toPtr("two"),
					IntParam:        2,
					IntPtrParam:     toPtr(3),
					Int64Param:      4,
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
		},
		{
			in: In{
				name: "success -- maps",
				FormData: url.Values{
					"map_string_slice[keyOne]":  []string{"valueOne"},
					"map_string[key1]":          []string{"value1"},
					"map_string_slice[keyTwo]":  []string{"valueTwo"},
					"map_string[key2]":          []string{"value2"},
					"map_string_int_slice[key]": []string{"1", "2"},
					"map_string_int[key]":       []string{"2"},
				},
			},
			out: Out{
				Resp: FormStruct{
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
		},
		{
			in: In{
				name: "failure -- bad int param",
				FormData: url.Values{
					"int_param": []string{"two"},
				},
			},
			out: Out{
				Resp: FormStruct{},
				Err:  "Unable to decode tag 'int_param': strconv.ParseInt: parsing \"two\": invalid syntax",
			},
		},
		{
			in: In{
				name: "failure -- bad uint param",
				FormData: url.Values{
					"uint_param": []string{"two"},
				},
			},
			out: Out{
				Resp: FormStruct{},
				Err:  "Unable to decode tag 'uint_param': strconv.ParseUint: parsing \"two\": invalid syntax",
			},
		},
		{
			in: In{
				name: "failure -- bad float param",
				FormData: url.Values{
					"float32_param": []string{"two"},
				},
			},
			out: Out{
				Resp: FormStruct{},
				Err:  "Unable to decode tag 'float32_param': strconv.ParseFloat: parsing \"two\": invalid syntax",
			},
		},
		{
			in: In{
				name: "failure -- bad bool param",
				FormData: url.Values{
					"bool_param": []string{"two"},
				},
			},
			out: Out{
				Resp: FormStruct{},
				Err:  "Unable to decode tag 'bool_param': strconv.ParseBool: parsing \"two\": invalid syntax",
			},
		},
		{
			in: In{
				name: "failure -- bad duration param",
				FormData: url.Values{
					"durationParam": []string{"not a duration"},
				},
			},
			out: Out{
				Resp: FormStruct{},
				Err:  "Unable to decode tag 'durationParam': time: invalid duration \"not a duration\"",
			},
		},
		{
			in: In{
				name: "failure -- bad time param",
				FormData: url.Values{
					"timeParam": []string{"not a time"},
				},
			},
			out: Out{
				Resp: FormStruct{},
				Err:  "Unable to decode tag 'timeParam': parsing time \"not a time\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"not a time\" as \"2006\"",
			},
		},
		{
			in: In{
				name: "failure -- bad map int param",
				FormData: url.Values{
					"map_string_int[keyOne]": []string{"not an int"},
				},
			},
			out: Out{
				Resp: FormStruct{},
				Err:  "Unable to decode tag 'map_string_int': error decoding map value: Unable to decode tag 'map_string_int': strconv.ParseInt: parsing \"not an int\": invalid syntax",
			},
		},
		{
			in: In{
				name: "failure -- bad map slice param",
				FormData: url.Values{
					"map_string_int_slice[keyOne]": []string{"not an int"},
				},
			},
			out: Out{
				Resp: FormStruct{},
				Err:  "Unable to decode tag 'map_string_int_slice': error decoding map slice: Unable to decode tag 'map_string_int_slice': strconv.ParseInt: parsing \"not an int\": invalid syntax",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.in.name, func(t *testing.T) {
			var err error
			var response any
			if _, ok := tt.out.Resp.(FormStruct); ok {
				tempResp := FormStruct{}
				err = Unmarshal(tt.in.FormData, &tempResp)
				response = tempResp
			} else if _, ok := tt.out.Resp.(TruckSettingsInput); ok {
				tempResp := TruckSettingsInput{}
				err = Unmarshal(tt.in.FormData, &tempResp)
				response = tempResp
			}

			if tt.out.Err == "" {
				assert.NoError(t, err, "expected nil error")
			} else {
				assert.ErrorContainsf(t, err, tt.out.Err, "expected equal errors")
			}
			assert.Equal(t, tt.out.Resp, response, "expected equal form struct")
		})
	}
}

func TestUnmarshal_StructPointer(t *testing.T) {
	input := url.Values{
		"stringParam": []string{"one"},
	}
	val := FormStruct{}
	err := Unmarshal(input, &val)
	assert.NoError(t, err, "unexpected error")
	assert.Equal(t, input["stringParam"][0], val.StringParam, "expected stringParams to equal")
}

func TestUnmarshal_NonStruct(t *testing.T) {
	err := Unmarshal(url.Values{}, []string{})
	assert.ErrorContains(t, err, "destination ([]string) must be a pointer to a struct", "unexpected error")
}
