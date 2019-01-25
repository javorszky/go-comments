package database

import (
	"fmt"
	"github.com/javorszky/go-comments/config"
	"github.com/jinzhu/gorm"
	"time"
)

// GetDb is a function to return an instance of an opened database after making sure it can be connected to
func GetDb(config *config.Config) (db *gorm.DB, err error) {
	tries := 1
	for tries < 5 {

		tries++

		db, err = gorm.Open("mysql", fmt.Sprintf("%v:%v@%v/?charset=utf8mb4&parseTime=True&loc=Local", config.DatabaseUser, config.DatabasePassword, config.DatabaseAddress))
		if err == nil {
			return db, nil
		}

		fmt.Println(fmt.Sprintf("Database not open yet on %v, sleeping for 2 seconds.", fmt.Sprintf("%v:%v@%v/%v?charset=utf8mb4&parseTime=True&loc=Local", config.DatabaseUser, config.DatabasePassword, config.DatabaseAddress, config.DatabaseTable)))
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("could not connect to database in 10 seconds: %v", err)
}
