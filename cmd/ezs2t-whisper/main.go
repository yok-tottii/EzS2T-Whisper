package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/yok-tottii/EzS2T-Whisper/internal/api"
	"github.com/yok-tottii/EzS2T-Whisper/internal/audio"
	"github.com/yok-tottii/EzS2T-Whisper/internal/clipboard"
	"github.com/yok-tottii/EzS2T-Whisper/internal/config"
	"github.com/yok-tottii/EzS2T-Whisper/internal/hotkey"
	"github.com/yok-tottii/EzS2T-Whisper/internal/logger"
	"github.com/yok-tottii/EzS2T-Whisper/internal/permissions"
	"github.com/yok-tottii/EzS2T-Whisper/internal/recognition"
	"github.com/yok-tottii/EzS2T-Whisper/internal/server"
	"github.com/yok-tottii/EzS2T-Whisper/internal/tray"
	"github.com/yok-tottii/EzS2T-Whisper/internal/wizard"
	hk "golang.design/x/hotkey"
)

const version = "0.3.0"

// App holds all application state
type App struct {
	logger      *logger.Logger
	config      *config.Config
	trayMgr     *tray.Manager
	httpServer  *server.Server
	apiHandler  *api.Handler
	hotkeyMgr   *hotkey.Manager
	audioDriver audio.AudioDriver
	audioConfig audio.Config
	recognizer  *recognition.WhisperRecognizer
	clipboard   *clipboard.Manager
	wizard      *wizard.SetupWizard

	micGranted  bool
	accGranted  bool
	modelLoaded bool
	isFirstRun  bool
}

func init() {
	// macOSのCGO呼び出しにはメインスレッドが必要
	runtime.LockOSThread()
}

func main() {
	app := &App{}

	// ロガーの初期化
	loggerConfig := logger.DefaultConfig()
	var err error
	app.logger, err = logger.New(loggerConfig)
	if err != nil {
		log.Fatalf("ロガーの初期化に失敗: %v", err)
	}
	defer app.logger.Close()

	app.logger.Info("EzS2T-Whisper v%s 起動", version)

	// 設定ファイルの読み込み
	configPath := config.GetConfigPath()
	app.config, err = config.Load(configPath)
	if err != nil {
		app.logger.Error("設定ファイルの読み込みに失敗: %v", err)
		log.Fatalf("設定ファイルの読み込みに失敗: %v", err)
	}
	app.logger.Info("設定ファイルを読み込みました: %s", configPath)

	// セットアップウィザード初期化
	app.wizard, err = wizard.NewSetupWizard()
	if err != nil {
		app.logger.Error("セットアップウィザード初期化エラー: %v", err)
	}

	// 初回起動判定
	app.isFirstRun = app.wizard != nil && app.wizard.ShouldShowWizard()

	// Clipboard Managerの初期化
	app.clipboard = clipboard.NewManager(clipboard.DefaultConfig())
	app.logger.Info("Clipboard Manager初期化完了")

	// Whisper Recognizerの初期化
	app.recognizer = recognition.NewWhisperRecognizer(recognition.DefaultConfig())
	defer app.recognizer.Close()

	// HTTPサーバーの初期化
	app.httpServer = server.New(server.DefaultConfig())
	app.apiHandler = api.New(app.config, app.wizard, app.ReloadHotkey)

	// APIルートを登録
	app.apiHandler.RegisterRoutes(app.httpServer.GetMux())
	app.logger.Info("APIルート登録完了")

	// システムトレイマネージャーの作成
	app.trayMgr = tray.NewManager(tray.Config{
		OnReady:        app.onReady,
		OnSettings:     app.handleOpenSettings,
		OnRescanModels: app.handleRescanModels,
		OnRecordTest:   app.handleRecordTest,
		OnAbout:        app.handleAbout,
		OnQuit:         app.handleQuit,
	})

	app.logger.Info("systray初期化開始")

	// systray.Run()を呼び出し - これはブロッキング呼び出し
	app.trayMgr.Run()
}

