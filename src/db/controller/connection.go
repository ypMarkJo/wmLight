package controller

import (
	"database/sql"
	"fmt"
	"github.com/ypMarkJo/wmLight/src/db/model"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

const (
	dbUser     = "root"
	dbPassword = "123456"
	dbAddress  = "localhost"
	dbPort     = "3306"
	dbName     = "wmdb"
)

// mysql 연결 초기화
func InitDB() (*model.DBConnection, error) {
	// DB 연결 설정
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbAddress, dbPort, dbName)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Connected to the database- %s:%s/%s\n", dbAddress, dbPort, dbName)
	return &model.DBConnection{DB: db}, nil
}
