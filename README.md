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

## Contributions

Contributions are welcome! Feel free to open issues or submit pull requests. For more information, please see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

This library is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
