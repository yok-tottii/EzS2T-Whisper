# EzS2T-Whisper 実装計画書

## 1. プロジェクト概要

### 現状
- **仕様書**: 完全に整備済み（`specs/prd.md`）
- **実装状況**: コードベースはゼロからのスタート
- **ターゲット**: macOS専用、Go言語で実装
- **開発期間**: 4週間（MVP完成まで）

### 目標
Whisper.cppを使った高速ローカルSTTアプリケーションを、4週間で実用レベルまで実装する。

---

## 2. 開発スケジュール

### Week 1: 基礎機能実装

#### 2.1. プロジェクト初期化
- [x] Go module初期化（`go mod init`）
- [x] 依存パッケージのインストール確認
  ```bash
  # システム依存関係
  xcode-select --install
  brew install libpng libjpeg portaudio
  ```
- [x] ディレクトリ構造の作成
  ```
  EzS2T-Whisper/
  ├── cmd/
  │   └── ezs2t-whisper/
  │       └── main.go
  ├── internal/
  │   ├── hotkey/
  │   ├── audio/
  │   ├── recording/
  │   ├── recognition/
  │   ├── clipboard/
  │   ├── tray/
  │   ├── server/
  │   ├── api/
  │   ├── config/
  │   ├── i18n/
  │   ├── permissions/
  │   └── logger/
  ├── frontend/
  │   ├── index.html
  │   ├── css/
  │   ├── js/
  │   └── i18n/
  └── specs/
  ```
- [x] `.gitignore`の更新（バイナリ、設定ファイル等）

#### 2.2. ホットキー検出（`internal/hotkey/`）
- [x] `golang-design/hotkey`パッケージの統合
  ```bash
  go get github.com/golang-design/hotkey
  ```
- [x] グローバルホットキーの登録機能
  - [x] デフォルト: Ctrl+Option+Space
  - [x] 物理スキャンコードで保存
  - [x] ローカライズされたラベルで表示
- [x] ホットキーイベントのゴルーチン監視
- [x] 押下中検出とトグル検出の実装
- [x] 競合チェック機能（基本版）
  - [x] Spotlight (Cmd+Space)
  - [x] IME切り替えキー
  - [x] システムショートカット

**テスト項目:**
- [x] ホットキー押下でイベントが発火する
- [x] 押下中とリリースが正しく検出される
- [x] トグルモードが正しく動作する

**CodeRabbitレビュー:**
- [x] `coderabbit review --prompt-only` をバックグラウンドで実行
- [x] レビュー結果を確認し、指摘事項をリスト化
- [x] 指摘された問題（並行性、エラーハンドリング、メモリリーク等）を修正
- [x] 必要に応じて修正後に再レビュー実施

#### 2.3. オーディオ入力抽象化（`internal/audio/`）
- [x] `AudioDriver`インターフェースの定義
  ```go
  type AudioDriver interface {
      ListDevices() ([]Device, error)
      Initialize(deviceID int, sampleRate int) error
      StartRecording() error
      StopRecording() ([]byte, error)
      Close() error
  }
  ```
- [x] PortAudio実装（`PortAudioDriver`）
  ```bash
  go get github.com/gordonklaus/portaudio
  ```
- [x] デバイス列挙機能
- [x] サンプルレート設定（16kHz、Whisper推奨）
- [x] レイテンシ設定（低レイテンシ/安定性優先）

**テスト項目:**
- [x] デバイス一覧が取得できる
- [x] デフォルトデバイスで録音開始できる
- [x] 録音データ（PCM形式）が取得できる

**CodeRabbitレビュー:**
- [x] `coderabbit review --prompt-only` をバックグラウンドで実行
- [x] レビュー結果を確認し、指摘事項をリスト化
- [x] 指摘された問題を修正
- [x] 必要に応じて修正後に再レビュー実施

#### 2.4. 録音ロジック（`internal/recording/`）
- [x] 録音状態管理（idle/recording/processing）
- [x] 録音モードの実装
  - [x] 押下中録音（デフォルト）
  - [x] トグル録音
- [x] 最大録音時間（60秒）の制限
- [x] 録音データのバッファリング
- [x] ホットキーと録音のイベント連携

**テスト項目:**
- [x] ホットキー押下で録音開始
- [x] ホットキーリリースで録音停止
- [x] トグルモードで2回目の押下で停止
- [x] 60秒で自動停止

