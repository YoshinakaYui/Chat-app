package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

type User struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Username     string `gorm:"unique" json:"username"`
	PasswordHash string `json:"-"`
}

// メッセージ情報の構造体
type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ハッシュ化したパスワードを生成
func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("パスワードハッシュ化失敗: %v", err)
	}
	return string(hashed), nil
}

// ユーザーを保存
func SaveUser(username, password string) error {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	user := User{Username: username, PasswordHash: hashedPassword}
	result := DB.Create(&user)
	return result.Error
}

// ログインチェック
func IsLogin(username, password string) bool {
	var user User
	result := DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// メッセージをデータベースに保存
func SaveMessage(sender, content string) error {
	message := Message{
		Sender:    sender,
		Content:   content,
		CreatedAt: time.Now(),
	}
	result := DB.Create(&message)
	if result.Error != nil {
		return fmt.Errorf("メッセージ保存エラー: %v", result.Error)
	}
	return nil
}

// 全ユーザーを取得
func GetAllUsers() ([]User, error) {
	var users []User
	result := DB.Select("id", "username").Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("ユーザー取得失敗: %v", result.Error)
	}
	return users, nil
}

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
