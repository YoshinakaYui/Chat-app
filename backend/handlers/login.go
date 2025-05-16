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

	if r.Method != http.MethodPost {
		http.Error(w, "ログイン：許可されていないメソッド", http.StatusMethodNotAllowed)
		return
	}

	var loginReq models.TsUser

	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "ログイン：リクエスト形式が不正", http.StatusBadRequest)
		return
	}
	log.Println("login-ddd")

	if loginReq.Username == "" || loginReq.Password == "" {
		http.Error(w, "ユーザー名またはパスワードが空です", http.StatusBadRequest)
		return
	}
	log.Println("login-eee")
	// クエリ実行：ユーザー名で検索
	var user db.Users
	result := db.DB.Where("username = ?", loginReq.Username).First(&user)
	if result.Error != nil {
		log.Println("ユーザーが見つかりません:", loginReq.Username) // ユーザー名をログに出力
		http.Error(w, "ユーザーが存在しません", http.StatusUnauthorized)
		return
	}
	log.Println("login-fff")
	log.Println("ログインID：", user.ID)
	log.Println(loginReq.Password)
	log.Println("ハッシュ化：", user.PasswordHash)
	// パスワード検証
	if !db.CheckPasswordHash(loginReq.Password, user.PasswordHash) {
		http.Error(w, "パスワードが違います", http.StatusUnauthorized)
		return
	}
	log.Println("login-ggg")
	token, err := auth.GenerateJWT(user.Username, user.PasswordHash)
	if err != nil {
		http.Error(w, "トークン生成エラー", http.StatusInternalServerError)
		return
	}
	log.Println("auth_token", token)

	// レスポンスを返す
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"username": loginReq.Username,
		"userId":   user.ID,
		"message":  "ログイン成功",
		"token":    token,
	})

	log.Println("ログインした人", user.ID, loginReq.Username)
}
