package models

import "time"

type User struct {
	Base
	Type               UserType `gorm:"type:user_type;not null"`
	Email              string   `json:"email" gorm:"uniqueIndex;not null"`
	Password           string   `json:"password" gorm:"not null"`
	EmailVerified      bool     `json:"emailVerfied" gorm:"default:false;not null"`
	ServiceName        string   `json:"serviceName"`
	ServiceWebsite     string   `json:"serviceWebsite"`
	CanRequestWork     bool     `json:"canRequestWork" gorm:"default:false;not null"`
	InvalidResultCount int      `json:"invalidResultCount" gorm:"default:0;not null"`
	// The work this user provider
	WorkResults        []WorkRequest `gorm:"foreignKey:ProvidedBy"`
	LastProvidedWorkAt time.Time     `json:"lastProvidedWorkAt"`
	// The work this user has requested
	WorkRequests        []WorkRequest `gorm:"foreignKey:RequestedBy"`
	LastRequestedWorkAt time.Time     `json:"lastRequestedWorkAt"`
}
