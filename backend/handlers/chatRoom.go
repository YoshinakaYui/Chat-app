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

//	type CreateChatRoomRequest struct {
//		User1ID int `json:"login_id"`
//		User2ID int `json:"user_id"`
//	}
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

	var is_group int = 0
	if reqGroup.GroupName == "" {
		// 個別チャット
		//log.Println("🟡CreateGroupRoom：", len(reqGroup.SelectedUsers))

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

	// メンバー登録
	// 作成したchat_roomsのidとreqGroupのuser_ids(複数ユーザー)をroom_membersに保存
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
	// GroupNameの確定.
	groupname := reqGroup.GroupName
	if is_group == 0 {
		// 個別チャット.

		var username string
		err := db.DB.
			Table("users").
			Select("username").
			Where("id = ?", reqGroup.SelectedUsers[0]).
			Scan(&username).Error

		if err != nil {
			log.Println("❌ ユーザー名の取得失敗:", err)
		}

		groupname = username

	}
	reqGroup.SelectedUsers = append(reqGroup.SelectedUsers, reqGroup.LoggedInUserID)

	roomBroadcast := map[string]interface{}{
		"type":       "createroom",
		"memberlist": reqGroup.SelectedUsers,
		"roomname":   groupname,
		"room_id":    room.ID,
		"is_group":   is_group,
	}
	roomJSON, _ := json.Marshal(roomBroadcast)
	// log.Println("NNN：", mentionJSON)

	var decoded map[string]interface{}
	err2 := json.Unmarshal(roomJSON, &decoded)
	if err2 != nil {
		log.Println("JSONデコード失敗:", err2)
	}

	broadcast <- roomJSON

	log.Println("新規グループルーム作成成功:", room.ID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "新規グループチャットルームを作成しました",
		"roomId":  room.ID,
	})

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
	joinBroadcast := map[string]interface{}{
		"type":    "leaveroom",
		"userids": userIDs,
		"room_id": req.RoomID,
	}
	joinJSON, _ := json.Marshal(joinBroadcast)

	log.Println("ルーム退出：")

	broadcast <- joinJSON

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
	log.Println("🐣AddMemberHandler：", r.Method)

	// リクエストボディからデータを取得
	var req addMember
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("リクエストボディのデコードエラー:", err)
		http.Error(w, "リクエスト形式が不正です", http.StatusBadRequest)
		return
	}
	log.Println("🐣AddMemberHandler_req：", req)

	// メンバー登録
	var AddMembers []db.RoomMember

	for i := 0; i < len(req.SelectedUsers); i++ {
		AddMembers = append(AddMembers, db.RoomMember{RoomID: req.RoomID, UserID: req.SelectedUsers[i], JoinedAt: time.Now()})
	}

	log.Println("追加できたと思う", AddMembers)

	if err := db.DB.Create(&AddMembers).Error; err != nil {
		log.Println("room_members作成エラー：", err)
		http.Error(w, "メンバー作成失敗", http.StatusInternalServerError)
		return
	}

	log.Println("追加しました", AddMembers)

	// message_readsに記録
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
	joinBroadcast := map[string]interface{}{
		"type":    "addmembers",
		"userids": req.SelectedUsers,
		"room_id": req.RoomID,
	}
	joinJSON, _ := json.Marshal(joinBroadcast)

	log.Println("ルーム退出：")

	broadcast <- joinJSON
}

// メンバー追加のためのルームに存在しないをユーザー取得
func UsersNotInRoomHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🐶AddMemberHandler：スタート")
	utils.EnableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	log.Println("🐶AddMemberHandler：スタート", r.Method)

	idStr := r.URL.Query().Get("room_id")
	log.Println("🐶AddMemberHandler ルームID：", idStr)
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ルームIDが不正です", http.StatusBadRequest)
		return
	}
	log.Println("🐶AddMemberHandler", id)

	// リクエストボディのパース
	var req struct {
		LoginUserID int   `json:"login_id"`
		Members     []int `json:"members"`
	}

	//utils.JsonRawDataDisplay(w, r)
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
