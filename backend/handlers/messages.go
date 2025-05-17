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

	// メッセージを格納するスライス
	var SendMessages []struct {
		MessageID  int    `json:"message_id"`
		Content    string `json:"content"`
		CreatedAt  string `json:"created_at"`
		SenderID   int    `json:"sender_id"`
		SenderName string `json:"sender_name"`
	}

	// GORMでSQLクエリを構築（Link形式）
	result := db.DB.Table("messages AS m").
		Select("m.id AS message_id, m.content, m.created_at, m.sender_id, u.username AS sender_name").
		Joins("JOIN users AS u ON m.sender_id = u.id").
		Where("m.room_id = ?", roomID).
		Find(&SendMessages)

		// エラー処理
	if result.Error != nil {
		log.Println("メッセージ取得エラー:", result.Error)
		http.Error(w, "メッセージが見つかりません", http.StatusNotFound)
		return
	}

	log.Println("⚫︎msg-44")
	// JSONレスポンス
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"messages": SendMessages,
	})
}

// ルームメンバー取得処理（LEFT JOINを使用）
func GetRoomMembersByUsers(user1ID int, user2ID int) *db.ChatRoom {
	// utils.EnableCORS(w)
	// log.Println("⚫︎msg-55")
	// // クエリパラメータからroom_idを取得
	// roomIDStr := r.URL.Query().Get("room_id")
	// if roomIDStr == "" {
	// 	http.Error(w, "ルームIDが必要です", http.StatusBadRequest)
	// 	return
	// }
	// log.Println("⚫︎msg-66")
	// roomID, err := strconv.Atoi(roomIDStr)
	// if err != nil {
	// 	http.Error(w, "ルームIDが不正です", http.StatusBadRequest)
	// 	return
	// }
	// log.Println("⚫︎msg-77")
	// var members []struct {
	// 	RoomID   int    `json:"room_id"`
	// 	RoomName string `json:"room_name"`
	// 	IsGroup  int    `json:"is_group"`
	// 	UserID   *int   `json:"user_id"` // NULL対応
	// }

	// LEFT JOINでルームとメンバーを取得
	// GORMのLink形式を使ってクエリを組み立て
	var chatroom db.ChatRoom

	result := db.DB.Table("chat_rooms AS cr").
		Select("cr.*").
		Joins(`JOIN (
                SELECT rm1.room_id
                FROM room_members AS rm1
                JOIN room_members AS rm2 ON rm1.room_id = rm2.room_id
                WHERE rm1.user_id = ? 
                  AND rm2.user_id = ? 
                  AND rm1.user_id <> rm2.user_id
            ) AS common_rooms ON cr.id = common_rooms.room_id`, user1ID, user2ID).
		Where("cr.is_group = ?", 0).
		First(&chatroom)

	if result.Error != nil {
		log.Println("チャットルームが見つかりません:", result.Error)
		return nil
	}

	return &chatroom
}
