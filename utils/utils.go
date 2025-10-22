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

func IndexPathToString(path []int) string {
	strs := make([]string, len(path))
	for i, v := range path {
		strs[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(strs, ".")
}

func StringToIndexPath(s string) []int {
	parts := strings.Split(strings.TrimSpace(s), ".")
	path := make([]int, len(parts))
	for i, p := range parts {
		path[i], _ = strconv.Atoi(p)
	}
	return path
}
