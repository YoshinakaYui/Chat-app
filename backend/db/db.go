// SQLとバックエンドを接続するコード(データベース接続を管理するGoコード)
package db

// ORMライブラリ：SQLを書かずにデータの構造体を変更
import (
	"gorm.io/driver/postgres"
	//"gorm.io/gorm/logger"
	"gorm.io/gorm"
	"os" // 環境変数を扱うための標準ライブラリ 
	"log"
	"fmt"
	"time"
	"golang.org/x/crypto/bcrypt"
)

var DB *gorm.DBz

// データベースとmain.goをつなげる関数
func Connect() error {
	// 環境変数から接続情報を接続情報を取得
	// os.Getenvは関数であり、環境変数の値を取得するために使う // os.Genenvで"DB_DSN"の環境変数の値を取得して、dsnに代入する
	dsn := os.Getenv("DB_DSN")
	// error型の変数を宣言
	var err error

	// GORMを使ってPostgreSQLに接続
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return err

    // SQL DBオブジェクトを取得
    sqlDB, err := DB.DB()
    if err != nil {
        return fmt.Errorf("GORMからSQL DBオブジェクト取得失敗: %v", err)
    }

    // Pingを実施して接続確認
    if err := sqlDB.Ping(); err != nil {
        return fmt.Errorf("データベースPing失敗: %v", err)
    }

	log.Println("DB接続成功！sssss")
	return nil
}

// パスワードをハッシュ化する関数
func hashPassword(password string) (string, error) {
    hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", fmt.Errorf("パスワードのハッシュ化に失敗: %w", err)
    }
    return string(hashed), nil
}

// ユーザーをデータベースに保存する関数
func SaveUser(username, password string) error {
    // パスワードをハッシュ化
    hashedPassword, err := hashPassword(password)
    if err != nil {
        log.Println("パスワードハッシュエラー:", err)
        return err
    }

    jst, err := time.LoadLocation("Asia/Tokyo")
    if err != nil {
        panic("タイムゾーンの取得に失敗しました")
    }

    // 現在日時を日本時間で取得
    now := time.Now().In(jst)
	log.Println(now)

    query := "INSERT INTO users (username, password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4)"
    result := DB.Exec(query, username, hashedPassword, now, now)
    if result.Error != nil {
        log.Println("データ保存エラー:", err)
        return fmt.Errorf("データ保存エラー")
    }

    log.Println("ユーザー情報が保存されました")
    return nil
}

