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

type RoomInfo struct {
	ID          int       `gorm:"column:id" json:"id"`
	RoomName    string    `gorm:"column:room_name" json:"room_name"`
	IsGroup     int       `gorm:"column:is_group" json:"is_group"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
	UnreadCount int       `gorm:"column:unread_count" json:"unread_count"`
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
	// ID        int       `gorm:"primaryKye"  json:"id"`
	MessageID int       `gorm:"not null;index" json:"message_id"`
	UserID    int       `json:"user_id"` // room_idだった
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

type Mentions struct {
	MessageID         int `json:"message_id"`
	MentionedTargetID int `json:"mentioned_target_id"`
}

type DeletedMessage struct {
	MessageID int       `json:"message_id"`
	UserID    int       `json:"user_id"`
	DeletedAt time.Time `json:"deleted_at"`
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

// 所属個別ルームと未読数を取得
func GetMyRooms(loginedUserID int) ([]RoomInfo, error) {
	log.Println("🟡GetOtherUsers：スタート")
	// var rooms []ChatRoom

	// // GORMクエリ
	// // room_nameには、相手の名前にして返す!
	// result := DB.Table("chat_rooms AS cr").
	// 	Select("cr.id AS id, u.username AS room_name, cr.is_group, cr.created_at, cr.updated_at").
	// 	Joins("JOIN room_members AS rm1 ON cr.id = rm1.room_id").
	// 	Joins("JOIN room_members AS rm2 ON cr.id = rm2.room_id AND rm2.user_id <> ?", loginedUserID).
	// 	Joins("JOIN users AS u ON rm2.user_id = u.id").
	// 	Where("cr.is_group = 0 AND rm1.user_id = ?", loginedUserID).
	// 	Group("cr.id, u.username, cr.is_group, cr.created_at, cr.updated_at").
	// 	Having("COUNT(DISTINCT rm2.user_id) = 1").
	// 	Order("cr.id ASC").
	// 	Scan(&rooms).Error

	var rooms []RoomInfo // 結果を格納する構造体

	subQuery := DB.Table("message_reads AS mr").
		Select("mr.message_id").
		Where("mr.user_id = ?", loginedUserID)

	result := DB.Table("chat_rooms AS cr").
		Select(`cr.id AS id, u.username AS room_name, cr.is_group, cr.created_at, cr.updated_at, COUNT(m.id) AS unread_count`).
		Joins("JOIN room_members AS rm1 ON cr.id = rm1.room_id").
		Joins("JOIN room_members AS rm2 ON cr.id = rm2.room_id AND rm2.user_id <> ?", loginedUserID).
		Joins("JOIN users AS u ON rm2.user_id = u.id").
		Joins("LEFT JOIN messages AS m ON m.room_id = cr.id AND m.id NOT IN (?)", subQuery).
		Where("cr.is_group = ? AND rm1.user_id = ?", 0, loginedUserID).
		Group("cr.id, u.username, cr.is_group, cr.created_at, cr.updated_at").
		Having("COUNT(DISTINCT rm2.user_id) = 1").
		Order("cr.id ASC").
		Scan(&rooms).Error

	if result != nil {
		fmt.Println("エラー:", result)
		return nil, fmt.Errorf("✖︎ルーム一覧取得エラー：%v", result)
	}
	return rooms, nil
}

// 所属グループルームを取得
func GetMyGroupRooms(userid int) ([]RoomInfo, error) {
	log.Println("GetMyGroupRooms")
	var rooms []RoomInfo

	// GORMクエリ
	result := DB.Table("chat_rooms cr").
		Select(`cr.*, COUNT(m.id) AS unread_count`).
		Joins("JOIN room_members rm ON cr.id = rm.room_id").
		Joins(`
        LEFT JOIN messages m ON m.room_id = cr.id
        AND m.sender_id <> ?
        AND m.id NOT IN (
            SELECT mr.message_id FROM message_reads mr WHERE mr.user_id = ?
        )`, userid, userid).
		Where("rm.user_id = ? AND cr.is_group = 1", userid).
		Group("cr.id").
		Order("cr.id ASC").
		Scan(&rooms).Error

	if result != nil {
		fmt.Println("エラー:", result)
		return nil, fmt.Errorf("ルーム一覧取得エラー：%v", result)
	}
	return rooms, nil
}
