package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
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
	client := &http.Client{
		Timeout: 8 * time.Second,
				Transport: &http.Transport{
					MaxIdleConns:        200,
					MaxIdleConnsPerHost: 200,
					IdleConnTimeout:     30 * time.Second,
				},
	}

	res, err := queryTenor(client, "👩‍👧‍👦")
	if err != nil {
		fmt.Println(err)
	}
	if res == nil || len(res.Results) == 0 {
			fmt.Println("No results found for query")
			return
		}
	fmt.Println(res)
	fmt.Println(res.Results[0].MediaFormats.PNGTransparent.URL)
	parsed, err := parseStickerURL(res.Results[0].MediaFormats.PNGTransparent.URL)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(parsed)
	a := emojiToCodePoint("👩‍👧‍👦")
	fmt.Println(codepointToEmoji(a))
	emojis, err := getAllEmojis(client)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(len(emojis))
	validCanEmojis, validAliases, err := getValidEmojis(client, emojis)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(len(validCanEmojis), len(validAliases))

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

func stripPrefix(hexStr string) string {
	hexStr = strings.TrimPrefix(hexStr, "u")
	return strings.ReplaceAll(hexStr, "-u", "-")
}

func parseStickerURL(url string) (intermediatePair, error) {
	matches := regexp.MustCompile(`/emojikitchen/(?P<date>\d{8})/u(?P<dir>[0-9a-fA-F-u]+)/u(?P<left>[0-9a-fA-F-u]+)_u(?P<right>[0-9a-fA-F-u]+)\.png`).FindStringSubmatch(url)
	if matches == nil {
		return intermediatePair{}, fmt.Errorf("invalid sticker URL: %s", url)
	}
	date := matches[1]
	// dir := matches[2]
	left := matches[3]
	right := matches[4]

	return intermediatePair{
		DateString: date,
		Left:       stripPrefix(left),
		Right:      stripPrefix(right),
	}, nil
}

func getCanonicalEmoji(tags1, tags2 []string) string {
	for _, x := range tags1 {
		if slices.Contains(tags2, x) {
			return x
		}
	}
	return ""
}

func queryTenor(client *http.Client, query string) (*tenorResponse, error) {
	url := fmt.Sprintf("https://tenor.googleapis.com/v2/featured?key=AIzaSyACvEq5cnT7AcHpDdj64SE3TJZRhW-iHuo&limit=2&media_filter=png_transparent&collection=emoji_kitchen_v6&client_key=emoji_kitchen_funbox&q=%s", url.QueryEscape(query))
	backoff := 100 * time.Millisecond
	maxRetries := 5

	for attempt := 1; attempt <= maxRetries; attempt++ {
		response, err := client.Get(url)
		if err == nil {
			defer response.Body.Close()
			switch response.StatusCode {
			case http.StatusOK:
				var tr tenorResponse
				if err := json.NewDecoder(response.Body).Decode(&tr); err == nil {
					return &tr, nil
				}
			case http.StatusNotFound:
				return nil, nil
			default:
				response.Body.Close()
			}
		}
		if attempt < maxRetries {
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	return nil, fmt.Errorf("failed to query tenor after %d attempts", maxRetries)
}

func getAllEmojis(client *http.Client) ([]string, error) {
	var response *http.Response
	var err error
	response, err = client.Get("https://unicode.org/Public/emoji/latest/emoji-test.txt")
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
		emojis = append(emojis, strings.Join(subparts, "-"))
	}
	if err := scanner.Err();
	err != nil {
		return nil, err
	}
	return emojis, nil
}

func getValidEmojis(client *http.Client, emojis []string) ([]string, map[string]string, error) {
	canonicalSet := make(map[string]string)
	aliases := make(map[string]string)

	WORKERS_COUNT := 50

	totalJobs := len(emojis)
	var processedCount atomic.Int64
	var validCount atomic.Int64

	var mu sync.Mutex
	jobs := make(chan string, totalJobs)
	var wg sync.WaitGroup

	for range WORKERS_COUNT {
			wg.Go(func() {
				for cp := range jobs {
					emoji := codepointToEmoji(cp)

					res, err := queryTenor(client, emoji)

					currCount := processedCount.Add(1)
					if currCount % 10 == 0 || int(currCount) == totalJobs {
						fmt.Printf("\r[ValidateEmojis] Processed: %d/%d (%.1f%%) | Valid: %d",
							currCount, totalJobs, float64(currCount) / float64(totalJobs) * 100, validCount.Load())
					}

					if err != nil || res == nil {
						continue
					}

					tags1, tags2 := res.Results[0].Tags, res.Results[1].Tags
					canonical := getCanonicalEmoji(tags1, tags2)
					if canonical == "" {
						continue
					}

					canonical_cp := emojiToCodePoint(canonical)
					validCount.Add(1)

					mu.Lock()
					canonicalSet[canonical_cp] = cp
					if cp != canonical_cp {
						aliases[cp] = canonical
					}
					mu.Unlock()
				}
			})
		}

	for _, emoji_codepoint := range emojis {
		jobs <- emoji_codepoint
	}
	close(jobs)
	wg.Wait()
	fmt.Printf("\r[ValidateEmojis] Total: %d | Valid: %d | Canonicals: %d | Aliases: %d\033[K\n",
			totalJobs, validCount.Load(), len(canonicalSet), len(aliases))

	canonicals := make([]string, 0, len(canonicalSet))
	for cp := range canonicalSet {
		canonicals = append(canonicals, cp)
	}

	sort.Strings(canonicals)
	return canonicals, aliases, nil
}
