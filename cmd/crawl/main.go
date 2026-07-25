package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
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
	Left      string
	Right     string
	DateIndex int
}

func (p PairTuple) MarshalJSON() ([]byte, error) {
	return json.Marshal(&[3]any{p.Left, p.Right, p.DateIndex})
}

type intermediatePair struct {
	Left       string
	Right      string
	DateString string
}

func main() {
	start := time.Now()
	defer func() {
		fmt.Printf("Total duration: %s\n", time.Since(start))
	}()
	client := &http.Client{
		Timeout: 8 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        1000,
			MaxConnsPerHost:     1000,
			MaxIdleConnsPerHost: 2000,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
			ForceAttemptHTTP2:   true,
		},
	}

	res, err := queryTenor(client, "👩‍👧‍👦")
	if err != nil {
		fmt.Println(err)
		return
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
		return
	}
	fmt.Println(parsed)
	a := emojiToCodePoint("👩‍👧‍👦")
	fmt.Println(codepointToEmoji(a))
	emojis, err := getAllEmojis(client)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(emojis))
	validCanEmojis, validAliases, err := getValidEmojis(client, emojis)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(validCanEmojis), len(validAliases))
	validCombinations, sortedDates, err := getValidCombinations(client, validCanEmojis)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(validCombinations), sortedDates)
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

func parseStickerURL(rawURL string) (intermediatePair, error) {
	_, path, ok := strings.Cut(rawURL, "/emojikitchen/")
	if !ok {
		return intermediatePair{}, fmt.Errorf("invalid sticker URL: %s", rawURL)
	}

	parts := strings.Split(path, "/")
	filename := strings.TrimSuffix(parts[2], ".png")
	left, right, _ := strings.Cut(filename, "_")

	return intermediatePair{
		DateString: parts[0],
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
	reqUrl := fmt.Sprintf("https://tenor.googleapis.com/v2/featured?key=AIzaSyACvEq5cnT7AcHpDdj64SE3TJZRhW-iHuo&limit=2&media_filter=png_transparent&collection=emoji_kitchen_v6&client_key=emoji_kitchen_funbox&q=%s", url.QueryEscape(query))
	backoff := 100 * time.Millisecond
	maxRetries := 5

	for attempt := 1; attempt <= maxRetries; attempt++ {
		response, err := client.Get(reqUrl)
		if err == nil {
			switch response.StatusCode {
			case http.StatusOK:
				var tr tenorResponse
				decErr := json.NewDecoder(response.Body).Decode(&tr)
				response.Body.Close()
				if decErr == nil {
					return &tr, nil
				}
			case http.StatusBadRequest:
				response.Body.Close()
				return nil, fmt.Errorf("bad request: %s", reqUrl)
			case http.StatusNotFound:
				response.Body.Close()
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
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return emojis, nil
}

func getValidEmojis(client *http.Client, emojis []string) ([]string, map[string]string, error) {
	canonicalSet := make(map[string]string)
	aliases := make(map[string]string)

	workersCount := 500

	totalJobs := len(emojis)
	var processedCount atomic.Int64
	var validCount atomic.Int64

	var mu sync.Mutex
	jobs := make(chan string, totalJobs)
	var wg sync.WaitGroup

	for range workersCount {
		wg.Go(func() {
			for cp := range jobs {
				emoji := codepointToEmoji(cp)

				res, err := queryTenor(client, emoji)

				currCount := processedCount.Add(1)
				if currCount%10 == 0 || int(currCount) == totalJobs {
					fmt.Printf("\r[ValidateEmojis] Processed: %d/%d (%.1f%%) | Valid: %d",
						currCount, totalJobs, float64(currCount)/float64(totalJobs)*100, validCount.Load())
				}

				if err != nil || res == nil {
					continue
				}

				if len(res.Results) < 2 {
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

	for _, emoji_cp := range emojis {
		jobs <- emoji_cp
	}
	close(jobs)
	wg.Wait()
	fmt.Printf("\r[ValidateEmojis] Total: %d | Valid: %d | Canonicals: %d | Aliases: %d\033[K\n",
		totalJobs, validCount.Load(), len(canonicalSet), len(aliases))

	canonicals := make([]string, 0, len(canonicalSet))
	for cp := range canonicalSet {
		canonicals = append(canonicals, cp)
	}

	slices.Sort(canonicals)
	return canonicals, aliases, nil
}

func getValidCombinations(client *http.Client, canonicals []string) ([]intermediatePair, []string, error) {
	n := len(canonicals)
	totalJobs := int64(n * (n + 1) / 2)
	if totalJobs == 0 {
		return nil, nil, nil
	}

	type job struct{ i, j int }
	workersCount := 1000

	jobs := make(chan job, totalJobs)
	for i := range n {
		for j := range i + 1 {
			jobs <- job{i, j}
		}
	}
	close(jobs)
	var (
		validCombinations []intermediatePair
		dateSet           = make(map[string]struct{})
		mu                sync.Mutex
		processedCount    atomic.Int64
		validCount        atomic.Int64
		wg                sync.WaitGroup
	)

	for range workersCount {
		wg.Go(func() {
			var localCombos []intermediatePair
			localDates := make(map[string]struct{})
			for job := range jobs {
				leftEmoji := codepointToEmoji(canonicals[job.i])
				rightEmoji := codepointToEmoji(canonicals[job.j])

				query := fmt.Sprintf("%s_%s", leftEmoji, rightEmoji)

				res, err := queryTenor(client, (query))

				currCount := processedCount.Add(1)
				if currCount%100 == 0 || currCount == totalJobs {
					fmt.Printf("\r[ValidateCombinations] Processed: %d/%d (%.1f%%) | Valid pairs: %d", currCount, totalJobs, float64(currCount)/float64(totalJobs)*100, validCount.Load())
				}

				if err != nil || res == nil {
					continue
				}

				if len(res.Results) < 1 {
					continue
				}

				stickerURL := res.Results[0].MediaFormats.PNGTransparent.URL
				combo, err := parseStickerURL(stickerURL)
				if err != nil {
					fmt.Println(err)
					continue
				}
				validCount.Add(1)
				localCombos = append(localCombos, combo)
				localDates[combo.DateString] = struct{}{}
			}
			mu.Lock()
			validCombinations = append(validCombinations, localCombos...)
			for d := range localDates {
				dateSet[d] = struct{}{}
			}
			mu.Unlock()

		})
	}
	wg.Wait()
	fmt.Printf("\r[ValidateCombinations] Total: %d | Valid: %d\033[K\n",
		totalJobs, validCount.Load())

	sortedDates := make([]string, 0, len(dateSet))
	for date := range dateSet {
		sortedDates = append(sortedDates, date)
	}
	slices.Sort(sortedDates)
	return validCombinations, sortedDates, nil
}
