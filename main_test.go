package main

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// 一時ディレクトリに小さなツリーを作って scanTree() を呼び、CSV 出力を検証する。
func TestScanTreeBasic(t *testing.T) {
	root := t.TempDir()

	mustWrite(t, filepath.Join(root, "a.txt"), "hello")
	mustWrite(t, filepath.Join(root, "b.LOG"), "log line")
	mustWrite(t, filepath.Join(root, "noext"), "x")
	mustMkdir(t, filepath.Join(root, "sub"))
	mustWrite(t, filepath.Join(root, "sub", "c.csv"), "1,2,3")
	mustMkdir(t, filepath.Join(root, "skipme"))
	mustWrite(t, filepath.Join(root, "skipme", "ignored.txt"), "should be skipped")

	cfg := config{
		excludes:      []string{normalizePattern(`/skipme/`)},
		minSize:       0,
		followSymlink: false,
	}

	var buf bytes.Buffer
	cw := csv.NewWriter(&buf)
	if err := cw.Write(csvHeader); err != nil {
		t.Fatal(err)
	}

	st := &stats{}
	last := time.Now()
	scanTree(cfg, root, cw, st, &last, time.Now())
	cw.Flush()
	if err := cw.Error(); err != nil {
		t.Fatal(err)
	}

	if st.files != 4 {
		t.Errorf("files=%d, want 4 (a.txt,b.LOG,noext,sub/c.csv)", st.files)
	}

	out := buf.String()
	for _, want := range []string{"a.txt", "b.LOG", "noext", "c.csv"} {
		if !strings.Contains(out, want) {
			t.Errorf("CSV に %q が含まれていない", want)
		}
	}
	if strings.Contains(out, "ignored.txt") {
		t.Error("除外したはずの ignored.txt が出力されている")
	}

	// 種類列が拡張子小文字なこと
	if !strings.Contains(out, ",log,") {
		t.Error("種類列が小文字化されていない (LOG -> log)")
	}
	if !strings.Contains(out, ",(なし),") {
		t.Error("拡張子なしが (なし) になっていない")
	}
}

func TestScanTreeMinSize(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "small.txt"), "x")        // 1 byte
	mustWrite(t, filepath.Join(root, "big.txt"), strings.Repeat("y", 1024)) // 1024 bytes

	cfg := config{minSize: 100}
	var buf bytes.Buffer
	cw := csv.NewWriter(&buf)
	cw.Write(csvHeader)

	st := &stats{}
	last := time.Now()
	scanTree(cfg, root, cw, st, &last, time.Now())
	cw.Flush()

	if st.files != 1 {
		t.Errorf("files=%d, want 1", st.files)
	}
	if !strings.Contains(buf.String(), "big.txt") {
		t.Error("big.txt が出力されていない")
	}
	if strings.Contains(buf.String(), "small.txt") {
		t.Error("min-size 未満の small.txt が出力されている")
	}
}

func TestFileKind(t *testing.T) {
	cases := map[string]string{
		"a.txt":     "txt",
		"a.TXT":     "txt",
		"a.tar.gz":  "gz",
		"noext":     "(なし)",
		".hidden":   "hidden",
	}
	for in, want := range cases {
		if got := fileKind(in); got != want {
			t.Errorf("fileKind(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestMatchesAny(t *testing.T) {
	pats := []string{
		normalizePattern(`\node_modules\`),
		normalizePattern(`\.git\objects\`),
	}
	// Windows 形式パスでもマッチすること
	if !matchesAny(`C:\proj\node_modules\foo.js`, pats) {
		t.Error("Windows パス: node_modules がマッチしない")
	}
	// Unix 形式パスでもマッチすること
	if !matchesAny(`/home/me/proj/node_modules/foo.js`, pats) {
		t.Error("Unix パス: node_modules がマッチしない")
	}
	// 大小文字無視
	if !matchesAny(`C:\Proj\Node_Modules\foo.js`, pats) {
		t.Error("大文字混在でマッチしない")
	}
	// 対象外
	if matchesAny(`C:\proj\src\foo.js`, pats) {
		t.Error("対象外がマッチしてしまった")
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
}
