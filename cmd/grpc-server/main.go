package main

import (
	"context"
	"log"
	"net"
	"time"

	"postaggregator/internal/models"
	"postaggregator/postpb"

	"google.golang.org/grpc"
)

type postServer struct {
	postpb.UnimplementedPostServiceServer
	users map[int32]*models.User
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

func main() {
	users := make(map[int32]*models.User)

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
