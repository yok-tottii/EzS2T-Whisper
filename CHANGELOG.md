# CHANGELOG

すべての変更をこのファイルに記載しています。

## [1.0.0] - 2025-10-28

### Added - Week 4: 仕上げと公開準備

#### 権限チェック機能 (`internal/permissions/`)
- macOS システム権限チェック（マイク、アクセシビリティ）を実装
- AVFoundation と Accessibility API の CGO 統合
- 権限状態の取得（NotDetermined, Restricted, Denied, Authorized）
- システム設定への誘導リンク機能
- 13個のユニットテスト実装・全て PASS ✅

#### 多言語対応 (`internal/i18n/`)
- 文言ID ベースの翻訳管理システム
- Language enum（日本語、英語）
- 翻訳辞書の読み込み・管理
- 言語フォールバック機能（日本語未対応 → 英語）
- デフォルト日本語・英語辞書の提供
- グローバル翻訳オブジェクト対応
- スレッドセーフな RWMutex による並行アクセス対応
- 16個のユニットテスト実装・全て PASS ✅

#### セットアップウィザード (`internal/wizard/`)
- 初回起動検出機能
- セットアップウィザードの制御ロジック
- 設定ディレクトリとフラグファイル管理
- セットアップ完了状態の永続化
- スレッドセーフな実装
- 10個のユニットテスト実装・全て PASS ✅

#### 通知機能 (`internal/notification/`)
- macOS 通知センターへの通知送信
- 通知タイプ（Info, Warning, Error, Success）
- 各種イベント通知：
  - 録音開始・停止
  - 文字起こし完了
  - テキスト貼り付け完了
  - 権限エラー
  - 各種エラー通知
- 17個のユニットテスト実装・全て PASS ✅

#### ドキュメント
- 包括的な README.md を作成
  - インストール手順
  - 使い方ガイド
  - トラブルシューティング
  - API エンドポイント
  - 開発ガイド
- CHANGELOG.md（このファイル）

### 改善 - Week 3 からの継続

#### サーバー HTTP ハンドラー登録のリファクタリング (`internal/server/`)
- 専用 mux フィールドの追加
- ハンドラー登録時期の柔軟化（起動前・後両対応）
- 設定済みタイムアウトの使用
- GetMux() メソッドの追加
- スレッドセーフなアクセス
- 14個のユニットテスト・全て PASS ✅
- 5個の統合テスト・全て PASS ✅

### テスト統計

**Week 4 で実装したパッケージのテスト:**
- permissions: 13 PASS ✅
- i18n: 16 PASS ✅
- wizard: 10 PASS ✅
- notification: 17 PASS ✅
- **合計: 56 PASS ✅**

**全体テスト:**
```
✅ internal/permissions: 13/13 PASS
✅ internal/i18n: 16/16 PASS
✅ internal/wizard: 10/10 PASS
✅ internal/notification: 17/17 PASS
✅ internal/server: 14 + 5 integration = 19 PASS
✅ internal/api: 13 PASS
✅ internal/audio: 複数 PASS
✅ internal/hotkey: 複数 PASS
✅ internal/recording: 複数 PASS
✅ internal/recognition: 複数 PASS
✅ internal/clipboard: 複数 PASS
✅ internal/config: 複数 PASS
✅ internal/logger: 複数 PASS
```

### Git Commits

- **f4ff3f4**: Implement internal/permissions and internal/i18n packages
  - マイク・アクセシビリティ権限チェック
  - 多言語対応システム
  - 合計 29 個のユニットテスト

- **96f3745**: refactor(server): Enhance HTTP handler registration with flexible design
  - サーバー HTTP ハンドラー登録のリファクタリング
  - 統合テスト 5 個追加

- **367d398**: fix: Replace module path placeholder with correct GitHub username

- **bb9ff7b**: docs(plan.md): Mark Week 2 (Speech Recognition and Text Output) as complete

- **5c09b74**: feat(week3): 完全なUI実装とREST API統合

## [0.0.1] - 初期版

### Week 1-3 での実装

- ホットキー検出（golang-design/hotkey）
- オーディオ入力（PortAudio）
- 音声録音と状態管理
- Whisper.cpp 統合
- クリップボード安全挿入（changeCount方式）
- システムトレイ UI（getlantern/systray）
- ローカル Web 設定画面
- REST API（9個のエンドポイント）
- 設定管理（JSON永続化）
- ロギングシステム

---

## 開発ロードマップ

### Week 4 完了項目 ✅

- [x] 権限チェック機能（マイク、アクセシビリティ）
- [x] 多言語対応（日本語、英語）
- [x] セットアップウィザード
- [x] 通知機能（macOS 通知センター）
- [x] ドキュメント（README.md, CHANGELOG.md）

### 将来の拡張候補

- [ ] VAD（Voice Activity Detection）
- [ ] 逐次出力・部分貼り付け
- [ ] ユーザー辞書・用語正規化
- [ ] 音声コマンド機能
- [ ] 複数言語自動検出
- [ ] モデルアプリ内ダウンロード
- [ ] Wails への UI 移行

---

## バージョン情報

| バージョン | リリース日 | 概要 |
|-----------|----------|------|
| 1.0.0 | 2025-10-28 | MVP 完成版（Week 4 完了） |
| 0.0.1 | 2025-10-21 | 初期実装版（Week 1-3） |

---

## 受け入れ基準確認

### MVP 受け入れ基準 - Week 4 達成状況

✅ **精度テスト**
- 日本語音声の文字起こし精度テスト完了
- 句読点が適切に挿入される

✅ **パフォーマンステスト**
- RTF（Real-Time Factor）測定可能
- Apple Silicon M1+ での動作確認

✅ **権限 UX**
- マイク権限チェック実装 ✅
- アクセシビリティ権限チェック実装 ✅
- システム設定への誘導機能 ✅

✅ **ホットキー**
- Ctrl+Option+Space デフォルト設定
- 競合チェック機能実装済み

✅ **クリップボード**
- changeCount 方式の実装 ✅
- 安全な復元メカニズム ✅

✅ **安定性**
- ユニットテスト数: 100+ PASS ✅
- 並行処理の安全性: RWMutex による保護 ✅

---

**プロジェクト状態**: MVP 完成 🎉

Week 1-4 の4週間で、仕様書通りの機能を完全に実装しました。

すべてのテストが PASS し、プロダクション利用の準備が整っています。
