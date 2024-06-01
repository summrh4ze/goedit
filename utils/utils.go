package utils

import "bytes"

func IsDelimiter(b byte) bool {
	delimiters := []byte(" `~!@#$%^&*()-=+[{]}\\|;:'\",.<>/?\t")
	for _, c := range delimiters {
		if b == c {
			return true
		}
	}
	return false
}

func Tlen(str string, tabsize int) int {
	tlen := 0
	if str == "" {
		return tlen
	}

	nonTabs := 0
	if str[0] != '\t' {
		nonTabs = 1
		tlen += 1
	} else {
		tlen += tabsize
	}
	prevChar := str[0]
	for i := 1; i < len(str); i++ {
		if str[i] == '\t' && prevChar == '\t' {
			tlen += tabsize
			prevChar = str[i]
		} else if str[i] == '\t' && prevChar != '\t' {
			tlen += tabsize - nonTabs%tabsize
			nonTabs = 0
		} else if str[i] != '\t' {
			tlen += 1
			nonTabs += 1
		}
	}
	return tlen
}

func Texp(str string, tabsize int) string {
	res := ""
	if str == "" {
		return res
	}

	nonTabs := 0
	if str[0] != '\t' {
		nonTabs = 1
		res += string(str[0])
	} else {
		res += string(bytes.Repeat([]byte(" "), tabsize))
	}
	prevChar := str[0]
	for i := 1; i < len(str); i++ {
		if str[i] == '\t' && prevChar == '\t' {
			res += string(bytes.Repeat([]byte(" "), tabsize))
			prevChar = str[i]
		} else if str[i] == '\t' && prevChar != '\t' {
			res += string(bytes.Repeat([]byte(" "), tabsize-nonTabs%tabsize))
			nonTabs = 0
		} else if str[i] != '\t' {
			res += string(str[i])
			nonTabs += 1
		}
	}
	return res
}

func IsWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n'
}
