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

	log.Println("🟡 メッセージ更新成功:", id)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "メッセージを更新しました",
	})

}

// メッセージ削除
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
	if r.Method != http.MethodDelete {
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

	log.Println("🟢DeleteMessageHandler：エンド")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("削除成功"))

	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(map[string]interface{}{
	// 	"status": "success",
	// })
}

// メッセージリアクション
func ReactionMessageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🔵EditMessageHandler：スタート")
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
	log.Println("🔵メソッド：", r.Method)

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	log.Println("🔵id：", id)
	if err != nil {
		http.Error(w, "無効なID形式です", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("リアクション成功"))

}
