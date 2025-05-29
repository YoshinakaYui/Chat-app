package handlers

import (
	"backend/db"
	"backend/utils"
	"strconv"

	"encoding/json"
	"log"
	"net/http"
	"time"
)

type CreateGroupRoomRequest struct {
	GroupName      string `json:"room_name"`
	LoggedInUserID int    `json:"login_id"`
	SelectedUsers  []int  `json:"user_ids"`
}

type RoomCreatedEvent struct {
	Type     string `json:"type"`      // "room_created"
	RoomID   int    `json:"room_id"`   // 作成されたルームID
	RoomName string `json:"room_name"` // グループ名（個別チャットなら空）
	IsGroup  int    `json:"is_group"`  // グループかどうか（0 or 1）
}

type LeaveRoom struct {
	RoomID         int `json:"room_id"`
	LoggedInUserID int `json:"user_id"`
}

type addMember struct {
	RoomID        int   `json:"room_id"`
	SelectedUsers []int `json:"user_ids"`
}

// チャットルーム作成（個別ルーム＆グループルーム）
func CreateChatRoom(w http.ResponseWriter, r *http.Request) {
	log.Println("CreateGroupRoom：スタート")
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

	var is_group int = 0
	if reqGroup.GroupName == "" {
		// 個別チャット
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

	// メンバー登録：作成したchat_roomsのidとreqGroupのuser_ids(複数ユーザー)をroom_membersに保存
	var GruopMembers []db.RoomMember

	// SelectedUsersに自分を追加する
	for i := 0; i < len(reqGroup.SelectedUsers); i++ {
		GruopMembers = append(GruopMembers, db.RoomMember{RoomID: room.ID, UserID: reqGroup.SelectedUsers[i], JoinedAt: time.Now()})
	}
	GruopMembers = append(GruopMembers, db.RoomMember{RoomID: room.ID, UserID: reqGroup.LoggedInUserID, JoinedAt: time.Now()})

	if err := db.DB.Create(&GruopMembers).Error; err != nil {
		log.Println("room_members作成エラー：", err)
		http.Error(w, "メンバー作成失敗", http.StatusInternalServerError)
		return
	}

	// ルーム作成を他のクライアントへブロードキャスト
	roomusers := make(map[int]string)

	groupname := reqGroup.GroupName
	if is_group == 0 {
		// 個別チャット.
		username1, err := db.GetUserName(reqGroup.SelectedUsers[0])
		if err != nil {
			log.Println("❌ ユーザー名の取得失敗:", err)
		}

		username2, err1 := db.GetUserName(reqGroup.LoggedInUserID)
		if err1 != nil {
			log.Println("❌ ユーザー名の取得失敗:", err1)
		}

		roomusers[reqGroup.SelectedUsers[0]] = username2
		roomusers[reqGroup.LoggedInUserID] = username1

	} else {
		// グループチャット
		for i := 0; i < len(reqGroup.SelectedUsers); i++ {
			roomusers[reqGroup.SelectedUsers[i]] = reqGroup.GroupName
		}
		roomusers[reqGroup.LoggedInUserID] = reqGroup.GroupName
	}

	//ルーム作成をブロードキャスト
	BroadcastCreateRomm(room.ID, groupname, roomusers, is_group)

	log.Println("新規グループルーム作成成功:", room.ID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "新規グループチャットルームを作成しました",
		"roomId":  room.ID,
	})

}

