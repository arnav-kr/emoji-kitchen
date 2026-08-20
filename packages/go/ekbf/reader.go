package ekbf

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

type Reader struct {
	data       []byte
	header     Header
	layout     Layout
	canonicals []string
}

type Stats struct {
	CanonicalCount int
	AliasCount     int
	DateCount      int
	TotalPairs     int
}

func addPrefix(hex string) string {
	els := strings.Split(hex, "-")
	for i, part := range els {
		els[i] = "u" + part
	}
	return strings.Join(els, "-")
}

func (r Pair) URL() string {
	return fmt.Sprintf("https://www.gstatic.com/android/keyboard/emojikitchen/%s/%s/%s_%s.png",
		r.Date, addPrefix(r.Left), addPrefix(r.Left), addPrefix(r.Right),
	)
}

func NewReader(data []byte) (*Reader, error) {
	header, err := ParseHeader(data)
	if err != nil {
		return nil, err
	}

	f := &Reader{
		data:   data,
		header: *header,
		layout: computeLayout(header),
	}

	f.canonicals = make([]string, f.header.CanonicalCount)
	table := f.data[f.layout.StringTableStart:]
	for i := range f.canonicals {
		recordStart := f.layout.CanonicalOffsetStart + i*CanonicalRecordSize
		offset := binary.LittleEndian.Uint32(f.data[recordStart : recordStart+CanonicalRecordSize])
		f.canonicals[i] = readFromStringTable(table, offset)
	}

	return f, nil
}

func (f *Reader) Header() Header {
	return f.header
}

func (f *Reader) Canonicals() []string {
	return f.canonicals
}

func (f *Reader) Dates() []string {
	dates := make([]string, f.header.DateCount)
	for i := range dates {
		recordStart := f.layout.DateArrayStart + i*DateSize
		value := binary.LittleEndian.Uint32(f.data[recordStart : recordStart+DateSize])
		dates[i] = fmt.Sprintf("%08d", value)
	}
	return dates
}

func (f *Reader) LastAmended() string {
	dates := f.Dates()
	if len(dates) == 0 {
		return ""
	}
	return dates[len(dates)-1]
}

func (f *Reader) Aliases() (map[string]string, error) {
	mapping := make(map[string]string, f.header.AliasCount)
	table := f.data[f.layout.StringTableStart:]
 
	for i := range int(f.header.AliasCount) {
		recordStart := f.layout.AliasRecordsStart + i*AliasRecordSize
		aliasOffset := binary.LittleEndian.Uint32(f.data[recordStart : recordStart+AliasRecordSize/2])
		targetID := binary.LittleEndian.Uint32(f.data[recordStart+AliasRecordSize/2 : recordStart+AliasRecordSize])
 
		if int(targetID) >= len(f.canonicals) {
			return nil, fmt.Errorf("ekbf: alias record %d points at canonical ID %d, but there are only %d canonicals", i, targetID, len(f.canonicals))
		}
 
		mapping[readFromStringTable(table, aliasOffset)] = f.canonicals[targetID]
	}
	return mapping, nil
}

func (f *Reader) Resolve(hex string) (id uint32, found bool) {
	// canonical check
	i := sort.SearchStrings(f.canonicals, hex)
	if i < len(f.canonicals) && f.canonicals[i] == hex {
		return uint32(i), true
	}

	// alias check
	table := f.data[f.layout.StringTableStart:]
	aliasCount := int(f.header.AliasCount)

	aliasStringAt := func(i int) string {
		recordStart := f.layout.AliasRecordsStart + i*AliasRecordSize
		offset := binary.LittleEndian.Uint32(f.data[recordStart : recordStart+AliasRecordSize/2])
		return readFromStringTable(table, offset)
	}

	aliasIdx := sort.Search(aliasCount, func(i int) bool {
		return aliasStringAt(i) >= hex
	})
	if aliasIdx < aliasCount && aliasStringAt(aliasIdx) == hex {
		recordStart := f.layout.AliasRecordsStart + aliasIdx*AliasRecordSize
		targetID := binary.LittleEndian.Uint32(f.data[recordStart+AliasRecordSize/2 : recordStart+AliasRecordSize])
		return uint32(targetID), true
	}

	return 0, false
}

func (f *Reader) IsCanonical(hexCode string) bool {
	i := sort.SearchStrings(f.canonicals, hexCode)
	return i < len(f.canonicals) && f.canonicals[i] == hexCode
}

func (f *Reader) IsAlias(hex string) bool {
	id, found := f.Resolve(hex)
	return found && f.canonicals[id] != hex
}

func (f *Reader) valueAt(a, b uint32) int16 {
	pos := MatrixPosition(a, b)
	byteOffset := f.layout.MatrixStart + int(pos)*MatrixValueSize
	return int16(binary.LittleEndian.Uint16(f.data[byteOffset : byteOffset+MatrixValueSize]))
}

func (f *Reader) Exists(left, right string) bool {
	l, ok := f.Resolve(left)
	if !ok {
		return false
	}
	r, ok := f.Resolve(right)
	if !ok {
		return false
	}
	return f.valueAt(l, r) != 0
}

func (f *Reader) Lookup(left, right string) (Pair, bool) {
	l, ok := f.Resolve(left)
	if !ok {
		return Pair{}, false
	}
	r, ok := f.Resolve(right)
	if !ok {
		return Pair{}, false
	}
	row, col := l, r
	if row < col {
		row, col = col, row
	}
	value := f.valueAt(row, col)
	if value == 0 {
		return Pair{}, false
	}

	dateIndex := abs16(value) - 1
	dateStart := f.layout.DateArrayStart + dateIndex*DateSize
	date := binary.LittleEndian.Uint32(f.data[dateStart : dateStart+DateSize])
	
	rowHex, colHex := f.canonicals[row], f.canonicals[col]
	primary, other := rowHex, colHex
	if value < 0 {
		primary, other = colHex, rowHex
	}
	return Pair{
		Date:  fmt.Sprintf("%08d", date),
		Left:  primary,
		Right: other,
	}, true
}

func abs16(v int16) int {
	wide := int(v)
	if wide < 0 {
		return -wide
	}
	return wide
}
 

func (f *Reader) PartnersOf(hex string) ([]string, error) {
	id, ok := f.Resolve(hex)
	if !ok {
		return nil, fmt.Errorf("ekbf: %q is not a known emoji", hex)
	}

	var partners []string
	for otherID := range f.canonicals {
		if f.valueAt(id, uint32(otherID)) != 0 {
			partners = append(partners, f.canonicals[otherID])
		}
	}
	return partners, nil
}

func (f *Reader) Pairs(yield func(Pair) bool) {
	n := int(f.header.CanonicalCount)
	for i := range n {
		for j := range i + 1 {
			result, ok := f.Lookup(f.canonicals[i], f.canonicals[j])
			if !ok {
				continue
			}
			if !yield(result) {
				return
			}
		}
	}
}

func (f *Reader) Stats() Stats {
	total := 0
	f.Pairs(func(Pair) bool {
		total++
		return true
	})
	return Stats{
		CanonicalCount: int(f.header.CanonicalCount),
		AliasCount:     int(f.header.AliasCount),
		DateCount:      int(f.header.DateCount),
		TotalPairs:     total,
	}
}