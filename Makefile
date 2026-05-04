BINARY      := searchdisk
WIN_BINARY  := $(BINARY).exe
DIST        := dist
PKG         := ./...

VERSION     ?= 0.0.0-dev

.PHONY: all build build-windows installer test test-race vet fmt fmt-check tidy clean help

all: vet test build

## build: 自分の OS 向けにビルド (検証/開発用)
build:
	go build -o $(BINARY) .

## build-windows: Windows amd64 向けにクロスビルド -> dist/searchdisk.exe
build-windows:
	mkdir -p $(DIST)
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o $(DIST)/$(WIN_BINARY) .
	@ls -lh $(DIST)/$(WIN_BINARY)

## installer: Inno Setup でインストーラを生成 (要 Windows + iscc)。VERSION=0.1.0 で上書き可
installer: build-windows
	iscc /DAppVersion=$(VERSION) installer.iss

## test: 単体テスト
test:
	go test -v $(PKG)

## test-race: race detector 付きテスト
test-race:
	go test -race -v $(PKG)

## vet: go vet
vet:
	go vet $(PKG)

## fmt: gofmt 適用
fmt:
	gofmt -s -w .

## fmt-check: gofmt 差分が出ないことを CI で確認したいとき
fmt-check:
	@diff=$$(gofmt -s -l .); \
	if [ -n "$$diff" ]; then \
		echo "gofmt 差分あり:"; echo "$$diff"; exit 1; \
	fi

## tidy: go.mod / go.sum を整える
tidy:
	go mod tidy

## clean: 生成物を削除
clean:
	rm -rf $(DIST) $(BINARY) $(WIN_BINARY) filelist_*.csv

## help: ターゲット一覧
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## //'