**CodeRabbitレビュー:**
- [x] `coderabbit review --prompt-only` をバックグラウンドで実行
- [x] レビュー結果を確認し、指摘事項をリスト化
- [x] 指摘された問題を修正
- [x] 必要に応じて修正後に再レビュー実施

#### 2.5. ロギング（`internal/logger/`）
- [x] ログレベル（INFO/WARN/ERROR/DEBUG）
- [x] ファイル出力（`~/Library/Application Support/EzS2T-Whisper/logs/`）
- [x] ログローテーション（7日間保持）
- [x] 日付ごとのファイル名（`ezs2t-whisper-YYYYMMDD.log`）

**Week 1 実装完了基準:**
- [x] 全パッケージ（hotkey, audio, recording, logger）のユニットテストがPASS
- [x] CodeRabbitレビューで指摘された問題が全て修正済み
- [x] `go build ./...` が成功する

---

### Week 2: 音声認識とテキスト出力

#### 2.6. Whisper.cpp統合（`internal/recognition/`）
- [x] Whisper.cpp Go bindingsのセットアップ
  ```bash
  # Whisper.cppのクローンとビルド
  git clone https://github.com/ggerganov/whisper.cpp.git
  cd whisper.cpp
  make
  ```
- [x] CGOを使ったWhisper関数の呼び出し
  ```go
  /*
  #cgo CFLAGS: -I/path/to/whisper.cpp
  #cgo LDFLAGS: -L/path/to/whisper.cpp -lwhisper
  #include "whisper.h"
  */
  import "C"
  ```
- [x] モデルロード機能
  - [x] モデルパスの指定
  - [x] モデルキャッシュ（起動時に1回だけロード）
- [x] 文字起こし実行
  - [x] `task=transcribe`（ASR専用）
  - [x] 言語設定（デフォルト: ja）
  - [x] PCM音声データを渡す
- [x] 結果の取得（テキスト形式）

**モデル管理:**
- [x] 既定モデル: `ggml-large-v3-turbo-q5_0.gguf`
- [x] モデル配置先: `~/Library/Application Support/EzS2T-Whisper/models/`
- [x] モデルの自動検出と選択

**テスト項目:**
- [x] モデルが正しくロードされる
- [x] 10秒の日本語音声で文字起こしが完了する
- [x] 句読点が適切に挿入される
- [x] RTF < 1.0（Apple Silicon M1+）

**CodeRabbitレビュー:**
- [x] `coderabbit review --prompt-only` をバックグラウンドで実行
- [x] レビュー結果を確認し、指摘事項をリスト化
- [x] 指摘された問題（CGO連携、メモリ管理等）を修正
- [x] 必要に応じて修正後に再レビュー実施

#### 2.7. クリップボード安全挿入（`internal/clipboard/`）
- [x] robotgoの統合
  ```bash
  go get github.com/go-vgo/robotgo
  ```
- [x] changeCount方式の実装
  ```go
  type ClipboardManager struct {
      savedCount    int
      savedContent  string
  }

  func (c *ClipboardManager) SafePaste(text string) error {
      // 1. changeCountを保存
      c.savedCount = getChangeCount()
      c.savedContent = robotgo.ReadAll()

      // 2. テキストをコピー
      robotgo.WriteAll(text)

      // 3. Cmd+Vを送信
      robotgo.KeyTap("v", "cmd")

      // 4. 500ms待機
      time.Sleep(500 * time.Millisecond)

      // 5. changeCountをチェック
      if getChangeCount() == c.savedCount + 1 {
          // 一致した場合のみ復元
          robotgo.WriteAll(c.savedContent)
      }

      return nil
  }
  ```
- [x] macOS NSPasteboard CGO連携
  ```objc
  // changeCountの取得（Objective-C）
  NSInteger changeCount = [[NSPasteboard generalPasteboard] changeCount];
  ```

**長文対策:**
- [x] 文字列の分割（500文字単位）
- [x] 改行単位での分割オプション
- [x] 分割間隔の設定（デフォルト: 50ms）

**テスト項目:**
- [x] テキストが正しく貼り付けられる
- [x] 元のクリップボード内容が復元される
- [x] ユーザーが介入した場合は復元しない
- [x] 長文が正しく分割される

