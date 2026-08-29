package emojikitchen

import (
	"testing"
)

func TestToCodepoint(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"😺", "1f63a"},
		{"1f63a", "1f63a"},
		{"1F63A", "1f63a"},
		{"1f600-1f3fb", "1f600-1f3fb"},
		{"", ""},
	}

	for _, tt := range tests {
		actual := ToCodepoint(tt.input)
		if actual != tt.expected {
			t.Errorf("ToCodepoint(%q) = %q; expected %q", tt.input, actual, tt.expected)
		}
	}
}

func TestFromCodepoint(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1f63a", "😺"},
		{"1F63A", "😺"},
		{"😺", "😺"},
		{"", ""},
	}

	for _, tt := range tests {
		actual := FromCodepoint(tt.input)
		if actual != tt.expected {
			t.Errorf("FromCodepoint(%q) = %q; expected %q", tt.input, actual, tt.expected)
		}
	}
}

func TestToAndFromCodepointRoundtrip(t *testing.T) {
	input := "😺"
	cp := ToCodepoint(input)
	if cp != "1f63a" {
		t.Errorf("expected 1f63a, got %q", cp)
	}

	emoji := FromCodepoint(cp)
	if emoji != input {
		t.Errorf("expected %q, got %q", input, emoji)
	}
}
