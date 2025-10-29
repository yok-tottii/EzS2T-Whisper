# EzS2T-Whisper

**高速ローカルSTTアプリケーション - macOS専用**

EzS2T-Whisperは、Whisper.cppを使用した完全オフライン音声文字起こしアプリケーションです。ホットキー（Ctrl+Option+Space）を押すだけで、音声を自動的に文字起こしして、アクティブなアプリケーションに貼り付けます。

![Version](https://img.shields.io/badge/version-0.3.0-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-macOS-lightgrey)

## 特徴

- 🔒 **完全オフライン処理**: インターネット接続なしで動作
- ⚡ **高速**: Apple Silicon M1+でRTF < 1.0を実現
- 🌍 **多言語対応**: 日本語（デフォルト）と英語の音声認識
- 🎙️ **自動貼り付け**: 文字起こし結果を自動的にペースト
- 🔐 **プライバシー重視**: すべての処理をローカルで完結
- 🎛️ **カスタマイズ可能**: ホットキー、モデル、言語、マイクデバイスを自由に設定
- 🌐 **多言語UI**: 日本語・英語のUI言語切り替えに対応
- 🔄 **安全なクリップボード復元**: changeCount方式で元のクリップボード内容を保護
- ⚙️ **ホットキー競合チェック**: Spotlight、Alfred、Raycast等との競合を検出

## システム要件

- **OS**: macOS 11 (Big Sur) 以上
- **推奨**: Apple Silicon (M1/M2/M3/M4)
- **Intel Mac**: サポート（パフォーマンスは劣る）
- **メモリ**: 8GB以上推奨
- **ディスク**: 2GB以上の空き容量

## インストール

### 前提条件

Xcode コマンドラインツールと必要なシステムライブラリをインストール：

```bash
# Xcode コマンドラインツール
xcode-select --install

# システム依存関係（Homebrew経由）
brew install libpng libjpeg portaudio
```

### Go からビルド

```bash
# リポジトリをクローン
git clone https://github.com/yok-tottii/EzS2T-Whisper.git
cd EzS2T-Whisper

# Go 依存関係をダウンロード
go mod download

# ビルド
go build -o ezs2t-whisper ./cmd/ezs2t-whisper

# 実行
./ezs2t-whisper
```

### リリースビルド

```bash
# サイズ最適化ビルド
go build -ldflags="-s -w" -o ezs2t-whisper ./cmd/ezs2t-whisper
```

## 使い方

### 初回起動時

1. アプリケーションを起動すると自動的に設定画面が開きます
2. システム権限を確認・付与（マイク、アクセシビリティ）
3. モデルを選択（推奨: `ggml-large-v3-turbo-q5_0`）
4. ホットキーを設定（デフォルト: Ctrl+Option+Space）
5. 必要に応じてマイクデバイスとUI言語を選択
6. 設定を保存してテスト録音を実行

### 基本的な使い方

1. **ホットキーを押す**: Ctrl+Option+Space を押すと録音が開始
2. **話す**: マイクに向かって話す（最大60秒）
3. **ホットキーを離す**: キーを離すと録音が停止し、文字起こしが開始
4. **自動貼り付け**: テキストが自動的にアクティブなアプリケーションに貼り付け

### 設定画面

メニューバーのアイコンから「設定を開く...」を選択すると、ブラウザで設定画面が開きます。

**設定項目:**
- **ホットキー設定**: 録音開始のキーコンビネーションをカスタマイズ
- **録音モード**: 押下中録音 / トグル切替を選択
- **モデル選択**: 推奨モデルまたはカスタムモデルを選択
- **音声認識言語**: 日本語（ja）/ 英語（en）
- **マイクデバイス選択**: システムデフォルトまたは利用可能なデバイスを選択
- **UI言語**: 設定画面の表示言語（日本語 / English）

### 録音モード

- **押下中録音（デフォルト）**: ホットキーを押している間のみ録音
- **トグル切替**: 1回目の押下で録音開始、2回目で停止

### クリップボードの安全性

EzS2T-Whisperは**changeCount方式**を採用し、ユーザーのクリップボード内容を保護します：

1. 文字起こし開始時にクリップボードの状態を保存
2. 文字起こしテキストを一時的にクリップボードにコピーして貼り付け
3. **ユーザーが変換中に別のコピー操作を行わなかった場合のみ**元のクリップボード内容を復元
4. ユーザーが介入した場合は復元をスキップし、新しい内容を保持

## モデル管理

### 推奨モデル

- **`ggml-large-v3-turbo-q5_0.bin`** (~1.5GB)
  - 用途: 高精度な音声文字起こし（ASR専用、翻訳なし）
  - 推奨: Apple Silicon M1以上
  - RTF < 0.5 (Apple Silicon M1+)

### 軽量モデル

- **`ggml-small-q5_1.bin`** (~200MB)
  - 用途: バッテリー節約、低性能端末向け
  - 推奨: Intel Mac
  - 発熱抑制、軽量動作

### モデル配置先

```
~/Library/Application Support/EzS2T-Whisper/models/
```

モデルファイル（`.bin` または `.gguf`）をこのディレクトリに配置すると、設定画面で自動的に検出されます。

### モデルのダウンロード方法

公式のWhisper.cppスクリプトを使用してダウンロードできます：

```bash
# Whisper.cppリポジトリをクローン
git clone https://github.com/ggerganov/whisper.cpp.git
cd whisper.cpp

# 推奨モデルをダウンロード
./models/download-ggml-model.sh large-v3-turbo

# モデルファイルを EzS2T-Whisper のディレクトリにコピー
mkdir -p ~/Library/Application\ Support/EzS2T-Whisper/models/
cp models/ggml-large-v3-turbo.bin ~/Library/Application\ Support/EzS2T-Whisper/models/
```

**注意**: ファイル名に `q5_0` が含まれている場合は、拡張子を確認して適宜コピーしてください。

## トラブルシューティング

### ホットキーが反応しない

**原因**: アクセシビリティ権限が未付与

**解決策**:
1. システム設定 → プライバシーとセキュリティ → アクセシビリティ
2. Terminal.app（またはこのアプリ）が許可されていることを確認
3. アプリを再起動

### 録音が開始されない

**原因**: マイク権限が未付与

**解決策**:
1. システム設定 → プライバシーとセキュリティ → マイク
2. Terminal.app（またはこのアプリ）が許可されていることを確認
3. アプリを再起動

### テキストが貼り付けられない

**原因**: アクセシビリティ権限がない、または別のアプリが入力を妨害

**解決策**:
1. アクセシビリティ権限を確認
2. テキストエディタなど単純なアプリで試す
3. アプリを再起動

### 文字起こしが遅い

**原因**: モデルが大きい、または処理能力が不足

**解決策**:
1. 軽量モデル（`ggml-small-q5_1.bin`）に変更
2. Activity Monitor で他の重い処理が実行されていないか確認
3. バックグラウンドアプリを終了

### ホットキーが他のアプリと競合する

**原因**: Spotlight、Alfred、Raycast等のランチャーアプリと競合

**解決策**:
1. 設定画面でホットキーを変更
2. 競合警告が表示される場合は、別のキーコンビネーションを試す
3. 推奨: Ctrl+Shift+Space、Ctrl+Option+R など

### ビルドエラー

#### `portaudio.h: No such file or directory`
```bash
brew install portaudio
```

#### `ld: library not found for -lwhisper`
Whisper.cppが正しくビルドされていません。[Whisper.cpp](https://github.com/ggerganov/whisper.cpp)のセットアップを確認してください。

## 動作確認

本アプリケーションは以下の環境で動作確認済みです：

### テスト済みマイクデバイス
- **MacBook Airのマイク** - 音声入力・文字起こし動作確認済み
- **iPhoneのマイク** - 音声入力・文字起こし動作確認済み

### テスト済み機能
- ✅ 複数マイクデバイス認識と切り替え
- ✅ 各デバイスからの音声入力
- ✅ Whisper.cppによる文字起こし
- ✅ 自動テキスト貼り付け
- ✅ changeCount方式のクリップボード復元
- ✅ ホットキー競合チェック
- ✅ 多言語UI（日本語/英語）

## API エンドポイント

| メソッド | エンドポイント | 説明 |
|---------|-----------|------|
| GET | `/api/settings` | 現在の設定を取得 |
| PUT | `/api/settings` | 設定を更新 |
| POST | `/api/hotkey/validate` | ホットキーの競合チェック |
| POST | `/api/hotkey/register` | ホットキーを登録 |
| GET | `/api/devices` | オーディオ入力デバイス一覧を取得 |
| GET | `/api/models` | 利用可能なモデル一覧を取得 |
| POST | `/api/models/rescan` | モデルディレクトリを再スキャン |
| POST | `/api/models/browse` | ネイティブファイル選択ダイアログを開く |
| POST | `/api/models/validate` | モデルファイルパスを検証 |
| POST | `/api/test/record` | テスト録音を実行 |
| GET | `/api/permissions` | 必要な権限の状態を確認 |

## 設定ファイル

設定は以下の場所に JSON形式で保存されます：

```
~/Library/Application Support/EzS2T-Whisper/config.json
```

### 設定項目

```json
{
  "hotkey": {
    "ctrl": true,
    "shift": false,
    "alt": true,
    "cmd": false,
    "key": "Space"
  },
  "recording_mode": "press-to-hold",
  "model_path": "~/Library/Application Support/EzS2T-Whisper/models/ggml-large-v3-turbo-q5_0.bin",
  "language": "ja",
  "audio_device_id": -1,
  "ui_language": "ja",
  "max_record_time": 60,
  "paste_split_size": 500
}
```

## ログ

アプリケーションのログは以下の場所に保存されます：

```
~/Library/Application Support/EzS2T-Whisper/logs/
```

ファイル形式: `ezs2t-whisper-YYYYMMDD.log`

ログレベル:
- **INFO**: 起動、設定変更、録音開始/終了等
- **WARN**: 権限不足、デバイスエラー等
- **ERROR**: クラッシュ、予期しないエラー
- **DEBUG**: デバッグモード時のみ詳細ログ

ログは7日間保持され、古いログは自動的に削除されます。

## 開発

### ディレクトリ構造

```
EzS2T-Whisper/
├── cmd/
│   └── ezs2t-whisper/
│       └── main.go              # エントリーポイント
├── internal/
│   ├── hotkey/                  # グローバルホットキー
│   ├── audio/                   # オーディオ入力
│   ├── recording/               # 録音ロジック
│   ├── recognition/             # Whisper.cpp 統合
│   ├── clipboard/               # クリップボード操作
│   ├── tray/                    # システムトレイ
│   ├── server/                  # HTTPサーバー
│   ├── api/                     # REST API
│   ├── config/                  # 設定管理
│   ├── i18n/                    # 多言語対応
│   ├── permissions/             # システム権限チェック
│   ├── notification/            # 通知機能
│   ├── wizard/                  # セットアップウィザード
│   └── logger/                  # ログ出力
├── assets/                      # 静的アセット
│   └── icon/                    # トレイアイコン
├── specs/                       # 仕様書
├── LICENSES/                    # 依存ライブラリのライセンス
└── go.mod
```

### テストの実行

```bash
# すべてのテストを実行
go test ./...

# 詳細出力
go test -v ./...

# 特定パッケージのテスト
go test -v ./internal/permissions

# カバレッジ付き
go test -cover ./...
```

### コード品質チェック

```bash
# フォーマット確認
go fmt ./...

# 静的解析
go vet ./...

# Linter（golangci-lintインストール後）
golangci-lint run ./...
```

## 依存関係

### Go パッケージ

| パッケージ | バージョン | 用途 |
|-----------|----------|------|
| `golang.design/x/hotkey` | v0.4.1 | グローバルホットキー |
| `github.com/gordonklaus/portaudio` | v0.0.0-20250206... | オーディオ入力 |
| `github.com/go-vgo/robotgo` | v0.110.8 | クリップボード、キーイベント |
| `github.com/getlantern/systray` | v1.2.2 | システムトレイ |

### システム依存関係

| ツール | インストール | 用途 |
|-------|-----------|------|
| Xcode CLI | `xcode-select --install` | CGOコンパイル |
| libpng | `brew install libpng` | robotgo依存 |
| libjpeg | `brew install libjpeg` | robotgo依存 |
| portaudio | `brew install portaudio` | オーディオ入力 |

### macOSフレームワーク（CGO経由）

- **AVFoundation** - マイク権限チェック
- **ApplicationServices** - アクセシビリティ権限チェック
- **Cocoa** - NSPasteboard（クリップボードchangeCount）
- **Whisper.cpp** - 音声認識エンジン

## 将来の拡張予定

以下の機能は将来のバージョンで実装予定です：

- **VAD（Voice Activity Detection）**: 無音検出による自動録音停止
- **逐次出力**: 文字起こし中の部分的なテキスト表示
- **辞書機能**: ユーザー定義の固有名詞・専門用語の補正
- **音声コマンド**: 「まる」「改行」等の発話での制御
- **録音波形表示**: メニューバーでの波形可視化
- **モデル自動ダウンロード**: アプリ内からの公式モデルダウンロード

## 配布とサポート

### 配布方針

- **個人プロジェクト**: 個人利用および興味を持ったmacOSユーザーを対象としています
- **オープンソース**: GitHub上でMIT Licenseの下で公開
- **ビルド済みバイナリ**: 現在は提供していません。ソースからビルドしてください

### サポートについて

**重要**: このプロジェクトは個人プロジェクトのため、以下の点にご留意ください：

- ✅ **バグ報告・機能リクエスト**: [GitHub Issues](https://github.com/yok-tottii/EzS2T-Whisper/issues) で歓迎します
- ✅ **プルリクエスト**: [GitHub Pull Requests](https://github.com/yok-tottii/EzS2T-Whisper/pulls) で歓迎しますが、対応は任意です
- ⚠️ **サポート**: 公式なサポートは保証しません
- ⚠️ **対応時期**: バグ修正や機能追加の対応時期は保証しません
- ℹ️ **コミュニティ**: ユーザー同士での情報交換・助け合いを推奨します

### セキュリティ

セキュリティ上の問題を発見した場合は、公開のIssueではなく、プロジェクトメンテナに直接ご連絡ください。

## ライセンス

このプロジェクトは **MIT License** の下で公開されています。詳細は [LICENSE](LICENSE) ファイルをご覧ください。

### 依存ライブラリのライセンス

各ライブラリのライセンスに従います。詳細は`LICENSES/`ディレクトリをご覧ください：

| ライブラリ | ライセンス | ファイル |
|-----------|----------|---------|
| Whisper.cpp | MIT | [LICENSES/Whisper_cpp_MIT.md](LICENSES/Whisper_cpp_MIT.md) |
| robotgo | Apache 2.0 | [LICENSES/robotgo_Apache_2.0.md](LICENSES/robotgo_Apache_2.0.md) |
| getlantern/systray | Apache 2.0 | [LICENSES/getlantern-systray_Apache_2.0.md](LICENSES/getlantern-systray_Apache_2.0.md) |
| golang-design/hotkey | MIT | [LICENSES/golang-design-hotkey_MIT.md](LICENSES/golang-design-hotkey_MIT.md) |
| gordonklaus/portaudio | MIT | [LICENSES/gordonklaus-portaudio_MIT.md](LICENSES/gordonklaus-portaudio_MIT.md) |

## 謝辞

このプロジェクトは以下のプロジェクトを利用しています：

- [Whisper.cpp](https://github.com/ggerganov/whisper.cpp) - 音声認識エンジン
- [robotgo](https://github.com/go-vgo/robotgo) - クリップボード・キーイベント
- [getlantern/systray](https://github.com/getlantern/systray) - システムトレイ
- [golang-design/hotkey](https://github.com/golang-design/hotkey) - ホットキー
- [gordonklaus/portaudio](https://github.com/gordonklaus/portaudio) - オーディオ入力
- [Material Symbols and Icons](https://fonts.google.com/icons) (Apache 2.0) - システムトレイアイコン

## アイコン

システムトレイアイコンには、Googleの [Material Symbols and Icons](https://fonts.google.com/icons) を使用しています（Apache License 2.0）。

使用アイコン：

- `speech_to_text` - 待機状態
- `graphic_eq` - 録音中
- `hourglass_empty` - 処理中

---

Made with ❤️ for macOS
