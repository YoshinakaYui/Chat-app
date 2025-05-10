package main

import (
	"log"
	"net/http" // HTTPサーバーを作成・操作するためのライブラリ
	"backend/db" // データベース接続を管理する自作パッケージ
	"os" // 環境変数を扱うための標準ライブラリ
	// "backend/db" // データベース接続用のパッケージ	
	"github.com/joho/godotenv" // .envファイルから環境変数をロードするための外部パッケージ
)

// init関数：環境変数をロード(init関数はプログラム起動時に必ず1回だけ自動で実行されるため、呼び出す必要がない)
func init() {
	err := godotenv.Load(".env")
	if err != nil {
		// ログメッセージを出力して、プログラムを強制終了させる(重大なエラーが発生したときに使う)(強すぎるため、本番環境では使いすぎない)
		log.Fatalf("環境変数の読み込みに失敗しました: %v", err)
	}
	log.Println("環境変数の読み込み成功")
}

// CORS対応を設定する関数(CORS：異なるポート番号でも通信できるようにする仕組み)
func enableCORS(w http.ResponseWriter) {
	// どこからのリクエストでもOKにする（全許可）
    w.Header().Set("Access-Control-Allow-Origin", "*")
	// 許可するHTTPメソッド（ブラウザが送れるリクエストの種類を指定）
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	// 許可するリクエストヘッダー（リクエストに含められるヘッダーの種類を指定）
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// リクエストが来たときに呼ばれる関数
func handler(w http.ResponseWriter, r *http.Request) {
	// CORS対応を有効にする
    enableCORS(w)
    w.Write([]byte("Hello!"))
}

// メイン関数：サーバーを起動する
func main() {
	// データベース接続確認(エラーがあれば失敗と表示)
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
