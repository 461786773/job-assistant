package resume

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func TestParseMarkdown(t *testing.T) {
	r, err := Parse("a.md", strings.NewReader("# 张三\n产品经理"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Format != "md" || !strings.Contains(r.Text, "张三") {
		t.Fatalf("unexpected: %+v", r)
	}
}

func TestParseDocx(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write([]byte(`<?xml version="1.0"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>简历正文</w:t></w:r></w:p></w:body></w:document>`))
	_ = zw.Close()

	r, err := Parse("cv.docx", bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if r.Format != "docx" || !strings.Contains(r.Text, "简历正文") {
		t.Fatalf("unexpected: %+v", r)
	}
}
