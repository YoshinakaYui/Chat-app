package handlers

import (
	"backend/db"
	"backend/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type ReadRequest struct {
	UserID    int `json:"login_id"`
	MessageID int `json:"msg_id"`
	RoomID    int `json:"room_id"`
}

type MessageRead struct {
	MessageID int       `json:"message_id"`
	UserID    int       `json:"user_id"`
	ReadAt    time.Time `json:"read_at"`
}

// 既読
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

	//utils.JsonRawDataDisplay(w, r)
	var req ReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		log.Println("read.go エラー：", err)
		return
	}
	log.Println("🔺🔺🔺req：", req)

	read := MessageRead{
		MessageID: req.MessageID,
		UserID:    req.UserID,
		ReadAt:    time.Now(),
	}

	if err := db.DB.Table("message_reads").Create(&read).Error; err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	var messageid = []int{req.MessageID}
	//i_roomid, err := strconv.Atoi(req.RoomID)
	// if err != nil {
	// 	http.Error(w, "strconv.Atoi error", http.StatusInternalServerError)
	// 	return
	// }
	BroadcastReadCountsToRoom(req.RoomID, messageid)

	// 未読のリアルタイム通知（roomSelect宛）
	type UnreadResult struct {
		UserID      int `json:"user_id" gorm:"column:user_id"`
		RoomID      int `json:"room_id" gorm:"column:room_id"`
		UnreadCount int `json:"unread_count" gorm:"column:unread_count"`
	}

	var results []UnreadResult

	err1 := db.DB.
		Raw(`
		SELECT
		  rm.user_id,
		  m.room_id,
		  COUNT(*) AS unread_count
		FROM messages m
		JOIN room_members rm ON rm.room_id = m.room_id
		WHERE m.room_id = ?
		  AND rm.user_id != ?
		  AND NOT EXISTS (
		    SELECT 1
		    FROM message_reads mr
		    WHERE mr.message_id = m.id
		      AND mr.user_id = rm.user_id
		  )
		GROUP BY rm.user_id, m.room_id
	`, req.RoomID, req.UserID).
		Scan(&results).Error

	if err1 != nil {
		log.Println("未読数の取得エラー:", err1)
	} else {
		for _, r := range results {
			fmt.Printf("room_id: %d, 未読数: %d\n", r.RoomID, r.UnreadCount)
		}
	}
	log.Println("🔺🔺🔺未読数取得-BB:", results)

	if len(results) != 0 {
		// 既読を他のクライアントへブロードキャスト
		joinBroadcast := map[string]interface{}{
			"type":          "unreadmessage",
			"unReadMessage": results,
			"room_id":       req.RoomID,
		}
		joinJSON, _ := json.Marshal(joinBroadcast)
		log.Println("未読ブロードキャスト：", joinJSON)

		// var decoded map[string]interface{}
		// err2 := json.Unmarshal(joinJSON, &decoded)
		// if err2 != nil {
		// 	log.Println("JSONデコード失敗:", err2)
		// }
		// log.Println("未読ブロードキャストPPP：", decoded)

		broadcast <- joinJSON
	}

	w.WriteHeader(http.StatusOK)
	// JSONレスポンス
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   read,
	})
}
