package main

import (
	"sync"
	"time"
)

type Post struct {
	postId    int
	userId    int // AUTHOR
	timeStamp time.Time
	content   string
	mu        sync.RWMutex
}
