package handlers

import (
	"backend/db"
	"backend/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// メッセージ編集
func EditMessageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🟡EditMessageHandler：スタート")
	utils.EnableCORS(w)

	// メソッド確認
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPut {
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}
	log.Println("🟡メソッド：", r.Method)

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	log.Println("🟡id：", id)
	if err != nil {
		http.Error(w, "無効なID形式です", http.StatusBadRequest)
		return
	}

	// リクエストボディのパース
	var reqBody struct {
		Content string `json:"content"`
		RoomID  string `json:"room_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "JSONの解析に失敗しました", http.StatusBadRequest)
		return
	}
	if reqBody.Content == "" {
		http.Error(w, "contentが空です", http.StatusBadRequest)
		return
	}

	// 更新処理
	if err := db.DB.Table("messages").
		Where("id = ?", id).
		UpdateColumns(map[string]interface{}{
			"content":    reqBody.Content,
			"updated_at": time.Now(),
		}).Error; err != nil {
		log.Println("更新失敗:", err)
		http.Error(w, "更新に失敗しました", http.StatusInternalServerError)
		return
	}

	// 既読を他のクライアントへブロードキャスト
	messageBroadcast := map[string]interface{}{
		"type":      "updataMessage",
		"messageid": id,
		"room_id":   reqBody.RoomID,
		"content":   reqBody.Content,
	}
	messageJSON, _ := json.Marshal(messageBroadcast)
	//log.Println("NNN：", joinJSON)

	broadcast <- messageJSON

	log.Println("🟡 メッセージ更新成功:", id)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "メッセージを更新しました",
	})

}

// メッセージ削除
func DeleteMyMessageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("DeleteMyMessageHandler")
	utils.EnableCORS(w)

	// メソッド確認
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		log.Println("🟢-000メソッド")
		return
	}
	if r.Method != http.MethodPut {
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		log.Println("🟢メソッド：", r.Method)
		return
	}
	log.Println("🟢メソッド2：", r.Method)

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ルームIDが不正です", http.StatusBadRequest)
		return
	}
	log.Println("🟢メッセージID：", idStr)

	// リクエストボディのパース
	var reqBody struct {
		UserID int    `json:"login_id"`
		RoomID string `json:"room_id"`
	}

	//utils.JsonRawDataDisplay(w, r)
	// リクエストボディのデコード
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Println("🟢デコード：", err)
		http.Error(w, "リクエスト形式が不正", http.StatusBadRequest)
		return
	}
	log.Println("🟢-111：")

	// 必須フィールドチェック
	if reqBody.RoomID == "" || reqBody.UserID == 0 || idStr == "" {
		http.Error(w, "必須フィールドが不足しています", http.StatusBadRequest)
		return
	}
	log.Println("🟢-222：")

	//roomid, err := strconv.Atoi(reqBody.RoomID)
	// if err != nil {
	// 	http.Error(w, "ルームIDが不正です", http.StatusBadRequest)
	// 	return
	// }

	log.Println("🟢-333：")

	// // メッセージの保存 // message_id, user_id, atでテーブル作る
	// message := db.Message{
	// 	RoomID:    roomid,
	// 	SenderID:  reqBody.UserID,
	// 	Content:   "DeleteOnlyMessage:" + idStr,
	// 	CreatedAt: time.Now(),
	// 	UpdatedAt: time.Now(),
	// }

	// // データベースに保存
	// if err := db.DB.Create(&message).Error; err != nil {
	// 	log.Println("メッセージ保存エラー:", err)
	// 	http.Error(w, "メッセージ保存失敗", http.StatusInternalServerError)
	// 	return
	// }

	delete := db.DeletedMessage{
		MessageID: id,
		UserID:    reqBody.UserID,
		DeletedAt: time.Now(),
	}

	// データベースに保存
	if err := db.DB.Create(&delete).Error; err != nil {
		log.Println("メッセージ保存エラー:", err)
		http.Error(w, "メッセージ保存失敗", http.StatusInternalServerError)
		return
	}
	log.Println("🟢deleteデータベースに保存完了")

}

// メッセージ送信取消
func DeleteMessageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🟢DeleteMessageHandler：スタート")
	utils.EnableCORS(w)

	type Message struct {
		ID uint `gorm:"primaryKey"`
		// 他のフィールドは省略
	}

	type MessageAttachment struct {
		ID        uint `gorm:"primaryKey"`
		MessageID uint
	}

	type Mention struct {
		ID        uint `gorm:"primaryKey"`
		MessageID uint
	}

	type MessageRead struct {
		ID        uint `gorm:"primaryKey"`
		MessageID uint
	}

	// メソッド確認
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPut {
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}
	log.Println("🟢メソッド：", r.Method)

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	log.Println("🟢id：", id)
	if err != nil {
		http.Error(w, "無効なID形式です", http.StatusBadRequest)
		return
	}

	// トランザクション開始
	tx := db.DB.Begin()
	if tx.Error != nil {
		log.Println("トランザクション開始失敗:", tx.Error)
		return
	}
	defer tx.Rollback()
	log.Println("🟢：AA")

	// 1. message_attachments
	if err := tx.Where("message_id = ?", id).Delete(&MessageAttachment{}).Error; err != nil {
		tx.Rollback()
		http.Error(w, "message_attachments 削除失敗", http.StatusInternalServerError)
		return
	}
	log.Println("🟢：BB")
	// 2. mentions
	if err := tx.Where("message_id = ?", id).Delete(&Mention{}).Error; err != nil {
		tx.Rollback()
		http.Error(w, "mentions 削除失敗", http.StatusInternalServerError)
		return
	}
	log.Println("🟢：CC")
	// 3. message_reads
	if err := tx.Where("message_id = ?", id).Delete(&MessageRead{}).Error; err != nil {
		tx.Rollback()
		http.Error(w, "message_reads 削除失敗", http.StatusInternalServerError)
		return
	}
	log.Println("🟢：DD")
	// 4. messages(本体)
	if err := tx.Delete(&Message{}, id).Error; err != nil {
		tx.Rollback()
		http.Error(w, "messages 削除失敗", http.StatusInternalServerError)
		return
	}
	log.Println("🟢：EE")
	if err := tx.Commit().Error; err != nil {
		http.Error(w, "コミット失敗", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "メッセージ %d を削除しました", id)

	// リクエストボディのパース
	var reqBody struct {
		RoomID string `json:"room_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "JSONの解析に失敗しました", http.StatusBadRequest)
		return
	}

	// 送信取消を他のクライアントへブロードキャスト
	joinBroadcast := map[string]interface{}{
		"type":      "updataMessage",
		"messageid": id,
		"room_id":   reqBody.RoomID,
		"content":   "（このメッセージは削除されました）",
	}
	joinJSON, _ := json.Marshal(joinBroadcast)
	//log.Println("NNN：", joinJSON)

	broadcast <- joinJSON

	log.Println("🟢DeleteMessageHandler：エンド")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("削除成功"))
}

