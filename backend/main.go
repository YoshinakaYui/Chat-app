package main

import (
	"backend/auth"
	"backend/db" // データベース接続を管理する自作パッケージ
	"encoding/json"

	//"fmt"
	"log"
	"net/http" // HTTPサーバーを作成・操作するライブラリ
)

type tsUser struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	PasswordHash string
}

type Response struct {
	Message string `json:"message"`
}

// CORS対応を設定する関数
func enableCORS(w http.ResponseWriter) {
	// w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3001")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func handler(w http.ResponseWriter, r *http.Request) {
	// CORS対応を有効にする
	enableCORS(w)

	log.Println("aaa")
	w.Write([]byte("Hello"))
}

// ユーザー登録ハンドラー
func addUserHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "サインアップ：許可されていないメソッド", http.StatusMethodNotAllowed)
		return
	}

	var user tsUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "サインアップ：リクエスト形式が不正", http.StatusBadRequest)
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
	log.Println("bbb")

	if r.Method != http.MethodPost {
		http.Error(w, "ログイン：許可されていないメソッド", http.StatusMethodNotAllowed)
		return
	}
	log.Println("ccc")

	var loginReq tsUser
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "ログイン：リクエスト形式が不正", http.StatusBadRequest)
		return
	}
	log.Println("ddd")

	if loginReq.Username == "" || loginReq.Password == "" {
		http.Error(w, "ユーザー名またはパスワードが空です", http.StatusBadRequest)
		return
	}

	json.NewDecoder(r.Body).Decode(&loginReq)

	var user db.User
	// 正しいクエリを使用
	result := db.DB.Where("username = ?", loginReq.Username).First(&user)
	if result.Error != nil {
		http.Error(w, "ユーザーが存在しません", http.StatusUnauthorized)
		return
	}
	log.Println("eee")

	// パスワード検証
	if !db.CheckPasswordHash(loginReq.Password, user.PasswordHash) {
		http.Error(w, "パスワードが違います", http.StatusUnauthorized)
		return
	}
	log.Println("fff")

	// JWTトークンを発行
	token, err := auth.GenerateJWT(user.Username)
	//token := db.GenerateJWT(user.Username) // 本来はJWTを生成する

	if err != nil {
		http.Error(w, "トークン生成エラー", http.StatusInternalServerError)
		return
	}
	log.Println("ggg")

	// クッキーにトークンをセット
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	})
	log.Println("hhh")

	response := Response{Message: "たぶんログイン成功"}
	log.Println("iii")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// メッセージハンドラー
func messageHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	log.Println("jjj")
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		http.Error(w, "未ログイン", http.StatusUnauthorized)
		return
	}

	// トークンを検証してユーザー名を取得
	username, err := auth.ValidateJWT(cookie.Value)
	if err != nil {
		log.Println("jjj11")
		http.Error(w, "トークンが無効です", http.StatusUnauthorized)
		return
	}
	w.Write([]byte("ようこそ " + username + " さん"))
	db.GetAllUsers()
	log.Println("kkk")
}

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
	//http.HandleFunc("/message", getRegisteredUsersHandler)
	http.HandleFunc("/message", messageHandler)

	log.Println("サーバー起動中 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
