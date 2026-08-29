package emojikitchen

import (
	_ "embed"
	"fmt"
	"iter"
	"strconv"
	"strings"
	"sync"

	"github.com/arnav-kr/emoji-kitchen/packages/go/ekbf"
)

//go:embed data.ekbf
var data []byte

var reader = sync.OnceValue(func() *ekbf.Reader {
	r, err := ekbf.NewReader(data)
	if err != nil {
		panic(fmt.Errorf("emoji-kitchen: embedded data is corrupt: %w", err))
	}
	return r
})

type Combo struct {
	Left  string
	Right string
	Date  string
	raw   ekbf.Pair
}

func (c Combo) URL() string {
	return c.raw.URL()
}

func Mix(left, right string) (Combo, bool) {
	pair, ok := reader().Lookup(ToCodepoint(left), ToCodepoint(right))
	if !ok {
		return Combo{}, false
	}

	return Combo{
		Left:  FromCodepoint(pair.Left),
		Right: FromCodepoint(pair.Right),
		Date:  pair.Date,
		raw:   pair,
	}, true
}

func CanMix(left, right string) bool {
	return reader().Exists(ToCodepoint(left), ToCodepoint(right))
}

func Partners(emoji string) []string {
	hexList, err := reader().PartnersOf(ToCodepoint(emoji))
	if err != nil {
		return nil
	}

	out := make([]string, len(hexList))
	for i, h := range hexList {
		out[i] = FromCodepoint(h)
	}
	return out
}

func AllEmoji() []string {
	canonicals := reader().Canonicals()
	out := make([]string, len(canonicals))
	for i, h := range canonicals {
		out[i] = FromCodepoint(h)
	}
	return out
}

func Combos() iter.Seq[Combo] {
	return func(yield func(Combo) bool) {
		r := reader()
		for pair := range r.Pairs {
			combo := Combo{
				Left:  FromCodepoint(pair.Left),
				Right: FromCodepoint(pair.Right),
				Date:  pair.Date,
				raw:   pair,
			}
			if !yield(combo) {
				return
			}
		}
	}
}

func LastAmended() string {
	return reader().LastAmended()
}

func Stats() ekbf.Stats {
	return reader().Stats()
}

func Reader() *ekbf.Reader {
	return reader()
}

func ToCodepoint(emojiOrCodepoint string) string {
	if isCodepointString(emojiOrCodepoint) {
		return strings.ToLower(emojiOrCodepoint)
	}

	var sb strings.Builder
	first := true
	for _, r := range emojiOrCodepoint {
		if r == 0xFE0F {
			continue
		}
		if !first {
			sb.WriteByte('-')
		}
		sb.WriteString(strconv.FormatInt(int64(r), 16))
		first = false
	}
	return sb.String()
}

func FromCodepoint(hexStrOrEmoji string) string {
	if hexStrOrEmoji == "" {
		return ""
	}
	if !isCodepointString(hexStrOrEmoji) {
		return hexStrOrEmoji
	}

	var sb strings.Builder
	for part := range strings.SplitSeq(hexStrOrEmoji, "-") {
		val, err := strconv.ParseInt(part, 16, 32)
		if err == nil {
			sb.WriteRune(rune(val))
		}
	}
	return sb.String()
}

func isCodepointString(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') || c == '-') {
			return false
		}
	}
	return true
}
