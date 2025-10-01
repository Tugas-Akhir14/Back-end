package models

type Admin struct {

    ID              uint   `gorm:"primaryKey"`
    FullName        string `form:"full_name" json:"full_name" binding:"required"`
    Email           string `form:"email" json:"email" binding:"required,email"`
    PhoneNumber     string `form:"phone_number" json:"phone_number" binding:"required"`
    Password        string `form:"password" json:"password" binding:"required,min=8"`
    ConfirmPassword string `form:"confirm_password" json:"confirm_password" binding:"required"`
    
}
