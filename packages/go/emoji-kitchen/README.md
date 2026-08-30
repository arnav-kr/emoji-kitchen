# emoji-kitchen

[![Go Reference](https://pkg.go.dev/badge/github.com/arnav-kr/emoji-kitchen/packages/go/emoji-kitchen.svg)](https://pkg.go.dev/github.com/arnav-kr/emoji-kitchen/packages/go/emoji-kitchen)

A Go library to check, mix, and search Google's Emoji Kitchen combinations. The dataset is bundled with the package, so there are no files to download and no API calls; everything works offline and instantly.

## Features

- No setup: import the library and start using it
- Fast: lookups run locally in microseconds
- Flexible input: pass raw emojis (`"😺"`) or hex codepoints (`"1f63a"`)
- Offline: no network calls at runtime

---

## Installation

```bash
go get github.com/arnav-kr/emoji-kitchen/packages/go/emoji-kitchen
```

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/arnav-kr/emoji-kitchen/packages/go/emoji-kitchen"
)

func main() {
	// Mix two emojis
	if combo, ok := emojikitchen.Mix("😺", "🐼"); ok {
		fmt.Printf("Mix: %s + %s merged on %s\n", combo.Left, combo.Right, combo.Date)
		fmt.Printf("Sticker URL: %s\n\n", combo.URL())
	}

	// You can use raw emojis or hex codepoints
	if combo, ok := emojikitchen.Mix("1f63a", "1f43c"); ok {
		fmt.Printf("Mix using codepoints succeeded: %s\n\n", combo.URL())
	}

	// Check if a combination exists
	canMix := emojikitchen.CanMix("😺", "💩")
	fmt.Printf("Can you mix Cat and Poop? %t\n\n", canMix)

	// Get all valid mixes for an emoji
	partners := emojikitchen.Partners("😺")
	fmt.Printf("😺 can mix with %d other emojis! First 5: %v...\n", len(partners), partners[:5])
}
```

## API Reference

### `Mix(left, right string) (Combo, bool)`

Combines two emojis and returns the result. Accepts raw emojis or hex codepoints. The boolean reports whether the combination exists.

```go
type Combo struct {
	Left  string
	Right string
	Date  string // YYYYMMDD
}

func (c Combo) URL() string // gstatic PNG URL for the sticker
```

### `CanMix(left, right string) bool`

Reports whether a combination exists for the two emojis.

### `Partners(emoji string) []string`

Returns all emojis that can be mixed with the given emoji.

### `AllEmoji() []string`

Returns all supported emojis.

### `Combos() iter.Seq[Combo]`

Iterates over every valid combination in the dataset:

```go
for combo := range emojikitchen.Combos() {
	// ...
}
```

### `Stats() ekbf.Stats`

Returns counts of emojis, aliases, and valid pairings.

## License

[AGPL-3.0](https://github.com/arnav-kr/emoji-kitchen/blob/main/LICENSE)