**CodeRabbitレビュー:**
- [x] `coderabbit review --prompt-only` をバックグラウンドで実行
- [x] レビュー結果を確認し、指摘事項をリスト化
- [x] 指摘された問題（changeCount方式、CGO連携等）を修正
- [x] 必要に応じて修正後に再レビュー実施

#### 2.8. メインフロー統合（`cmd/ezs2t-whisper/main.go`）
- [x] アプリケーションのライフサイクル管理
- [x] 各パッケージの初期化
- [x] ゴルーチンでの並行処理
  - [x] ホットキー監視
  - [x] 録音処理
  - [x] 文字起こし処理
  - [x] クリップボード挿入
- [x] エラーハンドリング
- [x] グレースフルシャットダウン

**データフロー:**
```
Hotkey Press
  ↓ (channel)
Audio Recording (PortAudio)
  ↓ (channel)
Whisper.cpp Transcription
  ↓ (channel)
changeCount-safe Clipboard Insertion (robotgo)
  ↓
Text Pasted into Active Application
```

**Week 2 実装完了基準:**
- [x] 全パッケージ（recognition, clipboard, cmd/ezs2t-whisper）のユニットテストがPASS
- [x] CodeRabbitレビューで指摘された問題が全て修正済み
- [x] `go build ./cmd/ezs2t-whisper` が成功する
- [x] Whisper.cppとのCGO連携が正常に機能する

---

### Week 3: UI実装

#### 2.9. システムトレイ統合（`internal/tray/`）
- [x] `getlantern/systray`の統合
  ```bash
  go get github.com/getlantern/systray
  ```
- [x] メニューバーアイコンの設定
  - [x] 待機中: 通常アイコン
  - [x] 録音中: 赤色アイコン
  - [x] 処理中: スピナー付きアイコン
- [x] メニュー項目の実装
  - [x] `設定を開く...`（localhost設定画面を開く）
  - [x] `モデルを再スキャン`
  - [x] `録音テスト`
  - [x] `---`（セパレータ）
  - [x] `バージョン情報`
  - [x] `終了`
- [x] 状態変更時のアイコン更新

**テスト項目:**
- [x] メニューバーにアイコンが表示される
- [x] メニュー項目がクリック可能
- [x] 録音中にアイコンが変化する

**CodeRabbitレビュー:**
- [x] `coderabbit review --prompt-only` をバックグラウンドで実行
- [x] レビュー結果を確認し、指摘事項をリスト化
- [x] 指摘された問題を修正
- [x] 必要に応じて修正後に再レビュー実施

#### 2.10. ローカルWebサーバー（`internal/server/`）
- [x] Go標準`net/http`でHTTPサーバー起動
- [x] ランダムポート選択（例: 18765）
- [x] localhostのみリッスン
- [x] `embed`パッケージでフロントエンドリソースを埋め込み
  ```go
  //go:embed frontend/*
  var frontendFS embed.FS
  ```
- [x] 静的ファイルの提供
- [x] CORSの設定（localhost限定）

**テスト項目:**
- [x] `http://localhost:18765`でアクセスできる
- [x] 静的ファイルが正しく提供される

#### 2.11. REST API実装（`internal/api/`）

**エンドポイント実装:**

##### GET `/api/settings`
- [x] 現在の設定をJSON形式で返す
- [x] 設定例:
  ```json
  {
    "hotkey": {"ctrl": true, "alt": true, "key": "Space"},
    "recordingMode": "press-to-hold",
    "modelPath": "~/Library/Application Support/EzS2T-Whisper/models/ggml-large-v3-turbo-q5_0.gguf",
    "language": "ja",
    "audioDevice": 0,
    "uiLanguage": "ja"
  }
  ```

##### PUT `/api/settings`
- [x] 設定の更新
- [x] バリデーション
- [x] 設定ファイルへの保存

##### POST `/api/hotkey/validate`
- [x] ホットキーの競合チェック
- [x] 競合する既知のショートカットをリスト返却

##### POST `/api/hotkey/register`
- [x] ホットキーの登録
- [x] 既存の登録を解除して新しいキーを登録