// onReady は systray が初期化完了後に呼ばれる
func (a *App) onReady() {
	a.logger.Info("systray初期化完了 - アプリケーション初期化開始")

	// 権限チェック
	permChecker := permissions.NewPermissionChecker()
	perms := permChecker.CheckAllPermissions()

	a.micGranted = perms["microphone"]
	a.accGranted = perms["accessibility"]

	if a.micGranted {
		a.logger.Info("マイク権限: 許可済み")
	} else {
		a.logger.Warn("マイク権限: 未許可 - 録音機能が無効化されます")
		a.trayMgr.ShowError("マイク権限が未許可です。システム設定で許可してください。")
	}

	if a.accGranted {
		a.logger.Info("アクセシビリティ権限: 許可済み")
	} else {
		a.logger.Warn("アクセシビリティ権限: 未許可 - ホットキーと貼り付け機能が無効化されます")
		a.trayMgr.ShowError("アクセシビリティ権限が未許可です。システム設定で許可してください。")
	}

	// モデルのロード（モデルパスが設定されている場合）
	if a.config.ModelPath != "" {
		modelPath, err := a.config.GetModelPath()
		if err != nil {
			a.logger.Error("モデルパスの展開に失敗: %v", err)
		} else if err := a.config.ValidateModelPath(); err != nil {
			a.logger.Warn("モデルパスの検証に失敗: %v", err)
		} else {
			a.logger.Info("モデルをロード中: %s", modelPath)
			if err := a.recognizer.LoadModel(modelPath); err != nil {
				a.logger.Warn("モデルのロードに失敗: %v", err)
				a.trayMgr.ShowError(fmt.Sprintf("モデルのロードに失敗: %v", err))
			} else {
				a.logger.Info("モデルロード完了")
				a.modelLoaded = true
			}
		}
	} else {
		a.logger.Warn("モデルパスが設定されていません")
	}

	// オーディオドライバの初期化（マイク権限がある場合のみ）
	if a.micGranted {
		var err error
		a.audioDriver, err = audio.NewPortAudioDriver()
		if err != nil {
			a.logger.Error("PortAudioドライバの作成に失敗: %v", err)
		} else {
			a.audioConfig = audio.DefaultConfig()
			// 設定ファイルのデバイスIDを反映（-1の場合はシステムデフォルト）
			a.audioConfig.DeviceID = a.config.AudioDeviceID
			a.logger.Info("設定からオーディオデバイスIDを適用: %d", a.config.AudioDeviceID)
			if err := a.audioDriver.Initialize(a.audioConfig); err != nil {
				a.logger.Error("オーディオドライバの初期化に失敗: %v", err)
			} else {
				a.logger.Info("オーディオドライバ初期化完了")
				// API HandlerにAudioDriverを設定
				a.apiHandler.SetAudioDriver(a.audioDriver)
			}
		}
	}

	// ホットキーマネージャーの初期化（アクセシビリティ権限がある場合のみ）
	if a.accGranted {
		a.hotkeyMgr = hotkey.New()

		// 設定ファイルからホットキー設定を読み込み
		hotkeyConfig := hotkey.Config{
			Modifiers: configToModifiers(a.config.Hotkey),
			Key:       stringToKey(a.config.Hotkey.Key),
			Mode:      hotkey.PressToHold, // TODO: RecordingModeから決定
		}

		// ホットキーの登録
		if err := a.hotkeyMgr.Register(hotkeyConfig); err != nil {
			a.logger.Error("ホットキーの登録に失敗: %v", err)
			a.trayMgr.ShowError(fmt.Sprintf("ホットキーの登録に失敗: %v", err))
		} else {
			hotkeyFormatted := hotkey.FormatHotkey(hotkeyConfig.Modifiers, hotkeyConfig.Key)
			a.logger.Info("ホットキー登録完了: %s", hotkeyFormatted)

			// ホットキーイベントループを開始
			go a.hotkeyEventLoop()
		}
	}

	// 初回起動時は自動的にセットアップ画面を開く
	if a.isFirstRun && a.wizard != nil {
		a.logger.Info("初回起動検出 - セットアップ画面を開きます")
		a.handleOpenSettings()
		// MarkSetupCompleted()はAPIハンドラで設定保存時に呼ばれる
	}

	a.logger.Info("アプリケーション初期化完了")

	// HTTPサーバーを起動
	if err := a.httpServer.Start(); err != nil {
		a.logger.Error("HTTPサーバーの起動に失敗: %v", err)
		a.trayMgr.ShowError("設定画面の起動に失敗しました")
	}

	// シグナルハンドリングを設定（Ctrl+Cでの適切な終了処理）
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		a.logger.Info("終了シグナルを受信しました")
		a.handleQuit()
		a.trayMgr.Quit() // systray.Quit()を呼び出してsystray.Run()を終了
	}()

	// ターミナルに設定画面URLを常に表示
	fmt.Println("\n" + "==========================================================")
	fmt.Println("[起動] EzS2T-Whisper が起動しました")
	fmt.Println("==========================================================")
	fmt.Printf("[設定] 設定画面URL: %s\n", a.httpServer.URL())
	fmt.Printf("[操作] メニューバーのアイコンをクリックしてメニューを開けます\n")

	// 現在のホットキー設定を表示
	currentHotkey := a.hotkeyMgr.GetConfig()
	hotkeyDisplay := hotkey.FormatHotkey(currentHotkey.Modifiers, currentHotkey.Key)
	fmt.Printf("[設定] ホットキー: %s\n", hotkeyDisplay)

	fmt.Printf("[終了] Ctrl+C またはメニューから「終了」\n")
	fmt.Println("==========================================================" + "\n")
}

