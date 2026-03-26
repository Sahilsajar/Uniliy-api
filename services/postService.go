package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/unilly-api/api"
	db "github.com/unilly-api/db/sqlc"
	"github.com/unilly-api/dto"
	"github.com/unilly-api/repositories"
)

type PostService struct {
	postRepo *repositories.PostRepo
}

func NewPostService(postRepo *repositories.PostRepo) *PostService {
	return &PostService{postRepo: postRepo}
}

func (ps *PostService) CreatePost(ctx context.Context, userID int64, req dto.CreatePostRequestDTO) (*dto.PostResponseDTO, error) {
	taggedUserIDs := uniqueIDs(req.TaggedUserIDs)

	post, err := ps.postRepo.CreatePostWithTags(ctx, db.CreatePostParams{
		Title:  pgtype.Text{String: req.Title, Valid: true},
		Body:   pgtype.Text{String: req.Body, Valid: true},
		Status: pgtype.Text{String: req.Status, Valid: true},
		UserID: userID,
	}, taggedUserIDs, userID)
	if err != nil {
		return nil, mapPostError(err)
	}

	return &dto.PostResponseDTO{
		ID:            post.ID,
		Title:         post.Title.String,
		Body:          post.Body.String,
		Status:        post.Status.String,
		UserID:        post.UserID,
		TaggedUserIDs: taggedUserIDs,
	}, nil
}

func (ps *PostService) TagUsersOnPost(ctx context.Context, userID, postID int64, req dto.TagUsersRequestDTO) error {
	post, err := ps.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.NotFound("POST_NOT_FOUND", "Post not found")
		}
		return api.Internal("POST_LOOKUP_FAILED", "Failed to retrieve post").WithCause(err)
	}

	if post.UserID != userID {
		return api.Forbidden("TAG_NOT_ALLOWED", "Only the post owner can tag users")
	}

	userIDs := uniqueIDs(req.UserIDs)
	if len(userIDs) == 0 {
		return api.BadRequest("INVALID_TAG_USERS", "At least one user id is required")
	}

	if err := ps.postRepo.TagUsers(ctx, postID, userIDs, userID); err != nil {
		return mapPostError(err)
	}

	return nil
}

func uniqueIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func mapPostError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			if strings.Contains(pgErr.ConstraintName, "post_tags_user_id_fkey") {
				return api.BadRequest("USER_NOT_FOUND", "One or more tagged users do not exist").WithCause(err)
			}
			if strings.Contains(pgErr.ConstraintName, "post_tags_post_id_fkey") {
				return api.NotFound("POST_NOT_FOUND", "Post not found").WithCause(err)
			}
			return api.BadRequest("INVALID_REFERENCE", "Invalid related resource").WithCause(err)
		case "23505":
			if strings.Contains(pgErr.ConstraintName, "unique_post_user_tag") {
				return api.Conflict("USER_ALREADY_TAGGED", "User is already tagged on this post").WithCause(err)
			}
			return api.Conflict("DUPLICATE_RESOURCE", "Duplicate resource").WithCause(err)
		}
	}

	return api.Internal("POST_OPERATION_FAILED", "Failed to process post request").WithCause(fmt.Errorf("post operation failed: %w", err))
}