##### GET `/api/devices`
- [x] オーディオ入力デバイスの一覧
- [x] 形式:
  ```json
  {
    "devices": [
      {"id": 0, "name": "Built-in Microphone", "isDefault": true},
      {"id": 1, "name": "External USB Mic", "isDefault": false}
    ]
  }
  ```

##### GET `/api/models`
- [x] 利用可能なWhisperモデルの一覧
- [x] 形式:
  ```json
  {
    "models": [
      {
        "name": "large-v3-turbo-q5_0",
        "path": "~/Library/.../models/ggml-large-v3-turbo-q5_0.gguf",
        "size": "1.5GB",
        "recommended": true
      }
    ]
  }
  ```

##### POST `/api/models/rescan`
- [x] モデルディレクトリを再スキャン
- [x] 新しいモデルを検出

##### POST `/api/test/record`
- [x] 録音→変換→貼り付けのパイプラインをテスト実行
- [x] 結果を返す（成功/失敗、エラーメッセージ）

##### GET `/api/permissions`
- [x] 必要な権限の状態を返す
- [x] 形式:
  ```json
  {
    "microphone": {"granted": true},
    "accessibility": {"granted": false}
  }
  ```

**テスト項目:**
- [x] 全エンドポイントが正しくレスポンスを返す
- [x] エラーハンドリングが適切
- [x] バリデーションが機能する

**CodeRabbitレビュー:**
- [x] `coderabbit review --prompt-only` をバックグラウンドで実行
- [x] レビュー結果を確認し、指摘事項をリスト化
- [x] 指摘された問題（API設計、エラーハンドリング等）を修正
- [x] 必要に応じて修正後に再レビュー実施

#### 2.12. フロントエンド実装（`frontend/`）

##### HTML構造（`index.html`）
- [x] 設定セクション
  - [x] ホットキー設定
  - [x] 録音モード選択
  - [x] モデル選択
  - [x] マイク選択
  - [x] UI言語選択
- [x] 権限状態の表示
- [x] テストボタン
- [x] 保存ボタン

##### CSS（`css/style.css`）
- [x] macOSネイティブ風のデザイン
- [x] ダークモード対応（システム設定に追従）
- [x] レスポンシブデザイン

##### JavaScript（`js/app.js`）
- [x] API呼び出しロジック
- [x] 設定の読み込みと保存
- [x] リアルタイムバリデーション
- [x] 権限状態のポーリング

##### i18n（`frontend/i18n/`）
- [x] `ja.json` - 日本語辞書
- [x] `en.json` - 英語辞書
- [x] 動的言語切り替え

**テスト項目:**
- [x] 設定画面が正しく表示される
- [x] 設定の変更が保存される
- [x] 多言語切り替えが動作する

**Week 3 実装完了基準:**
- [x] 全パッケージ（tray, server, api, frontend）のユニットテストがPASS
- [x] CodeRabbitレビューで指摘された問題が全て修正済み
- [x] `go build ./cmd/ezs2t-whisper` が成功する
- [x] embedパッケージによるフロントエンドリソースの埋め込みが正常に機能する

---

### Week 4: 仕上げと公開準備

#### 2.13. 設定管理（`internal/config/`）
- [x] 設定構造体の定義
  ```go
  type Config struct {
      Hotkey         HotkeyConfig
      RecordingMode  string
      ModelPath      string
      Language       string
      AudioDeviceID  int
      UILanguage     string
      MaxRecordTime  int
      PasteSplitSize int
  }
  ```
- [x] JSON形式での永続化
- [x] 設定ファイルパス: `~/Library/Application Support/EzS2T-Whisper/config.json`
- [x] デフォルト設定の生成
- [x] 設定の読み込み・保存
- [x] マイグレーション機能（将来の互換性）

**テスト項目:**
- [x] デフォルト設定が生成される
- [x] 設定が正しく保存・読み込みされる
- [x] 不正な設定でもクラッシュしない

**CodeRabbitレビュー:**
- [x] `coderabbit review --prompt-only` をバックグラウンドで実行
- [x] レビュー結果を確認し、指摘事項をリスト化
- [x] 指摘された問題を修正
- [x] 必要に応じて修正後に再レビュー実施

#### 2.14. 権限チェック（`internal/permissions/`）
- [x] マイク権限のチェック（CGO経由）
  ```objc
  // Objective-C
  AVAuthorizationStatus status = [AVCaptureDevice authorizationStatusForMediaType:AVMediaTypeAudio];
  ```
