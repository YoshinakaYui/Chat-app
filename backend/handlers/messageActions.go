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
	log.Println("EditMessageHandler：スタート")
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

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
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

	// 編集を他のクライアントへブロードキャスト
	BroadcastEditMessage(reqBody.RoomID, id, reqBody.Content)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "メッセージを更新しました",
	})

}

// メッセージ削除
func DeleteMyMessageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("DeleteMyMessageHandler：スタート")
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

	// リクエストボディのパース
	var reqBody struct {
		UserID int    `json:"login_id"`
		RoomID string `json:"room_id"`
	}

	// リクエストボディのデコード
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "リクエスト形式が不正", http.StatusBadRequest)
		return
	}

	// 必須フィールドチェック
	if reqBody.RoomID == "" || reqBody.UserID == 0 || idStr == "" {
		http.Error(w, "必須フィールドが不足しています", http.StatusBadRequest)
		return
	}

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
}

// メッセージ送信取消
func DeleteMessageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("DeleteMessageHandler：スタート")
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

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
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

	// 1. message_attachments
	if err := tx.Where("message_id = ?", id).Delete(&MessageAttachment{}).Error; err != nil {
		tx.Rollback()
		http.Error(w, "message_attachments 削除失敗", http.StatusInternalServerError)
		return
	}

	// 2. mentions
	if err := tx.Where("message_id = ?", id).Delete(&Mention{}).Error; err != nil {
		tx.Rollback()
		http.Error(w, "mentions 削除失敗", http.StatusInternalServerError)
		return
	}

	// 3. message_reads
	if err := tx.Where("message_id = ?", id).Delete(&MessageRead{}).Error; err != nil {
		tx.Rollback()
		http.Error(w, "message_reads 削除失敗", http.StatusInternalServerError)
		return
	}

	// 4. messages(本体)
	if err := tx.Delete(&Message{}, id).Error; err != nil {
		tx.Rollback()
		http.Error(w, "messages 削除失敗", http.StatusInternalServerError)
		return
	}

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
	BroadcastDeleteMessage(reqBody.RoomID, id)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("削除成功"))
}

// メッセージリアクション
func ReactionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("ReactionHandler：スタート")
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

	var req struct {
		MessageID int    `json:"message_id"`
		UserID    int    `json:"user_id"`
		RoomID    int    `json:"room_id"`
		Reaction  string `json:"reaction"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "リクエスト形式が不正", http.StatusBadRequest)
		return
	}

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

	// リアクションをブロードキャスト
	BroadcastReaction(req.RoomID, req.UserID, req.MessageID, req.Reaction)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	w.Write([]byte("リアクション成功"))
}
