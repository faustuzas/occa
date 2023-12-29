package db

import (
	"gorm.io/gorm"
)

type UsersDB struct {
	db *gorm.DB
}

func (u *UsersDB) Get() (User, error) {
	//TODO implement me
	panic("implement me")
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