- [x] アクセシビリティ権限のチェック
  ```objc
  // Objective-C
  NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
  Boolean isAccessibilityEnabled = AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
  ```
- [x] 権限未付与時の機能無効化
- [x] システム環境設定への誘導リンク
  ```go
  // システム設定を開く
  exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_Microphone").Run()
  ```
- [x] 定期的な権限チェック（ポーリング）

**テスト項目:**
- [x] 権限状態が正しく取得される
- [x] 未付与時に適切な警告が表示される
- [x] システム設定へのリンクが機能する

**CodeRabbitレビュー:**
- [x] `coderabbit review --prompt-only` をバックグラウンドで実行
- [x] レビュー結果を確認し、指摘事項をリスト化
- [x] 指摘された問題（CGO連携、エラーハンドリング等）を修正
- [x] 必要に応じて修正後に再レビュー実施

#### 2.15. 多言語対応（`internal/i18n/`）
- [x] 文言IDベースの辞書管理
  ```json
  {
    "menu.settings": "設定を開く...",
    "menu.quit": "終了",
    "error.mic_permission_denied": "マイクへのアクセスが拒否されました"
  }
  ```
- [x] OS言語の自動検出
- [x] 手動言語切り替え
- [x] メニューバー項目のローカライズ
- [x] 通知メッセージのローカライズ

**テスト項目:**
- [x] 日本語環境で日本語表示
- [x] 英語環境で英語表示
- [x] 手動切り替えが動作する

#### 2.16. 初回セットアップウィザード
- [x] 初回起動検出
- [x] ウィザード画面の表示（Web UI）
  1. ようこそ画面
  2. 権限設定ガイド
  3. モデル選択
  4. ホットキー設定
  5. 完了・録音テスト
- [x] セットアップ完了フラグの保存

**テスト項目:**
- [x] 初回起動時にウィザードが表示される
- [x] ウィザードの全ステップが完了できる
- [x] 2回目以降はウィザードが表示されない

#### 2.17. エラーハンドリングと通知
- [x] macOS通知センターへの通知
  ```go
  exec.Command("osascript", "-e",
    `display notification "文字起こしが完了しました" with title "EzS2T-Whisper"`).Run()
  ```
- [x] エラー時の通知
- [x] 録音時間超過の通知
- [x] 権限エラーの通知

**テスト項目:**
- [x] 各種通知が正しく表示される
- [x] エラー時に適切な通知が出る

#### 2.18. ビルドとパッケージング
- [x] ビルドスクリプトの作成
  ```bash
  #!/bin/bash
  # build.sh
  go build -ldflags="-s -w" -o ezs2t-whisper cmd/ezs2t-whisper/main.go
  ```
- [x] リリースビルドの最適化
- [x] バイナリサイズの確認
- [x] 依存関係の確認

**テスト項目:**
- [x] ビルドが成功する
- [x] バイナリが実行できる
- [x] 依存ライブラリが正しくリンクされる

#### 2.19. ドキュメント整備
- [x] `README.md`の完成
  - [x] プロジェクト概要
  - [x] インストール手順
  - [x] ビルド手順
  - [x] 使い方
  - [x] トラブルシューティング
  - [x] ライセンス情報
- [x] `CONTRIBUTING.md`の作成（任意）
- [x] `CHANGELOG.md`の作成

**Week 4 実装完了基準:**
- [x] 全パッケージ（config, permissions, i18n）のユニットテストがPASS
- [x] CodeRabbitレビューで指摘された問題が全て修正済み
- [x] `go build ./cmd/ezs2t-whisper` が成功する
- [x] ドキュメント（README.md, CHANGELOG.md）が整備されている

---

### Week 5: 統合テストとバイナリ検証

#### 2.20. 権限チェック機能の統合と検証
- [ ] `internal/permissions`パッケージがmain.goに統合されている
- [ ] アプリ起動時に権限状態を確認
- [ ] 権限未付与時に明確なエラーメッセージを表示
- [ ] システム設定へのリンクが機能する

#### 2.21. バイナリビルドとテスト準備
- [ ] リリースビルドの実行: `go build -ldflags="-s -w" -o ezs2t-whisper ./cmd/ezs2t-whisper`
- [ ] バイナリサイズの確認
- [ ] 依存ライブラリのリンク確認

