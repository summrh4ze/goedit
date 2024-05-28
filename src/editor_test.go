package main

import "testing"

func TestTlen(t *testing.T) {
	testData := []struct {
		str      string
		tabsize  int
		expected int
	}{
		{"\tHello", 8, 13},
		{"\t\tGood morning", 8, 28},
		{"Hello\tGood morning", 8, 20},
		{"Te\t\tst", 8, 18},
		{"Te\tst is pa\t!", 8, 25},
		{"Te\tst is p\ta!", 8, 18},
		{"", 8, 0},
		{"\t", 8, 8},
		{"\t\t", 8, 16},
		{"\te", 8, 9},
		{"\tes", 8, 10},
		{"\test", 8, 11},
		{"test with final tab\t", 8, 24},
	}

	for _, data := range testData {
		l := tlen(data.str, data.tabsize)
		if l != data.expected {
			t.Errorf("\"%s\": expected len %d, found %d\n", data.str, data.expected, l)
		}
	}
}
