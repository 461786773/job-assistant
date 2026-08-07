package resume

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

type Result struct {
	Text   string `json:"text"`
	Format string `json:"format"`
}

func Parse(filename string, r io.Reader) (*Result, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".md", ".markdown", ".txt":
		text := string(data)
		if !utf8.ValidString(text) {
			return nil, fmt.Errorf("文本编码无效，请使用 UTF-8")
		}
		format := "txt"
		if ext == ".md" || ext == ".markdown" {
			format = "md"
		}
		return &Result{Text: strings.TrimSpace(text), Format: format}, nil
	case ".docx":
		text, err := parseDocx(data)
		if err != nil {
			return nil, err
		}
		return &Result{Text: strings.TrimSpace(text), Format: "docx"}, nil
	case ".pdf":
		text, err := parsePDF(data)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(text) == "" {
			return nil, fmt.Errorf("未能从 PDF 提取文本（可能是扫描件）。请改用可选中文本的 PDF，或粘贴 Markdown/纯文本")
		}
		return &Result{Text: strings.TrimSpace(text), Format: "pdf"}, nil
	default:
		return nil, fmt.Errorf("不支持的格式 %s，请上传 md / txt / docx / pdf", ext)
	}
}

func parseDocx(data []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("解析 docx 失败: %w", err)
	}
	var docFile *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			docFile = f
			break
		}
	}
	if docFile == nil {
		return "", fmt.Errorf("docx 中缺少 word/document.xml")
	}
	rc, err := docFile.Open()
	if err != nil {
		return "", fmt.Errorf("打开 docx 内容失败: %w", err)
	}
	defer rc.Close()

	decoder := xml.NewDecoder(rc)
	var b strings.Builder
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("解析 docx XML 失败: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "tab" {
				b.WriteByte('\t')
			}
		case xml.EndElement:
			if t.Name.Local == "p" || t.Name.Local == "br" {
				b.WriteByte('\n')
			}
		case xml.CharData:
			b.Write(t)
		}
	}
	return collapseBlankLines(b.String()), nil
}

// parsePDF extracts visible text from text-based PDFs via common literal string patterns.
// Scanned/image PDFs will yield empty text and should fall back to paste.
func parsePDF(data []byte) (string, error) {
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		return "", fmt.Errorf("不是有效的 PDF 文件")
	}
	// Prefer uncompressed literal strings: (text) Tj / TJ
	reLiteral := regexp.MustCompile(`\((?:\\.|[^\\()])*\)\s*Tj`)
	reArray := regexp.MustCompile(`\[(.*?)\]\s*TJ`)
	reInside := regexp.MustCompile(`\((?:\\.|[^\\()])*\)`)

	var parts []string
	for _, m := range reLiteral.FindAll(data, -1) {
		s := string(m)
		s = strings.TrimSpace(strings.TrimSuffix(s, "Tj"))
		parts = append(parts, unescapePDFString(s))
	}
	for _, m := range reArray.FindAllSubmatch(data, -1) {
		var line []string
		for _, piece := range reInside.FindAll(m[1], -1) {
			line = append(line, unescapePDFString(string(piece)))
		}
		if len(line) > 0 {
			parts = append(parts, strings.Join(line, ""))
		}
	}
	text := strings.Join(parts, "\n")
	text = strings.ReplaceAll(text, "\r", "")
	return collapseBlankLines(text), nil
}

func unescapePDFString(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '(' && s[len(s)-1] == ')' {
		s = s[1 : len(s)-1]
	}
	replacer := strings.NewReplacer(
		`\n`, "\n",
		`\r`, "\r",
		`\t`, "\t",
		`\(`, "(",
		`\)`, ")",
		`\\`, `\`,
	)
	return replacer.Replace(s)
}

func collapseBlankLines(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	prevBlank := false
	for _, line := range lines {
		line = strings.TrimRight(line, " \t")
		if strings.TrimSpace(line) == "" {
			if prevBlank {
				continue
			}
			prevBlank = true
			out = append(out, "")
			continue
		}
		prevBlank = false
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
