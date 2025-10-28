package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/YOURUSERNAME/EzS2T-Whisper/internal/audio"
	"github.com/YOURUSERNAME/EzS2T-Whisper/internal/clipboard"
	"github.com/YOURUSERNAME/EzS2T-Whisper/internal/hotkey"
	"github.com/YOURUSERNAME/EzS2T-Whisper/internal/logger"
	"github.com/YOURUSERNAME/EzS2T-Whisper/internal/recognition"
	hk "golang.design/x/hotkey"
)

func init() {
	// macOSのCGO呼び出しにはメインスレッドが必要
	runtime.LockOSThread()
}

func main() {
	fmt.Println("EzS2T-Whisper v0.2.0 (Week 2 統合ビルド)")
	fmt.Println("=========================================")

	// ロガーの初期化
	loggerConfig := logger.DefaultConfig()
	appLogger, err := logger.New(loggerConfig)
	if err != nil {
		log.Fatalf("ロガーの初期化に失敗: %v", err)
	}
	defer appLogger.Close()

	appLogger.Info("EzS2T-Whisper起動")

	// オーディオドライバの作成と初期化
	audioDriver, err := audio.NewPortAudioDriver()
	if err != nil {
		appLogger.Error("PortAudioドライバの作成に失敗: %v", err)
		log.Fatalf("PortAudioドライバの作成に失敗: %v", err)
	}
	defer audioDriver.Close()

	audioConfig := audio.DefaultConfig()
	if err := audioDriver.Initialize(audioConfig); err != nil {
		appLogger.Error("オーディオドライバの初期化に失敗: %v", err)
		log.Fatalf("オーディオドライバの初期化に失敗: %v", err)
	}

	appLogger.Info("オーディオドライバ初期化完了")

	// Whisper Recognizerの初期化
	recognizer := recognition.NewWhisperRecognizer(recognition.DefaultConfig())
	defer recognizer.Close()

	// モデルパスの設定（デフォルトモデルを使用）
	homeDir, err := os.UserHomeDir()
	if err != nil {
		appLogger.Error("ホームディレクトリの取得に失敗: %v", err)
		log.Fatalf("ホームディレクトリの取得に失敗: %v", err)
	}
	modelPath := filepath.Join(homeDir, "Library", "Application Support", "EzS2T-Whisper", "models", "ggml-large-v3-turbo-q5_0.gguf")

	// モデルのロード（モデルが存在する場合のみ）
	if _, err := os.Stat(modelPath); err == nil {
		fmt.Printf("モデルをロード中: %s\n", modelPath)
		if err := recognizer.LoadModel(modelPath); err != nil {
			appLogger.Warn("モデルのロードに失敗（スキップ）: %v", err)
			fmt.Printf("[警告] モデルのロードに失敗: %v\n", err)
			fmt.Println("  → 録音データの文字起こしはスキップされます")
		} else {
			appLogger.Info("モデルロード完了")
			fmt.Println("モデルロード完了")
		}
	} else {
		appLogger.Warn("モデルファイルが見つかりません: %s", modelPath)
		fmt.Printf("[警告] モデルファイルが見つかりません: %s\n", modelPath)
		fmt.Println("  → 録音データの文字起こしはスキップされます")
	}

	// Clipboard Managerの初期化
	clipboardManager := clipboard.NewManager(clipboard.DefaultConfig())
	appLogger.Info("Clipboard Manager初期化完了")

	// ホットキーマネージャーの作成
	hotkeyManager := hotkey.New()
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
		log.Fatalf("ホットキーの登録に失敗: %v", err)
	}

	hotkeyFormatted := hotkey.FormatHotkey(hotkeyConfig.Modifiers, hotkeyConfig.Key)
	appLogger.Info("ホットキー登録完了: %s", hotkeyFormatted)
	fmt.Printf("\nホットキー: %s\n", hotkeyFormatted)
	fmt.Println("ホットキーを押すと録音が開始されます")
	fmt.Println("Ctrl+C で終了します")

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
				fmt.Println("[イベント] ホットキー押下 - 録音開始")
				appLogger.Info("ホットキー押下検出 - 録音開始")
				if err := audioDriver.StartRecording(); err != nil {
					appLogger.Error("録音開始エラー: %v", err)
					fmt.Printf("[エラー] 録音開始に失敗: %v\n", err)
				}

			case hotkey.Released:
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

				// クリップボードに貼り付け
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
