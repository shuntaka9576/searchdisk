//go:build !windows

package main

import (
	"os"
	"time"
)

// listFixedDrives は非 Windows では空。-path で明示してもらう。
func listFixedDrives() []string {
	return nil
}

// creationTime は ModTime をフォールバックとして返す。
func creationTime(fi os.FileInfo) time.Time {
	return fi.ModTime()
}

// isReparse は非 Windows では常に false。
func isReparse(fi os.FileInfo) bool {
	return false
}
