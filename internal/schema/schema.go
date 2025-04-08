package schema

import (
	"postaggregator/internal/grpcclient"
	"postaggregator/internal/models"
	"sort"
	"strconv"
	"time"

	"github.com/graphql-go/graphql"
)

func SetupSchema(grpcClient *grpcclient.GRPCClient) (graphql.Schema, error) {
	postType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Post",
		Fields: graphql.Fields{
			"postId": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return strconv.Itoa(p.Source.(*models.Post).PostId), nil
				},
			},
			"userId": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return strconv.Itoa(p.Source.(*models.Post).UserId), nil
				},
			},
			"content": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(*models.Post).Content, nil
				},
			},
			"timestamp": &graphql.Field{
				Type: graphql.DateTime,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(*models.Post).TimeStamp, nil
				},
			},
		},
	})

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"getTimeline": &graphql.Field{
				Type: graphql.NewList(postType),
				Args: graphql.FieldConfigArgument{
					"userId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, err := strconv.Atoi(p.Args["userId"].(string))
					if err != nil {
						return nil, err
					}

					// 1. Get who this user follows
					followingIDs, err := grpcClient.ListFollowing(int32(userID))
					if err != nil {
						return nil, err
					}

					// 2. For each followed user, get their posts
					var allPosts []*models.Post
					for _, followedID := range followingIDs {
						pbPosts, err := grpcClient.ListPostsByUser(followedID)
						if err != nil {
							continue
						}

						// Convert protobuf posts to our model
						for _, pbPost := range pbPosts {
							timestamp, _ := time.Parse(time.RFC3339, pbPost.Timestamp)
							allPosts = append(allPosts, &models.Post{
								PostId:    int(pbPost.PostId),
								UserId:    int(pbPost.UserId),
								Content:   pbPost.Content,
								TimeStamp: timestamp,
							})
						}
					}

					// 3. Sort by timestamp (newest first) and limit to 20
					sort.Slice(allPosts, func(i, j int) bool {
						return allPosts[i].TimeStamp.After(allPosts[j].TimeStamp)
					})

					if len(allPosts) > 20 {
						return allPosts[:20], nil
					}
					return allPosts, nil
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query: rootQuery,
	})
}
