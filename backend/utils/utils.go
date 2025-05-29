package utils

import (
	"io"
	"log"
	"net/http" // HTTPサーバーを作成・操作するライブラリ
)

// CORS対応を設定する関数
func EnableCORS(w http.ResponseWriter) {
	// w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3001")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func JsonRawDataDisplay(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}

	// 生のJSONデータを表示
	log.Println("🔍 生のJSONデータ:", string(body))

}
