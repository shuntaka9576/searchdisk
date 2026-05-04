package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// パターンはすべて小文字 + スラッシュ区切り。matchesAny がパスを同じ形に正規化して比較する。
var defaultExcludes = []string{
	"/windows/",
	"/program files/",
	"/program files (x86)/",
	"/programdata/",
	"/perflogs/",
	"/$recycle.bin/",
	"/system volume information/",
	"/$winreagent/",
	"/recovery/",
	"/config.msi/",
	"/appdata/local/temp/",
	"/appdata/local/microsoft/windows/inetcache/",
	"/appdata/local/microsoft/windows/webcache/",
	"/appdata/local/packages/",
	"/hiberfil.sys",
	"/pagefile.sys",
	"/swapfile.sys",
	"/node_modules/",
	"/.git/objects/",
}

var csvHeader = []string{
	"ディスク", "ファイルネーム", "種類", "作成日時", "更新日時", "サイズ", "フルパス",
}

type config struct {
	roots         []string
	excludes      []string
	minSize       int64
	followSymlink bool
	out           string
	bom           bool
	progressEvery int64
}

type stats struct {
	files   int64
	bytes   int64
	skipDir int64
}

func main() {
	var (
		pathFlag     = flag.String("path", "", "対象パス (カンマ区切り)。空なら固定ディスクを自動検出")
		outFlag      = flag.String("out", "", "出力 CSV パス。空ならカレントに filelist_<日時>.csv")
		excludeFlag  = flag.String("exclude", "", "追加の除外パス (カンマ区切り、部分一致、大小文字無視)")
		minSizeFlag  = flag.Int64("min-size", 0, "この値未満のファイルは除外 (バイト)")
		followFlag   = flag.Bool("follow-symlink", false, "シンボリックリンク/ジャンクションを辿る")
		bomFlag      = flag.Bool("bom", true, "UTF-8 BOM を付ける (Excel 互換)")
		useDefaultEx = flag.Bool("default-exclude", true, "Windows 既定除外を有効化")
	)
	flag.Parse()

	cfg := config{
		minSize:       *minSizeFlag,
		followSymlink: *followFlag,
		bom:           *bomFlag,
		progressEvery: 10000,
	}

	if *pathFlag != "" {
		for _, p := range strings.Split(*pathFlag, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				cfg.roots = append(cfg.roots, p)
			}
		}
	} else {
		cfg.roots = listFixedDrives()
	}
	if len(cfg.roots) == 0 {
		die("対象パスがありません。-path で指定してください")
	}

	if *useDefaultEx {
		cfg.excludes = append(cfg.excludes, defaultExcludes...)
	}
	if *excludeFlag != "" {
		for _, e := range strings.Split(*excludeFlag, ",") {
			e = strings.TrimSpace(e)
			if e != "" {
				cfg.excludes = append(cfg.excludes, normalizePattern(e))
			}
		}
	}

	cfg.out = *outFlag
	if cfg.out == "" {
		cfg.out = fmt.Sprintf("filelist_%s.csv", time.Now().Format("20060102_150405"))
	}

	if err := run(cfg); err != nil {
		die(err.Error())
	}
}

func die(s string) {
	fmt.Fprintln(os.Stderr, "error:", s)
	os.Exit(1)
}

func run(cfg config) error {
	f, err := os.Create(cfg.out)
	if err != nil {
		return fmt.Errorf("出力ファイル作成失敗: %w", err)
	}
	defer f.Close()

	bw := bufio.NewWriterSize(f, 1<<20)
	defer bw.Flush()

	if cfg.bom {
		if _, err := bw.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
			return err
		}
	}

	cw := csv.NewWriter(bw)
	cw.UseCRLF = true

	if err := cw.Write(csvHeader); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "対象 : %s\n出力 : %s\n除外 : %d パターン\n\n",
		strings.Join(cfg.roots, ", "), cfg.out, len(cfg.excludes))

	st := &stats{}
	start := time.Now()
	last := start

	for _, root := range cfg.roots {
		scanTree(cfg, root, cw, st, &last, start)
	}

	cw.Flush()
	if err := cw.Error(); err != nil {
		return err
	}
	if err := bw.Flush(); err != nil {
		return err
	}

	elapsed := time.Since(start).Truncate(time.Second)
	fmt.Fprintf(os.Stderr, "\n完了: %d 件 / %.2f GB / 経過 %s\n除外dir: %d\n出力: %s\n",
		st.files, float64(st.bytes)/(1<<30), elapsed, st.skipDir, cfg.out)
	return nil
}

// scanTree はルートを再帰スキャンして CSV にストリーミング書き込みする。
func scanTree(cfg config, root string, cw *csv.Writer, st *stats, last *time.Time, start time.Time) {
	abs, err := filepath.Abs(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "パス解決失敗: %s: %v\n", root, err)
		return
	}
	info, err := os.Lstat(abs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "パスが開けません: %s: %v\n", abs, err)
		return
	}
	if !info.IsDir() {
		return
	}

	drive := filepath.VolumeName(abs)
	if drive != "" {
		drive += string(filepath.Separator)
	}

	stack := []string{abs}
	for len(stack) > 0 {
		dir := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if matchesAny(dir, cfg.excludes) {
			st.skipDir++
			continue
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			// 権限不足等は静かにスキップ
			continue
		}

		for _, e := range entries {
			full := filepath.Join(dir, e.Name())
			if matchesAny(full, cfg.excludes) {
				continue
			}

			fi, err := e.Info()
			if err != nil {
				continue
			}

			if e.IsDir() {
				if !cfg.followSymlink && (fi.Mode()&fs.ModeSymlink != 0 || isReparse(fi)) {
					continue
				}
				stack = append(stack, full)
				continue
			}

			if fi.Size() < cfg.minSize {
				continue
			}

			if err := cw.Write([]string{
				drive,
				fi.Name(),
				fileKind(fi.Name()),
				creationTime(fi).Format("2006-01-02 15:04:05"),
				fi.ModTime().Format("2006-01-02 15:04:05"),
				fmt.Sprintf("%d", fi.Size()),
				full,
			}); err != nil {
				fmt.Fprintf(os.Stderr, "書き込み失敗: %v\n", err)
				return
			}

			st.files++
			st.bytes += fi.Size()

			if cfg.progressEvery > 0 && st.files%cfg.progressEvery == 0 {
				now := time.Now()
				if now.Sub(*last) >= time.Second {
					elapsed := now.Sub(start).Seconds()
					rate := 0.0
					if elapsed > 0 {
						rate = float64(st.files) / elapsed
					}
					fmt.Fprintf(os.Stderr, "\r%d 件 / %.2f GB / %.0f 件/s   ",
						st.files, float64(st.bytes)/(1<<30), rate)
					*last = now
				}
			}
		}
	}
}

// normalizePattern はパターンを「小文字 + 全部 '/' 区切り」に正規化する。
// filepath.ToSlash は Windows 以外では '\' を変換しないので明示的に置換する。
func normalizePattern(p string) string {
	return strings.ToLower(strings.ReplaceAll(p, `\`, "/"))
}

// matchesAny はパスを正規化して、いずれかのパターンに部分一致するか判定する。
func matchesAny(path string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}
	p := normalizePattern(path)
	for _, pat := range patterns {
		if strings.Contains(p, pat) {
			return true
		}
	}
	return false
}

// fileKind は拡張子を小文字で返す (ドットなし)。拡張子なしは "(なし)"。
func fileKind(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return "(なし)"
	}
	return strings.ToLower(strings.TrimPrefix(ext, "."))
}
