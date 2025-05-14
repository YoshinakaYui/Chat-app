package models

import (
	"time"
)

//"backend/auth"
//"log"

// front -> end へのfetchメッセージの構造体
type TsUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	//PasswordHash string `json:"password_hash"`
}

type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Response struct {
	Message string `json:"message"`
}

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `json:"username"`
}
