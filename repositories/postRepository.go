package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/unilly-api/db/sqlc"
)

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

func (pr *PostRepo) CreatePostWithTags(ctx context.Context, arg db.CreatePostParams, taggedUserIDs []int64, taggedBy int64) (db.Post, error) {
	tx, err := pr.pool.Begin(ctx)
	if err != nil {
		return db.Post{}, err
	}
	defer tx.Rollback(ctx)

	qtx := pr.q.WithTx(tx)

	post, err := qtx.CreatePost(ctx, arg)
	if err != nil {
		return db.Post{}, err
	}

	for _, taggedUserID := range taggedUserIDs {
		err := qtx.CreatePostTag(ctx, db.CreatePostTagParams{
			PostID:   post.ID,
			UserID:   taggedUserID,
			TaggedBy: taggedBy,
		})
		if err != nil {
			return db.Post{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Post{}, err
	}

	return post, nil
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
