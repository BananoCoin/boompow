package models

type User struct {
	Base
	Email         string `json:"email" gorm:"unique;not null"`
	Password      string `json:"password" gorm:"not null"`
	EmailVerified bool   `json:"emailVerfied" gorm:"default:false;not null"`
}
