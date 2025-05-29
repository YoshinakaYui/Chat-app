package handlers

import (
	"backend/db" // データベース接続を管理する自作パッケージ
	"backend/models"
	"backend/utils"
	"encoding/json"
	"log"
	"net/http" // HTTPサーバーを作成・操作するライブラリ
)

// ユーザー登録ハンドラー
func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("AddUserHandler：スタート")
	utils.EnableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "サインアップ：許可されていないメソッド", http.StatusMethodNotAllowed)
		return
	}

	var req models.TsUser
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "サインアップ：リクエスト形式が不正", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "ユーザー名またはパスワードが空です", http.StatusBadRequest)
		return
	}

	var existingUser db.Users
	result := db.DB.Where("username = ?", req.Username).First(&existingUser)

	if result.RowsAffected > 0 {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "すでにそのユーザー名は使われています。",
		})
		return
	}

	if err := db.SaveUser(req.Username, req.Password); err != nil {
		http.Error(w, "データベース保存エラー", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("保存成功"))

}
