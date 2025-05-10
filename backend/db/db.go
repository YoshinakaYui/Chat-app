// SQLとバックエンドを接続するコード(データベース接続を管理するGoコード)
package db

// ORMライブラリ：SQLを書かずにデータの構造体を変更
import (
	//"gorm.io/driver/postgres"
	//"gorm.io/gorm/logger"
	//"gorm.io/gorm"
	"os" // 環境変数を扱うための標準ライブラリ 
)

// var DB *gorm.DB

// データベースとmain.goをつなげる関数
func Connect() error {
	// 環境変数から接続情報を接続情報を取得
	// os.Getenvは関数であり、環境変数の値を取得するために使う // os.Genenvで"DB_DSN"の環境変数の値を取得して、dsnに代入する
	dsn := os.Getenv("DB_DSN")
	// error型の変数を宣言
	var err error

	// GORMを使ってPostgreSQLに接続
	//DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return err
}

