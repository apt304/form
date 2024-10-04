# Form

A Go library for parsing HTTP form data into struct fields with concrete types, including support for dynamic form data.

- **Parse Form Data into Structs**: Parse HTTP form data into Go structs with concrete types.
- **Support for Dynamic Fields**: Parse dynamic form data into a `map[string]<type>` embedded in structs.
- **Reduce Boilerplate**: Simplify development by removing manual type conversions in HTTP handlers.

## Installation

To install the library, run:

```sh
go get github.com/apt304/form
```

## Usage

Here's an example to get started:

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/apt304/form"
)

type SampleForm struct {
	ID          int               `form:"id"`
	DynamicData map[string]string `form:"dynamicData"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var sample SampleForm
	err = form.Unmarshal(r.Form, &sample)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	encodedForm, err := form.Marshal(sample)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Parsed form: %+v\nEncoded form: %+v", sample, encodedForm)
}

func main() {
	http.HandleFunc("/submit", handler)
	http.ListenAndServe(":8080", nil)
}
```

Form data is flat, with key/value pairings: `field = val`, where `field` is matched to a struct tag. Forms allow the same key to be reused: `field = val1, field = val2`. Multiple values can be handled by using a slice in the struct. Dynamic values are encoded in the keys with a `field[key] = val` syntax. Use `map[string]<type>` as the struct type to unmarshal these dynamic pairs.

```
These form values:
    id:               "123"
    dynamicData[one]: "val1"
    dynamicData[two]: "val2"

Unmarshal into:
    SampleForm {
        ID: 123,
        DynamicData: map[string]string{
            "one": "val1",
            "two": "val2"
        }
    }
```

## Comparison to `gorilla/schema`

`gorilla/schema` enables marshaling and unmarshaling form values to and from typed structs. However, it does not support dynamic fields that map key/value pairs. This library was created to expand on `gorilla/schema`'s base functionality by supporting typed struct conversion, as well as dynamic data pairs.

## Benchmarks

This library decodes and encodes form values to and from structs. Performance for flat struct processing is compared to `gorilla/schema`.

- **Decode**: `apt304/form` decodes form values into structs faster than `gorilla/schema` and has fewer memory allocations.
- **Encode**: Both libraries have comparable speed when encoding structs into form values. `gorilla/schema` encodes with fewer memory allocations.

```sh
go test -bench=. -benchtime 10s -benchmem
goos: darwin
goarch: arm64
pkg: github.com/apt304/form
BenchmarkDecode-10                      	10825464	      1083   ns/op	     640 B/op	      17 allocs/op
BenchmarkGorillaSchemaDecode-10         	 5848353	      2057   ns/op	    1104 B/op	      49 allocs/op
BenchmarkDecodeLarge-10                 	 3880474	      3120   ns/op	     480 B/op	      40 allocs/op
BenchmarkGorillaSchemaDecodeLarge-10    	 1000000	     11523   ns/op	    3392 B/op	     184 allocs/op
BenchmarkEncode-10                      	12548168	       939.4 ns/op	     782 B/op	      22 allocs/op
BenchmarkGorillaSchemaEncode-10         	13078416	       922.2 ns/op	     776 B/op	      19 allocs/op
BenchmarkEncodeLarge-10                 	 2776134	      4310   ns/op	    4196 B/op	      81 allocs/op
BenchmarkGorillaSchemaEncodeLarge-10    	 2687842	      4424   ns/op	    3831 B/op	      66 allocs/op
PASS
ok  	github.com/apt304/form	113.075s
```

_Benchmark run on Macbook Pro M1 Max_

## Contributions

Contributions are welcome! Feel free to open issues or submit pull requests. For more information, please see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

This library is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
