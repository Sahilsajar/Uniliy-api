package repositories

import (
	"context"
	"errors"

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

func (pr *PostRepo) CreatePostWithAssets(ctx context.Context, arg db.CreatePostParams, taggedUserIDs []int64, taggedBy int64, mediaIDs []int64, ownerID int64) (db.Post, []string, error) {
	tx, err := pr.pool.Begin(ctx)
	if err != nil {
		return db.Post{}, nil, err
	}
	defer tx.Rollback(ctx)

	qtx := pr.q.WithTx(tx)

	mediaURLs := make([]string, 0, len(mediaIDs))
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
		for _, m := range tempMedia {
			mediaURLs = append(mediaURLs, m.Url)
		}
	}

	post, err := qtx.CreatePost(ctx, arg)
	if err != nil {
		return db.Post{}, nil, err
	}

	for _, taggedUserID := range taggedUserIDs {
		err := qtx.CreatePostTag(ctx, db.CreatePostTagParams{
			PostID:   post.ID,
			UserID:   taggedUserID,
			TaggedBy: taggedBy,
		})
		if err != nil {
			return db.Post{}, nil, err
		}
	}

	for i, mediaID := range mediaIDs {
		err := qtx.CreatePostImage(ctx, db.CreatePostImageParams{
			ImageUrl: mediaURLs[i],
			MediaID:  pgtype.Int8{Int64: mediaID, Valid: true},
			PostID:   post.ID,
		})
		if err != nil {
			return db.Post{}, nil, err
		}
	}

	if len(mediaIDs) > 0 {
		err = qtx.MarkMediaPermanentByIDs(ctx, db.MarkMediaPermanentByIDsParams{
			UserID:  ownerID,
			Column2: mediaIDs,
		})
		if err != nil {
			return db.Post{}, nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Post{}, nil, err
	}

	return post, mediaURLs, nil
}

func (pr *PostRepo) GetPostByID(ctx context.Context, postID int64) (db.Post, error) {
	return pr.q.GetPostByID(ctx, postID)
}

func (pr *PostRepo) TagUsers(ctx context.Context, postID int64, userIDs []int64, taggedBy int64) error {
	for _, userID := range userIDs {
		err := pr.q.CreatePostTag(ctx, db.CreatePostTagParams{
			PostID:   postID,
			UserID:   userID,
			TaggedBy: taggedBy,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (pr *PostRepo) CreateTempMedia(ctx context.Context, publicID, url string, userID int64) (db.Medium, error) {
	return pr.q.CreateMedia(ctx, db.CreateMediaParams{
		PublicID: publicID,
		Url:      url,
		UserID:   userID,
		IsTemp:   true,
	})
}
