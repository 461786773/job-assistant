package hr

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type ApplyResult struct {
	Index   int    `json:"index"`
	Target  string `json:"target"`
	OK      bool   `json:"ok"`
	Method  string `json:"method,omitempty"` // exact | trim | whitespace
	Message string `json:"message,omitempty"`
}

// ApplyRewrites applies selected rewrite items to resume text in order.
func ApplyRewrites(resume string, items []RewriteItem, indexes []int) (string, []ApplyResult) {
	out := resume
	results := make([]ApplyResult, 0, len(indexes))
	for _, idx := range indexes {
		res := ApplyResult{Index: idx}
		if idx < 0 || idx >= len(items) {
			res.Message = "索引无效"
			results = append(results, res)
			continue
		}
		item := items[idx]
		res.Target = item.Target
		if strings.TrimSpace(item.After) == "" {
			res.Message = "缺少改写后文本"
			results = append(results, res)
			continue
		}
		if strings.TrimSpace(item.Before) == "" {
			res.Message = "缺少原文定位，无法自动替换"
			results = append(results, res)
			continue
		}
		next, ok, method := replaceOnce(out, item.Before, item.After)
		if !ok {
			res.Message = "未在简历中找到对应原文"
			results = append(results, res)
			continue
		}
		out = next
		res.OK = true
		res.Method = method
		results = append(results, res)
	}
	return out, results
}

func replaceOnce(resume, before, after string) (string, bool, string) {
	if before == "" {
		return resume, false, ""
	}
	if i := strings.Index(resume, before); i >= 0 {
		return resume[:i] + after + resume[i+len(before):], true, "exact"
	}
	trimmed := strings.TrimSpace(before)
	if trimmed != "" && trimmed != before {
		if i := strings.Index(resume, trimmed); i >= 0 {
			return resume[:i] + after + resume[i+len(trimmed):], true, "trim"
		}
	}
	if trimmed == "" {
		return resume, false, ""
	}
	return replaceCollapsedWS(resume, trimmed, after)
}

func replaceCollapsedWS(resume, before, after string) (string, bool, string) {
	cResume, mapResume := collapseWS(resume)
	cBefore, _ := collapseWS(before)
	if cBefore == "" {
		return resume, false, ""
	}
	pos := strings.Index(cResume, cBefore)
	if pos < 0 {
		return resume, false, ""
	}
	last := pos + len(cBefore) - 1
	if last >= len(mapResume) || pos >= len(mapResume) {
		return resume, false, ""
	}
	start := mapResume[pos]
	end := mapResume[last] + 1
	if start < 0 || end > len(resume) || start > end {
		return resume, false, ""
	}
	return resume[:start] + after + resume[end:], true, "whitespace"
}

// collapseWS collapses unicode whitespace runs to a single space.
// indexMap[i] is the byte offset in the original string of collapsed[i].
func collapseWS(s string) (string, []int) {
	var b strings.Builder
	indexMap := make([]int, 0, len(s))
	prevSpace := false
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if unicode.IsSpace(r) {
			if !prevSpace && b.Len() > 0 {
				b.WriteByte(' ')
				indexMap = append(indexMap, i)
				prevSpace = true
			}
			i += size
			continue
		}
		prevSpace = false
		b.WriteString(s[i : i+size])
		for j := 0; j < size; j++ {
			indexMap = append(indexMap, i+j)
		}
		i += size
	}
	return b.String(), indexMap
}
