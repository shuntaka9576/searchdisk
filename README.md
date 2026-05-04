# searchdisk

Windows 上の固定ディスクを再帰スキャンしてファイル一覧を CSV に出力する Go 製 CLI。
2TB 級のディスクでもメモリ一定のままストリーミングで書き出します。

## CSV 仕様

| 列 | 内容 |
|---|---|
| ディスク | ボリュームルート (例: `C:\`) |
| ファイルネーム | ベース名 (例: `report.xlsx`) |
| 種類 | 拡張子を小文字化 (例: `xlsx`)。拡張子なしは `(なし)` |
| 作成日時 | `yyyy-MM-dd HH:mm:ss` |
| 更新日時 | `yyyy-MM-dd HH:mm:ss` |
| サイズ | バイト数 (整数) |
| フルパス | フル絶対パス |

- 文字コード: UTF-8 (既定で BOM あり、Excel で文字化けしない)
- 改行: CRLF
- 値内の `,` `"` 改行は `encoding/csv` 標準のクォート/エスケープで処理

## 既定の除外パス

普段使わないシステム/キャッシュ系を既定で除外します (`-default-exclude=false` で無効化)。

```
\Windows\        \Program Files\        \Program Files (x86)\
\ProgramData\    \PerfLogs\             \$Recycle.Bin\
\System Volume Information\             \$WinREAgent\
\Recovery\       \Config.Msi\
\AppData\Local\Temp\
\AppData\Local\Microsoft\Windows\INetCache\
\AppData\Local\Microsoft\Windows\WebCache\
\AppData\Local\Packages\
\hiberfil.sys    \pagefile.sys          \swapfile.sys
\node_modules\   \.git\objects\
```

マッチは大小文字無視・パス区切り (`\`/`/`) 無視の部分一致です。

## ビルド

```sh
# Windows 向け (Mac/Linux からクロスコンパイル)
GOOS=windows GOARCH=amd64 go build -o searchdisk.exe .

# 自分の OS 向け (検証/開発用)
go build -o searchdisk .
```

依存ゼロ (Go stdlib のみ) なので生成された exe を Windows にコピーするだけで動きます。

## 使い方

```powershell
# 全固定ディスクをスキャン (既定。Excel で開ける CSV を作る)
.\searchdisk.exe

# パス指定 (カンマ区切りで複数可)
.\searchdisk.exe -path "D:\,E:\Data"

# 出力先指定
.\searchdisk.exe -path "D:\" -out "D:\out.csv"

# 1MB 未満を除外
.\searchdisk.exe -min-size 1048576

# 追加の除外パターン (部分一致)
.\searchdisk.exe -exclude "\Backup\,\Temp\"

# 既定除外を無効化して全部出す
.\searchdisk.exe -default-exclude=false

# Excel で開かない (BOM なし)
.\searchdisk.exe -bom=false

# シンボリックリンク/ジャンクションを辿る (既定は辿らない)
.\searchdisk.exe -follow-symlink
```

`-help` で全フラグを確認できます。

## 出力例

```csv
ディスク,ファイルネーム,種類,作成日時,更新日時,サイズ,フルパス
C:\,report.xlsx,xlsx,2025-04-01 09:12:33,2025-04-12 18:02:11,184320,C:\Users\me\Documents\report.xlsx
C:\,README,(なし),2025-01-15 10:00:00,2025-01-15 10:00:00,512,C:\Users\me\Projects\README
```

## 実行中の挙動

- 1 万件ごとに進捗 (件数 / 累計サイズ / 件数/秒) を stderr に表示
- 権限不足 / ロック中ファイル / 開けないディレクトリは静かにスキップ
- 完了時に合計件数・総サイズ・経過時間・除外ディレクトリ数を表示

2TB / 数百万ファイル想定で、メモリ消費は数十 MB に収まります。

## テスト

ローカルで:

```sh
go test -v ./...
go vet ./...
```

軽い単体テスト (`main_test.go`) で以下を検証しています。

- 一時ディレクトリに小さなツリーを作って `scanTree()` を呼び、CSV 行数とエントリを確認
- `-min-size` 未満の除外
- 拡張子の小文字化と「拡張子なし → `(なし)`」
- 除外パターンの大小文字 / パス区切り無視マッチ

CI:

- `.github/workflows/ci.yml` で Ubuntu / macOS / Windows の 3 OS で `go test -race` を実行
- Windows 向け exe をクロスビルドしてアーティファクトとしてアップロード

### 実機での簡単チェック

PowerShell から:

```powershell
# 適当な小さいフォルダで動作確認
.\searchdisk.exe -path "C:\Users\$env:USERNAME\Documents" -out test.csv

# CSV を Excel で開くか PowerShell で確認
Import-Csv test.csv | Select-Object -First 10 | Format-Table
```

予期せず `C:\Windows` 配下まで歩くことがないか、全体件数が常識的か (Documents 配下なら数千〜数万) を見れば OK です。

## 制限事項

- Windows 以外で `-path` を指定しない場合は対象パスが空になります (固定ディスク自動検出は Windows のみ)。Mac/Linux で動作確認したいときは `-path /some/dir` を明示してください
- 種類列は拡張子ベースのため、Windows エクスプローラーが表示する「テキスト ドキュメント」のような表示名とは異なります
- ハードリンクは別ファイルとしてそれぞれ計上されます (重複検出はしません)

## ライセンス

MIT
