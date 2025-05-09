package database

import (
	"fmt"
	"log"

	"belajar-via-dev.to/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "123"
	dbname   = "test1"
)

var dsn string = fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta", host, port, user, password, dbname)

var DB *gorm.DB

func DBconn() {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	DB = db

	db.AutoMigrate(&models.Users{}, &models.Kategori_Soal{}, &models.Tingkatan{}, models.Kelas{}, models.Kuis{})
}
