package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
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
	emojis, err := getAllEmojis()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(len(emojis))
	fmt.Println(emojis)
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

func getAllEmojis() ([]string, error) {
	var response *http.Response
	var err error
	response, err = http.Get("https://unicode.org/Public/emoji/latest/emoji-test.txt")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var emojis []string
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "; fully-qualified") {
			continue
		}
		if !strings.Contains(line, "#") {
			continue
		}
		if strings.Contains(line, "skin tone") {
			continue
		}
		codePoint := strings.TrimSpace(strings.Split(line, ";")[0])
		subparts := strings.Fields(codePoint)
		for i, part := range subparts {
			subparts[i] = strings.ToLower(part)
		}
		emojis = append(emojis, codepointToEmoji(strings.Join(subparts, "-")))
	}
	if err := scanner.Err();
	err != nil {
		return nil, err
	}
	return emojis, nil
}
