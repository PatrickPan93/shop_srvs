package main

import (
	"crypto/sha512"
	"fmt"
	"github.com/anaskhan96/go-password-encoder"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"shop_srvs/user_srv/model"
	"time"
)

var (
	db *gorm.DB
)

func init() {
	var (
		err error
	)

	dsn := "root:root@tcp(127.0.0.1:3306)/shop_user_srv?charset=utf8mb4&parseTime=True&loc=Local"

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			Colorful:      true,
			LogLevel:      logger.Info,
		})

	if db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
		// gorm 定义schema命名规则
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: true,
			NameReplacer:  nil,
			NoLowerCase:   false,
		},
	}); err != nil {
		log.Fatalln(err)
	}
}

func doMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&model.User{})
}

func mockUsers(db *gorm.DB) {
	options := &password.Options{SaltLen: 10, Iterations: 100, KeyLen: 32, HashFunction: sha512.New}

	salt, encodedPwd := password.Encode("HardCodePassWord", options)
	newPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)

	for i := 0; i < 10; i++ {

		user := model.User{
			NickName: fmt.Sprintf("bobby%d", i),
			Mobile:   fmt.Sprintf("1599996251%d", i),
			Password: newPassword,
		}
		db.Save(&user)
	}
}

func main() {
	_ = doMigrate(db)
	mockUsers(db)
}
