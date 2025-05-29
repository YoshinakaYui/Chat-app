package handlers

import (
	"backend/db"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type IncomingMessage struct {
	MessageID  int    `json:"id"`
	SenderID   int    `json:"sender"`
	SenderName string `json:"sendername"`
	Content    string `json:"content"`
}

type JoinEvent struct {
	Type   string `json:"type"` // "join"
	RoomID int    `json:"roomId"`
	UserID int    `json:"userId"`
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan []byte)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocket接続を初期化する関数
func InitWebSocketConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	ws, err := upgrader.Upgrade(w, r, nil)
	log.Println("##")
	if err != nil {
		log.Println("接続のアップグレードエラー:", err)
		return nil, err
	}
	return ws, nil
}

// WebSocketでメッセージ、入室管理
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Println("HandleWebSocket")

	// WebSocket接続を初期化
	ws, err := InitWebSocketConnection(w, r)
	if err != nil {
		log.Println("WebSocket接続エラー:", err)
		return
	}
	defer func() {
		log.Println("🛑 WebSocket切断:", ws.RemoteAddr())
		ws.Close()
		delete(clients, ws)
	}()
	//defer ws.Close()

	// クライアントをmapに追加
	clients[ws] = true
	log.Println("WebSocket接続確立:", ws.RemoteAddr())

	for {
		// メッセージの受信
		_, msg, err := ws.ReadMessage()
		log.Printf("📥 raw受信メッセージ（%dバイト）: %s", len(msg), string(msg))

		if err != nil {
			fmt.Println("受信エラー:", err)
			delete(clients, ws)
			break
		}

		str := string(msg)
		if strings.Contains(str, `"type":"join"`) {
			// 入室通知のブロードキャスト
			var join JoinEvent
			if err := json.Unmarshal(msg, &join); err != nil {
				fmt.Println("JSON解析エラー:", err)
				continue
			}

			log.Printf("🟦： ユーザー %d がルーム %d に入室", join.UserID, join.RoomID)

			// 入室通知を他のクライアントへブロードキャスト
			joinBroadcast := map[string]interface{}{
				"type":   "user_joined",
				"userId": join.UserID,
				"roomId": join.RoomID,
			}
			joinJSON, _ := json.Marshal(joinBroadcast)
			broadcast <- joinJSON

		}

	}

}

