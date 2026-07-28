package ekbf

import (
	"encoding/binary"
	"fmt"
	"slices"
)

var Magic = [4]byte{'E', 'K', 'B', 'F'}

var Versions = []uint16{1}
var CurrentVersion uint16 = 1

// size in bytes
const (
	HeaderSize           = 12
	CanonicalRecordSize  = 4
	AliasRecordSize      = 8
	DateSize             = 4
	MatrixValueSize      = 2
	StringTableAlignment = 4
)

const MaxDate = 0x7FFF

type Header struct {
	Version        uint16
	CanonicalCount uint16
	AliasCount     uint16
	DateCount      uint16
}

type Layout struct {
	CanonicalOffsetStart,
	AliasRecordsStart,
	DateArrayStart,
	MatrixStart,
	StringTableStart int
}

func ParseHeader(data []byte) (*Header, error) {
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("file only %d bytes, too short for a %d byte EKBF header", len(data), HeaderSize)
	}

	var magic [4]byte
	copy(magic[:], data[:4])
	if magic != Magic {
		return nil, fmt.Errorf("invalid magic string, got %v, expected %v", magic, Magic)
	}

	version := binary.LittleEndian.Uint16(data[4:6])
	if !slices.Contains(Versions, version) {
		return nil, fmt.Errorf("unsupported version, got %d, expected one of %v", version, Versions)
	}

	header := &Header{
		Version:        version,
		CanonicalCount: binary.LittleEndian.Uint16(data[6:8]),
		AliasCount:     binary.LittleEndian.Uint16(data[8:10]),
		DateCount:      binary.LittleEndian.Uint16(data[10:12]),
	}
	return header, nil
}

func (h *Header) WriteTo(buf []byte) {
	copy(buf[0:4], Magic[:])
	binary.LittleEndian.PutUint16(buf[4:6], h.Version)
	binary.LittleEndian.PutUint16(buf[6:8], h.CanonicalCount)
	binary.LittleEndian.PutUint16(buf[8:10], h.AliasCount)
	binary.LittleEndian.PutUint16(buf[10:12], h.DateCount)
}

func computeLayout(h *Header) Layout {
	canonicalOffsetStart := HeaderSize
	aliasRecordsStart := canonicalOffsetStart + int(h.CanonicalCount)*CanonicalRecordSize
	dateArrayStart := aliasRecordsStart + int(h.AliasCount)*AliasRecordSize
	matrixStart := dateArrayStart + int(h.DateCount)*DateSize
	stringTableStart := byteAlign(matrixStart+MatrixSize(int(h.CanonicalCount))*MatrixValueSize, StringTableAlignment)

	return Layout{
		CanonicalOffsetStart: canonicalOffsetStart,
		AliasRecordsStart:    aliasRecordsStart,
		DateArrayStart:       dateArrayStart,
		MatrixStart:          matrixStart,
		StringTableStart:     stringTableStart,
	}
}

func byteAlign(n, alignment int) int {
	if alignment&(alignment-1) != 0 {
		panic("alignment must be a power of 2")
	}
	return (n + alignment - 1) &^ (alignment - 1)
}

func MatrixSize(canonicalCount int) int {
	n := canonicalCount
	return n * (n + 1) / 2
}

func MatrixPosition(a, b uint32) uint32 {
	row, col := a, b
	if row < col {
		row, col = col, row
	}
	return row*(row+1)/2 + col
}
