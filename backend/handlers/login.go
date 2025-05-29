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

// ログイン処理
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("LoginHandler：スタート")
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

	if loginReq.Username == "" || loginReq.Password == "" {
		http.Error(w, "ユーザー名またはパスワードが空です", http.StatusBadRequest)
		return
	}

	// クエリ実行：ユーザー名で検索
	var user db.Users
	result := db.DB.Where("username = ?", loginReq.Username).First(&user)
	if result.Error != nil {
		log.Println("ユーザーが見つかりません:", loginReq.Username) // ユーザー名をログに出力
		http.Error(w, "ユーザーが存在しません", http.StatusUnauthorized)
		return
	}

	log.Println("ログインID：", user.ID, "ハッシュ化：", user.PasswordHash)

	// パスワード検証
	if !db.CheckPasswordHash(loginReq.Password, user.PasswordHash) {
		http.Error(w, "パスワードが違います", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateJWT(user.Username, user.PasswordHash)
	if err != nil {
		http.Error(w, "トークン生成エラー", http.StatusInternalServerError)
		return
	}
	log.Println("auth_token：", token)

	// レスポンスを返す
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"username": loginReq.Username,
		"userId":   user.ID,
		"message":  "ログイン成功",
		"token":    token,
	})
}
