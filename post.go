package main

import "time"

type Post struct {
	postId    int
	userId    int // AUTHOR
	timeStamp time.Time
	content   string
}
