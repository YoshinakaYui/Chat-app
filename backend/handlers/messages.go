package handlers

import (
	"backend/db"
	"backend/utils"

	"encoding/json"
	//"log"
	"net/http"
)

// メッセージ送受信ハンドラー
func MessageHandler(w http.ResponseWriter, r *http.Request) {
	// CORS対応
	utils.EnableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// メッセージ送信
	if r.Method == http.MethodPost {
		var msg db.Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, "無効なリクエスト", http.StatusBadRequest)
			return
		}

		if err := db.SaveMessage(msg.Sender, msg.Content, msg.MessagesID); err != nil {
			http.Error(w, "メッセージ保存エラー", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("メッセージ送信成功"))
		return
	}

	// メッセージ取得
	if r.Method == http.MethodGet {
		recipientID := r.URL.Query().Get("user")
		messages, err := db.GetMessagesByRecipient(recipientID)
		if err != nil {
			http.Error(w, "メッセージ取得エラー", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
		return
	}

	//var users db.Users
	db.GetAllUsers()

	// メソッドが許可されていない場合
	http.Error(w, "許可されていないメソッド", http.StatusMethodNotAllowed)

}

// package handlers

// import (
// 	"backend/utils"
// 	"net/http" // HTTPサーバーを作成・操作するライブラリ
// )

// // メッセージハンドラー
// // func MessageHandler(w http.ResponseWriter, r *http.Request) {
// // 	utils.EnableCORS(w)

// 	// // ヘッダーを正しく設定
// 	// w.Header().Set("Content-Type", "application/json")

// 	// // JSON形式でレスポンスを作成
// 	// response := map[string]string{
// 	// 	"message": "Hello, world!",
// 	// }

// 	// // JSONとして出力
// 	// json.NewEncoder(w).Encode(response)

// 	// log.Println("messages-jjj")

// 	// cookie, err := r.Cookie("auth_token")
// 	// if err != nil {
// 	// 	log.Println("クッキーが存在しません:", err)
// 	// 	http.Error(w, "未ログイン", http.StatusUnauthorized)
// 	// 	return
// 	// }

// 	// var user db.Users
// 	// user.Username, _, err = auth.ValidateJWT(cookie.Value)
// 	// log.Println("クッキー：" + cookie.Value)
// 	// if err != nil {
// 	// 	log.Println("JWT検証エラー:", err)
// 	// 	http.Error(w, "トークンが無効です", http.StatusUnauthorized)
// 	// 	return
// 	// }

// 	// // クッキーが取得できた場合
// 	// log.Println("クッキー取得成功:", cookie.Value)
// 	// w.Write([]byte("クッキー取得成功: " + cookie.Value))
// //}