// hotkeyEventLoop はホットキーイベントを処理するループ
func (a *App) hotkeyEventLoop() {
	a.logger.Info("ホットキーイベントループ開始")

	eventChan := a.hotkeyMgr.Events()

	for event := range eventChan {
		switch event.Type {
		case hotkey.Pressed:
			if !a.micGranted || a.audioDriver == nil {
				a.logger.Warn("ホットキー押下検出しましたが、マイク権限がないため無視します")
				continue
			}

			a.logger.Info("ホットキー押下検出 - 録音開始")
			a.trayMgr.SetState(tray.StateRecording)

			if err := a.audioDriver.StartRecording(); err != nil {
				a.logger.Error("録音開始エラー: %v", err)
				a.trayMgr.ShowError(fmt.Sprintf("録音開始に失敗: %v", err))
				a.trayMgr.SetState(tray.StateIdle)
			}

		case hotkey.Released:
			if !a.micGranted || a.audioDriver == nil {
				continue
			}

			a.logger.Info("ホットキー解放検出 - 録音停止")
			a.trayMgr.SetState(tray.StateProcessing)

			audioData, err := a.audioDriver.StopRecording()
			if err != nil {
				a.logger.Error("録音停止エラー: %v", err)
				a.trayMgr.ShowError(fmt.Sprintf("録音停止に失敗: %v", err))
				a.trayMgr.SetState(tray.StateIdle)
				continue
			}

			dataSize := len(audioData)
			a.logger.Info("録音データ受信: %d バイト", dataSize)

			// データが空の場合はスキップ
			if dataSize == 0 {
				a.logger.Warn("録音データが空です")
				a.trayMgr.SetState(tray.StateIdle)
				continue
			}

			// モデルがない場合はスキップ
			if !a.modelLoaded {
				a.logger.Warn("モデル未読み込みのため文字起こしをスキップ")
				a.trayMgr.ShowError("モデルが読み込まれていません。設定画面でモデルを選択してください。")
				a.trayMgr.SetState(tray.StateIdle)
				continue
			}

			// 文字起こし処理
			a.logger.Info("文字起こし処理開始")

			transcription, err := a.recognizer.Transcribe(audioData, a.audioConfig.SampleRate)
			if err != nil {
				a.logger.Error("文字起こしエラー: %v", err)
				a.trayMgr.ShowError(fmt.Sprintf("文字起こしに失敗: %v", err))
				a.trayMgr.SetState(tray.StateIdle)
				continue
			}

			a.logger.Info("文字起こし完了: %s", transcription)

			// 文字起こし結果が空の場合はスキップ
			if transcription == "" {
				a.logger.Warn("文字起こし結果が空です")
				a.trayMgr.SetState(tray.StateIdle)
				continue
			}

			// クリップボードに貼り付け（アクセシビリティ権限が必要）
			if !a.accGranted {
				a.logger.Warn("アクセシビリティ権限なしのため貼り付けをスキップ")
				a.trayMgr.ShowError("アクセシビリティ権限がありません。システム設定で許可してください。")
				a.trayMgr.SetState(tray.StateIdle)
				continue
			}

			a.logger.Info("クリップボード貼り付け開始")

			if err := a.clipboard.SafePasteWithSplit(transcription); err != nil {
				a.logger.Error("貼り付けエラー: %v", err)
				a.trayMgr.ShowError(fmt.Sprintf("貼り付けに失敗: %v", err))
				a.trayMgr.SetState(tray.StateIdle)
				continue
			}

			a.logger.Info("貼り付け完了")
			a.trayMgr.SetState(tray.StateIdle)
		}
	}

	a.logger.Info("ホットキーイベントループ終了")
}

