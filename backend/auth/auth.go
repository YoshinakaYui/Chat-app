package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 環境変数からJWT秘密鍵を取得
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// トークン生成関数
func GenerateJWT(username string, passwordHash string) (string, error) {
	claims := jwt.MapClaims{
		"username":     username,
		"passwordHash": passwordHash,
		"exp":          time.Now().Add(time.Hour * 24).Unix(), // 24時間有効
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// トークン検証関数　リクエストが来たときに検証
func ValidateJWT(tokenString string) (string, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})

	if err != nil {
		return "", "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, _ := claims["username"].(string)
		passwordHash, _ := claims["passwordHash"].(string)
		return username, passwordHash, nil
	}

	return "", "", jwt.ErrSignatureInvalid
}

// ログイン中のユーザー名を取得するハンドラー
// func GetLoggedInUserHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	// リクエストヘッダーからトークンを取得
// 	token := r.Header.Get("Authorization")
// 	if token == "" {
// 		http.Error(w, "トークンがありません", http.StatusUnauthorized)
// 		return
// 	}

// 	// "Bearer " を削除
// 	if len(token) > 7 && token[:7] == "Bearer " {
// 		token = token[7:]
// 	}

// 	// 正しく3つの値で受け取る
// 	username, _, err := ValidateJWT(token)
// 	if err != nil {
// 		log.Println("JWT検証エラー:", err)
// 		http.Error(w, "無効なトークン", http.StatusUnauthorized)
// 		return
// 	}

// 	// ユーザー名をレスポンスとして返す
// 	json.NewEncoder(w).Encode(map[string]string{
// 		"username": username,
// 	})
// }
