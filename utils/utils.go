package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func MapKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func GetFileExtension(fileName string) string {
	if len(fileName) == 0 {
		return ""
	}
	parts := strings.Split(fileName, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}

func Contains(slice []string, element string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, strings.TrimSpace(element)) {
			return true
		}
	}
	return false
}

func IntToRoman(num int) string {
	conversions := []struct {
		value  int
		symbol string
	}{
		{1000, "M"},
		{900, "CM"},
		{500, "D"},
		{400, "CD"},
		{100, "C"},
		{90, "XC"},
		{50, "L"},
		{40, "XL"},
		{10, "X"},
		{9, "IX"},
		{5, "V"},
		{4, "IV"},
		{1, "I"},
	}

	var builder strings.Builder
	for _, conv := range conversions {
		for num >= conv.value {
			builder.WriteString(conv.symbol)
			num -= conv.value
		}
	}

	return builder.String()
}

func IndexPathToString(path int64) string {
	if path == 0 {
		return ""
	}

	strs := make([]string, 0, 10)
	// Extract from highest bits (54-59) down to lowest (0-5)
	// Using bits 0-59 to avoid sign bit (bit 63)
	for shift := 54; shift >= 0; shift -= 6 {
		part := (path >> shift) & 0x3F
		if part == 0 {
			break
		}
		strs = append(strs, fmt.Sprintf("%d", part))
	}
	if len(strs) == 0 {
		return ""
	}
	return strings.Join(strs, ".")
}

func StringToIndexPath(s string) (int64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty index path string")
	}
	parts := strings.Split(strings.TrimSpace(s), ".")
	if len(parts) > 10 {
		return 0, fmt.Errorf("index path too deep (max 10 levels)")
	}
	var path int64 = 0
	for i, p := range parts {
		val, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			return 0, err
		}
		if val > 63 || val < 0 {
			return 0, fmt.Errorf("index part %q out of range (0-63)", p)
		}
		// Encode from left to right using bits 54-59, 48-53, 42-47, ...
		// Avoids using sign bit (bit 63)
		path = path | (val << (54 - 6*i))
	}
	return path, nil
}