// メッセージリアクション
func ReactionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🔵ReactionHandler：スタート")
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
	log.Println("🔵メソッド：", r.Method)

	var req struct {
		MessageID int    `json:"message_id"`
		UserID    int    `json:"user_id"`
		Reaction  string `json:"reaction"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("🔵デコード：", err)
		http.Error(w, "リクエスト形式が不正", http.StatusBadRequest)
		return
	}
	log.Println("🔵-111：")

	//var read db.MessageReads

	err := db.DB.Model(&db.MessageReads{}).
		Where("message_id = ? AND user_id = ?", req.MessageID, req.UserID).
		Updates(map[string]interface{}{
			"reaction": req.Reaction,
			"read_at":  time.Now(),
		}).Error

	if err != nil {
		log.Println("リアクション更新エラー:", err)
		http.Error(w, "リアクションの更新に失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	w.Write([]byte("リアクション成功"))
	log.Println("🔵リアクション成功")
}

// メンションのためのルームメンバー一覧取得
func GetRoomMembersHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🟢GetRoomMembersHandler：スタート")
	w.Header().Set("Content-Type", "application/json")

	utils.EnableCORS(w)

	// メソッド確認
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	log.Println("🟢GetRoomMembers メソッド：", r.Method)
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

	log.Println("🟢GetRoomMembers ユーザーID：", req.LoginUserID)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("🟢GetRoomMembers JSON：", req)
		http.Error(w, "JSONの解析に失敗しました", http.StatusBadRequest)
		return
	}

	log.Println("🟢GetRoomMembers-1：", req.LoginUserID)
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
	log.Println("🟢GetRoomMembers-2")
	log.Println("🟢GetRoomMembers：", members)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"members": members,
	})
}
