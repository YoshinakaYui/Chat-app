package handlers

import (
	"backend/db"
	"backend/utils"
	"encoding/json"
	"log"
	"net/http"
)

// ユーザー一覧を取得するハンドラー
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(w)

	w.Header().Set("Content-Type", "application/json")

	users, err := db.GetAllUsers()
	if err != nil {
		log.Println("ユーザー一覧取得エラー:", err)
		log.Println("romm.go-111")
		http.Error(w, "ユーザー一覧の取得に失敗しました", http.StatusInternalServerError)
		return
	}
	// レスポンスを返す
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)

}
