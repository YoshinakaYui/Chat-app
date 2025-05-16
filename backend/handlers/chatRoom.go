package handlers

import (
	"backend/db"
	"backend/models"
	"backend/utils"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type CreateChatRoomRequest struct {
	User1ID int `json:"user1"`
	User2ID int `json:"user2"`
}

// チャットルーム作成ハンドラー
func CreateChatRoom(w http.ResponseWriter, r *http.Request) {
	log.Println("チャットルーム作成処理開始")
	utils.EnableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	// リクエストボディからデータを取得
	var req CreateChatRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("リクエストボディのデコードエラー:", err)
		http.Error(w, "リクエスト形式が不正です", http.StatusBadRequest)
		return
	}

	if req.User1ID == 0 || req.User2ID == 0 {
		http.Error(w, "ユーザーIDが不正です", http.StatusBadRequest)
		return
	}

	// ユーザーIDを昇順にソートして比較基準を統一
	userIDs := []int{req.User1ID, req.User2ID}
	if userIDs[0] > userIDs[1] {
		userIDs[0], userIDs[1] = userIDs[1], userIDs[0]
	}

	// room_membersテーブルから既存ルームを検索//room_membersとchat_roomsを繋げた上で、セレクトをかける
	// 🔴SQL文でroom_membersとchat_roomsを繋げる
	var roomIDs []int
	err := db.DB.Table("room_members").
		Select("room_id").
		Where("user_id IN (?, ?)", userIDs[0], userIDs[1]).
		Group("room_id").
		Having("COUNT(DISTINCT user_id) = 2").
		Pluck("room_id", &roomIDs).Error

	if err == nil && len(roomIDs) > 0 {
		// 既存のルームが見つかった場合
		var existingRoom models.TsChatRoom
		err := db.DB.Where("id = ?", roomIDs[0]).First(&existingRoom).Error
		if err == nil {
			log.Println("既存ルームID:", existingRoom.ID)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "success",
				"message": "既存のチャットルームを取得しました",
				"roomId":  existingRoom.ID,
			})
			return
		}
	}

	// 新規チャットルーム作成
	room := models.TsChatRoom{
		RoomName:  "", // チャットルーム名は空欄
		IsGroup:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.DB.Create(&room).Error; err != nil {
		log.Println("chat_rooms作成エラー：", err)
		http.Error(w, "チャットルーム作成失敗", http.StatusInternalServerError)
		return
	}

	// メンバー登録
	members := []models.TsRoomMember{
		{RoomID: room.ID, UserID: req.User1ID, JoinedAt: time.Now()},
		{RoomID: room.ID, UserID: req.User2ID, JoinedAt: time.Now()},
	}
	if err := db.DB.Create(&members).Error; err != nil {
		log.Println("room_members作成エラー：", err)
		http.Error(w, "メンバー作成失敗", http.StatusInternalServerError)
		return
	}

	log.Println("新規ルーム作成成功:", room.ID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "新規チャットルームを作成しました",
		"roomId":  room.ID,
	})
}
