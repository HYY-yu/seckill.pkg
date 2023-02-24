package db

import (
	"gorm.io/gorm"
)

type MockRepo struct {
	GDB *gorm.DB
}

func (m MockRepo) GetDb() *gorm.DB {
	return m.GDB
}

func (m MockRepo) DbClose() error {
	return nil
}
