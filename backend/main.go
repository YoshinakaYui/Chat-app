package main

import (
	"log"
	"net/http" // HTTPサーバーを作成・操作するためのライブラリ
	"backend/db" // データベース接続を管理する自作パッケージ
	//"os" // 環境変数を扱うための標準ライブラリ
	// "backend/db" // データベース接続用のパッケージ	
	//"github.com/joho/godotenv" // .envファイルから環境変数をロードするための外部パッケージ
	"encoding/json"
)

// init関数：環境変数をロード(init関数はプログラム起動時に必ず1回だけ自動で実行されるため、呼び出す必要がない)
func init() {
	// err := godotenv.Load(".env")
	// if err != nil {
	// 	// ログメッセージを出力して、プログラムを強制終了させる(重大なエラーが発生したときに使う)(強すぎるため、本番環境では使いすぎない)
	// 	log.Fatalf("環境変数の読み込みに失敗しました: %v", err)
	// }
	// log.Println("環境変数の読み込み成功")
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

	log.Println("aaa")
    w.Write([]byte("Hello!"))
}


type User struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

// ユーザー登録ハンドラー
func addUserHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	log.Println("bbb")
	w.Write([]byte("write-bbb"))
    if r.Method != http.MethodPost {
        http.Error(w, "許可されていないメソッド", http.StatusMethodNotAllowed)
        return
    }

	var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "リクエスト形式が不正", http.StatusBadRequest)
        return
    }

	log.Println(user)
	log.Println(user.Username)
    // w.Write([]byte(user))

    // user := User{Username: username, Password: hashedPassword}
    if err := db.SaveUser(user.Username, user.Password); err != nil {
        http.Error(w, "データベース保存エラー", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("保存成功"))
}

// メイン関数：サーバーを起動する
func main() {
	// データベース接続確認(エラーがあれば失敗と表示)
	if err := db.Connect(); err != nil {
		log.Fatal("DB接続失敗:", err)

	}
	log.Println("DB接続成功！aaaaaaa")

	// HTTPリクエストを処理するハンドラー関数
	http.HandleFunc("/login", addUserHandler)
	http.HandleFunc("/", handler)

	// サーバー起動メッセージ
	log.Println("aaaaサーバー起動中 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

