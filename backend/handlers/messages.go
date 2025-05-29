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

// メッセージ取得処理
func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("GetMessagesHandler：スタート")
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
	if roomIDStr == "" {
		http.Error(w, "ルームIDが必要です", http.StatusBadRequest)
		return
	}

	// 文字列を整数に変換
	roomID, err := strconv.Atoi(roomIDStr)
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

	// メッセージID、内容、作成時、送信者ID、送信者、既読フラグ+カウント、リアクションをセレクト
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

	// IDの配列を作る
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
		"status":   "success",
		"messages": filtered,
	})

}

// 未読処理
func UpdateMessageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("UpdateMessageHandler：スタート")
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

// メッセージ送信処理
func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("SendMessageHandler：スタート")
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

	var msg models.TsMessage

	// リクエストボディのデコード
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("JSONデコードエラー: %v", err)
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

	// メッセージを他のクライアントへブロードキャスト
	BroadcastMessage(message.RoomID, message)

	// 成功レスポンス
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "メッセージ保存完了",
		"data":    message,
	})
}

// メンション
func MentionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("MentionHandler：スタート")
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

	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("JSONデコードエラー: %v", err)
		http.Error(w, "リクエスト形式が不正", http.StatusBadRequest)
		return
	}

	for _, targetID := range msg.MentionedUserID {
		mention := db.Mentions{
			MessageID:         msg.MessageID,
			MentionedTargetID: targetID,
		}
		db.DB.Create(&mention)
	}

	var mentions []db.UnreadMentionCount

	err := db.DB.
		Table("mentions AS m").
		Select("msg.room_id, m.mentioned_target_id AS user_id, COUNT(*) AS unread_mentions").
		Joins("JOIN messages AS msg ON msg.id = m.message_id").
		Joins("LEFT JOIN message_reads AS mr ON mr.message_id = m.message_id AND mr.user_id = m.mentioned_target_id").
		Where("m.mentioned_target_id = ? AND mr.message_id IS NULL", msg.MentionedUserID).
		Group("msg.room_id, m.mentioned_target_id").
		Scan(&mentions).Error

	if err != nil {
		log.Println("❌ メンション未読取得失敗:", err)
	}

	// メッセージを他のクライアントへブロードキャスト
	if len(mentions) != 0 {
		BroadcastMention(msg.RoomID, mentions)
	}

	log.Println("メンション保存完了")
}

// メンションのためのルームメンバー一覧取得
func GetRoomMembersHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("GetRoomMembersHandler：スタート")
	w.Header().Set("Content-Type", "application/json")

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

	roomIDStr := r.URL.Query().Get("room_id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		http.Error(w, "不正なルームIDです", http.StatusBadRequest)
		return
	}

	// リクエストボディのパース
	var req struct {
		LoginUserID int `json:"login_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSONの解析に失敗しました", http.StatusBadRequest)
		return
	}

	type User struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	}

	var members []User

	err = db.DB.
		Table("users").
		Joins("JOIN room_members ON users.id = room_members.user_id").
		Where("room_members.room_id = ? AND users.id <> ?", roomID, req.LoginUserID).
		Select("users.id, users.username").
		Scan(&members).Error

	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"members": members,
	})
}

// ファイル送受信処理
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("UploadHandler：スタート")
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
	log.Printf("アップロード成功: %s", fileURL)

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

	message.Content = fileURL

	BroadcastMessage(message.RoomID, message)

	// 成功レスポンス
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "ファイル保存完了",
		"data":    message,
		"image":   fileURL,
	})
}
