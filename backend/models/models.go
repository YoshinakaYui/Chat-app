package models

import (
	"time"
)

// front -> end へのfetchメッセージの構造体
type TsUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	//PasswordHash string `json:"password_hash"`
}

type TsMessage struct {
	ID           int       `gorm:"primaryKey"`
	RoomID       int       `gorm:"not null;index"`
	SenderID     int       `gorm:"not null;index"`
	Content      string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
	ThreadRootID int       `gorm:"index"` // 親メッセージID（スレッド）
}

type TsResponse struct {
	Message string `json:"message"`
}

// type User struct {
// 	ID       int    `gorm:"primaryKey" json:"id"`
// 	Username string `json:"username"`
// }

// チャットルーム構造体
type TsChatRoom struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	RoomName  string    `json:"room_name"`
	IsGroup   int       `json:"is_group"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TsRoomMember struct {
	ID       int `gorm:"primaryKey"`
	RoomID   int `json:"room_name"`
	UserID   int `json:"user_id"`
	JoinedAt time.Time
}

// テーブル名指定
func (TsChatRoom) TableName() string {
	return "chat_rooms"
}
func (TsRoomMember) TableName() string {
	return "room_members"
}
