package db

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// usersテーブルの構造体
type Users struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

var DB *gorm.DB

// データベース接続
func Connect() error {
	dsn := os.Getenv("DB_DSN")
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("データベース接続エラー: %v", err)
	}
	log.Println("データベース接続成功")
	return nil
}

// ハッシュ化パスワードと入力パスワードを比較する関数
func CheckPasswordHash(password, hash string) bool {
	// bcrypt.CompareHashAndPasswordで比較
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ユーザーを保存
func SaveUser(username, password string) error {
	log.Println("11111", password)
	// パスワードをハッシュ化
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("パスワードハッシュ化エラー: %v", err)
	}

	log.Println("22222")
	// ハッシュ化成功時にユーザーを保存（仮にDBに保存する処理とする）
	user := Users{Username: username, PasswordHash: hashedPassword}
	result := DB.Create(&user)
	return result.Error

	// ここでは、ハッシュ化されたパスワードを利用してDBに保存
	// fmt.Println("ユーザー保存成功:", username, hashedPassword)
	// log.Println("33333")
	// // 処理が成功した場合、nilを返す
	// return nil
}

// ハッシュ化したパスワードを生成
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("パスワードハッシュ化失敗: %v", err)
	}
	return string(hashed), nil
}

// 	log.Println("islogin-3")
// 	log.Println(user)
// 	return true
// }

// メッセージをデータベースに保存
// func SaveMessage(sender, content string) error {
// 	message := Message{
// 		Sender:    sender,
// 		Content:   content,
// 		CreatedAt: time.Now(),
// 	}
// 	result := DB.Create(&message)
// 	if result.Error != nil {
// 		return fmt.Errorf("メッセージ保存エラー: %v", result.Error)
// 	}
// 	return nil
// }

// 全ユーザーを取得
// func GetAllUsers() ([]User, error) {
// 	var users []User
// 	result := DB.Select("id", "username").Find(&users)
// 	if result.Error != nil {
// 		return nil, fmt.Errorf("ユーザー取得失敗: %v", result.Error)
// 	}
// 	return users, nil
// }

/*
// すべてのメッセージを取得
func GetAllMessages() ([]Message, error) {
	var messages []Message
	result := DB.Order("created_at asc").Find(&messages)
	if result.Error != nil {
		log.Println("メッセージ取得エラー:", result.Error)
		return nil, fmt.Errorf("メッセージ取得エラー: %v", result.Error)
	}
	log.Println("メッセージ一覧取得成功:", messages)
	return messages, nil
}
*/
