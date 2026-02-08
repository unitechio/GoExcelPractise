package exporter

import "strings"

func beautifyHeader(col string) string {
	return strings.ToUpper(strings.ReplaceAll(col, "_", " "))
}

func softWrapEmail(s string) string {
	s = strings.ReplaceAll(s, "@", "@\u200B")
	s = strings.ReplaceAll(s, ".", ".\u200B")
	return s
}
