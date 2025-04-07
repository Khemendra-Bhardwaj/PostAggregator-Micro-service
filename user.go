package main

import (
	"container/heap"
	"sync"
)

type User struct {
	userId    int
	userName  string
	posts     map[int]*PostHeap
	Following []*User
	Followers []*User
}

func (u *User) AddPost(post *Post) {
	if u.posts == nil {
		u.posts = make(map[int]*PostHeap)
	}

	if _, ok := u.posts[post.userId]; !ok {
		h := &PostHeap{}
		heap.Init(h)
		u.posts[post.userId] = h
	}
	heap.Push(u.posts[post.userId], post)
}

// latest post for a specific user with userID
func (u *User) GetLatestPostByUser(userId int) *Post {
	if h, ok := u.posts[userId]; ok && h.Len() > 0 {
		return heap.Pop(h).(*Post)
	}
	return nil
}

// 20 recent posts for a specific user with userID
func (u *User) GetRecentPostsByUser(userId int) []*Post {
	count := 20
	if h, ok := u.posts[userId]; ok && h.Len() > 0 {
		var recent []*Post
		var temp []*Post

		// Pop up to `count` posts
		for i := 0; i < count && h.Len() > 0; i++ {
			post := heap.Pop(h).(*Post)
			recent = append(recent, post)
			temp = append(temp, post)
		}

		// Push the posts back to restore the heap
		for _, post := range temp {
			heap.Push(h, post)
		}

		return recent
	}
	return nil
}

/*
// 20 recent posts for a specific user with userID
func (u *User) GetUserFeed() []*Post {
	var wg sync.WaitGroup
	var mu sync.Mutex
	allPosts := []*Post{}

	// Start a goroutine for each followed user
	for _, followed := range u.Following {
		wg.Add(1)

		go func(f *User) {
			defer wg.Done()
			posts := f.GetRecentPostsByUser(f.userId)

			mu.Lock()
			allPosts = append(allPosts, posts...)
			mu.Unlock()
		}(followed)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Sort all collected posts by timestamp (descending)
	sort.Slice(allPosts, func(i, j int) bool {
		return allPosts[i].timeStamp.After(allPosts[j].timeStamp)
	})

	// Return top 20
	if len(allPosts) > 20 {
		return allPosts[:20]
	}
	return allPosts
}
*/

// Optmized approach using Heap to Generate User Feed from its followers
func (u *User) GetUserFeed() []*Post {
	var wg sync.WaitGroup
	var mu sync.Mutex
	feedHeap := &PostHeap{}
	heap.Init(feedHeap)

	// Fetch recent posts from all followings concurrently
	for _, followed := range u.Following {
		wg.Add(1)
		go func(f *User) {
			defer wg.Done()
			posts := f.GetRecentPostsByUser(f.userId)

			mu.Lock()
			for _, post := range posts {
				if feedHeap.Len() < 20 {
					heap.Push(feedHeap, post)
				} else if post.timeStamp.After((*feedHeap)[0].timeStamp) {
					heap.Pop(feedHeap) // remove the oldest
					heap.Push(feedHeap, post)
				}
			}
			mu.Unlock()
		}(followed)
	}

	wg.Wait()

	// Extract posts from heap and reverse to get newest-first order
	var result []*Post
	for feedHeap.Len() > 0 {
		result = append(result, heap.Pop(feedHeap).(*Post))
	}
	// Reverse result to get most recent first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}
