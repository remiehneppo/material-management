package utils

import "strings"

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
