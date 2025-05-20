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

	log.Println("サーバー起動中 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
