package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/arnav-kr/emoji-kitchen/packages/go/ekbf"
)

type PairTuple struct {
	Left      string 
	Right     string
	DateIndex int
}

func (p *PairTuple) UnmarshalJSON(data []byte) error {
	var raw [3]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("PairTuple size should be 3: %w", err)
	}

	if err := json.Unmarshal(raw[0], &p.Left); err != nil {
		return fmt.Errorf("invalid left: %w", err)
	}
	if err := json.Unmarshal(raw[1], &p.Right); err != nil {
		return fmt.Errorf("invalid right: %w", err)
	}
	if err := json.Unmarshal(raw[2], &p.DateIndex); err != nil {
		return fmt.Errorf("invalid date index: %w", err)
	}

	return nil
}

type emojiKitchenData struct {
	LastAmended string            `json:"last_amended"`
	Canonicals  []string          `json:"canonicals"`
	Dates       []string          `json:"dates"`
	Aliases     map[string]string `json:"aliases"`
	Pairs       []PairTuple       `json:"pairs"`
}

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

	var sourceJSON emojiKitchenData
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
			Left:      p.Left,
			Right:     p.Right,
			DateIndex: p.DateIndex,
		})
	}

	if outPath == "" {
		outPath = fmt.Sprintf("./dist/emoji-kitchen-%s.ekbf", sourceJSON.LastAmended)
	}

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
