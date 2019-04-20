package database

import (
	"fmt"
	"github.com/javorszky/go-comments/config"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
	"time"
)

// GetInstance is a function to return an instance of an opened database after making sure it can be connected to
func GetInstance(config *config.Config) (db *gorm.DB, err error) {
	tries := 1
	for tries < 5 {

		tries++

		db, err = gorm.Open("mysql", fmt.Sprintf("%v:%v@%v/?charset=utf8mb4&parseTime=True&loc=Local", config.DatabaseRootUser, config.DatabaseRootPassword, config.DatabaseAddress))
		if err == nil {
			fmt.Println(fmt.Sprintf("Connected to %s, creating database %s.", fmt.Sprintf("%v:%v@%v/?charset=utf8mb4&parseTime=True&loc=Local", config.DatabaseUser, config.DatabasePassword, config.DatabaseAddress), config.DatabaseTable))

			db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", config.DatabaseTable))
			db.Exec(fmt.Sprintf("USE `%s`", config.DatabaseTable))
			db.Close()
			db, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&parseTime=True&loc=Local", config.DatabaseUser, config.DatabasePassword, config.DatabaseAddress, config.DatabaseTable))

			return db, nil
		}

		fmt.Println(fmt.Sprintf("Database not open yet on %v, sleeping for 2 seconds.", fmt.Sprintf("%v:%v@%v/?charset=utf8mb4&parseTime=True&loc=Local", config.DatabaseUser, config.DatabasePassword, config.DatabaseAddress)))
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("could not connect to database in 10 seconds: %v", err)
}

func RunMigrations(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		// create persons table
		{
			ID: "201903032031",
			Migrate: func(tx *gorm.DB) error {
				// it's a good practice to copy the struct inside the function,
				// so side effects are prevented if the original struct changes during the time
				type User struct {
					gorm.Model
					Email          string `json:"email" form:"email" gorm:"type:varchar(191);unique_index:email"`
					PasswordOne    string `form:"password1" gorm:"-" json:"-"`
					PasswordTwo    string `form:"password2" gorm:"-" json:"-"`
					HashedPassword string `json:"passwordHash" gorm:"type:varchar(255)"`
				}
				return tx.AutoMigrate(&User{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("users").Error
			},
		},
		{
			ID: "201903231810",
			Migrate: func(tx *gorm.DB) error {
				type Session struct {
					ID        string `gorm:"type:varchar(36);primary_key"`
					UserID    uint
					CreatedAt time.Time `gorm:"index:created_at"`
					IP        string
					UserAgent string
				}

				type User struct {
					gorm.Model
					Email          string    `json:"email" form:"email" gorm:"type:varchar(191);unique_index:email"`
					PasswordOne    string    `form:"password1" gorm:"-" json:"-"`
					PasswordTwo    string    `form:"password2" gorm:"-" json:"-"`
					HashedPassword string    `json:"passwordHash" gorm:"type:varchar(255)"`
					Sessions       []Session `gorm:"auto_preload"`
				}
				tx.AutoMigrate(&User{}, &Session{})

				return tx.Model(&Session{}).AddForeignKey("user_id", "users(id)", "CASCADE", "RESTRICT").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("sessions").Error
			},
		},
		{
			ID: "201903251327",
			Migrate: func(tx *gorm.DB) error {
				type Session struct {
					ID        string `gorm:"type:varchar(36);primary_key"`
					UserID    uint
					CreatedAt time.Time `gorm:"index:created_at"`
					IP        string
					UserAgent string
					Hash      string
				}

				return tx.AutoMigrate(&Session{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				type Session struct {
					ID        string `gorm:"type:varchar(36);primary_key"`
					UserID    uint
					CreatedAt time.Time `gorm:"index:created_at"`
					IP        string
					UserAgent string
					Hash      string
				}
				return tx.Model(&Session{}).DropColumn("hash").Error
			},
		},
		{
			ID: "201904201411",
			Migrate: func(tx *gorm.DB) error {
				type Site struct {
					gorm.Model
					UserID      uint
					Designation string `gorm:"type:varchar(191);not null;unique"`
					Domain      string `gorm:"type:varchar(191)"`
				}

				type Session struct {
					ID        string `gorm:"type:varchar(36);primary_key"`
					UserID    uint
					CreatedAt time.Time `gorm:"index:created_at"`
					IP        string
					UserAgent string
					Hash      string
				}

				type User struct {
					gorm.Model
					Email          string    `json:"email" form:"email" gorm:"type:varchar(191);unique_index:email"`
					PasswordOne    string    `form:"password1" gorm:"-" json:"-"`
					PasswordTwo    string    `form:"password2" gorm:"-" json:"-"`
					HashedPassword string    `json:"passwordHash" gorm:"type:varchar(255)"`
					Sessions       []Session `gorm:"auto_preload"`
					Sites          []Site    `gorm:"auto_preload"`
				}

				tx.AutoMigrate(&Site{})

				return tx.Model(&Site{}).AddForeignKey("user_id", "users(id)", "CASCADE", "RESTRICT").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("sites").Error
			},
		},
	})

	return m.Migrate()
}
