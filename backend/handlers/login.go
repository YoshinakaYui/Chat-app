package handlers

import (
	"backend/auth"
	"backend/db"
	"backend/models"
	"backend/utils"
	"encoding/json"
	"log"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	//log.Println("login-bbb")

	if r.Method != http.MethodPost {
		http.Error(w, "ログイン：許可されていないメソッド", http.StatusMethodNotAllowed)
		return
	}
	//log.Println("login-ccc")

	var loginReq models.TsUser

	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "ログイン：リクエスト形式が不正", http.StatusBadRequest)
		return
	}
	//log.Println("login-ddd")

	if loginReq.Username == "" || loginReq.Password == "" {
		http.Error(w, "ユーザー名またはパスワードが空です", http.StatusBadRequest)
		return
	}
	//log.Println("login-eee")
	// クエリ実行：ユーザー名で検索
	var user db.Users
	result := db.DB.Where("username = ?", loginReq.Username).First(&user)
	if result.Error != nil {
		log.Println("ユーザーが見つかりません:", loginReq.Username) // ユーザー名をログに出力
		http.Error(w, "ユーザーが存在しません", http.StatusUnauthorized)
		return
	}
	//log.Println("login-fff")
	log.Println(loginReq.Password)
	log.Println(user.PasswordHash)
	// パスワード検証
	if !db.CheckPasswordHash(loginReq.Password, user.PasswordHash) {
		http.Error(w, "パスワードが違います", http.StatusUnauthorized)
		return
	}
	//log.Println("login-ggg")
	token, err := auth.GenerateJWT(user.Username, user.PasswordHash)
	if err != nil {
		http.Error(w, "トークン生成エラー", http.StatusInternalServerError)
		return
	}
	//log.Println("login-hhh")
	log.Println("auth_token", token)
	// // ユーザーがログインに成功した場合
	// http.SetCookie(w, &http.Cookie{
	// 	Name:     "auth_token",
	// 	Value:    token,
	// 	Path:     "/",
	// 	Domain:   "localhost",             // ドメイン指定
	// 	HttpOnly: true,                    // JavaScriptでアクセス不可
	// 	Secure:   false,                   // 開発環境ではfalse, 本番環境ではtrue
	// 	SameSite: http.SameSiteStrictMode, // CSRF対策
	// 	//Expires:  time.Now().Add(24 * time.Hour), // 24時間有効
	// })
	//log.Println("login-iii")

	// レスポンスを返す
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	log.Println("auth_token", token)

	json.NewEncoder(w).Encode(map[string]string{
		"username": loginReq.Username,
		"message":  "ログイン成功",
		"token":    token,
	})
}
