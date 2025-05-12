package main

import (
	"log"
	"net/http" // HTTPサーバーを作成・操作するためのライブラリ
	"backend/db" // データベース接続を管理する自作パッケージ
    //"backend/login" // データベース接続を管理する自作パッケージ
	//"os" // 環境変数を扱うための標準ライブラリ
	// "backend/db" // データベース接続用のパッケージ	
	//"github.com/joho/godotenv" // .envファイルから環境変数をロードするための外部パッケージ
	"encoding/json"
)

type tsUser struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

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
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	// 許可するリクエストヘッダー（リクエストに含められるヘッダーの種類を指定）
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// リクエストが来たときに呼ばれる関数
func handler(w http.ResponseWriter, r *http.Request) {
	// CORS対応を有効にする
    enableCORS(w)

	log.Println("aaa")
    w.Write([]byte("Hello!!"))
}


// ユーザー登録ハンドラー
func addUserHandler(w http.ResponseWriter, r *http.Request) {

	enableCORS(w)

	log.Println(r.Method)
    if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
    }

	log.Println(r.Method)
    if r.Method != http.MethodPost {
        http.Error(w, "サインアップ：許可されていないメソッド", http.StatusMethodNotAllowed)
        return
    }

	var user tsUser

    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "サインアップ：リクエスト形式が不正", http.StatusBadRequest)
        return
    }

	log.Println(user)
	log.Println(user.Username)
	log.Println(user.Password)

	// エラーチェック（userbane, password)
	if user.Username == "" || user.Password == "" {
		log.Println("サインアップ：ユーザー名またはパスワードが空です")
		http.Error(w, "ユーザー名を入力してください", http.StatusBadRequest)
 		return 
	}

	// おなじユーザー名は登録不可
	//if xxx {
	//
	//}


    // user := User{Username: username, Password: hashedPassword}
    if err := db.SaveUser(user.Username, user.Password); err != nil {
        http.Error(w, "サインアップ：データベース保存エラー", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("保存成功"))
}

// ログインハンドラー
func loginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("ログインハンドラー実行")
    enableCORS(w)

    // OPTIONSメソッド対応
	log.Println(r.Method)
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    // POSTメソッド以外は拒否
    if r.Method != http.MethodPost {
        http.Error(w, "ログイン：許可されていないメソッド", http.StatusMethodNotAllowed)
        return
    }

    var loginReq LoginRequest

    if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
        http.Error(w, "ログイン：リクエスト形式が不正", http.StatusBadRequest)
        return
    }

    // ユーザー名とパスワードが空でないかチェック
    if loginReq.Username == "" || loginReq.Password == "" {
		log.Println("ログイン：ユーザー名またはパスワードが空です")
		//w.Write([]byte("空です"))
        http.Error(w, "ユーザー名またはパスワードが空です", http.StatusBadRequest)
        return
    }

    // ログインチェック
    if db.IsLogin(loginReq.Username, loginReq.Password) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ログインチェックが成功しました"))
    } else {
        http.Error(w, "ログイン失敗", http.StatusUnauthorized)
		w.Write([]byte("ログインチェックに失敗しました"))
    }
}



// メイン関数：サーバーを起動する
func main() {
	//enableCORS(w)
	// データベース接続確認(エラーがあれば失敗と表示)
	// if err := db.Connect(); err != nil {
	// 	log.Fatal("DB接続失敗:", err)

	// }
	// log.Println("DB接続成功！")
	if err := db.Connect(); err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	} else {
		log.Println("DB接続成功")
	}

	// HTTPリクエストを処理するハンドラー関数
	http.HandleFunc("/signup", addUserHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/", handler)
	//http.HandleFunc("/chat-home", chatHandler)

	// サーバー起動メッセージ
	log.Println("サーバー起動中 http://localhost:8080")
	// わからない
	log.Fatal(http.ListenAndServe(":8080", nil))
}

