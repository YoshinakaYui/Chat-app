package handlers

import (
	"backend/db" // データベース接続を管理する自作パッケージ
	"backend/models"
	"backend/utils"
	"encoding/json"
	"net/http" // HTTPサーバーを作成・操作するライブラリ
)

// ユーザー登録ハンドラー
func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "サインアップ：許可されていないメソッド", http.StatusMethodNotAllowed)
		return
	}

	var user models.TsUser
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
