package models

import (
	"time"
)

// front -> end へのfetchメッセージの構造体
type TsUser struct {
	ID       int    `json:"id"`
	Username string `json:"username" gorm:"unique"`
	Password string `json:"password"`
}

type TsMessage struct {
	ID           int       `gorm:"primaryKey"`
	RoomID       int       `json:"room_id"`
	SenderID     int       `json:"sender_id"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ThreadRootID int       `json:"thread_root_id"`
}

type TsResponse struct {
	Message string `json:"message"`
}

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
