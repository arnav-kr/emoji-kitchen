package ekbf

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

type Pair struct {
	Left      string
	Right     string
	DateIndex int
}

type Builder struct {
	Canonicals []string
	Aliases    map[string]string
	Dates      []string
	Pairs      []Pair
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
	
	if _, err := w.Write(matrix); err != nil {
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
