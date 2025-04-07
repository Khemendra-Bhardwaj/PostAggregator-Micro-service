package main

import (
	"fmt"
	"time"
)

func main() {
	user := &User{userId: 1, userName: "Alice"}
	// Add some posts
	user.AddPost(&Post{postId: 1, userId: 1, timeStamp: time.Now().Add(-1 * time.Minute), content: "First post"})
	user.AddPost(&Post{postId: 2, userId: 1, timeStamp: time.Now().Add(-5 * time.Minute), content: "Second post"})
	user.AddPost(&Post{postId: 3, userId: 1, timeStamp: time.Now().Add(-10 * time.Minute), content: "Latest post"})

	// Get latest post
	// latest := user.GetLatestPost(1)
	// if latest != nil {
	// 	fmt.Printf("Latest post by user %d: \"%s\" at %s\n", latest.userId, latest.content, latest.timeStamp.Format(time.RFC822))
	// }

	recentPosts := user.GetRecentPosts(1)

	for _, post := range recentPosts {
		// fmt.Printf("%+v\n", post)
		fmt.Printf("• %s (at %s)\n", post.content, post.timeStamp.Format(time.RFC822))
	}

}
