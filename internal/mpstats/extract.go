package mpstats

import (
	"fmt"
	"strings"
)

func norm(s string) string {
	return strings.TrimSpace(strings.ToLower(strings.ReplaceAll(s, "ё", "е")))
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

func ExtractByMapping(fp *FullPage, mapping map[string]string) map[string]string {
	out := map[string]string{
		"Название товара": fp.FullName,
		"Описание":        strings.TrimSpace(firstNonEmpty(fp.Description, fp.FullText)),
	}
	for i, name := range fp.ParamNames {
		val := ""
		if i < len(fp.ParamValues) {
			val = fmt.Sprint(fp.ParamValues[i])
		}
		for source, target := range mapping {
			if strings.Contains(norm(name), norm(source)) {
				if strings.TrimSpace(val) != "" {
					out[target] = strings.TrimSpace(val)
				}
			}
		}
	}
	return out
}
