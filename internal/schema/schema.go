package schema

import (
	"postaggregator/internal/grpcclient"
	"postaggregator/internal/models"
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

					// Create a temporary user with proper following relationships
					tempUser := &models.User{UserId: userID}

					// Set following relationships based on userID
					switch userID {
					case 1: // Alice follows Bob and Charlie
						tempUser.Following = []*models.User{
							{UserId: 2},
							{UserId: 3},
						}
					case 2: // Bob follows Charlie
						tempUser.Following = []*models.User{
							{UserId: 3},
						}
					case 3: // Charlie follows Bob
						tempUser.Following = []*models.User{
							{UserId: 2},
						}
					default:
						return []*models.Post{}, nil
					}

					// For each followed user, fetch their posts via gRPC
					for _, followedUser := range tempUser.Following {
						pbPosts, err := grpcClient.ListPostsByUser(int32(followedUser.UserId))
						if err != nil {
							continue
						}

						// Add posts to the followed user
						for _, pbPost := range pbPosts {
							timestamp, _ := time.Parse(time.RFC3339, pbPost.Timestamp)
							followedUser.AddPost(&models.Post{
								PostId:    int(pbPost.PostId),
								UserId:    int(pbPost.UserId),
								Content:   pbPost.Content,
								TimeStamp: timestamp,
							})
						}
					}

					return tempUser.GetUserFeed(), nil
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query: rootQuery,
	})
}
