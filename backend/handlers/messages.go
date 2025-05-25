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

// メッセージ送信・取得ハンドラー
func MessageHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		handleSendMessage(w, r)
	case http.MethodPut:
		handleGetMessages(w, r)
	case http.MethodOptions:
		w.WriteHeader(http.StatusOK)
		return
	default:
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
	}
}

// メッセージ更新処理
func handleSendMessage(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(w)

	var msg models.TsMessage

	// リクエストボディのデコード
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "リクエスト形式が不正", http.StatusBadRequest)
		return
	}

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

	// 🟣メッセージを他のクライアントへブロードキャスト
	sendBroadcast := map[string]interface{}{
		"type":        "postmessage",
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

	log.Println("🟣：", message)

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

	// クエリパラメータからroom_idを取得
	roomIDStr := r.URL.Query().Get("room_id")
	log.Println("🟣：", roomIDStr)
	if roomIDStr == "" {
		http.Error(w, "ルームIDが必要です", http.StatusBadRequest)
		return
	}

	// 文字列を整数に変換
	roomID, err := strconv.Atoi(roomIDStr)
	log.Println("🟣：", roomID)
	if err != nil {
		http.Error(w, "ルームIDが不正です", http.StatusBadRequest)
		return
	}

	var user struct {
		Userid int `json:"login_id"`
	}

	//utils.JsonRawDataDisplay(w, r)

	// リクエストボディのデコード
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "リクエスト形式が不正", http.StatusBadRequest)
		return
	}

	// メッセージを格納するスライス
	// var SendMessages []struct {
	// 	MessageID  int    `json:"message_id"`
	// 	Content    string `json:"content"`
	// 	CreatedAt  string `json:"created_at"`
	// 	Sender     int    `json:"sender"`
	// 	SenderName string `json:"sendername"`
	// 	AllRead    bool   `json:"allread"` // 既読のカウント変数、（SQLに変数＝１しとく）0以外は未読
	// }
	type SendMessages struct {
		MessageID  int       `json:"id"`
		Content    string    `json:"content"`
		CreatedAt  time.Time `json:"created_at"`
		Sender     int       `json:"sender"`
		SenderName string    `json:"sendername" gorm:"column:sendername"`
		AllRead    bool      `json:"allread" gorm:"column:allread"`
		ReadCount  int       `json:"readcount" gorm:"column:readcount"`
	}

	var messages []SendMessages

	// result := db.DB.Table("messages AS m").
	// 	Select(`
	// 		m.id AS message_id,
	// 		COALESCE(a.file_name, m.content) AS content,
	// 		m.created_at,
	// 		m.sender_id AS sender,
	// 		u.username AS sendername,
	// 		COUNT(DISTINCT mr.user_id) = COUNT(DISTINCT rm.user_id) AS allread,
	// 		COUNT(DISTINCT mr.user_id) AS readcount`).
	// 	Joins("JOIN users AS u ON m.sender_id = u.id").
	// 	Joins("LEFT JOIN message_attachments AS a ON m.id = a.message_id").
	// 	Joins("JOIN room_members AS rm ON m.room_id = rm.room_id").
	// 	Joins("LEFT JOIN message_reads AS mr ON mr.message_id = m.id AND mr.user_id = rm.user_id").
	// 	Where("m.room_id = ?", roomID).
	// 	Group("m.id, a.file_name, m.content, m.created_at, m.sender_id, u.username").
	// 	Order("m.created_at ASC").
	// 	Scan(&messages)

	result := db.DB.Table("messages AS m").
		Select(`
		m.id AS message_id,
		COALESCE(a.file_name, m.content) AS content,
		m.created_at,
		m.sender_id AS sender,
		u.username AS sendername,
		COUNT(DISTINCT mr.user_id) = COUNT(DISTINCT rm.user_id) AS allread,
		COUNT(DISTINCT mr.user_id) AS readcount`).
		Joins("JOIN users AS u ON m.sender_id = u.id").
		Joins("LEFT JOIN message_attachments AS a ON m.id = a.message_id").
		Joins("JOIN room_members AS rm ON m.room_id = rm.room_id").
		Joins("LEFT JOIN message_reads AS mr ON mr.message_id = m.id AND mr.user_id = rm.user_id").
		Where("m.room_id = ?", roomID).
		Where("m.content NOT LIKE ?", "DeleteOnlyMessage:%").
		Group("m.id, a.file_name, m.content, m.created_at, m.sender_id, u.username").
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

// 既読未読処理
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
			"roomId":         roomID,
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

// func DeleteMessageHandler(w http.ResponseWriter, r *http.Request) {
// 	log.Println("🟢DeleteMessageHandler：スタート")
// 	utils.EnableCORS(w)

// 	type Message struct {
// 		ID uint `gorm:"primaryKey"`
// 		// 他のフィールドは省略
// 	}

