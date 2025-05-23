package main

import (
	"backend/db" // データベース接続を管理する自作パッケージ
	"backend/handlers"

	// 正しいパッケージパス
	"log"
	"net/http" // HTTPサーバーを作成・操作するライブラリ
	// gorilla/websocketを別名でインポート
)

// メイン関数：サーバーを起動する
func main() {
	if err := db.Connect(); err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	} else {
		log.Println("DB接続成功")
	}

	// WebSocketエンドポイント
	http.HandleFunc("/ws", handlers.HandleWebSocket)
	go handlers.BroadcastMessages()

	http.HandleFunc("/", handlers.Handler)

	http.HandleFunc("/signup", handlers.AddUserHandler)
	http.HandleFunc("/login", handlers.LoginHandler)

	http.HandleFunc("/roomSelect", handlers.GetUsersHandler)
	http.HandleFunc("/groupRoomSelect", handlers.GetGroupRoomsHandlers)
	http.HandleFunc("/PersonalRoomSelect", handlers.GetPersonalRoomsHandlers)

	http.HandleFunc("/createRooms", handlers.CreateGroupRoom)
	http.HandleFunc("/createGroup", handlers.CreateGroupRoom)
	http.HandleFunc("/getRooms", handlers.CreateGroupRoom)

	http.HandleFunc("/getRoomMessages", handlers.MessageHandler)
	http.HandleFunc("/message", handlers.MessageHandler)
	http.HandleFunc("/sendFile", handlers.UploadHandler)
	http.HandleFunc("/updataUnReadMessage", handlers.UpdataMessageHandler)

	http.HandleFunc("/read", handlers.MarkMessageAsRead)

	// メッセージ削除、編集
	http.HandleFunc("/deleteMessage", handlers.DeleteMessageHandler)
	http.HandleFunc("/EditMessage", handlers.DeleteMessageHandler)

	log.Println("サーバー起動中 http://localhost:8080")
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
