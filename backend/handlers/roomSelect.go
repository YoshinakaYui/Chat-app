package handlers

import (
	"backend/db"
	"backend/utils"
	"encoding/json"
	"log"
	"net/http"
)

type GetGroupRoomsRequest struct {
	LoggedInUserID int `json:"login_id"`
}

// 個別ルーム一覧を取得するハンドラー
func GetPersonalRoomsHandlers(w http.ResponseWriter, r *http.Request) {
	log.Println("GetGroupRoomsHandlers：スタート")
	utils.EnableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req GetGroupRoomsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("リクエストボディのデコードエラー:", err)
		http.Error(w, "リクエスト形式が不正です", http.StatusBadRequest)
		return
	}

	roomInfo, err := db.GetMyRooms(req.LoggedInUserID)

	if err != nil {
		log.Println("個別ルーム一覧取得エラー:", err)
		http.Error(w, "個別ルーム一覧の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// レスポンスを返す
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(roomInfo)
}

// グループルーム一覧を取得するハンドラー
func GetGroupRoomsHandlers(w http.ResponseWriter, r *http.Request) {
	log.Println("GetGroupRoomsHandlers：スタート")
	utils.EnableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req GetGroupRoomsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("リクエストボディのデコードエラー:", err)
		http.Error(w, "リクエスト形式が不正です", http.StatusBadRequest)
		return
	}

	rooms, err := db.GetMyGroupRooms(req.LoggedInUserID)
	if err != nil {
		log.Println("ルーム一覧取得エラー:", err)
		http.Error(w, "ルーム一覧の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// レスポンスを返す
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rooms)
}

// ユーザー一覧を取得するハンドラー
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("GetUsersHandler：スタート")
	utils.EnableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "メソッドが許可されていません", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req GetGroupRoomsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("リクエストボディのデコードエラー:", err)
		http.Error(w, "リクエスト形式が不正です", http.StatusBadRequest)
		return
	}

	users, err := db.GetOtherUsers(req.LoggedInUserID)
	if err != nil {
		log.Println("ユーザー一覧取得エラー:", err)
		http.Error(w, "ユーザー一覧の取得に失敗しました", http.StatusInternalServerError)
		return
	}
	// レスポンスを返す
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}
