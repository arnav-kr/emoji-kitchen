package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type tenorResult struct {
	MediaFormats struct {
		PNGTransparent struct {
			URL string `json:"url"`
		} `json:"png_transparent"`
	} `json:"media_formats"`
	Tags []string `json:"tags"`
}

type tenorResponse struct {
	Results []tenorResult `json:"results"`
}

type PairTuple struct {
	Left string
	Right string
	DateIndex int
}

func (p PairTuple) MarshalJSON() ([]byte, error) {
	return json.Marshal(&[3]any{p.Left, p.Right, p.DateIndex})
}

type intermediatePair struct {
	Left string
	Right string
	DateString string
}

func main() {
	apiKey := os.Getenv("TENOR_API_KEY")
	if apiKey == "" {
		apiKey = "AIzaSyACvEq5cnT7AcHpDdj64SE3TJZRhW-iHuo"
	}

	baseURL := "https://tenor.googleapis.com/v2/featured?key=" + apiKey + "&limit=2&media_filter=png_transparent&collection=emoji_kitchen_v6&client_key=emoji_kitchen_funbox&q=%s"
	fmt.Printf(baseURL, "👩‍👧‍👦")
	a := emojiToCodePoint("👩‍👧‍👦")
	fmt.Println(codepointToEmoji(a))
}

func emojiToCodePoint(emoji string) string {
	parts := make([]string, 0, len(emoji))

	for _, r := range emoji {
		parts = append(parts, strconv.FormatInt(int64(r), 16))
	}
	return strings.Join(parts, "-")
}

func codepointToEmoji(hexStr string) string {
	var b strings.Builder
	b.Grow(len(hexStr))

	for p := range strings.SplitSeq(hexStr, "-") {
		v, err := strconv.ParseInt(p, 16, 32)
		if err == nil {
			b.WriteRune(rune(v))
		}
	}
	return b.String()
}