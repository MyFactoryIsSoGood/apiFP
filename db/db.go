package db

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"os"
)

var DB *gorm.DB

func ConnectDB() {
	fmt.Println(os.Getenv("DB_USERNAME"))
	fmt.Println(os.Getenv("DB_NAME"))
	fmt.Println(os.Getenv("DB_PASSWORD"))
	args := fmt.Sprintf("user=%v dbname=%v password=%v sslmode=disable host=db port=5432",
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"))
	//args := "user=postgres dbname=Fingerprints password=123qwe123 sslmode=disable host=localhost port=5432"
	db, err := gorm.Open("postgres", args)

	if err != nil {
		panic(err.Error())
	}

	db.AutoMigrate(&ISOTemplatesSTR{})
	DB = db
}
