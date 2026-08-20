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
	var raw [3]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("pair should have exactly 3 elements: %w", err)
	}

	if err := json.Unmarshal(raw[0], &p.Left); err != nil {
		return fmt.Errorf("invalid left: %w", err)
	}
	if err := json.Unmarshal(raw[1], &p.Right); err != nil {
		return fmt.Errorf("invalid right: %w", err)
	}
	if err := json.Unmarshal(raw[2], &p.Date); err != nil {
		return fmt.Errorf("invalid date: %w", err)
	}
	return nil
}

type EmojiKitchenData struct {
	LastAmended string            `json:"last_amended"`
	Canonicals  []string          `json:"canonicals"`
	Dates       []string          `json:"dates"`
	Aliases     map[string]string `json:"aliases"`
	Pairs       []PairTuple       `json:"pairs"`
}
