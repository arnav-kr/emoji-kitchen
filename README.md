# Emoji Kitchen

<p align="center">
  <img src="logo.png" alt="Emoji Kitchen Logo" width="160"/>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/github/license/arnav-kr/emoji-kitchen" alt="License"/></a>
  <a href="https://pkg.go.dev/github.com/arnav-kr/emoji-kitchen/packages/go/emoji-kitchen"><img src="https://pkg.go.dev/badge/github.com/arnav-kr/emoji-kitchen/packages/go/emoji-kitchen.svg" alt="Go Reference"/></a>
  <img src="https://img.shields.io/badge/typescript-WIP-orange" alt="TypeScript Status"/>
</p>

## About

Libraries and tools for working with Google's Emoji Kitchen combinations entirely offline. No network requests and no external API dependencies; the data ships with the library.

This project grew out of [emoji-kitchen.vercel.app](https://emoji-kitchen.vercel.app), an Emoji Kitchen web app I created. As its user base grew, the architecture needed an overhaul to keep up, and that work evolved into this standalone set of libraries. The website is great for quick, casual testing, but production apps should rely on these libraries instead. That also takes load off the underlying API, helping it stay within its free quota.

## Why not just use JSON?

Storing 100,000+ combinations as JSON takes around 8.7 MB. You can bring that down to 2-3 MB by storing dates in an array and referencing indices, but JSON still has to be fully parsed and stringified at runtime, which costs startup time and memory.

EKBF (Emoji Kitchen Binary Format) is a custom binary format built for this dataset:

| Format | File Size | Runtime Cost |
| :--- | :---: | :--- |
| Standard JSON | ~8.7 MB | Full parse on load, high memory use |
| Optimized JSON (date indices) | ~2.3 MB | Still fully parsed on load |
| EKBF | ~391 KB | No parsing, data read on demand |

Since mixing A + B is the same as mixing B + A, EKBF stores each combination only once and resolves the order at query time.

## Repository Layout

| Path | Description |
| :--- | :--- |
| [`packages/go/emoji-kitchen`](packages/go/emoji-kitchen) | The main library. Data is bundled in, so it works out of the box. Use this if you just want to look up and mix emojis in a Go application. |
| [`packages/go/ekbf`](packages/go/ekbf) | Reader and writer for the EKBF binary format. Use this if you want to compile your own data files or build custom tooling around the format. |
| [`packages/ts/emoji-kitchen`](packages/ts/emoji-kitchen) | TypeScript port of the main library. Work in progress. |
| [`packages/ts/ekbf`](packages/ts/ekbf) | TypeScript port of the EKBF format. Work in progress. |
| [`cmd/crawl`](cmd/crawl) | Crawls Google/Tenor to fetch and validate combinations. |
| [`cmd/compile`](cmd/compile) | Compiles the crawled JSON into an `.ekbf` binary. |
| [`cmd/decompile`](cmd/decompile) | Converts an `.ekbf` binary back to JSON. |

## Quick Start

```bash
go get github.com/arnav-kr/emoji-kitchen/packages/go/emoji-kitchen
```

```go
package main

import (
	"fmt"

	"github.com/arnav-kr/emoji-kitchen/packages/go/emoji-kitchen"
)

func main() {
	if combo, ok := emojikitchen.Mix("😺", "🐼"); ok {
		fmt.Println("Sticker URL:", combo.URL())
	}
}
```

## Docs

- [emoji-kitchen package](packages/go/emoji-kitchen/README.md): API reference and examples for the main library.
- [ekbf package](packages/go/ekbf/README.md): Usage examples and the binary format layout.
- [spec.pdf](spec.pdf): Full specification of the EKBF binary format.

## License

[AGPL-3.0](LICENSE)
