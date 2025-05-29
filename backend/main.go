package main

import (
	"backend/db" // データベース接続を管理する自作パッケージ
	"backend/handlers"

	"log"
	"net/http" // HTTPサーバーを作成・操作するライブラリ
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

	// サインアップ、ログイン
	http.HandleFunc("/signup", handlers.AddUserHandler)
	http.HandleFunc("/login", handlers.LoginHandler)

	// 所属ルームの取得（roomSelect.go）
	http.HandleFunc("/PersonalRoomSelect", handlers.GetPersonalRoomsHandlers)
	http.HandleFunc("/groupRoomSelect", handlers.GetGroupRoomsHandlers)
	http.HandleFunc("/roomSelect", handlers.GetUsersHandler)

	// ルームの作成、取得（chatRoom.go）
	http.HandleFunc("/createRooms", handlers.CreateChatRoom)
	http.HandleFunc("/createGroup", handlers.CreateChatRoom)
	http.HandleFunc("/getRooms", handlers.CreateChatRoom)

	// メッセージ履歴の取得（messages.go）
	http.HandleFunc("/getRoomMessages", handlers.GetMessagesHandler)
	http.HandleFunc("/updateUnReadMessage", handlers.UpdateMessageHandler)

	// 既読（read.go）
	http.HandleFunc("/read", handlers.MarkMessageAsRead)

	// メッセージ送信（messages.go）
	http.HandleFunc("/message", handlers.SendMessageHandler)
	http.HandleFunc("/addMention", handlers.MentionHandler)
	http.HandleFunc("/getRoomMembers", handlers.GetRoomMembersHandler)
	http.HandleFunc("/sendFile", handlers.UploadHandler)

	// メッセージ編集、削除、送信取消、リアクション
	http.HandleFunc("/editMessage", handlers.EditMessageHandler)
	http.HandleFunc("/deleteMyMessage", handlers.DeleteMyMessageHandler)
	http.HandleFunc("/deleteMessage", handlers.DeleteMessageHandler)
	http.HandleFunc("/addReaction", handlers.ReactionHandler)

	// 退出、メンバー追加（chatRoom.go）
	http.HandleFunc("/leaveRoom", handlers.LeaveRoomHandler)
	http.HandleFunc(("/usersNotInRoom"), handlers.UsersNotInRoomHandler)
	http.HandleFunc("/addMember", handlers.AddMemberHandler)

	log.Println("サーバー起動中 http://localhost:8080")
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