// handleOpenSettings は設定画面を開く
func (a *App) handleOpenSettings() {
	a.logger.Info("設定画面を開く要求")

	// サーバーが起動していない場合はエラー
	if !a.httpServer.IsRunning() {
		a.logger.Error("HTTPサーバーが起動していません")
		a.trayMgr.ShowError("設定画面が利用できません。アプリケーションを再起動してください。")
		return
	}

	// ブラウザで設定画面を開く
	url := a.httpServer.URL()
	a.logger.Info("ブラウザを開きます: %s", url)

	// goroutineで非同期実行
	go func() {
		cmd := exec.Command("open", url)
		if err := cmd.Run(); err != nil {
			a.logger.Error("ブラウザの起動に失敗: %v", err)
			a.trayMgr.ShowError(fmt.Sprintf("ブラウザの起動に失敗: %v", err))

			// フォールバック: ターミナルにURLを表示
			fmt.Printf("\n[警告] ブラウザが自動で開きませんでした\n")
			fmt.Printf("[情報] 設定画面URL: %s\n", url)
			fmt.Printf("[ヒント] 上記URLをブラウザで開いてください\n\n")
		}
	}()
}

// handleRescanModels はモデルディレクトリを再スキャン
func (a *App) handleRescanModels() {
	a.logger.Info("モデル再スキャン要求")
	a.trayMgr.ShowNotification("モデル再スキャン", "モデルディレクトリを再スキャンしています...")
	// TODO: 実装
}

// handleRecordTest は録音テストを実行
func (a *App) handleRecordTest() {
	a.logger.Info("録音テスト要求")

	// goroutineで非同期実行（UIブロックを防ぐ）
	go func() {
		// 1. 権限チェック
		if !a.micGranted {
			a.logger.Warn("録音テスト: マイク権限がありません")
			a.trayMgr.ShowError("マイク権限がありません。システム設定で許可してください。")
			return
		}

		if !a.accGranted {
			a.logger.Warn("録音テスト: アクセシビリティ権限がありません")
			a.trayMgr.ShowError("アクセシビリティ権限がありません。システム設定で許可してください。")
			return
		}

		if a.audioDriver == nil {
			a.logger.Error("録音テスト: オーディオドライバが初期化されていません")
			a.trayMgr.ShowError("オーディオドライバの初期化に失敗しています。")
			return
		}

		if !a.modelLoaded {
			a.logger.Warn("録音テスト: モデルが読み込まれていません")
			a.trayMgr.ShowError("モデルが読み込まれていません。設定画面でモデルを選択してください。")
			return
		}

		// 2. 録音開始
		a.logger.Info("録音テスト: 録音開始（5秒間）")
		a.trayMgr.ShowNotification("録音テスト", "録音を開始します（5秒間話してください）")
		a.trayMgr.SetState(tray.StateRecording)

		if err := a.audioDriver.StartRecording(); err != nil {
			a.logger.Error("録音テスト: 録音開始エラー: %v", err)
			a.trayMgr.ShowError(fmt.Sprintf("録音開始に失敗: %v", err))
			a.trayMgr.SetState(tray.StateIdle)
			return
		}

		// 3. 5秒間録音
		time.Sleep(5 * time.Second)

		// 4. 録音停止
		a.logger.Info("録音テスト: 録音停止")
		a.trayMgr.SetState(tray.StateProcessing)

		audioData, err := a.audioDriver.StopRecording()
		if err != nil {
			a.logger.Error("録音テスト: 録音停止エラー: %v", err)
			a.trayMgr.ShowError(fmt.Sprintf("録音停止に失敗: %v", err))
			a.trayMgr.SetState(tray.StateIdle)
			return
		}

		dataSize := len(audioData)
		a.logger.Info("録音テスト: 録音データ受信: %d バイト", dataSize)

		// データが空の場合
		if dataSize == 0 {
			a.logger.Warn("録音テスト: 録音データが空です")
			a.trayMgr.ShowError("録音データが空です。マイクが正しく動作しているか確認してください。")
			a.trayMgr.SetState(tray.StateIdle)
			return
		}

		// 5. 文字起こし処理
		a.logger.Info("録音テスト: 文字起こし処理開始")
		a.trayMgr.ShowNotification("録音テスト", "文字起こし処理中...")

		transcription, err := a.recognizer.Transcribe(audioData, a.audioConfig.SampleRate)
		if err != nil {
			a.logger.Error("録音テスト: 文字起こしエラー: %v", err)
			a.trayMgr.ShowError(fmt.Sprintf("文字起こしに失敗: %v", err))
			a.trayMgr.SetState(tray.StateIdle)
			return
		}

		a.logger.Info("録音テスト: 文字起こし完了: %s", transcription)

		// 文字起こし結果が空の場合
		if transcription == "" {
			a.logger.Warn("録音テスト: 文字起こし結果が空です")
			a.trayMgr.ShowError("文字起こし結果が空です。音声が短すぎるか、ノイズが多い可能性があります。")
			a.trayMgr.SetState(tray.StateIdle)
			return
		}

		// 6. 結果を通知
		a.logger.Info("録音テスト: テスト完了")
		a.trayMgr.ShowNotification("録音テスト完了", fmt.Sprintf("文字起こし結果:\n%s", transcription))
		a.trayMgr.SetState(tray.StateIdle)
	}()
}