// ルームメンバー取得処理（チャットルーム作成で呼び出される）
func GetRoomMembersByUsers(user1ID int, user2ID int) *db.ChatRoom {
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

// ルーム退出
func LeaveRoomHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("LeaveRoomHandler：スタート")
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
	var req LeaveRoom
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("リクエストボディのデコードエラー:", err)
		http.Error(w, "リクエスト形式が不正です", http.StatusBadRequest)
		return
	}

	// is_groupを取得
	var isGroup int
	err := db.DB.
		Table("chat_rooms").
		Select("is_group").
		Where("id = ?", req.RoomID).
		Scan(&isGroup).Error

	if err != nil {
		log.Println("❌ is_group取得失敗:", err)
	}

	var userIDs []int

	if isGroup == 0 {

		err := db.DB.
			Table("room_members").
			Select("user_id").
			Where("room_id = ?", req.RoomID).
			Scan(&userIDs).Error

		if err != nil {
			log.Println("❌ user_id一覧の取得失敗:", err)
		}

		err1 := db.DB.
			Where("room_id = ?", req.RoomID).
			Delete(&db.RoomMember{}).Error

		if err1 != nil {
			log.Println("❌ ルームメンバー一括削除失敗:", err1)
		}
	} else {
		userIDs = append(userIDs, req.LoggedInUserID)
		err := db.DB.
			Where("room_id = ? AND user_id = ?", req.RoomID, req.LoggedInUserID).
			Delete(&db.RoomMember{}).Error

		if err != nil {
			log.Println("❌ メンバー削除失敗:", err)
		}
	}

	// 既読を他のクライアントへブロードキャスト
	BroadcastLeaveRoom(req.RoomID, userIDs)
}

// メンバー追加のためのルームに存在しないをユーザー取得
func UsersNotInRoomHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("AddMemberHandler：スタート")
	utils.EnableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("room_id")
	log.Println("AddMemberHandler ルームID：", idStr)
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ルームIDが不正です", http.StatusBadRequest)
		return
	}

	// リクエストボディのパース
	var req struct {
		LoginUserID int   `json:"login_id"`
		Members     []int `json:"members"`
	}

	// リクエストボディからデータを取得
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("リクエストボディのデコードエラー:", err)
		http.Error(w, "リクエスト形式が不正です", http.StatusBadRequest)
		return
	}

	log.Println("members取得：", req)

	var users []struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	}

	err1 := db.DB.
		Table("users").
		Select("id, username").
		Where("id NOT IN (?)",
			db.DB.Table("room_members").
				Select("user_id").
				Where("room_id = ?", id),
		).
		Scan(&users).Error

	if err1 != nil {
		log.Println("❌ ルーム外ユーザーの取得失敗:", err)
	}

	log.Println("メンバー以外のユーザー：", users)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"members": users,
	})

}

// メンバー追加
func AddMemberHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("AddMemberHandler：スタート")
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
	var req addMember
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("リクエストボディのデコードエラー:", err)
		http.Error(w, "リクエスト形式が不正です", http.StatusBadRequest)
		return
	}

	// メンバー登録
	var AddMembers []db.RoomMember

	for i := 0; i < len(req.SelectedUsers); i++ {
		AddMembers = append(AddMembers, db.RoomMember{RoomID: req.RoomID, UserID: req.SelectedUsers[i], JoinedAt: time.Now()})
	}

	if err := db.DB.Create(&AddMembers).Error; err != nil {
		log.Println("room_members作成エラー：", err)
		http.Error(w, "メンバー作成失敗", http.StatusInternalServerError)
		return
	}

	log.Println("メンバーを追加しました", AddMembers)

	// message_readsに記録し、既読状態にする
	for _, userID := range req.SelectedUsers {
		err := db.DB.Exec(`
			INSERT INTO message_reads (message_id, user_id, read_at)
			SELECT m.id, ?, ?
			FROM messages m
			WHERE m.room_id = ?
			  AND m.id NOT IN (
				SELECT mr.message_id FROM message_reads mr WHERE mr.user_id = ?
			  )`,
			userID, time.Now(), req.RoomID, userID).Error

		if err != nil {
			log.Println("❌ 既読データの挿入失敗:", err)
		}
	}

	log.Println("新メンバー既読")

	// 追加を他のクライアントへブロードキャスト
	BroadcastAddMember(req.RoomID, req.SelectedUsers)
}
