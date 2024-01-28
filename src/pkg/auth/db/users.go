package db

import (
	"gorm.io/gorm"
)

type UsersDB struct {
	db *gorm.DB
}

func (u *UsersDB) Create(user User) error {
	return u.db.Create(&user).Error
}

func (u *UsersDB) FindByUsername(username string) (User, error) {
	var user User
	return user, u.db.Find(&user, "username = ?", username).Error
}

func (u *UsersDB) Start() error {
	return u.db.AutoMigrate(User{})
}

func (u *UsersDB) Close() error {
	if db, _ := u.db.DB(); db != nil {
		return db.Close()
	}
	return nil
}

func NewUsersDB(db *gorm.DB) *UsersDB {
	return &UsersDB{
		db: db,
	}
}
