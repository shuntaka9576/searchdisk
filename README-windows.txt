============================================================
searchdisk - ディスク内ファイル一覧 CSV 出力ツール (Windows 版)
============================================================

■ 概要
固定ディスクを再帰スキャンしてファイル情報を CSV に出力します。
2TB 級の大容量ディスクでもメモリ一定のままストリーミング書き込みします。

  CSV 列: ディスク, ファイルネーム, 種類, 作成日時, 更新日時, サイズ, フルパス


■ セットアップ
1. searchdisk-setup.exe を実行してインストール（管理者権限不要）
2. searchdisk.exe が Program Files\searchdisk\ に配置され、PATH に自動追加されます
3. インストール後、コマンドプロンプトまたは PowerShell から使えます

  ※ 既に開いている端末では PATH が反映されないため、再起動してください。


■ 使い方

  searchdisk [options]

コマンドプロンプト (cmd) または PowerShell を開いて実行します。

  全固定ディスクをスキャン (既定):
    searchdisk

  パスを指定:
    searchdisk -path "D:\,E:\Data"

  出力先を指定:
    searchdisk -path "D:\" -out "D:\out.csv"

  1MB 未満のファイルを除外:
    searchdisk -min-size 1048576

  追加の除外パス (部分一致):
    searchdisk -exclude "\Backup\,\Temp\"

  既定の除外を無効化 (Windows / Program Files など):
    searchdisk -default-exclude=false

  シンボリックリンク/ジャンクションを辿る:
    searchdisk -follow-symlink

  ヘルプ:
    searchdisk -help


■ 既定で除外されるパス
\Windows\, \Program Files\, \Program Files (x86)\, \ProgramData\,
\PerfLogs\, \$Recycle.Bin\, \System Volume Information\, \$WinREAgent\,
\Recovery\, \Config.Msi\, \AppData\Local\Temp\, \AppData\Local\Microsoft\Windows\INetCache\,
\AppData\Local\Microsoft\Windows\WebCache\, \AppData\Local\Packages\,
\hiberfil.sys, \pagefile.sys, \swapfile.sys, \node_modules\, \.git\objects\


■ CSV 仕様
- 文字コード: UTF-8 BOM (Excel で文字化けしません)
- 改行: CRLF
- 既定の出力ファイル名: filelist_yyyyMMdd_HHmmss.csv (カレントディレクトリ)


■ アンインストール
「設定」→「アプリ」→「インストールされているアプリ」から
"searchdisk" をアンインストールしてください。


■ ライセンス
MIT
