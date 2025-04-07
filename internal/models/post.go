package models

import (
	"sync"
	"time"
)

type Post struct {
	PostId    int       // Changed from postId to PostId
	UserId    int       // Changed from userId to UserId
	TimeStamp time.Time // Changed from timeStamp to TimeStamp
	Content   string    // Already exported
	mu        sync.RWMutex
}
