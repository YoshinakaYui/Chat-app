package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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

		} else {
			// メッセージのブロードキャスト
			// ② 通常のチャットメッセージ処理
			var incoming IncomingMessage
			if err := json.Unmarshal(msg, &incoming); err != nil {
				fmt.Println("JSON解析エラー:", err)
				continue
			}

			log.Println("🟦：", incoming)

			outJSON, err := json.Marshal(incoming)
			if err != nil {
				fmt.Println("JSON変換エラー:", err)
				continue
			}

			// 皆にメッセージを送信（ブロードキャスト）
			broadcast <- outJSON

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
