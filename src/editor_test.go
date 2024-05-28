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

func TestTexp(t *testing.T) {
	testData := []struct {
		str      string
		tabsize  int
		expected string
	}{
		{"\tHello", 8, "        Hello"},
		{"\t\tGood morning", 8, "                Good morning"},
		{"Hello\tGood morning", 8, "Hello   Good morning"},
		{"Te\t\tst", 8, "Te              st"},
		{"Te\tst is pa\t!", 8, "Te      st is pa        !"},
		{"Te\tst is p\ta!", 8, "Te      st is p a!"},
		{"", 8, ""},
		{"\t", 8, "        "},
		{"\t\t", 8, "                "},
		{"\te", 8, "        e"},
		{"\tes", 8, "        es"},
		{"\test", 8, "        est"},
		{"test with final tab\t", 8, "test with final tab     "},
	}

	for _, data := range testData {
		l := texp(data.str, data.tabsize)
		if l != data.expected {
			t.Errorf("\"%s\": expected str %s, found %s\n", data.str, data.expected, l)
		}
	}
}
