package handlers

import (
	"backend/utils"
	"log"
	"net/http" // HTTPサーバーを作成・操作するライブラリ
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// CORS対応を有効にする
	utils.EnableCORS(w)

	log.Println("handler-aaa")
	//w.Write([]byte("Hello!!"))
}