// handleAbout はバージョン情報を表示
func (a *App) handleAbout() {
	a.logger.Info("バージョン情報表示要求")

	// バージョン情報をダイアログで表示
	info := []string{
		"EzS2T-Whisper",
		"",
		fmt.Sprintf("Version: %s", version),
		"",
		"高速ローカル音声認識アプリケーション",
		"",
		"Copyright © 2025 yoktotti",
		"MIT License",
	}

	dialogText := strings.Join(info, "\\n")
	script := fmt.Sprintf(`display dialog "%s" buttons {"OK"} default button "OK" with title "バージョン情報"`, dialogText)

	// goroutineで非同期実行（UIブロックを防ぐ）
	go exec.Command("osascript", "-e", script).Run()
}

// handleQuit はアプリケーションを終了
func (a *App) handleQuit() {
	a.logger.Info("終了要求")

	// HTTPサーバーを停止
	if a.httpServer != nil && a.httpServer.IsRunning() {
		if err := a.httpServer.Stop(); err != nil {
			a.logger.Error("HTTPサーバーの停止に失敗: %v", err)
		}
	}

	// ホットキーマネージャーをクローズ
	if a.hotkeyMgr != nil {
		a.hotkeyMgr.Close()
	}

	// オーディオドライバをクローズ
	if a.audioDriver != nil {
		a.audioDriver.Close()
	}

	a.logger.Info("アプリケーション終了")
}

