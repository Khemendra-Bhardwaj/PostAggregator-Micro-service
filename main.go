package main

import (
	"fmt"
	"time"
)

func main() {
	// Create users
	alice := &User{userId: 1, userName: "Alice"}
	bob := &User{userId: 2, userName: "Bob"}
	charlie := &User{userId: 3, userName: "Charlie"}

	// Establish following
	alice.Following = []*User{bob, charlie}

	// Add posts to Bob and Charlie
	now := time.Now()
	bob.AddPost(&Post{postId: 1, userId: 2, timeStamp: now.Add(-2 * time.Minute), content: "Bob's post"})
	charlie.AddPost(&Post{postId: 2, userId: 3, timeStamp: now.Add(-1 * time.Minute), content: "Charlie’s post"})

	// Get feed for Alice
	feed := alice.GetUserFeed()
	for _, post := range feed {
		fmt.Printf("Post from user %d: %s at %s\n", post.userId, post.content, post.timeStamp.Format(time.RFC822))
	}

}
