package schema

import (
	"encoding/json"
	"fmt"
)

type PairTuple struct {
	Left  string
	Right string
	Date  string // YYYYMMDD
}

func (p PairTuple) MarshalJSON() ([]byte, error) {
	return json.Marshal(&[3]any{p.Left, p.Right, p.Date})
}

func (p *PairTuple) UnmarshalJSON(data []byte) error {
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return fmt.Errorf("invalid pair tuple format: %w", err)
	}
	if len(arr) != 3 {
		return fmt.Errorf("pair tuple should have exactly 3 elements, got %d", len(arr))
	}
	p.Left = arr[0]
	p.Right = arr[1]
	p.Date = arr[2]
	return nil
}

type EmojiKitchenData struct {
	LastAmended string            `json:"last_amended"`
	Canonicals  []string          `json:"canonicals"`
	Dates       []string          `json:"dates"`
	Aliases     map[string]string `json:"aliases"`
	Pairs       []PairTuple       `json:"pairs"`
}