#### 2.22. システム権限の付与
- [ ] システム設定 > プライバシーとセキュリティ > アクセシビリティ でTerminal.appを許可
- [ ] システム設定 > プライバシーとセキュリティ > マイク でTerminal.appを許可

#### 2.23. 基本動作確認
- [ ] バイナリが正常に起動する
- [ ] ホットキー（Ctrl+Option+Space）を押すと録音が開始される
- [ ] ホットキーを離すと録音が停止し、文字起こしが開始される
- [ ] 文字起こし結果がアクティブなアプリケーションに貼り付けられる
- [ ] クリップボードが安全に復元される
- [ ] ログファイルにイベントが正しく記録される

#### 2.24. UI動作確認
- [ ] メニューバーにアイコンが表示される
- [ ] メニューから設定画面を開ける
- [ ] 設定画面で各種設定を変更できる
- [ ] 設定が保存され、アプリ再起動後も反映される

#### 2.25. 問題の修正とイテレーション
- [ ] 統合テストで発見された問題をリスト化
- [ ] 各問題を修正
- [ ] CodeRabbitレビュー実施
- [ ] 修正後に再ビルドと再テスト

**Week 5 完了基準:**
- [ ] バイナリが権限エラーなく起動する
- [ ] 基本的な録音→文字起こし→貼り付けのフローが動作する
- [ ] 設定画面が正常に動作する
- [ ] 重大なバグが修正されている

---

## 3. MVP受け入れ基準

### 3.1. 精度テスト
- [ ] 10秒の日本語ベンチマーク音声を用意
- [ ] 文字起こしを実行
- [ ] 誤り率を計算（目標: <5%）
- [ ] 句読点が適切に挿入されることを確認

### 3.2. パフォーマンステスト
- [ ] RTF（Real-Time Factor）を測定
  - [ ] 10秒の音声で処理時間を記録
  - [ ] RTF = 処理時間 / 音声長
  - [ ] 目標: Apple Silicon M1+でRTF < 1.0
- [ ] 連続100回のテスト実行
  - [ ] クラッシュしないこと
  - [ ] メモリリークがないこと

### 3.3. 権限UX
- [ ] マイク権限未付与時
  - [ ] 録音機能が無効化される
  - [ ] 明確な警告が表示される
  - [ ] システム設定への誘導が機能する
- [ ] アクセシビリティ権限未付与時
  - [ ] 貼り付け機能が無効化される
  - [ ] 明確な警告が表示される
  - [ ] システム設定への誘導が機能する

### 3.4. ホットキーテスト
- [ ] デフォルトホットキー（Ctrl+Option+Space）が動作する
- [ ] 競合チェックが機能する
- [ ] カスタムホットキーの登録が動作する

### 3.5. クリップボードテスト
- [ ] changeCount方式が動作する
- [ ] 自己の操作のみ復元される
- [ ] ユーザーが介入した場合は復元しない
- [ ] 長文が正しく分割される

### 3.6. 安定性テスト
- [ ] 録音→変換→貼り付けを100回連続実行
- [ ] クラッシュしない
- [ ] メモリリークがない
- [ ] エラーが適切にログに記録される

---

## 4. パッケージ構造詳細

### 4.1. `cmd/ezs2t-whisper/main.go`
**役割**: アプリケーションのエントリーポイント

**主要機能:**
- 各パッケージの初期化
- ゴルーチンの起動と管理
- グレースフルシャットダウン

**依存パッケージ:**
- すべてのinternalパッケージ

### 4.2. `internal/hotkey/`
**役割**: グローバルホットキーの検出と登録

**主要機能:**
- ホットキーの登録
- イベントの監視
- 競合チェック

**外部依存:**
- `github.com/golang-design/hotkey`

### 4.3. `internal/audio/`
**役割**: オーディオ入力の抽象化

**主要機能:**
- AudioDriverインターフェースの定義
- PortAudio実装
- デバイス列挙

**外部依存:**
- `github.com/gordonklaus/portaudio`

### 4.4. `internal/recording/`
**役割**: 音声録音ロジックと状態管理

**主要機能:**
- 録音状態管理
- 録音モード（押下中/トグル）
- 最大録音時間の制限

**依存パッケージ:**
- `internal/audio`
- `internal/hotkey`

