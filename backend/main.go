package main

import (
	"backend/db" // データベース接続を管理する自作パッケージ
	"encoding/json"
	"log"
	"net/http" // HTTPサーバーを作成・操作するライブラリ
)

type tsUser struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	// CORS対応を有効にする
	enableCORS(w)

	log.Println("aaa")
	w.Write([]byte("Hello!"))
}

// CORS対応を設定する関数
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// ユーザー登録ハンドラー
func addUserHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "許可されていないメソッド", http.StatusMethodNotAllowed)
		return
	}

	var user tsUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "リクエスト形式が不正", http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Password == "" {
		http.Error(w, "ユーザー名またはパスワードが空です", http.StatusBadRequest)
		return
	}

	if err := db.SaveUser(user.Username, user.Password); err != nil {
		http.Error(w, "データベース保存エラー", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("保存成功"))
}

// ログインハンドラー
func loginHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "許可されていないメソッド", http.StatusMethodNotAllowed)
		return
	}

	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "リクエスト形式が不正", http.StatusBadRequest)
		return
	}

	if loginReq.Username == "" || loginReq.Password == "" {
		http.Error(w, "ユーザー名またはパスワードが空です", http.StatusBadRequest)
		return
	}

	if db.IsLogin(loginReq.Username, loginReq.Password) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ログイン成功"))
	} else {
		http.Error(w, "ログイン失敗", http.StatusUnauthorized)
	}
}

// ユーザー一覧取得ハンドラー
func getRegisteredUsersHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	users, err := db.GetAllUsers()
	if err != nil {
		http.Error(w, "ユーザー一覧の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// メッセージ送信ハンドラー
func postMessageHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "送信：許可されていないメソッド", http.StatusMethodNotAllowed)
		return
	}

	var msg db.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "送信：リクエスト形式が不正", http.StatusBadRequest)
		return
	}

	if msg.Sender == "" || msg.Content == "" {
		http.Error(w, "送信者またはメッセージが空です", http.StatusBadRequest)
		return
	}

	if err := db.SaveMessage(msg.Sender, msg.Content); err != nil {
		http.Error(w, "メッセージ保存エラー", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("メッセージ送信成功"))
}

// メッセージ一覧を取得するハンドラー
// func getMessagesHandler(w http.ResponseWriter, r *http.Request) {
// 	enableCORS(w)

// 	if r.Method == "OPTIONS" {
// 		w.WriteHeader(http.StatusOK)
// 		return
// 	}

// 	if r.Method != http.MethodGet {
// 		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	messages, err := db.GetAllMessages()
// 	if err != nil {
// 		http.Error(w, "メッセージ一覧の取得に失敗しました", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(messages)
// }

// メイン関数：サーバーを起動する
func main() {
	if err := db.Connect(); err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	} else {
		log.Println("DB接続成功")
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/signup", addUserHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/message", getRegisteredUsersHandler)
	//http.HandleFunc("/message", postMessageHandler)

	log.Println("サーバー起動中 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