// ブロードキャスト処理
func BroadcastMessages() {
	for {
		msg := <-broadcast
		log.Println("🟦ブロード：", string(msg))
		for client := range clients {

			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// メンションをクライアントへブロードキャスト
func BroadcastMention(roomID int, mentions []db.UnreadMentionCount) {
	log.Println("BroadcastMention：スタート")
	mentionBroadcast := map[string]interface{}{
		"type":    "mention",
		"room_id": roomID,
		"Mention": mentions,
	}
	mentionJSON, _ := json.Marshal(mentionBroadcast)
	// log.Println("NNN：", mentionJSON)

	var decoded map[string]interface{}
	err2 := json.Unmarshal(mentionJSON, &decoded)
	if err2 != nil {
		log.Println("JSONデコード失敗:", err2)
	}

	broadcast <- mentionJSON
}

// ルーム作成をブロードキャスト
func BroadcastCreateRomm(roomID int, groupName string, members map[int]string, isGroup int) {
	log.Println("BroadcastCreateRomm：スタート")

	type memberObj struct {
		UserID    int    `json:"user_id"`
		GroupName string `json:"group_name"`
	}

	var memberlist []memberObj
	for key, value := range members {
		fmt.Printf("キー: %d, 値: %s\n", key, value)
		memberlist = append(memberlist, memberObj{UserID: key, GroupName: value})
	}
	log.Println("memberlist：", memberlist)

	roomBroadcast := map[string]interface{}{
		"type":       "createroom",
		"memberlist": memberlist,
		"room_id":    roomID,
		"is_group":   isGroup,
	}
	roomJSON, _ := json.Marshal(roomBroadcast)

	broadcast <- roomJSON
}

// 既読のブロードキャスト
func BroadcastReadCountsToRoom(roomID int, unreadIDs []int) {
	log.Println("BroadcastReadCountsToRoom：スタート")

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

		var decoded map[string]interface{}
		err2 := json.Unmarshal(joinJSON, &decoded)
		if err2 != nil {
			log.Println("JSONデコード失敗:", err2)
		}

		broadcast <- joinJSON
	}
}

// 未読をブロードキャスト
func BroadcastUnreadMessage(roomID int, message []db.UnreadResult) {
	log.Println("BroadcastUnreadMessage：スタート")

	joinBroadcast := map[string]interface{}{
		"type":          "unreadmessage",
		"unReadMessage": message,
		"room_id":       roomID,
	}
	joinJSON, _ := json.Marshal(joinBroadcast)
	log.Println("未読ブロードキャスト：", joinJSON)

	broadcast <- joinJSON
}

// メッセージをブロードキャスト
func BroadcastMessage(roomID int, message db.Message) {
	log.Println("BroadcastMessage：スタート")

	type postmsgObj struct {
		ID           int       `json:"ID"`
		RoomID       int       `json:"RoomId"`
		SenderID     int       `json:"SenderID"`
		SenderName   string    `json:"SenderName"`
		Content      string    `json:"Content"`
		CreatedAt    time.Time `json:"CreatedAt"`
		UpdatedAt    time.Time `json:"UpdatedAt"`
		ThreadRootID int       `json:"ThreadRootID"`
	}

	sendername, err := db.GetUserName(message.SenderID)
	if err != nil {
		log.Println("❌ ユーザー名の取得失敗:", err)
	}

	var postmsg postmsgObj
	postmsg.ID = message.ID
	postmsg.RoomID = message.RoomID
	postmsg.SenderID = message.SenderID
	postmsg.SenderName = sendername
	postmsg.Content = message.Content
	postmsg.CreatedAt = message.CreatedAt
	postmsg.UpdatedAt = message.UpdatedAt
	postmsg.ThreadRootID = message.ThreadRootID

	sendBroadcast := map[string]interface{}{
		"type":        "postmessage",
		"room_id":     roomID,
		"postmessage": postmsg,
	}
	sendJSON, _ := json.Marshal(sendBroadcast)

	var decoded map[string]interface{}
	err2 := json.Unmarshal(sendJSON, &decoded)
	if err2 != nil {
		log.Println("JSONデコード失敗:", err2)
	}

	broadcast <- sendJSON

}

// メッセージ編集をブロードキャスト
func BroadcastEditMessage(roomID string, id int, content string) {
	log.Println("BroadcastEditMessage：スタート")

	messageBroadcast := map[string]interface{}{
		"type":      "updateMessage",
		"messageid": id,
		"room_id":   roomID,
		"content":   content,
	}
	messageJSON, _ := json.Marshal(messageBroadcast)

	broadcast <- messageJSON
}

// 送信取消をブロードキャスト
func BroadcastDeleteMessage(roomID string, id int) {
	log.Println("BroadcastDeleteMessage：スタート")

	joinBroadcast := map[string]interface{}{
		"type":      "updateMessage",
		"messageid": id,
		"room_id":   roomID,
		"content":   "（このメッセージは削除されました）",
	}
	joinJSON, _ := json.Marshal(joinBroadcast)

	broadcast <- joinJSON
}

// リアクションをブロードキャスト
func BroadcastReaction(roomID int, userID int, messageID int, reaction string) {
	log.Println("BroadcastReaction：スタート")

	messageBroadcast := map[string]interface{}{
		"type":      "reaction",
		"messageid": messageID,
		"room_id":   roomID,
		"user_id":   userID,
		"reaction":  reaction,
	}
	messageJSON, _ := json.Marshal(messageBroadcast)

	broadcast <- messageJSON
}

// ルーム退出をクライアントへブロードキャスト
func BroadcastLeaveRoom(roomID int, users []int) {
	log.Println("BroadcastLeaveRoom：スタート")

	joinBroadcast := map[string]interface{}{
		"type":    "leaveroom",
		"userids": users,
		"room_id": roomID,
	}
	joinJSON, _ := json.Marshal(joinBroadcast)

	log.Println("ルーム退出：")

	broadcast <- joinJSON
}

// メンバー追加をブロードキャスト
func BroadcastAddMember(roomID int, users []int) {
	log.Println("BroadcastAddMember：スタート")

	joinBroadcast := map[string]interface{}{
		"type":    "addmembers",
		"userids": users,
		"room_id": roomID,
	}
	joinJSON, _ := json.Marshal(joinBroadcast)

	log.Println("ルーム退出：")

	broadcast <- joinJSON
}