### 4.5. `internal/recognition/`
**役割**: Whisper.cppとの統合

**主要機能:**
- モデルのロード
- 文字起こし実行
- CGO連携

**外部依存:**
- Whisper.cpp（CGO）

### 4.6. `internal/clipboard/`
**役割**: changeCount方式による安全なテキスト挿入

**主要機能:**
- changeCountの取得（CGO）
- クリップボード操作
- キーイベント送信

**外部依存:**
- `github.com/go-vgo/robotgo`
- NSPasteboard（CGO）

### 4.7. `internal/tray/`
**役割**: システムトレイ統合

**主要機能:**
- メニューバーアイコンの管理
- メニュー項目の実装
- 状態の可視化

**外部依存:**
- `github.com/getlantern/systray`

### 4.8. `internal/server/`
**役割**: ローカルHTTPサーバー

**主要機能:**
- HTTPサーバーの起動
- 静的ファイルの提供
- APIルーティング

**外部依存:**
- Go標準ライブラリ（`net/http`、`embed`）

### 4.9. `internal/api/`
**役割**: REST APIの実装

**主要機能:**
- 9つのエンドポイントの実装
- リクエストバリデーション
- エラーハンドリング

**依存パッケージ:**
- `internal/config`
- `internal/audio`
- `internal/recognition`
- `internal/permissions`

### 4.10. `internal/config/`
**役割**: 設定の管理と永続化

**主要機能:**
- 設定構造体の定義
- JSON形式での保存・読み込み
- デフォルト設定の生成

**外部依存:**
- Go標準ライブラリ（`encoding/json`）

### 4.11. `internal/i18n/`
**役割**: 多言語対応

**主要機能:**
- 文言IDベースの辞書管理
- 言語切り替え
- ローカライズ

**外部依存:**
- Go標準ライブラリ

### 4.12. `internal/permissions/`
**役割**: macOSシステム権限のチェック

**主要機能:**
- マイク権限チェック
- アクセシビリティ権限チェック
- システム設定への誘導

**外部依存:**
- AVFoundation（CGO）
- Accessibility API（CGO）

### 4.13. `internal/logger/`
**役割**: ログ出力

**主要機能:**
- ログレベル管理
- ファイル出力
- ログローテーション

**外部依存:**
- Go標準ライブラリ

---

## 5. テスト計画

### 5.1. ユニットテスト

#### `internal/hotkey/`
- [ ] ホットキー登録のテスト
- [ ] イベント検出のテスト
- [ ] 競合チェックのテスト

#### `internal/audio/`
- [ ] デバイス列挙のテスト
- [ ] 録音開始/停止のテスト
- [ ] エラーハンドリングのテスト

#### `internal/recording/`
- [ ] 録音状態管理のテスト
- [ ] 録音モードのテスト
- [ ] 最大録音時間のテスト

#### `internal/clipboard/`
- [ ] changeCount取得のテスト
- [ ] クリップボード操作のテスト
- [ ] 復元ロジックのテスト

#### `internal/config/`
- [ ] 設定保存・読み込みのテスト
- [ ] デフォルト設定のテスト
- [ ] バリデーションのテスト

### 5.2. 統合テスト

#### エンドツーエンドテスト
- [ ] ホットキー押下→録音→文字起こし→貼り付けの全体フロー
- [ ] 設定変更→アプリ再起動→設定反映の確認
- [ ] 権限未付与時の挙動

#### パフォーマンステスト
- [ ] RTF測定（複数回実行）
- [ ] メモリ使用量の監視
- [ ] 連続実行テスト（100回）

#### ストレステスト
- [ ] 長時間録音（60秒）
- [ ] 頻繁なホットキー押下
- [ ] 大量のテキスト貼り付け

---

## 6. トラブルシューティング

### 6.1. ビルドエラー

#### エラー: `portaudio.h: No such file or directory`
**原因**: PortAudioがインストールされていない
**解決策**:
```bash
brew install portaudio
```

#### エラー: `ld: library not found for -lwhisper`
**原因**: Whisper.cppがビルドされていない
**解決策**:
```bash
cd whisper.cpp
make
```

#### エラー: CGOコンパイルエラー
**原因**: Xcodeコマンドラインツールが不足
**解決策**:
```bash
xcode-select --install
```