// ReloadHotkey は設定ファイルから読み込んだ内容で、ホットキーを再登録する
func (a *App) ReloadHotkey() error {
	a.logger.Info("ホットキー再登録要求")

	// 権限チェック
	if !a.accGranted {
		a.logger.Warn("ホットキー再登録: アクセシビリティ権限がありません")
		return fmt.Errorf("accessibility permission not granted")
	}

	if a.hotkeyMgr == nil {
		a.logger.Warn("ホットキー再登録: ホットキーマネージャーが初期化されていません")
		return fmt.Errorf("hotkey manager not initialized")
	}

	// 設定ファイルを再読み込み（最新の設定を取得）
	configPath := config.GetConfigPath()
	freshConfig, err := config.Load(configPath)
	if err != nil {
		a.logger.Error("設定ファイルの再読み込みに失敗: %v", err)
		return fmt.Errorf("failed to reload config: %w", err)
	}

	// 新しいホットキー設定を作成
	newConfig := hotkey.Config{
		Modifiers: configToModifiers(freshConfig.Hotkey),
		Key:       stringToKey(freshConfig.Hotkey.Key),
		Mode:      hotkey.PressToHold, // TODO: RecordingModeから決定
	}

	a.logger.Info("新しいホットキー設定: Modifiers=%v, Key=%v", newConfig.Modifiers, newConfig.Key)

	// 既存の設定をバックアップ（ロールバック用）
	var oldConfig hotkey.Config
	needsRollback := false

	// 既存のホットキーを解除
	if a.hotkeyMgr.IsRunning() {
		a.logger.Info("既存のホットキーを解除します")
		oldConfig = a.hotkeyMgr.GetConfig()
		needsRollback = true

		if err := a.hotkeyMgr.Close(); err != nil {
			a.logger.Error("既存のホットキー解除に失敗: %v", err)
			return fmt.Errorf("failed to unregister old hotkey: %w", err)
		}
		// イベントループが完全に終了するまで待機
		time.Sleep(200 * time.Millisecond)
	}

	// 新しいホットキーを登録
	a.logger.Info("新しいホットキーを登録します")
	if err := a.hotkeyMgr.Register(newConfig); err != nil {
		a.logger.Error("新しいホットキー登録に失敗: %v", err)

		// ロールバック: 旧ホットキーを再登録
		if needsRollback {
			a.logger.Warn("ロールバック: 旧ホットキーを再登録します")
			if rollbackErr := a.hotkeyMgr.Register(oldConfig); rollbackErr != nil {
				a.logger.Error("ロールバック失敗: %v", rollbackErr)
				a.trayMgr.ShowError("ホットキーの登録に失敗しました。アプリケーションを再起動してください。")
				return fmt.Errorf("failed to register new hotkey and rollback failed: %w, rollback error: %v", err, rollbackErr)
			}
			go a.hotkeyEventLoop()
			a.logger.Info("ロールバック完了")
		}

		return fmt.Errorf("failed to register new hotkey: %w", err)
	}

	// イベントループを再起動
	go a.hotkeyEventLoop()

	// アプリケーションの設定を更新
	a.config = freshConfig

	hotkeyFormatted := hotkey.FormatHotkey(newConfig.Modifiers, newConfig.Key)
	a.logger.Info("ホットキー再登録完了: %s", hotkeyFormatted)
	a.trayMgr.ShowNotification("ホットキー変更", fmt.Sprintf("新しいホットキー: %s", hotkeyFormatted))

	return nil
}

// configToModifiers は HotkeyConfig を golang.design/x/hotkey の Modifier スライスに変換
func configToModifiers(hkConfig config.HotkeyConfig) []hk.Modifier {
	var mods []hk.Modifier
	if hkConfig.Ctrl {
		mods = append(mods, hk.ModCtrl)
	}
	if hkConfig.Shift {
		mods = append(mods, hk.ModShift)
	}
	if hkConfig.Alt {
		mods = append(mods, hk.ModOption)
	}
	if hkConfig.Cmd {
		mods = append(mods, hk.ModCmd)
	}
	return mods
}

// stringToKey は文字列をキーコードに変換
func stringToKey(keyStr string) hk.Key {
	keyMap := map[string]hk.Key{
		"Space":  hk.KeySpace,
		"A":      hk.KeyA,
		"B":      hk.KeyB,
		"C":      hk.KeyC,
		"D":      hk.KeyD,
		"E":      hk.KeyE,
		"F":      hk.KeyF,
		"G":      hk.KeyG,
		"H":      hk.KeyH,
		"I":      hk.KeyI,
		"J":      hk.KeyJ,
		"K":      hk.KeyK,
		"L":      hk.KeyL,
		"M":      hk.KeyM,
		"N":      hk.KeyN,
		"O":      hk.KeyO,
		"P":      hk.KeyP,
		"Q":      hk.KeyQ,
		"R":      hk.KeyR,
		"S":      hk.KeyS,
		"T":      hk.KeyT,
		"U":      hk.KeyU,
		"V":      hk.KeyV,
		"W":      hk.KeyW,
		"X":      hk.KeyX,
		"Y":      hk.KeyY,
		"Z":      hk.KeyZ,
		"0":      hk.Key0,
		"1":      hk.Key1,
		"2":      hk.Key2,
		"3":      hk.Key3,
		"4":      hk.Key4,
		"5":      hk.Key5,
		"6":      hk.Key6,
		"7":      hk.Key7,
		"8":      hk.Key8,
		"9":      hk.Key9,
		"Escape": hk.KeyEscape,
		"Return": hk.KeyReturn,
		"Tab":    hk.KeyTab,
	}

	if key, ok := keyMap[keyStr]; ok {
		return key
	}

	// デフォルトはSpace
	return hk.KeySpace
}
