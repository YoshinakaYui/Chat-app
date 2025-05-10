// SQLとバックエンドを接続するコード
package db

// ORMライブラリ：SQLを書かずにデータの構造体を変更
import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm/logger"
	"gorm.io/gorm"
	"os"
    "fmt"
    "log"
)

var DB *gorm.DB

// データベースとmain.goをつなげる関数
func Connect() error {
	// 環境変数から接続情報を接続情報を取得
	dsn := os.Getenv("DB_DSN")
	var err error

	// GORMを使ってPostgreSQLに接続
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return err
}


// 12344