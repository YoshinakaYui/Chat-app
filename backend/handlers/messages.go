package handlers

import (
	"backend/db"
	"backend/models"
	"backend/utils"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

// メッセージ送信・取得ハンドラー
func MessageHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(w)
	w.Header().Set("Content-Type", "application/json")
	log.Println("⚫︎msg-1")
	switch r.Method {
	case http.MethodPost:
		handleSendMessage(w, r)
	case http.MethodGet:
		handleGetMessages(w, r)
	case http.MethodOptions:
		w.WriteHeader(http.StatusOK)
		return
	default:
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
	}
}

// メッセージ送信処理
func handleSendMessage(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(w)
	var msg models.TsMessage
	log.Println("⚫︎msg-2")
	// リクエストボディのデコード
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "リクエスト形式が不正", http.StatusBadRequest)
		return
	}
	log.Println("⚫︎msg-3")
	// 必須フィールドチェック
	if msg.RoomID == 0 || msg.SenderID == 0 || msg.Content == "" {
		http.Error(w, "必須フィールドが不足しています", http.StatusBadRequest)
		return
	}
	log.Println("⚫︎msg-4")
	// メッセージの保存
	msg.CreatedAt = time.Now()
	msg.UpdatedAt = time.Now()
	message := db.Message{
		RoomID:    msg.RoomID,
		SenderID:  msg.SenderID,
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt,
		UpdatedAt: msg.UpdatedAt,
	}
	log.Println("⚫︎msg-5")
	// データベースに保存
	if err := db.DB.Create(&message).Error; err != nil {
		log.Println("メッセージ保存エラー:", err)
		http.Error(w, "メッセージ保存失敗", http.StatusInternalServerError)
		return
	}
	log.Println("⚫︎msg-6")
	// 成功レスポンス
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "メッセージ保存完了",
		"data":    message,
	})
}

// メッセージ取得処理
func handleGetMessages(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(w)
	log.Println("⚫︎msg-11")
	// クエリパラメータからroom_idを取得
	roomIDStr := r.URL.Query().Get("room_id")
	if roomIDStr == "" {
		http.Error(w, "ルームIDが必要です", http.StatusBadRequest)
		return
	}
	log.Println("⚫︎msg-22")
	// 文字列を整数に変換
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		http.Error(w, "ルームIDが不正です", http.StatusBadRequest)
		return
	}
	log.Println("⚫︎msg-33")
	var messages []db.Message
	// データベースからメッセージを取得（作成日時順にソート）
	if err := db.DB.Where("room_id = ?", roomID).Order("created_at ASC").Find(&messages).Error; err != nil {
		http.Error(w, "メッセージ取得失敗", http.StatusInternalServerError)
		return
	}
	log.Println("⚫︎msg-44")
	// JSONレスポンス
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"messages": messages,
	})
}
