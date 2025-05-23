package handlers

import (
	"backend/db"
	"backend/utils"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type ReadRequest struct {
	UserID    int `json:"login_id"`
	MessageID int `json:"msg_id"`
}

type MessageRead struct {
	MessageID int       `json:"message_id"`
	UserID    int       `json:"user_id"`
	ReadAt    time.Time `json:"read_at"`
}

func MarkMessageAsRead(w http.ResponseWriter, r *http.Request) {
	log.Println("🟩MarkMessageAsRead")

	utils.EnableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "ログイン：許可されていないメソッド", http.StatusMethodNotAllowed)
		return
	}

	var req ReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	read := MessageRead{
		MessageID: req.MessageID,
		UserID:    req.UserID,
		ReadAt:    time.Now(),
	}
	log.Println("🟩：", read)

	if err := db.DB.Table("message_reads").Create(&read).Error; err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	log.Println("🟩2：", read)

	w.WriteHeader(http.StatusOK)
	// JSONレスポンス
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   read,
	})
}
