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
	ID           uint   `gorm:"primaryKey" json:"id"`
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
	log.Println("db-11111", password)
	// パスワードをハッシュ化
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("パスワードハッシュ化エラー: %v", err)
	}

	log.Println("db-22222")
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

// メッセージ画面の色々な取得
type Message struct {
	ID         uint `gorm:"primaryKey"`
	Sender     string
	Content    string
	MessagesID uint
}

func SaveMessage(sender, content string, messagesID uint) error {
	message := Message{Sender: sender, Content: content, MessagesID: messagesID}
	return DB.Create(&message).Error
}

func GetMessagesByRecipient(messagesID string) ([]Message, error) {
	var messages []Message
	result := DB.Where("messages = ?", messagesID).Find(&messages)
	return messages, result.Error
}

// 全ユーザーを取得する関数
func GetAllUsers() ([]Users, error) {
	var users []Users
	result := DB.Select("id", "username").Find(&users)
	if result.Error != nil {
		log.Println("ユーザー一覧取得エラー:", result.Error)
		return nil, fmt.Errorf("ユーザー一覧取得エラー: %v", result.Error)
	}
	return users, nil
}
