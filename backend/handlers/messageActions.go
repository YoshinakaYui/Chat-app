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
	joinBroadcast := map[string]interface{}{
		"type":      "updataMessage",
		"messageid": id,
		"roomId":    reqBody.RoomID,
		"content":   reqBody.Content,
	}
	joinJSON, _ := json.Marshal(joinBroadcast)
	//log.Println("NNN：", joinJSON)

	broadcast <- joinJSON

	log.Println("🟡 メッセージ更新成功:", id)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "メッセージを更新しました",
	})

}

// メッセージ削除
func DeleteOnlyMessageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🟢DeleteOnlyMessageHandler：スタート")
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

	roomid, err := strconv.Atoi(reqBody.RoomID)
	if err != nil {
		http.Error(w, "ルームIDが不正です", http.StatusBadRequest)
		return
	}
	//userid, err := strconv.Atoi(reqBody.UserID)
	// if err != nil {
	// 	http.Error(w, "ユーザーIDが不正です", http.StatusBadRequest)
	// 	return
	// }
	log.Println("🟢-333：")

	// メッセージの保存
	message := db.Message{
		RoomID:    roomid,
		SenderID:  reqBody.UserID,
		Content:   "DeleteOnlyMessage:" + idStr,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// データベースに保存
	if err := db.DB.Create(&message).Error; err != nil {
		log.Println("メッセージ保存エラー:", err)
		http.Error(w, "メッセージ保存失敗", http.StatusInternalServerError)
		return
	}
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
		"roomId":    reqBody.RoomID,
		"content":   "（このメッセージは削除されました）",
	}
	joinJSON, _ := json.Marshal(joinBroadcast)
	//log.Println("NNN：", joinJSON)

	broadcast <- joinJSON

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
