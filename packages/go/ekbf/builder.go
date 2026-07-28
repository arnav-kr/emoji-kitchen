package ekbf

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"slices"
	"strconv"
)

type Pair struct {
	Left      string
	Right     string
	DateIndex int
}

type Builder struct {
	Version    uint16
	Canonicals []string
	Aliases    map[string]string
	Dates      []string
	Pairs      []Pair
}

func (b *Builder) Encode(w io.Writer) error {
	if !slices.Contains(Versions, b.Version) {
		return fmt.Errorf("unsupported version: %d", b.Version)
	}

	if b.Version == 0 {
		b.Version = uint16(CurrentVersion)
	}

	if len(b.Dates) > MaxDate {
		return fmt.Errorf("too many dates: %d > %d", len(b.Dates), MaxDate)
	}

	canonicals := append([]string(nil), b.Canonicals...)
	slices.Sort(canonicals)

	canonicalMap := make(map[string]uint32, len(canonicals))
	for i, c := range canonicals {
		canonicalMap[c] = uint32(i)
	}

	alises := make([]string, 0, len(b.Aliases))
	for k := range b.Aliases {
		alises = append(alises, k)
	}
	slices.Sort(alises)

	var stringTableBuf bytes.Buffer

	canonicalRecords := make([]uint32, len(canonicals))
	for i, s := range canonicals {
		canonicalRecords[i] = writeToStringTable(&stringTableBuf, s)
	}

	aliasRecords := make([]byte, len(alises)*AliasRecordSize)
	for i, aliasValue := range alises {
		targetValue := b.Aliases[aliasValue]
		if _, ok := canonicalMap[targetValue]; !ok {
			return fmt.Errorf("alias %q points at %q, which is not in Canonicals", aliasValue, targetValue)
		}

		aliasOffset := writeToStringTable(&stringTableBuf, aliasValue)
		targetOffset := writeToStringTable(&stringTableBuf, targetValue)

		recordStart := i * AliasRecordSize
		binary.LittleEndian.PutUint32(aliasRecords[recordStart:], aliasOffset)
		binary.LittleEndian.PutUint32(aliasRecords[recordStart+4:], targetOffset)
	}

	dates := make([]uint32, len(b.Dates))
	for i, date := range b.Dates {
		value, err := parseDate(date)
		if err != nil {
			return err
		}
		dates[i] = value
	}

	matrix := make([]byte, MatrixSize(len(canonicals))*MatrixValueSize)
	for _, pair := range b.Pairs {
		if err := writePair(matrix, canonicalMap, pair); err != nil {
			return err
		}
	}

	header := Header{
		Version:        b.Version,
		CanonicalCount: uint16(len(canonicals)),
		AliasCount:     uint16(len(b.Aliases)),
		DateCount:      uint16(len(b.Dates)),
	}

	return writeSections(w, header, canonicalRecords, aliasRecords, dates, matrix, stringTableBuf.Bytes())
}

var zeroPadding [StringTableAlignment]byte

func writeSections(w io.Writer, header Header, canonicalRecords []uint32, aliasRecords []byte, dates []uint32, matrix []byte, stringTable []byte) error {
	bw, ok := w.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(w)
		defer bw.Flush()
	}

	var headerBuf [HeaderSize]byte
	header.WriteTo(headerBuf[:])
	if _, err := bw.Write(headerBuf[:]); err != nil {
		return err
	}

	if err := writeUint32Slice(bw, canonicalRecords); err != nil {
		return err
	}

	if _, err := bw.Write(aliasRecords); err != nil {
		return err
	}

	if err := writeUint32Slice(bw, dates); err != nil {
		return err
	}

	if _, err := bw.Write(matrix); err != nil {
		return err
	}

	written := HeaderSize
	written += len(canonicalRecords) * CanonicalRecordSize
	written += len(aliasRecords)
	written += len(dates) * DateSize
	written += len(matrix)

	if padding := byteAlign(written, StringTableAlignment) - written; padding > 0 {
		if _, err := bw.Write(zeroPadding[:padding]); err != nil {
			return err
		}
	}

	if _, err := bw.Write(stringTable); err != nil {
		return err
	}

	return nil
}

func writeUint32Slice(w io.Writer, data []uint32) error {
	var scratch [4]byte
	for _, value := range data {
		binary.LittleEndian.PutUint32(scratch[:], value)
		if _, err := w.Write(scratch[:]); err != nil {
			return err
		}
	}
	return nil
}

func parseDate(s string) (uint32, error) {
	if len(s) != 8 {
		return 0, fmt.Errorf("invalid Date, should be YYYYMMDD")
	}
	value, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(value), nil
}

func writePair(matrix []byte, canonicalMap map[string]uint32, pair Pair) error {
	left, ok := canonicalMap[pair.Left]
	if !ok {
		return fmt.Errorf("left emoji not found: %s", pair.Left)
	}
	right, ok := canonicalMap[pair.Right]
	if !ok {
		return fmt.Errorf("right emoji not found: %s", pair.Right)
	}

	row, col := left, right

	isSwapped := false
	if row < col {
		row, col = col, row
		isSwapped = true
	}

	value := int16(pair.DateIndex + 1)
	if isSwapped {
		value = -value
	}

	pos := MatrixPosition(row, col)
	byteOffset := pos * MatrixValueSize
	existing := int16(binary.LittleEndian.Uint16(matrix[byteOffset:]))
	if existing != 0 && existing != value {
		return fmt.Errorf("conflicting value at position %d: %d vs %d", pos, existing, value)
	}
	binary.LittleEndian.PutUint16(matrix[byteOffset:], uint16(value))
	return nil
}

func readFromStringTable(table []byte, offset uint32) string {
	if int(offset) >= len(table) {
		return ""
	}
	idx := bytes.IndexByte(table[offset:], 0x00)
	if idx == -1 {
		return string(table[offset:])
	}
	return string(table[offset : int(offset)+idx])
}

func writeToStringTable(buf *bytes.Buffer, s string) uint32 {
	offset := uint32(buf.Len())
	buf.WriteString(s)
	buf.WriteByte(0x00)
	return offset
}
