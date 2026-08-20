package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/arnav-kr/emoji-kitchen/cmd/internal/schema"
	"github.com/arnav-kr/emoji-kitchen/packages/go/ekbf"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run ./cmd/decompile <file.ekbf> [file.json]")
		os.Exit(1)
	}

	ekbfPath := os.Args[1]
	bytes, err := os.ReadFile(ekbfPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed reading ekbf: %v\n", err)
		os.Exit(1)
	}
	reader, err := ekbf.NewReader(bytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed creating reader: %v\n", err)
		os.Exit(1)
	}

	canonicals := reader.Canonicals()
	dates := reader.Dates()
	lastAmended := dates[len(dates)-1]

	aliases, err := reader.Aliases()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed getting aliases: %v\n", err)
		os.Exit(1)
	}

	var pairs []schema.PairTuple
	reader.Pairs(func(pair ekbf.Pair) bool {
		pairs = append(pairs, schema.PairTuple{
			Left:  pair.Left,
			Right: pair.Right,
			Date:  pair.Date,
		})
		return true
	})

	ekbfData := schema.EmojiKitchenData{
		LastAmended: lastAmended,
		Canonicals:  canonicals,
		Dates:       dates,
		Aliases:     aliases,
		Pairs:       pairs,
	}

	jsonData, err := json.MarshalIndent(ekbfData, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed marshalling: %v\n", err)
		os.Exit(1)
	}

	jsonPath := strings.TrimSuffix(ekbfPath, ".ekbf") + "-decompiled.json"
	err = os.WriteFile(jsonPath, jsonData, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed writing json: %v\n", err)
		os.Exit(1)
	}
}
