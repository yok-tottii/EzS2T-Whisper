package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/yok-tottii/EzS2T-Whisper/internal/audio"
	"github.com/yok-tottii/EzS2T-Whisper/internal/clipboard"
	"github.com/yok-tottii/EzS2T-Whisper/internal/hotkey"
	"github.com/yok-tottii/EzS2T-Whisper/internal/logger"
	"github.com/yok-tottii/EzS2T-Whisper/internal/permissions"
	"github.com/yok-tottii/EzS2T-Whisper/internal/recognition"
	"github.com/yok-tottii/EzS2T-Whisper/internal/wizard"
	hk "golang.design/x/hotkey"
)

func init() {
	// macOSのCGO呼び出しにはメインスレッドが必要
	runtime.LockOSThread()
}

func main() {
	fmt.Println("EzS2T-Whisper v0.2.0 (Week 4 完成版)")
	fmt.Println("=========================================")

	// ロガーの初期化
	loggerConfig := logger.DefaultConfig()
	appLogger, err := logger.New(loggerConfig)
	if err != nil {
		log.Fatalf("ロガーの初期化に失敗: %v", err)
	}
	defer appLogger.Close()

	appLogger.Info("EzS2T-Whisper起動")

	// セットアップウィザード初期化
	setupWizard, err := wizard.NewSetupWizard()
	if err != nil {
		appLogger.Error("セットアップウィザード初期化エラー: %v", err)
		fmt.Printf("[エラー] セットアップウィザード初期化に失敗: %v\n", err)
	}

	// 初回起動判定
	isFirstRun := setupWizard != nil && setupWizard.ShouldShowWizard()
	if isFirstRun {
		fmt.Println("\n=== 初回セットアップ ===")
		fmt.Println("このアプリケーションを初めて使用しています。")
		fmt.Println("以下の権限が必要です：")
		fmt.Println("  1. マイク: 音声を録音するため")
		fmt.Println("  2. アクセシビリティ: ホットキー登録と文字貼り付けのため")
		fmt.Println("\nシステム設定で以下の手順で許可してください：")
		fmt.Println("  → プライバシーとセキュリティ > マイク")
		fmt.Println("  → プライバシーとセキュリティ > アクセシビリティ")
		fmt.Println("")
	}

	// 権限チェック
	permChecker := permissions.NewPermissionChecker()
	perms := permChecker.CheckAllPermissions()

	// 権限状態の表示
	micGranted := perms["microphone"]
	accGranted := perms["accessibility"]

	fmt.Println("=== 権限状態 ===")
	if micGranted {
		fmt.Println("✅ マイク: 許可済み")
		appLogger.Info("マイク権限: OK")
	} else {
		fmt.Println("❌ マイク: 未許可")
		fmt.Println("  → システム設定で許可してください")
		appLogger.Warn("マイク権限: 未許可 - 録音機能が無効化されます")
	}

	if accGranted {
		fmt.Println("✅ アクセシビリティ: 許可済み")
		appLogger.Info("アクセシビリティ権限: OK")
	} else {
		fmt.Println("❌ アクセシビリティ: 未許可")
		fmt.Println("  → システム設定で許可してください")
		appLogger.Warn("アクセシビリティ権限: 未許可 - ホットキーと貼り付け機能が無効化されます")
	}
	fmt.Println("")

	// Whisper Recognizerの初期化（常に初期化）
	recognizer := recognition.NewWhisperRecognizer(recognition.DefaultConfig())
	defer recognizer.Close()

	// モデルパスの設定
	homeDir, err := os.UserHomeDir()
	if err != nil {
		appLogger.Error("ホームディレクトリの取得に失敗: %v", err)
		fmt.Printf("[エラー] ホームディレクトリが取得できません\n")
		return
	}
	modelPath := filepath.Join(homeDir, "Library", "Application Support", "EzS2T-Whisper", "models", "ggml-large-v3-turbo-q5_0.gguf")

	// モデルのロード（モデルが存在する場合のみ）
	modelLoaded := false
	if _, err := os.Stat(modelPath); err == nil {
		fmt.Printf("モデルをロード中: %s\n", modelPath)
		if err := recognizer.LoadModel(modelPath); err != nil {
			appLogger.Warn("モデルのロードに失敗（スキップ）: %v", err)
			fmt.Printf("[警告] モデルのロードに失敗: %v\n", err)
			fmt.Println("  → 録音データの文字起こしはスキップされます")
		} else {
			appLogger.Info("モデルロード完了")
			fmt.Println("✅ モデルロード完了")
			modelLoaded = true
		}
	} else {
		appLogger.Warn("モデルファイルが見つかりません: %s", modelPath)
		fmt.Printf("[警告] モデルファイルが見つかりません\n")
		fmt.Printf("  → 配置先: %s\n", modelPath)
		fmt.Println("  → モデルをダウンロードして配置してください")
		fmt.Println("  → 録音データの文字起こしはスキップされます")
	}
	fmt.Println("")

	// Clipboard Managerの初期化
	clipboardManager := clipboard.NewManager(clipboard.DefaultConfig())
	appLogger.Info("Clipboard Manager初期化完了")

	// オーディオドライバの初期化（マイク権限がある場合のみ）
	var audioDriver audio.AudioDriver
	var audioConfig audio.Config
	audioAvailable := false

	if micGranted {
		audioDriver, err = audio.NewPortAudioDriver()
		if err != nil {
			appLogger.Error("PortAudioドライバの作成に失敗: %v", err)
			fmt.Printf("[警告] オーディオドライバの初期化に失敗: %v\n", err)
		} else {
			defer audioDriver.Close()
			audioConfig = audio.DefaultConfig()
			if err := audioDriver.Initialize(audioConfig); err != nil {
				appLogger.Error("オーディオドライバの初期化に失敗: %v", err)
				fmt.Printf("[警告] オーディオドライバの初期化に失敗: %v\n", err)
			} else {
				appLogger.Info("オーディオドライバ初期化完了")
				audioAvailable = true
				fmt.Println("✅ オーディオドライバ初期化完了")
			}
		}
	}

	// ホットキーマネージャーの初期化（アクセシビリティ権限がある場合のみ）
	var hotkeyManager *hotkey.Manager
	var hotkeyFormatted string
	hotkeyAvailable := false

	if accGranted {
		hotkeyManager = hotkey.New()
		defer hotkeyManager.Close()

		// ホットキーの設定（Ctrl+Option+Space）
		hotkeyConfig := hotkey.Config{
			Modifiers: []hk.Modifier{hk.ModCtrl, hk.ModOption},
			Key:       hk.KeySpace,
			Mode:      hotkey.PressToHold,
		}

		// ホットキーの登録
		if err := hotkeyManager.Register(hotkeyConfig); err != nil {
			appLogger.Error("ホットキーの登録に失敗: %v", err)
			fmt.Printf("[警告] ホットキーの登録に失敗: %v\n", err)
		} else {
			hotkeyFormatted = hotkey.FormatHotkey(hotkeyConfig.Modifiers, hotkeyConfig.Key)
			appLogger.Info("ホットキー登録完了: %s", hotkeyFormatted)
			fmt.Printf("✅ ホットキー: %s\n", hotkeyFormatted)
			hotkeyAvailable = true
		}
	}

	// 必要な権限がない場合の警告
	if !audioAvailable && !hotkeyAvailable {
		fmt.Println("\n[エラー] 必要な権限がありません。アプリケーションを終了します。")
		fmt.Println("システム設定で以下の権限を許可してください：")
		fmt.Println("  → プライバシーとセキュリティ > マイク")
		fmt.Println("  → プライバシーとセキュリティ > アクセシビリティ")
		appLogger.Error("必要な権限なし - 終了")
		return
	}

	// 初回起動フラグを設定
	if isFirstRun && setupWizard != nil {
		if err := setupWizard.MarkSetupCompleted(); err != nil {
			appLogger.Error("セットアップ完了フラグの設定に失敗: %v", err)
		}
	}

	fmt.Println("=" + "=========================================")
	if hotkeyAvailable {
		fmt.Printf("ホットキーを押すと録音が開始されます\n")
	}
	fmt.Println("Ctrl+C で終了します\n")

	// ホットキーが利用不可の場合は待機モード
	if !hotkeyAvailable {
		fmt.Println("[注意] アクセシビリティ権限がないため、ホットキーは無効化されています。")
		fmt.Println("システム設定で権限を許可してからアプリケーションを再起動してください。")
		fmt.Println("Ctrl+C で終了します...")
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		return
	}

	// ホットキーイベントチャネルの取得
	eventChan := hotkeyManager.Events()

	// シグナルハンドリング
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// イベントループ
	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				appLogger.Info("ホットキーイベントチャネルがクローズされました")
				return
			}

			switch event.Type {
			case hotkey.Pressed:
				if !audioAvailable {
					appLogger.Warn("ホットキー押下検出しましたが、マイク権限がないため無視します")
					continue
				}

				fmt.Println("[イベント] ホットキー押下 - 録音開始")
				appLogger.Info("ホットキー押下検出 - 録音開始")
				if err := audioDriver.StartRecording(); err != nil {
					appLogger.Error("録音開始エラー: %v", err)
					fmt.Printf("[エラー] 録音開始に失敗: %v\n", err)
				}

			case hotkey.Released:
				if !audioAvailable {
					continue
				}

				fmt.Println("[イベント] ホットキー解放 - 録音停止")
				appLogger.Info("ホットキー解放検出 - 録音停止")
				audioData, err := audioDriver.StopRecording()
				if err != nil {
					appLogger.Error("録音停止エラー: %v", err)
					fmt.Printf("[エラー] 録音停止に失敗: %v\n", err)
					continue
				}

				dataSize := len(audioData)
				appLogger.Info("録音データ受信: %d バイト", dataSize)
				fmt.Printf("[データ] 録音完了: %d バイト受信\n", dataSize)

				// データが空の場合はスキップ
				if dataSize == 0 {
					fmt.Println("[警告] 録音データが空です")
					appLogger.Warn("録音データが空です")
					continue
				}

				// モデルがない場合はスキップ
				if !modelLoaded {
					fmt.Println("[警告] モデルが読み込まれていないため、文字起こしをスキップします")
					appLogger.Warn("モデル未読み込みのため文字起こしをスキップ")
					continue
				}

				// 文字起こし処理
				fmt.Println("[処理] 文字起こし中...")
				appLogger.Info("文字起こし処理開始")

				transcription, err := recognizer.Transcribe(audioData, audioConfig.SampleRate)
				if err != nil {
					appLogger.Error("文字起こしエラー: %v", err)
					fmt.Printf("[エラー] 文字起こしに失敗: %v\n", err)
					continue
				}

				appLogger.Info("文字起こし完了: %s", transcription)
				fmt.Printf("[結果] %s\n", transcription)

				// 文字起こし結果が空の場合はスキップ
				if transcription == "" {
					fmt.Println("[警告] 文字起こし結果が空です")
					appLogger.Warn("文字起こし結果が空です")
					continue
				}

				// クリップボードに貼り付け（アクセシビリティ権限が必要）
				if !accGranted {
					fmt.Println("[警告] アクセシビリティ権限がないため、テキストを貼り付けられません")
					appLogger.Warn("アクセシビリティ権限なしのため貼り付けをスキップ")
					continue
				}

				fmt.Println("[処理] テキストを貼り付け中...")
				appLogger.Info("クリップボード貼り付け開始")

				if err := clipboardManager.SafePasteWithSplit(transcription); err != nil {
					appLogger.Error("貼り付けエラー: %v", err)
					fmt.Printf("[エラー] 貼り付けに失敗: %v\n", err)
					continue
				}

				appLogger.Info("貼り付け完了")
				fmt.Println("[完了] テキストが貼り付けられました")
			}

		case <-sigChan:
			fmt.Println("\n\n終了シグナル受信 - クリーンアップ中...")
			appLogger.Info("終了シグナル受信")
			return
		}
	}
}
