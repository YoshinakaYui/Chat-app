package main

import (
	"log"
	"net/http"
	"backend/db" // 正しいパッケージパス
	"os"
	"backend/db" // データベース接続用のパッケージ	
	"github.com/joho/godotenv" // 環境変数の読み込み
)

// init関数：環境変数をロード
func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("環境変数の読み込みに失敗しました: %v", err)
	}
	log.Println("環境変数の読み込み成功")
}

// CORS対応を設定する関数
func enableCORS(w http.ResponseWriter) {
	// "誰でもアクセスしていいよ"と設定
    w.Header().Set("Access-Control-Allow-Origin", "*")
	// どんな操作ができるかを教える
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	// どんな情報を送れるかを教える
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// リクエストが来たときに呼ばれる関数
func handler(w http.ResponseWriter, r *http.Request) {
	// CORS対応を有効にする
    enableCORS(w)
    w.Write([]byte("Hello from backend"))
}

// メイン関数：サーバーを起動する
func main() {
	// データベース接続確認
	if err := db.Connect(); err != nil {
		log.Fatal("DB接続失敗:", err)
	}
	log.Println("DB接続成功！")

	// HTTPリクエストを処理するハンドラー関数
	http.HandleFunc("/", handler)

	// サーバー起動メッセージ
	log.Println("サーバー起動中 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
