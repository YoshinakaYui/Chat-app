package db

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// usersテーブルの構造体
type Users struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Username     string `json:"username" gorm:"unique"`
	PasswordHash string `json:"password_hash"`
}

type ChatRoom struct {
	ID        int       `gorm:"primaryKey;column:id" json:"id"`
	RoomName  string    `gorm:"column:room_name" json:"room_name"`
	IsGroup   int       `gorm:"column:is_group" json:"is_group"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}
type RoomMember struct {
	ID       int `gorm:"primaryKey"`
	RoomID   int `json:"room_id"` // チャットルームのID
	UserID   int `json:"user_id"` // 参加ユーザーのID
	JoinedAt time.Time
}

type Message struct {
	ID           int       `gorm:"primaryKey"`
	RoomID       int       `gorm:"not null;index"`
	SenderID     int       `gorm:"not null;index"`
	Content      string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
	ThreadRootID int       `gorm:"index"` // 親メッセージID（スレッド）
}

type MessageAttachment struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	MessageID int       `gorm:"not null;index" json:"message_id"`   // 関連メッセージID
	FileName  string    `gorm:"type:varchar(255)" json:"file_name"` // ファイル名
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`   // 作成日時
}

type MessageReads struct {
	//ID        int       `gorm:"primaryKye"  json:"id"`
	MessageID int       `gorm:"not null;index" json:"message_id"`
	UserID    int       `json:"room_id"`
	Reaction  string    `gorm:"type:varchar" json:"reaction"`
	ReadAt    time.Time `gorm:"autoCreateTime" json:"read_at"`
}

// 既読者カウントの構造体
type MessageReadCount struct {
	MessageID   int    `json:"message_id"`
	Content     string `json:"content"`
	SenderID    int    `json:"sender_id"`
	ReadCount   int    `json:"read_count"`
	UnreadCount int    `json:"unread_count"`
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
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ユーザーを保存
func SaveUser(username, password string) error {
	log.Println("db.パスワード：", password)
	// パスワードをハッシュ化
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("パスワードハッシュ化エラー: %v", err)
	}

	// ハッシュ化成功時にユーザーを保存（仮にDBに保存する処理とする）
	user := Users{Username: username, PasswordHash: hashedPassword}
	result := DB.Create(&user)
	return result.Error
}

// ハッシュ化したパスワードを生成
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("パスワードハッシュ化失敗: %v", err)
	}
	return string(hashed), nil
}

// 全ユーザーを取得する関数
func GetOtherUsers(loginedUserID int) ([]Users, error) {
	log.Println("🟡GetOtherUsers")
	var users []Users
	result := DB.Table("users").
		Select("id, username").
		Where("id != ?", loginedUserID).
		Order("ID ASC").
		Scan(&users).Error
	if result != nil {
		fmt.Println("エラー:", result)
		return nil, fmt.Errorf("ユーザー一覧取得エラー：%v", result)
	}
	return users, nil
}

// 所属個別ルームを取得
func GetMyRooms(loginedUserID int) ([]ChatRoom, error) {
	log.Println("🟡GetOtherUsers")
	var rooms []ChatRoom

	// GORMクエリ
	// room_nameには、相手の名前にして返す!
	result := DB.Table("chat_rooms AS cr").
		Select("cr.id AS id, u.username AS room_name, cr.is_group, cr.created_at, cr.updated_at").
		Joins("JOIN room_members AS rm1 ON cr.id = rm1.room_id").
		Joins("JOIN room_members AS rm2 ON cr.id = rm2.room_id AND rm2.user_id <> ?", loginedUserID).
		Joins("JOIN users AS u ON rm2.user_id = u.id").
		Where("cr.is_group = 0 AND rm1.user_id = ?", loginedUserID).
		Group("cr.id, u.username, cr.is_group, cr.created_at, cr.updated_at").
		Having("COUNT(DISTINCT rm2.user_id) = 1").
		Order("cr.id ASC").
		Scan(&rooms).Error
	log.Println("🍅：", rooms)

	if result != nil {
		fmt.Println("エラー:", result)
		return nil, fmt.Errorf("✖︎ルーム一覧取得エラー：%v", result)
	}
	return rooms, nil
}

// 所属グループルームを取得
func GetMyGroupRooms(userid int) ([]ChatRoom, error) {
	log.Println("GetMyGroupRooms")
	var rooms []ChatRoom

	// GORMクエリ
	result := DB.Table("chat_rooms cr").
		Select("cr.*").
		Joins("JOIN room_members rm ON cr.id = rm.room_id").
		Where("rm.user_id = ? and cr.is_group = 1", userid).
		Order("cr.id ASC").
		Scan(&rooms).Error

	if result != nil {
		fmt.Println("エラー:", result)
		return nil, fmt.Errorf("ルーム一覧取得エラー：%v", result)
	}
	return rooms, nil
}

// 入室したユーザーの人数取得
// var memberCount int

// var err = db.Raw(`
//   SELECT COUNT(*)
//   FROM room_members
//   WHERE room_id = ?
// `, roomID).Row().Scan(&memberCount)

// if err != nil {
//     log.Println("エラー:", err)
// } else {
//     fmt.Println("部屋の参加者数:", memberCount)
// }
