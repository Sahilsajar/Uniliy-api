package models

type College struct {
	Base
	CollegeName string `gorm:"size:255;not null"`
	CollegeEmail string `gorm:"size:255"`
	State       string `gorm:"size:100"`
	City        string `gorm:"size:100"`
	CollegeType string `gorm:"size:100"`
	Website     string `gorm:"size:255"`

	Users []User `gorm:"foreignKey:CollegeID"`
}