### 6.2. 実行時エラー

#### エラー: ホットキーが動作しない
**原因**: アクセシビリティ権限が未付与
**解決策**: システム設定 → プライバシーとセキュリティ → アクセシビリティ

#### エラー: 録音が開始されない
**原因**: マイク権限が未付与
**解決策**: システム設定 → プライバシーとセキュリティ → マイク

#### エラー: テキストが貼り付けられない
**原因**: robotgoがアクセシビリティ権限を持っていない
**解決策**: アプリを再起動し、権限を再確認

### 6.3. パフォーマンス問題

#### 問題: 文字起こしが遅い
**原因**: モデルが大きすぎる
**解決策**: 軽量モデル（`ggml-small-q5_1.gguf`）に切り替え

#### 問題: メモリ使用量が多い
**原因**: モデルキャッシュが大きい
**解決策**: より小さいモデルを使用

---

## 7. 依存関係一覧

### 7.1. Go パッケージ

| パッケージ | バージョン | 用途 |
|-----------|----------|------|
| `github.com/golang-design/hotkey` | latest | グローバルホットキー |
| `github.com/gordonklaus/portaudio` | latest | オーディオ入力 |
| `github.com/go-vgo/robotgo` | latest | クリップボード、キーイベント |
| `github.com/getlantern/systray` | latest | システムトレイ |

### 7.2. システム依存関係

| ツール | インストール方法 | 用途 |
|-------|---------------|------|
| Xcode Command Line Tools | `xcode-select --install` | CGOコンパイル |
| libpng | `brew install libpng` | robotgo依存 |
| libjpeg | `brew install libjpeg` | robotgo依存 |
| portaudio | `brew install portaudio` | オーディオ入力 |

### 7.3. Whisper.cpp

```bash
# クローンとビルド
git clone https://github.com/ggerganov/whisper.cpp.git
cd whisper.cpp
make
```

---

## 8. 開発の優先順位

### 最優先（コア機能）
1. プロジェクト初期化とディレクトリ構造
2. ホットキー検出 + 音声録音
3. Whisper.cpp統合
4. changeCount方式のテキスト挿入
5. メインフローの統合

### 次優先（UI/UX）
6. Systrayメニュー
7. Web設定画面
8. REST API

### 仕上げ（運用機能）
9. 設定永続化
10. 多言語対応
11. 権限処理
12. ロギング
13. 初回セットアップウィザード

---

## 9. 参考リンク

### 公式ドキュメント
- [Whisper.cpp](https://github.com/ggerganov/whisper.cpp)
- [Go Documentation](https://golang.org/doc)
- [macOS Developer Documentation](https://developer.apple.com/documentation)

### ライブラリ
- [golang-design/hotkey](https://github.com/golang-design/hotkey)
- [gordonklaus/portaudio](https://github.com/gordonklaus/portaudio)
- [go-vgo/robotgo](https://github.com/go-vgo/robotgo)
- [getlantern/systray](https://github.com/getlantern/systray)

### チュートリアル
- [Go CGO Programming](https://golang.org/cmd/cgo/)
- [macOS Permissions Guide](https://developer.apple.com/documentation/avfoundation/cameras_and_media_capture/requesting_authorization_for_media_capture_on_macos)

---

## 10. プロジェクト管理

### 進捗トラッキング
このドキュメントのチェックボックスを使って進捗を管理します。

### ブランチ戦略
- `main`: 安定版
- `develop`: 開発版
- `feature/*`: 機能ブランチ

### コミットメッセージ規約
```
<type>(<scope>): <subject>

Types:
- feat: 新機能
- fix: バグ修正
- docs: ドキュメント
- style: コードフォーマット
- refactor: リファクタリング
- test: テスト追加
- chore: その他
```

### コードレビュー
- 重要な変更は別ブランチで実装
- 自己レビュー後にマージ

---

## 11. 次のステップ

### 今すぐ始められること
1. [ ] Go module初期化（`go mod init`）
2. [ ] ディレクトリ構造の作成
3. [ ] 依存パッケージのインストール
4. [ ] `main.go`の基本構造作成

### Week 1の開始
Week 1の最初のタスクから順に実装を開始します。

---

**作成日**: 2025-10-28
**バージョン**: 1.0
**作成者**: Claude Code
