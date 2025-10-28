# EzS2T-Whisper

**高速ローカルSTTアプリケーション - macOS専用**

EzS2T-Whisperは、Whisper.cppを使用した完全オフライン音声文字起こしアプリケーションです。ホットキー（Ctrl+Option+Space）を押すだけで、音声を自動的に文字起こしして、アクティブなアプリケーションに貼り付けます。

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-macOS-lightgrey)

## 特徴

- 🔒 **完全オフライン処理**: インターネット接続なしで動作
- ⚡ **高速**: Apple Silicon M1+でRTF < 1.0を実現
- 🌍 **多言語対応**: 日本語（デフォルト）と英語に対応
- 🎙️ **自動貼り付け**: 文字起こし結果を自動的にペースト
- 🔐 **プライバシー重視**: すべての処理をローカルで完結
- 🎛️ **カスタマイズ可能**: ホットキー、モデル、言語を自由に設定

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

1. アプリケーションを起動するとセットアップウィザードが表示されます
2. システム権限を確認（マイク、アクセシビリティ）
3. モデルを選択（推奨: `ggml-large-v3-turbo-q5_0`）
4. ホットキーを設定（デフォルト: Ctrl+Option+Space）
5. テスト録音を実行

### 基本的な使い方

1. **ホットキーを押す**: Ctrl+Option+Space を押すと録音が開始
2. **話す**: マイクに向かって話す
3. **ホットキーを離す**: キーを離すと録音が停止し、文字起こしが開始
4. **自動貼り付け**: テキストが自動的にアクティブなアプリケーションに貼り付け

### 設定画面

メニューバーのアイコンから「設定を開く...」を選択すると、ブラウザで設定画面が開きます。

**設定項目:**
- ホットキー設定
- 録音モード（押下中 / トグル）
- モデル選択
- マイクデバイス選択
- UI言語

## モデル管理

### デフォルトモデル

- **`ggml-large-v3-turbo-q5_0.gguf`** (~1.5GB)
  - 用途: 精度優先
  - 推奨: Apple Silicon M1以上

### 軽量モデル

- **`ggml-small-q5_1.gguf`** (~200MB)
  - 用途: バッテリー節約、低性能端末向け
  - 推奨: Intel Mac

### モデル配置先

```
~/Library/Application Support/EzS2T-Whisper/models/
```

モデルファイルをこのディレクトリに配置すると、設定画面で自動的に検出されます。

## トラブルシューティング

### ホットキーが反応しない

**原因**: アクセシビリティ権限が未付与

**解決策**:
1. システム設定 → プライバシーとセキュリティ → アクセシビリティ
2. Terminal.app（またはこのアプリ）が許可されていることを確認

### 録音が開始されない

**原因**: マイク権限が未付与

**解決策**:
1. システム設定 → プライバシーとセキュリティ → マイク
2. Terminal.app（またはこのアプリ）が許可されていることを確認

### テキストが貼り付けられない

**原因**: アクセシビリティ権限がない、または別のアプリが入力を妨害

**解決策**:
1. アクセシビリティ権限を確認
2. テキストエディタなど単純なアプリで試す
3. アプリを再起動

### 文字起こしが遅い

**原因**: モデルが大きい、または処理能力が不足

**解決策**:
1. 軽量モデル（`ggml-small-q5_1.gguf`）に変更
2. Activity Monitor で他の重い処理が実行されていないか確認

### ビルドエラー

#### `portaudio.h: No such file or directory`
```bash
brew install portaudio
```

#### `ld: library not found for -lwhisper`
Whisper.cppが正しくビルドされていません。[Whisper.cpp](https://github.com/ggerganov/whisper.cpp)のセットアップを確認してください。

## API エンドポイント

| メソッド | エンドポイント | 説明 |
|---------|-----------|------|
| GET | `/api/settings` | 現在の設定を取得 |
| PUT | `/api/settings` | 設定を更新 |
| POST | `/api/hotkey/validate` | ホットキーの競合チェック |
| POST | `/api/hotkey/register` | ホットキーを登録 |
| GET | `/api/devices` | オーディオ入力デバイス一覧 |
| GET | `/api/models` | 利用可能なモデル一覧 |
| POST | `/api/models/rescan` | モデルディレクトリを再スキャン |
| POST | `/api/test/record` | テスト録音を実行 |
| GET | `/api/permissions` | 必要な権限の状態を確認 |

## 設定ファイル

設定は以下の場所に JSON形式で保存されます：

```
~/Library/Application Support/EzS2T-Whisper/config.json
```

## ログ

アプリケーションのログは以下の場所に保存されます：

```
~/Library/Application Support/EzS2T-Whisper/logs/
```

ファイル形式: `ezs2t-whisper-YYYYMMDD.log`

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
├── frontend/                    # Web UI（embed済み）
├── specs/                       # 仕様書
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
```

### コード品質チェック

```bash
# フォーマット確認
go fmt ./...

# 静的解析
go vet ./...
```

## 依存関係

### Go パッケージ

| パッケージ | 用途 |
|-----------|------|
| `github.com/golang-design/hotkey` | グローバルホットキー |
| `github.com/gordonklaus/portaudio` | オーディオ入力 |
| `github.com/go-vgo/robotgo` | クリップボード、キーイベント |
| `github.com/getlantern/systray` | システムトレイ |

### システム依存関係

| ツール | インストール | 用途 |
|-------|-----------|------|
| Xcode CLI | `xcode-select --install` | CGOコンパイル |
| libpng | `brew install libpng` | robotgo依存 |
| libjpeg | `brew install libjpeg` | robotgo依存 |
| portaudio | `brew install portaudio` | オーディオ入力 |

## ライセンス

このプロジェクトは **MIT License** の下で公開されています。

## 貢献

バグ報告や機能リクエストは [GitHub Issues](https://github.com/yok-tottii/EzS2T-Whisper/issues) でお願いします。

## 謝辞

このプロジェクトは以下のプロジェクトを利用しています：

- [Whisper.cpp](https://github.com/ggerganov/whisper.cpp) - 音声認識エンジン
- [robotgo](https://github.com/go-vgo/robotgo) - クリップボード・キーイベント
- [getlantern/systray](https://github.com/getlantern/systray) - システムトレイ
- [golang-design/hotkey](https://github.com/golang-design/hotkey) - ホットキー
- [gordonklaus/portaudio](https://github.com/gordonklaus/portaudio) - オーディオ入力

---

**Made with ❤️ for macOS**
