package handlers

import (
	"backend/db"
	"backend/models"
	"backend/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type UnReadMsg struct {
	RoomID         int `json:"room_id"`
	LoggedInUserID int `json:"login_id"`
}

type SendMessages struct {
	MessageID  int       `json:"id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	Sender     int       `json:"sender"`
	SenderName string    `json:"sendername" gorm:"column:sendername"`
	AllRead    bool      `json:"allread" gorm:"column:allread"`
	ReadCount  int       `json:"readcount" gorm:"column:readcount"`
	Reactions  string    `json:"reaction" gorm:"colum:reactions"`
}

// メッセージ送信処理
func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	// メソッド確認
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}
	log.Println("🟣-11:", r.Method)

	var msg models.TsMessage

	// リクエストボディのデコード
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("JSONデコードエラー: %v", err)
		http.Error(w, "リクエスト形式が不正", http.StatusBadRequest)
		log.Println("🟣-22")
		return
	}
	log.Println("🟣-33")

	// 必須フィールドチェック
	if msg.RoomID == 0 || msg.SenderID == 0 || msg.Content == "" {
		http.Error(w, "必須フィールドが不足しています", http.StatusBadRequest)
		return
	}

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

	// データベースに保存
	if err := db.DB.Create(&message).Error; err != nil {
		log.Println("メッセージ保存エラー:", err)
		http.Error(w, "メッセージ保存失敗", http.StatusInternalServerError)
		return
	}

	// メッセージを他のクライアントへブロードキャスト
	sendBroadcast := map[string]interface{}{
		"type":        "postmessage",
		"room_id":     message.RoomID,
		"postmessage": message,
	}
	sendJSON, _ := json.Marshal(sendBroadcast)
	log.Println("NNN：", sendJSON)

	var decoded map[string]interface{}
	err2 := json.Unmarshal(sendJSON, &decoded)
	if err2 != nil {
		log.Println("JSONデコード失敗:", err2)
	}
	log.Println("PPP：", decoded)

	broadcast <- sendJSON

	log.Println("🟣-44：", message)

	// 成功レスポンス
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "メッセージ保存完了",
		"data":    message,
	})
}

// メッセージ取得処理
func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	// メソッド確認
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	// クエリパラメータからroom_idを取得
	roomIDStr := r.URL.Query().Get("room_id")
	log.Println("🟣-1：", roomIDStr)
	if roomIDStr == "" {
		http.Error(w, "ルームIDが必要です", http.StatusBadRequest)
		return
	}

	// 文字列を整数に変換
	roomID, err := strconv.Atoi(roomIDStr)
	log.Println("🟣-2：", roomID)
	if err != nil {
		http.Error(w, "ルームIDが不正です", http.StatusBadRequest)
		return
	}

	var user struct {
		Userid int `json:"login_id"`
	}

	// リクエストボディのデコード
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "リクエスト形式が不正", http.StatusBadRequest)
		return
	}

	var messages []SendMessages

	// メッセージID、内容、作成時、送信者ID、送信者、既読フラグ+カウント、リアクション
	selectFields := `
	m.id AS message_id,
	COALESCE(a.file_name, m.content) AS content,
	m.created_at,
	m.sender_id AS sender,
	u.username AS sendername,
	COUNT(DISTINCT mr.user_id) = COUNT(DISTINCT rm.user_id) AS allread,
	COUNT(DISTINCT mr.user_id) AS readcount,
	STRING_AGG(mr.reaction, ',') AS reactions
	`

	result := db.DB.Table("messages AS m").
		Select(selectFields).

		// JOIN句（送信者・添付・ルーム・既読・削除）
		Joins("JOIN users AS u ON m.sender_id = u.id").
		Joins("LEFT JOIN message_attachments AS a ON m.id = a.message_id").
		Joins("JOIN room_members AS rm ON m.room_id = rm.room_id").
		Joins("LEFT JOIN message_reads AS mr ON mr.message_id = m.id AND mr.user_id = rm.user_id").
		Joins("LEFT JOIN deleted_messages AS dm ON dm.message_id = m.id AND dm.user_id = ?", user.Userid).

		// WHERE＋GROUP＋ORDER
		Where("m.room_id = ? AND dm.id IS NULL", roomID).
		Group(`
			m.id, a.file_name, m.content,
			m.created_at, m.sender_id, u.username
		`).
		Order("m.created_at ASC").
		Scan(&messages)

	//log.Println("🟣ルームメッセージ一覧：", messages)

	//エラー処理
	if result.Error != nil {
		log.Println("メッセージ取得エラー:", result.Error)
		http.Error(w, "メッセージが見つかりません", http.StatusNotFound)
		return
	}

	type Message struct {
		ID               int
		Content          string
		CreatedAt        time.Time
		SenderID         int
		DeletedMessageID string `gorm:"column:deleted_message_id"`
	}

	var deletemessages []Message

	err1 := db.DB.Table("messages").
		Select(`
			messages.*,
			SUBSTRING(content, LENGTH('DeleteOnlyMessage:') + 1) AS deleted_message_id
		`).
		Where("room_id = ?", roomID).
		Where("sender_id = ?", user.Userid).
		Where("content LIKE ?", "DeleteOnlyMessage:%").
		Scan(&deletemessages).Error

	if err1 != nil {
		log.Println("DBエラー:", err1)
	}

	// 2. IDの配列を作る
	var deletedIDs []int
	for _, d := range deletemessages {
		var delmsgid, err3 = strconv.Atoi(d.DeletedMessageID)

		if err3 != nil {
			log.Println("DBエラー:", err3)
		}
		deletedIDs = append(deletedIDs, delmsgid)
	}

	filtered := make([]SendMessages, 0)
	for _, msg := range messages {
		found := false
		for _, delID := range deletedIDs {
			if msg.MessageID == delID {
				found = true
				break
			}
		}
		if !found {
			filtered = append(filtered, msg)
		}
	}

	// JSONレスポンス
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		// "messages": messages,
		"messages": filtered,
	})
	//log.Println("🟣ルームメッセージ一覧xxxxxx：", messages)

}

// ルームメンバー取得処理（LEFT JOINを使用）
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

// 未読処理
func UpdataMessageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🟠UpdataMessageHandler：スタート")
	utils.EnableCORS(w)

	// メソッド確認
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	// リクエストボディのデコード
	var msg UnReadMsg
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "リクエスト形式が不正", http.StatusBadRequest)
		return
	}

	// 必須フィールドチェック
	if msg.RoomID == 0 || msg.LoggedInUserID == 0 {
		http.Error(w, "必須フィールドが不足しています", http.StatusBadRequest)
		return
	}

	log.Println("🟠未読メッセージ WHERE：roomID:", msg.RoomID, "loggedinuserid:", msg.LoggedInUserID)

	var unreadIDs []int
	err := db.DB.Table("messages AS m").
		Select("m.id").
		Where("m.room_id = ?", msg.RoomID).
		Where("NOT EXISTS ("+
			"SELECT 1 FROM message_reads AS mr "+
			"WHERE mr.message_id = m.id AND mr.user_id = ?)", msg.LoggedInUserID).
		Order("m.id ASC").
		Scan(&unreadIDs).Error

	if err != nil {
		log.Println("❌ 未読メッセージIDの取得失敗:", err)
	} else {
		fmt.Println("📩 未読メッセージID:", unreadIDs)
	}

	log.Println("🟠未読メッセージID：", msg.LoggedInUserID, unreadIDs)

	var msgR db.MessageReads
	msgR.UserID = msg.LoggedInUserID
	msgR.ReadAt = time.Now()
	for i := 0; i < len(unreadIDs); i++ {
		log.Println(i)

		msgR.MessageID = unreadIDs[i]

		// データベースに保存
		if err := db.DB.Create(&msgR).Error; err != nil {
			log.Println("メッセージ保存エラー:", err)
			http.Error(w, "メッセージ保存失敗", http.StatusInternalServerError)
			return
		}
	}

	// unreadIDsのメッセージのallreadとreadcountをwebsocketに送信
	BroadcastReadCountsToRoom(msg.RoomID, unreadIDs)

	log.Println("KK：", unreadIDs)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "success",
		"message":       "既読完了",
		"readMessageID": unreadIDs,
	})
}

// 既読のブロードキャスト
func BroadcastReadCountsToRoom(roomID int, unreadIDs []int) {
	type SendMessages struct {
		RoomID    int  `json:"room_id" gorm:"column:room_id"`
		MessageID int  `json:"message_id" gorm:"column:message_id"`
		ReadCount int  `json:"readcount" gorm:"column:read_count"`
		AllRead   bool `json:"allread" gorm:"column:all_read"`
	}

	var result []SendMessages
	if len(unreadIDs) != 0 {
		err1 := db.DB.Table("messages AS m").
			Select(`
				m.room_id,
				m.id AS message_id,
				COUNT(DISTINCT r.user_id) AS read_count,
				COUNT(DISTINCT r.user_id) = COUNT(DISTINCT rm.user_id) AS all_read
				`).
			Joins("JOIN room_members rm ON m.room_id = rm.room_id").
			Joins("LEFT JOIN message_reads r ON m.id = r.message_id AND rm.user_id = r.user_id").
			Where("m.room_id = ? AND m.id IN ?", roomID, unreadIDs). // messageIDsは[]uintや[]intのスライス
			Group("m.id").
			Order("m.created_at ASC").
			Scan(&result)

		log.Println("MMM：", result)
		if err1 != nil {
			log.Println("❌ 新しい既読メッセージの取得失敗:", err1.Error)
		}
	}

	if len(result) != 0 {
		// 既読を他のクライアントへブロードキャスト
		joinBroadcast := map[string]interface{}{
			"type":           "newreadmessage",
			"newReadMessage": result,
			"room_id":        roomID,
		}
		joinJSON, _ := json.Marshal(joinBroadcast)
		//log.Println("NNN：", joinJSON)

		var decoded map[string]interface{}
		err2 := json.Unmarshal(joinJSON, &decoded)
		if err2 != nil {
			log.Println("JSONデコード失敗:", err2)
		}
		log.Println("PPP：", decoded)

		broadcast <- joinJSON
	}

}

// ファイル送受信処理
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🟠UploadHandler：スタート")
	utils.EnableCORS(w)

	// メソッド確認
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	// フォームの最大メモリサイズを指定
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		http.Error(w, "フォームのパースに失敗しました", http.StatusBadRequest)
		return
	}

	// ファイルの解析
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "ファイルを受け取れませんでした: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	senderID, err1 := strconv.Atoi(r.FormValue("senderID"))
	if err1 != nil {
		http.Error(w, "ファイルを受け取れませんでした: "+err1.Error(), http.StatusBadRequest)
		return
	}
	roomID, err := strconv.Atoi(r.FormValue("roomID"))
	if err != nil {
		http.Error(w, "ファイルを受け取れませんでした: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Println(senderID, roomID)

	// ファイル名を取得して表示
	fmt.Println("ファイル名:", handler.Filename)

	// 保存パスを作成
	saveDir := "./uploads/"
	os.MkdirAll(saveDir, os.ModePerm) // ディレクトリを作成

	savePath := saveDir + handler.Filename

	// ファイルを作成
	dst, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "ファイルを作成できません", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// ファイルをコピーして保存
	_, err = dst.ReadFrom(file)
	if err != nil {
		http.Error(w, "ファイルの保存に失敗しました", http.StatusInternalServerError)
		return
	}

	// アップロード成功
	fileURL := "http://localhost:8080/uploads/" + handler.Filename
	// fmt.Fprintf(w, "アップロード成功: %s\n", fileURL)
	log.Printf("アップロード成功: %s", fileURL)

	// // メッセージIDの取得
	// メッセージの保存
	message := db.Message{
		RoomID:    roomID,
		SenderID:  senderID,
		Content:   "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// データベースに保存
	if err := db.DB.Create(&message).Error; err != nil {
		log.Println("メッセージ保存エラー:", err)
		http.Error(w, "メッセージ保存失敗", http.StatusInternalServerError)
		return
	}

	att := db.MessageAttachment{
		MessageID: message.ID,
		FileName:  fileURL,
		CreatedAt: time.Now(),
	}

	if err := db.DB.Create(&att).Error; err != nil {
		log.Println("ファイル保存エラー:", err)
		http.Error(w, "ファイル保存失敗", http.StatusInternalServerError)
		return
	}

	// 🟣ファイルを他のクライアントへブロードキャスト
	message.Content = fileURL

	sendBroadcast := map[string]interface{}{
		"type":        "postmessage",
		"room_id":     message.RoomID,
		"postmessage": message,
	}
	sendJSON, _ := json.Marshal(sendBroadcast)
	log.Println("NNN：", sendJSON)

	var decoded map[string]interface{}
	err2 := json.Unmarshal(sendJSON, &decoded)
	if err2 != nil {
		log.Println("JSONデコード失敗:", err2)
	}
	log.Println("PPP：", decoded)

	broadcast <- sendJSON

	log.Println("🟠SendFileHandler：エンド")

	// 成功レスポンス
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "ファイル保存完了",
		"data":    message,
		"image":   fileURL,
	})
}

// メンション
func MentionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🟢MentionHandler：スタート")
	utils.EnableCORS(w)

	// メソッド確認
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}

	var msg struct {
		MessageID       int   `json:"message_id"`
		MentionedUserID []int `json:"mentioned_target_id"`
		RoomID          int   `json:"room_id"`
		SenderID        int   `json:"sender_id"`
	}

	// utils.JsonRawDataDisplay(w, r)
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("JSONデコードエラー: %v", err)
		http.Error(w, "リクエスト形式が不正", http.StatusBadRequest)
		return
	}
	log.Println("🟢-BB", msg)

	for _, targetID := range msg.MentionedUserID {
		mention := db.Mentions{
			MessageID:         msg.MessageID,
			MentionedTargetID: targetID,
		}
		db.DB.Create(&mention)
	}

	// type MentionUnread struct {
	// 	MessageID int       `json:"message_id"`
	// 	UserID    int       `json:"user_id"`
	// 	RoomID    int       `json:"room_id"`
	// 	Content   string    `json:"content"`
	// 	CreatedAt time.Time `json:"created_at"`
	// }

	// var mentions []MentionUnread

	// 未読のメンション通知
	type UnreadMentionCount struct {
		UserID         int   `json:"user_id"`
		RoomID         int   `json:"room_id"`
		UnreadMentions int64 `json:"unread_mentions"`
	}

	var mentions []UnreadMentionCount

	err := db.DB.
		Table("mentions AS m").
		Select("msg.room_id, m.mentioned_target_id AS user_id, COUNT(*) AS unread_mentions").
		Joins("JOIN messages AS msg ON msg.id = m.message_id").
		Joins("LEFT JOIN message_reads AS mr ON mr.message_id = m.message_id AND mr.user_id = m.mentioned_target_id").
		Where("m.mentioned_target_id = ? AND mr.message_id IS NULL", msg.MentionedUserID).
		Group("msg.room_id, m.mentioned_target_id").
		Scan(&mentions).Error

	// err := db.DB.
	// Raw(`
	// 	SELECT
	// 	  m.room_id,
	// 	  me.mentioned_target_id AS user_id,
	// 	  COUNT(*) AS unread_mentions
	// 	FROM mentions me
	// 	JOIN messages m ON m.id = me.message_id
	// 	WHERE me.mentioned_target_id = ?
	// 	  AND NOT EXISTS (
	// 	    SELECT 1
	// 	    FROM message_reads mr
	// 	    WHERE mr.message_id = me.message_id
	// 	      AND mr.user_id = me.mentioned_target_id
	// 	  )
	// 	GROUP BY m.room_id, me.mentioned_target_id
	// `, msg.MentionedUserID).
	// Scan(&mentions).Error

	if err != nil {
		log.Println("❌ メンション未読取得失敗:", err)
	}

	log.Println("🟢-DD", mentions)

	// メッセージを他のクライアントへブロードキャスト
	if len(mentions) != 0 {
		// メンションをクライアントへブロードキャスト
		mentionBroadcast := map[string]interface{}{
			"type":    "mention",
			"Mention": mentions,
			"room_id": msg.RoomID,
		}
		mentionJSON, _ := json.Marshal(mentionBroadcast)
		// log.Println("NNN：", mentionJSON)

		var decoded map[string]interface{}
		err2 := json.Unmarshal(mentionJSON, &decoded)
		if err2 != nil {
			log.Println("JSONデコード失敗:", err2)
		}
		log.Println("🟢メンションデコード：", decoded)

		broadcast <- mentionJSON
	}

	log.Println("メンション保存完了")
}
