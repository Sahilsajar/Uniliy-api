package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/unilly-api/db/sqlc"
)

var ErrInvalidMediaSelection = errors.New("invalid media ids")

type PostRepo struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewPostRepo(pool *pgxpool.Pool, q *db.Queries) *PostRepo {
	return &PostRepo{
		pool: pool,
		q:    q,
	}
}

func (pr *PostRepo) CreatePost(
	ctx context.Context,
	arg db.CreatePostParams,
	taggedUserIDs []int64,
	taggedBy int64,
	mediaIDs []int64,
	ownerID int64,
) (db.Post, []string, error) {

	tx, err := pr.pool.Begin(ctx)
	if err != nil {
		return db.Post{}, nil, err
	}

	// safer rollback handling
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	qtx := pr.q.WithTx(tx)

	var mediaURLs []string

	// Validate media ownership + fetch URLs
	if len(mediaIDs) > 0 {
		tempMedia, err := qtx.GetUserTempMediaByIDs(ctx, db.GetUserTempMediaByIDsParams{
			UserID:  ownerID,
			Column2: mediaIDs,
		})
		if err != nil {
			return db.Post{}, nil, err
		}

		if len(tempMedia) != len(mediaIDs) {
			return db.Post{}, nil, ErrInvalidMediaSelection
		}

		mediaURLs = make([]string, 0, len(tempMedia))
		for _, m := range tempMedia {
			mediaURLs = append(mediaURLs, m.Url)
		}
	}
	// Create post
	post, err := qtx.CreatePost(ctx, arg)
	if err != nil {
		return db.Post{}, nil, err
	}

	// Bulk tag users
	if len(taggedUserIDs) > 0 {
		err = qtx.CreatePostTagsBulk(ctx, db.CreatePostTagsBulkParams{
			PostID:   post.ID,
			Column2:  taggedUserIDs,
			TaggedBy: taggedBy,
		})
		if err != nil {
			return db.Post{}, nil, err
		}
	}

	// Bulk attach media
	if len(mediaIDs) > 0 {
		err = qtx.CreatePostImagesBulk(ctx, db.CreatePostImagesBulkParams{
			PostID:  post.ID,
			Column2: mediaIDs,
		})
		if err != nil {
			return db.Post{}, nil, err
		}

		// mark media permanent
		err = qtx.MarkMediaPermanentByIDs(ctx, db.MarkMediaPermanentByIDsParams{
			UserID:  ownerID,
			Column2: mediaIDs,
		})
		if err != nil {
			return db.Post{}, nil, err
		}
	}

	// Commit
	if err = tx.Commit(ctx); err != nil {
		return db.Post{}, nil, err
	}

	return post, mediaURLs, nil
}

func (pr *PostRepo) GetPostByID(ctx context.Context, postID int64) (db.Post, error) {
	return pr.q.GetPostByID(ctx, postID)
}

func (pr *PostRepo) GetPostDetailsByID(ctx context.Context, postID, viewerUserID int64) (db.GetPostDetailsByIDRow, error) {
	return pr.q.GetPostDetailsByID(ctx, db.GetPostDetailsByIDParams{
		ID:     postID,
		UserID: viewerUserID,
	})
}

func (pr *PostRepo) ListFeedPosts(ctx context.Context, viewerUserID int64, scope string, page, limit int32) ([]db.ListFeedPostsRow, int64, error) {
	offset := (page - 1) * limit

	posts, err := pr.q.ListFeedPosts(ctx, db.ListFeedPostsParams{
		ViewerUserID: viewerUserID,
		Scope:        scope,
		OffsetCount:  offset,
		LimitCount:   limit,
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := pr.q.CountFeedPosts(ctx, db.CountFeedPostsParams{
		Scope:        scope,
		ViewerUserID: viewerUserID,
	})
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (pr *PostRepo) GetTaggedUserIDs(ctx context.Context, postID int64) ([]int64, error) {
	return pr.q.GetTaggedUserIDs(ctx, postID)
}

func (pr *PostRepo) GetPostImageURLs(ctx context.Context, postID int64) ([]string, error) {
	return pr.q.GetPostImageURLs(ctx, postID)
}

func (pr *PostRepo) GetPostImageURLsByPostIDs(ctx context.Context, postIDs []int64) ([]db.GetPostImageURLsByPostIDsRow, error) {
	return pr.q.GetPostImageURLsByPostIDs(ctx, postIDs)
}

func (pr *PostRepo) GetTaggedUsersByPostIDs(ctx context.Context, postIDs []int64) ([]db.GetTaggedUsersByPostIDsRow, error) {
	return pr.q.GetTaggedUsersByPostIDs(ctx, postIDs)
}

// func (pr *PostRepo) TagUsers(ctx context.Context, postID int64, userIDs []int64, taggedBy int64) error {
// 	for _, userID := range userIDs {
// 		err := pr.q.CreatePostTag(ctx, db.CreatePostTagParams{
// 			PostID:   postID,
// 			UserID:   userID,
// 			TaggedBy: taggedBy,
// 		})
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func (pr *PostRepo) CreateTempMedia(ctx context.Context, publicID, url string, userID int64) (db.Medium, error) {
	return pr.q.CreateMedia(ctx, db.CreateMediaParams{
		PublicID: publicID,
		Url:      url,
		UserID:   userID,
		IsTemp:   true,
	})
}

func (pr *PostRepo) AddComment(
	ctx context.Context,
	postID, userID int64,
	message string,
	parentCommentID *int64,
) (db.AddCommentRow, error) {

	var parentID pgtype.Int8

	if parentCommentID != nil {
		// validate parent exists
		comment, err := pr.q.GetCommentByID(ctx, *parentCommentID)
		if err != nil {
			return db.AddCommentRow{}, err
		}

		// ensure same post
		if comment.PostID != postID {
			return db.AddCommentRow{}, fmt.Errorf("invalid parent comment")
		}

		parentID = pgtype.Int8{
			Int64: comment.ID,
			Valid: true,
		}
	} else {
		parentID = pgtype.Int8{Valid: false}
	}

	arg := db.AddCommentParams{
		PostID:          postID,
		UserID:          userID,
		Message:         message,
		ParentCommentID: parentID,
	}

	return pr.q.AddComment(ctx, arg)
}
func (pr *PostRepo) LikePost(ctx context.Context, postID, userID int64) error {
	return pr.q.LikePost(ctx, db.LikePostParams{
		PostID: postID,
		UserID: userID,
	})
}

func (pr *PostRepo) UnlikePost(ctx context.Context, postID, userID int64) error {
	return pr.q.UnlikePost(ctx, db.UnlikePostParams{
		PostID: postID,
		UserID: userID,
	})
}

func (pr *PostRepo) CheckPostLikeExists(ctx context.Context, postID, userID int64) (bool, error) {
	return pr.q.CheckPostLikeExists(ctx, db.CheckPostLikeExistsParams{
		PostID: postID,
		UserID: userID,
	})

}