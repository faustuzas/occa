package db

import (
	"context"

	"gorm.io/gorm"
)

type UsersDB struct {
	db *gorm.DB
}

func (u *UsersDB) Create(ctx context.Context, user User) error {
	return u.db.WithContext(ctx).Create(&user).Error
}

func (u *UsersDB) FindByUsername(ctx context.Context, username string) (User, error) {
	var user User
	return user, u.db.WithContext(ctx).Find(&user, "username = ?", username).Error
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
