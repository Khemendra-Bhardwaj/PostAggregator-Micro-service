package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"postaggregator/internal/models"
	postpb "postaggregator/proto"

	"google.golang.org/grpc"
)

type postServer struct {
	postpb.UnimplementedPostServiceServer
	users map[int32]*models.User
}

func (s *postServer) ListFollowing(ctx context.Context, req *postpb.ListFollowingRequest) (*postpb.ListFollowingResponse, error) {
	user, exists := s.users[req.UserId]
	if !exists {
		return &postpb.ListFollowingResponse{}, nil
	}

	var followingIds []int32
	for _, u := range user.Following {
		followingIds = append(followingIds, int32(u.UserId))
	}

	return &postpb.ListFollowingResponse{FollowingIds: followingIds}, nil
}

func (s *postServer) ListPostsByUser(ctx context.Context, req *postpb.ListPostsRequest) (*postpb.ListPostsResponse, error) {
	user, exists := s.users[req.UserId]
	if !exists {
		return &postpb.ListPostsResponse{Posts: []*postpb.Post{}}, nil
	}

	posts := user.GetRecentPostsByUser(int(req.UserId))
	var pbPosts []*postpb.Post
	for _, p := range posts {
		pbPosts = append(pbPosts, &postpb.Post{
			PostId:    int32(p.PostId),
			UserId:    int32(p.UserId),
			Content:   p.Content,
			Timestamp: p.TimeStamp.Format(time.RFC3339),
		})
	}

	return &postpb.ListPostsResponse{Posts: pbPosts}, nil
}

func (s *postServer) GetUserFeed(ctx context.Context, req *postpb.ListPostsRequest) (*postpb.ListPostsResponse, error) {
	user, exists := s.users[req.UserId]
	if !exists {
		return &postpb.ListPostsResponse{Posts: []*postpb.Post{}}, nil
	}

	posts := user.GetUserFeed(int(req.UserId))

	var pbPosts []*postpb.Post
	for _, p := range posts {
		pbPosts = append(pbPosts, &postpb.Post{
			PostId:    int32(p.PostId),
			UserId:    int32(p.UserId),
			Content:   p.Content,
			Timestamp: p.TimeStamp.Format(time.RFC3339),
		})
	}

	return &postpb.ListPostsResponse{Posts: pbPosts}, nil
}

// Struct for reading from users.json
// JSON structures
type userData struct {
	Users []userJSON `json:"users"`
}

type userJSON struct {
	UserId    int32      `json:"user_id"`
	UserName  string     `json:"user_name"`
	Following []int32    `json:"following"`
	Posts     []postJSON `json:"posts"`
}

type postJSON struct {
	PostId    int32  `json:"post_id"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

func loadUsersFromFile(filepath string) (map[int32]*models.User, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var userData userData
	if err := json.Unmarshal(data, &userData); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	users := make(map[int32]*models.User)

	// First pass: create all users
	for _, u := range userData.Users {
		users[u.UserId] = &models.User{
			UserId:   int(u.UserId),
			UserName: u.UserName,
		}
	}

	// Second pass: establish relationships and add posts
	for _, u := range userData.Users {
		user := users[u.UserId]

		// Add following relationships
		for _, followingID := range u.Following {
			if followedUser, exists := users[followingID]; exists {
				user.Following = append(user.Following, followedUser)
			}
		}

		// Add posts
		for _, p := range u.Posts {
			timestamp, err := time.Parse(time.RFC3339, p.Timestamp)
			if err != nil {
				log.Printf("Error parsing timestamp for post %d: %v", p.PostId, err)
				continue
			}

			post := &models.Post{
				PostId:    int(p.PostId),
				UserId:    int(u.UserId),
				Content:   p.Content,
				TimeStamp: timestamp,
			}
			user.AddPost(post)
		}
	}

	return users, nil
}

func main() {
	users := make(map[int32]*models.User)
	users, err := loadUsersFromFile("/home/khemendra/Desktop/PostAggregator/data.json")
	if err != nil {
		log.Fatalf("Failed to load users from file: %v", err)
	}
	log.Printf("Successfully loaded %d users", len(users))
	/*
		alice := &models.User{UserId: 1, UserName: "Alice"}
		bob := &models.User{UserId: 2, UserName: "Bob"}
		charlie := &models.User{UserId: 3, UserName: "Charlie"}

		bob.Following = []*models.User{charlie}

		alice.Following = []*models.User{bob, charlie}
		charlie.Following = []*models.User{bob}

		now := time.Now()
		bob.AddPost(&models.Post{PostId: 1, UserId: 2, TimeStamp: now.Add(-2 * time.Minute), Content: "Bob's post"})
		charlie.AddPost(&models.Post{PostId: 2, UserId: 3, TimeStamp: now.Add(-1 * time.Minute), Content: "Charlie's post"})

		users[1] = alice
		users[2] = bob
		users[3] = charlie

	*/

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	postpb.RegisterPostServiceServer(s, &postServer{users: users})

	log.Println("gRPC server started on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
