package hotel

import (
	"time"

	"gorm.io/gorm"
)

type Gallery struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	RoomID   *uint  `gorm:"index;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"room_id,omitempty"` 
	Room     *Room  `json:"room,omitempty"`

	Title    string `gorm:"size:150" json:"title"`
	Caption  string `gorm:"type:text" json:"caption"`
	URL      string `gorm:"size:255;not null" json:"url"`     
	MimeType string `gorm:"size:64" json:"mime_type"`
	Size     int64  `json:"size"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}


type UpdateGalleryRequest struct {
	Title   *string `json:"title" binding:"omitempty"`
	Caption *string `json:"caption" binding:"omitempty"`
	RoomID  *uint   `json:"room_id" binding:"omitempty"`
}

