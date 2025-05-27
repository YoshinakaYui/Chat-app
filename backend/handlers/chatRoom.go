package handlers

import (
	"backend/db"
	"backend/utils"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type CreateChatRoomRequest struct {
	User1ID int `json:"login_id"`
	User2ID int `json:"user_id"`
}
type CreateGroupRoomRequest struct {
	GroupName      string `json:"room_name"`
	LoggedInUserID int    `json:"login_id"`
	SelectedUsers  []int  `json:"user_ids"`
}

// チャットルーム作成（個別ルーム＆グループルーム）
func CreateGroupRoom(w http.ResponseWriter, r *http.Request) {
	log.Println("🟡CreateGroupRoom：スタート")
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
	var reqGroup CreateGroupRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&reqGroup); err != nil {
		log.Println("リクエストボディのデコードエラー:", err)
		http.Error(w, "リクエスト形式が不正です", http.StatusBadRequest)
		return
	}
	log.Println("🟡CreateGroupRoom：A")
	log.Println("🟡CreateGroupRoom：", reqGroup.GroupName)

	var is_group int = 0
	if reqGroup.GroupName == "" {
		// 個別チャット
		log.Println("🟡CreateGroupRoom：", len(reqGroup.SelectedUsers))

		if len(reqGroup.SelectedUsers) != 1 {
			http.Error(w, "メンバーを選択してください", http.StatusBadRequest)
			return
		}
		var existroom *db.ChatRoom = nil
		existroom = GetRoomMembersByUsers(reqGroup.LoggedInUserID, reqGroup.SelectedUsers[0])
		if existroom != nil {
			// 既存のルームが見つかった場合
			log.Println("既存ルームID:", existroom.ID)
			http.Error(w, "すでに作成されたルームです", http.StatusBadRequest)
			return
		}

		is_group = 0
	} else {
		// グループチャット
		if len(reqGroup.SelectedUsers) < 2 {
			http.Error(w, "グループ名または選択されているユーザー(2人以上の選択が必要です)が不正です", http.StatusBadRequest)
			return
		}

		// ルーム名の重複をチェック
		var groupRoom db.ChatRoom
		result := db.DB.Where("room_name = ?", reqGroup.GroupName).First(&groupRoom)
		if result.RowsAffected > 0 {
			http.Error(w, "すでに同じ名前のルームが作成されています。", http.StatusBadRequest)
			return
		}

		is_group = 1
	}

	log.Println("🟡CreateGroupRoom：B")

	room := db.ChatRoom{
		RoomName:  reqGroup.GroupName, // チャットルーム名は空欄
		IsGroup:   is_group,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.DB.Create(&room).Error; err != nil {
		log.Println("chat_rooms作成エラー：", err)
		http.Error(w, "チャットルーム作成失敗", http.StatusInternalServerError)
		return
	}
	log.Println("🟡CreateGroupRoom：C")

	// メンバー登録
	// 作成したchat_roomsのidとreqGroupのuser_ids(複数ユーザー)をroom_membersに保存
	var GruopMembers []db.RoomMember
	for i := 0; i < len(reqGroup.SelectedUsers); i++ {
		GruopMembers = append(GruopMembers, db.RoomMember{RoomID: room.ID, UserID: reqGroup.SelectedUsers[i], JoinedAt: time.Now()})
	}
	GruopMembers = append(GruopMembers, db.RoomMember{RoomID: room.ID, UserID: reqGroup.LoggedInUserID, JoinedAt: time.Now()})

	if err := db.DB.Create(&GruopMembers).Error; err != nil {
		log.Println("room_members作成エラー：", err)
		http.Error(w, "メンバー作成失敗", http.StatusInternalServerError)
		return
	}

	log.Println("🟡CreateGroupRoom：D")
	log.Println("新規グループルーム作成成功:", room.ID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "新規グループチャットルームを作成しました",
		"roomId":  room.ID,
	})
	log.Println("🟡CreateGroupRoom：エンド")
}
