package model

import "gorm.io/gorm"

type Identity struct {
	gorm.Model
	UUID     string `gorm:"uniqueIndex"`
	CPF      string `gorm:"index"`
	CNPJ     string `gorm:"index"`
	Name     string
	Verified bool
}

type CEP struct {
	gorm.Model
	CEP    string `gorm:"uniqueIndex"`
	Street string
	City   string
	State  string
}
