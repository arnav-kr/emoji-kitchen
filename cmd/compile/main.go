package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/arnav-kr/emoji-kitchen/cmd/internal/schema"
	"github.com/arnav-kr/emoji-kitchen/packages/go/ekbf"
)

func main() {
	var sourcePath, outPath string
	if len(os.Args) < 2 {
		fmt.Println("usage: compile <source> <output>")
		os.Exit(1)
	}
	sourcePath = os.Args[1]
	if len(os.Args) > 2 {
		outPath = os.Args[2]
	} else {
		outPath = ""
	}

	if err := compile(sourcePath, outPath); err != nil {
		fmt.Fprintf(os.Stderr, "compile: %v\n", err)
		os.Exit(1)
	}
}

func compile(sourcePath, outPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	var sourceJSON schema.EmojiKitchenData
	if err := json.NewDecoder(sourceFile).Decode(&sourceJSON); err != nil {
		return err
	}

	builder := &ekbf.Builder{
		Version:    ekbf.CurrentVersion,
		Canonicals: sourceJSON.Canonicals,
		Dates:      sourceJSON.Dates,
		Aliases:    sourceJSON.Aliases,
	}

	for _, p := range sourceJSON.Pairs {
		builder.Pairs = append(builder.Pairs, ekbf.Pair{
			Left:  p.Left,
			Right: p.Right,
			Date:  p.Date,
		})
	}

	if outPath == "" {
		outPath = fmt.Sprintf("./dist/emoji-kitchen-%s.ekbf", sourceJSON.LastAmended)
	}

	os.MkdirAll(filepath.Dir(outPath), 0755)
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if err := builder.Encode(outFile); err != nil {
		return err
	}
	return nil
}
