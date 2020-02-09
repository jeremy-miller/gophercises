package main

import "testing"

func TestNormalize(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"1234567890", "1234567890"},
		{"123 456 7891", "1234567891"},
		{"(123) 456 7892", "1234567892"},
		{"(123) 456-7893", "1234567893"},
		{"123-456-7894", "1234567894"},
		{"(123)456-7892", "1234567892"},
	}
	for _, test := range testCases {
		t.Run(test.input, func(t *testing.T) {
			actual := normalize(test.input)
			if actual != test.expected {
				t.Errorf("expected: %s; actual: %s", test.expected, actual)
			}
		})
	}
}
