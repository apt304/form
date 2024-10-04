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

	fmt.Fprintf(w, "Parsed form: %+v", sample)
}

func main() {
	http.HandleFunc("/submit", handler)
	http.ListenAndServe(":8080", nil)
}
```

## Comparison to `gorilla/schema`

`gorilla/schema` enables marshaling and unmarshaling form values to and from typed structs. However, it does not support dynamic fields that map key/value pairs. This library was created to extend the base functionality `gorilla/schema` provides by supporting the base typed struct conversion, as well as dynamic data pairs.

## Benchmarks

This library marshals and unmarshals form values to and from structs, much like `gorilla/schema`. Flat struct processing (without dynamic map fields) is compared between the two libraries.

```shell
go test -bench=. -benchtime 10s
goos: darwin
goarch: arm64
pkg: github.com/apt304/form
BenchmarkUnmarshal-10              	11018934	        1077 ns/op
BenchmarkGorillaSchemaDecode-10    	 5794143	        2058 ns/op
BenchmarkMarshal-10                	12946148	       922.5 ns/op
BenchmarkGorillaSchemaEncode-10    	12842407	       935.8 ns/op
PASS
ok  	github.com/apt304/form	53.294s
```

_Benchmark run on Macbook Pro M1 Max_

## Contributions

Contributions are welcome! Feel free to open issues or submit pull requests. For more information, please see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

This library is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