// 	type MessageAttachment struct {
// 		ID        uint `gorm:"primaryKey"`
// 		MessageID uint
// 	}

// 	type Mention struct {
// 		ID        uint `gorm:"primaryKey"`
// 		MessageID uint
// 	}

// 	type MessageRead struct {
// 		ID        uint `gorm:"primaryKey"`
// 		MessageID uint
// 	}

// 	// メソッド確認
// 	if r.Method == http.MethodOptions {
// 		w.WriteHeader(http.StatusOK)
// 		return
// 	}
// 	if r.Method != http.MethodDelete {
// 		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
// 		return
// 	}
// 	log.Println("🟢メソッド：", r.Method)

// 	idStr := r.URL.Query().Get("id")
// 	if idStr == "" {
// 		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
// 		return
// 	}

// 	id, err := strconv.Atoi(idStr)
// 	log.Println("🟢id：", id)
// 	if err != nil {
// 		http.Error(w, "無効なID形式です", http.StatusBadRequest)
// 		return
// 	}

// 	// トランザクション開始
// 	tx := db.DB.Begin()
// 	if tx.Error != nil {
// 		log.Println("トランザクション開始失敗:", tx.Error)
// 		return
// 	}
// 	defer tx.Rollback()
// 	log.Println("🟢：AA")

// 	// 1. message_attachments
// 	if err := tx.Where("message_id = ?", id).Delete(&MessageAttachment{}).Error; err != nil {
// 		tx.Rollback()
// 		http.Error(w, "message_attachments 削除失敗", http.StatusInternalServerError)
// 		return
// 	}
// 	log.Println("🟢：BB")
// 	// 2. mentions
// 	if err := tx.Where("message_id = ?", id).Delete(&Mention{}).Error; err != nil {
// 		tx.Rollback()
// 		http.Error(w, "mentions 削除失敗", http.StatusInternalServerError)
// 		return
// 	}
// 	log.Println("🟢：CC")
// 	// 3. message_reads
// 	if err := tx.Where("message_id = ?", id).Delete(&MessageRead{}).Error; err != nil {
// 		tx.Rollback()
// 		http.Error(w, "message_reads 削除失敗", http.StatusInternalServerError)
// 		return
// 	}
// 	log.Println("🟢：DD")
// 	// 4. messages(本体)
// 	if err := tx.Delete(&Message{}, id).Error; err != nil {
// 		tx.Rollback()
// 		http.Error(w, "messages 削除失敗", http.StatusInternalServerError)
// 		return
// 	}
// 	log.Println("🟢：EE")
// 	if err := tx.Commit().Error; err != nil {
// 		http.Error(w, "コミット失敗", http.StatusInternalServerError)
// 		return
// 	}
// 	fmt.Fprintf(w, "メッセージ %d を削除しました", id)

// 	log.Println("🟢DeleteMessageHandler：エンド")

// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte("削除成功"))

// 	// w.WriteHeader(http.StatusOK)
// 	// json.NewEncoder(w).Encode(map[string]interface{}{
// 	// 	"status": "success",
// 	// })
// }

// // メッセージ編集
// func EditMessageHandler(w http.ResponseWriter, r *http.Request) {
// 	log.Println("🟡EditMessageHandler：スタート")
// 	utils.EnableCORS(w)

// 	// メソッド確認
// 	if r.Method == http.MethodOptions {
// 		w.WriteHeader(http.StatusOK)
// 		return
// 	}
// 	if r.Method != http.MethodPut {
// 		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
// 		return
// 	}
// 	log.Println("🟡メソッド：", r.Method)

// 	idStr := r.URL.Query().Get("id")
// 	if idStr == "" {
// 		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
// 		return
// 	}

// 	id, err := strconv.Atoi(idStr)
// 	log.Println("🟡id：", id)
// 	if err != nil {
// 		http.Error(w, "無効なID形式です", http.StatusBadRequest)
// 		return
// 	}

// 	// リクエストボディのパース
// 	var reqBody struct {
// 		Content string `json:"content"`
// 	}
// 	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
// 		http.Error(w, "JSONの解析に失敗しました", http.StatusBadRequest)
// 		return
// 	}
// 	if reqBody.Content == "" {
// 		http.Error(w, "contentが空です", http.StatusBadRequest)
// 		return
// 	}

// 	// 更新処理
// 	if err := db.DB.Table("messages").
// 		Where("id = ?", id).
// 		UpdateColumns(map[string]interface{}{
// 			"content":    reqBody.Content,
// 			"updated_at": time.Now(),
// 		}).Error; err != nil {
// 		log.Println("更新失敗:", err)
// 		http.Error(w, "更新に失敗しました", http.StatusInternalServerError)
// 		return
// 	}

// 	log.Println("🟡 メッセージ更新成功:", id)
// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"status":  "success",
// 		"message": "メッセージを更新しました",
// 	})

// }
