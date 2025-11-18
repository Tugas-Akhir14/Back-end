// internal/models/hotel/room.go
package hotel

import (
	"mime/multipart"
	"time"
	"gorm.io/gorm"
)

type RoomStatus string

const (
	RoomStatusAvailable RoomStatus = "available"
	RoomStatusBooked    RoomStatus = "booked"
	RoomStatusCleaning  RoomStatus = "cleaning"
)

func (s RoomStatus) String() string {
	return string(s)
}

type Room struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Number      string         `gorm:"size:10;unique;not null" json:"number"`
	RoomTypeID  uint           `gorm:"not null" json:"room_type_id"`
	RoomType    RoomType       `gorm:"foreignKey:RoomTypeID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"room_type"`
	Capacity    int            `gorm:"not null" json:"capacity"`
	Description string         `gorm:"type:text" json:"description,omitempty"` 
	Image       string         `json:"image,omitempty"`
	Status      RoomStatus     `gorm:"type:varchar(20);default:'available'" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}


type CreateRoomRequest struct {
	Number      string                `json:"number" binding:"required" form:"number"`
	RoomTypeID  uint                  `json:"room_type_id" binding:"required" form:"room_type_id"`
	Capacity    int                   `json:"capacity" binding:"required,gt=0" form:"capacity"`
	Description string                `json:"description,omitempty" form:"description"` 
	Image       *multipart.FileHeader `form:"image" json:"-"` 
	Status      string                `json:"status,omitempty" form:"status"`
}


type UpdateRoomRequest struct {
	Number      *string               `json:"number,omitempty" form:"number"`
	RoomTypeID  *uint                 `json:"room_type_id,omitempty" form:"room_type_id"`
	Capacity    *int                  `json:"capacity,omitempty,gte=1" form:"capacity"`
	Description *string               `json:"description,omitempty" form:"description"` 
	Image       *multipart.FileHeader `form:"image" json:"-"` 
	Status      *string               `json:"status,omitempty" form:"status"`
}