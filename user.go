package main

import "container/heap"

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

func (u *User) GetLatestPost(userId int) *Post {
	// modify to send 20 posts by user
	if h, ok := u.posts[userId]; ok && h.Len() > 0 {
		return heap.Pop(h).(*Post)
	}
	return nil
}

func (u *User) GetRecentPosts(userId int) []*Post {
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
