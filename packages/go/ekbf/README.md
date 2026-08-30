# ekbf

[![Go Reference](https://pkg.go.dev/badge/github.com/arnav-kr/emoji-kitchen/packages/go/ekbf.svg)](https://pkg.go.dev/github.com/arnav-kr/emoji-kitchen/packages/go/ekbf)

Reader and writer for the Emoji Kitchen Binary Format (EKBF), the format used to store Emoji Kitchen combination data.

Use this package if you want to compile your own `.ekbf` data files, read existing ones, or build tooling around the format. If you just want to look up emoji combinations in a Go application, use the [emoji-kitchen package](../emoji-kitchen/README.md) instead; it bundles the data and handles everything for you.

## Why a custom format?

Emoji combinations are symmetric (A + B is the same as B + A), so storing them as individual records in JSON or SQLite duplicates every pair. EKBF stores each combination once, keeps the full dataset of 100,000+ pairs at ~391 KB (versus ~8.7 MB as JSON), and reads values directly from the binary data instead of parsing the whole file into memory.

## Installation

```bash
go get github.com/arnav-kr/emoji-kitchen/packages/go/ekbf
```

## Reading an `.ekbf` file

```go
package main

import (
	"fmt"
	"os"

	"github.com/arnav-kr/emoji-kitchen/packages/go/ekbf"
)

func main() {
	data, _ := os.ReadFile("data.ekbf")
	reader, _ := ekbf.NewReader(data)

	pair, ok := reader.Lookup("1f63a", "1f43c")
	if ok {
		fmt.Printf("Date: %s | URL: %s\n", pair.Date, pair.URL())
	}
}
```

## Compiling an `.ekbf` file

```go
package main

import (
	"os"

	"github.com/arnav-kr/emoji-kitchen/packages/go/ekbf"
)

func main() {
	builder := &ekbf.Builder{
		// Must be sorted alphabetically
		Canonicals: []string{"1f43c", "1f63a"},

		// Maps aliases back to canonicals
		Aliases: map[string]string{
			"1f600": "1f63a",
		},

		// Sorted unique release dates (YYYYMMDD)
		Dates: []string{"20201001", "20210515"},

		Pairs: []ekbf.Pair{
			{Left: "1f63a", Right: "1f43c", Date: "20210515"},
		},
	}

	outFile, _ := os.Create("custom.ekbf")
	defer outFile.Close()

	_ = builder.Encode(outFile)
}
```

## Format layout

A compiled `.ekbf` file consists of six sequential sections:

1. **Header** (12 bytes): magic bytes `EKBF`, format version, and counts of canonicals, aliases, and dates.
2. **Canonical records**: offsets to each canonical emoji's string in the string table.
3. **Alias records**: mappings from alias strings to canonical IDs.
4. **Dates array**: release dates stored as `YYYYMMDD` integers.
5. **Matrix**: one entry per unique emoji pair, storing the date index and pair order.
6. **String table**: null-terminated emoji codepoint strings.

For the full byte-level specification, see [spec.pdf](../../../spec.pdf).

## License

[AGPL-3.0](https://github.com/arnav-kr/emoji-kitchen/blob/main/LICENSE)
