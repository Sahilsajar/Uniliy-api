package services

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

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
	client   *http.Client
}

func NewPostService(postRepo *repositories.PostRepo) *PostService {
	return &PostService{
		postRepo: postRepo,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (ps *PostService) CreatePost(ctx context.Context, userID int64, req dto.CreatePostRequestDTO) (*dto.PostResponseDTO, error) {
	taggedUserIDs := uniqueIDs(req.TaggedUserIDs)

	mediaIDs := uniqueIDs(req.MediaIDs)
	post, mediaURLs, err := ps.postRepo.CreatePost(ctx, db.CreatePostParams{
		Title:  pgtype.Text{String: req.Title, Valid: true},
		Body:   pgtype.Text{String: req.Body, Valid: true},
		Status: db.NullPostStatus{PostStatus: db.PostStatusPublished, Valid: true},
		UserID: userID,
	}, taggedUserIDs, userID, mediaIDs, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrInvalidMediaSelection) {
			return nil, api.BadRequest("INVALID_MEDIA_IDS", "One or more media ids are invalid or not temporary")
		}
		return nil, mapPostError(err)
	}

	return &dto.PostResponseDTO{
		ID:          post.ID,
		Title:       post.Title.String,
		Body:        post.Body.String,
		Status:      string(post.Status.PostStatus),
		UserID:      post.UserID,
		ImageURLs:   mediaURLs,
		TaggedUsers: buildTaggedUsersFromIDs(taggedUserIDs),
		CreatedAt:   post.CreatedAt.Time,
		UpdatedAt:   post.UpdatedAt.Time,
	}, nil
}

// func (ps *PostService) TagUsersOnPost(ctx context.Context, userID, postID int64, req dto.TagUsersRequestDTO) error {
// 	post, err := ps.postRepo.GetPostByID(ctx, postID)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			return api.NotFound("POST_NOT_FOUND", "Post not found")
// 		}
// 		return api.Internal("POST_LOOKUP_FAILED", "Failed to retrieve post").WithCause(err)
// 	}

// 	if post.UserID != userID {
// 		return api.Forbidden("TAG_NOT_ALLOWED", "Only the post owner can tag users")
// 	}

// 	userIDs := uniqueIDs(req.UserIDs)
// 	if len(userIDs) == 0 {
// 		return api.BadRequest("INVALID_TAG_USERS", "At least one user id is required")
// 	}

// 	if err := ps.postRepo.TagUsers(ctx, postID, userIDs, userID); err != nil {
// 		return mapPostError(err)
// 	}

// 	return nil
// }

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

func (ps *PostService) UploadTempPostMedia(ctx context.Context, userID int64, fileHeader *multipart.FileHeader) (*dto.UploadPostMediaResponseDTO, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, api.BadRequest("INVALID_FILE", "Failed to read uploaded file").WithCause(err)
	}
	defer file.Close()

	publicID, secureURL, err := ps.uploadToCloudinary(ctx, userID, fileHeader.Filename, file)
	if err != nil {
		return nil, err
	}

	media, err := ps.postRepo.CreateTempMedia(ctx, publicID, secureURL, userID)
	if err != nil {
		return nil, api.Internal("TEMP_MEDIA_SAVE_FAILED", "Failed to store uploaded media").WithCause(err)
	}

	return &dto.UploadPostMediaResponseDTO{
		MediaID:  media.ID,
		PublicID: media.PublicID,
		URL:      media.Url,
	}, nil
}

func (ps *PostService) uploadToCloudinary(ctx context.Context, userID int64, filename string, file io.Reader) (string, string, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return "", "", api.Internal("CLOUDINARY_CONFIG_MISSING", "Cloudinary is not configured")
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	folder := "unilly/posts/temp"
	publicID := fmt.Sprintf("user_%d_%d", userID, time.Now().UnixNano())
	paramsToSign := map[string]string{
		"public_id": publicID,
		"folder":    folder,
		"timestamp": timestamp,
	}
	signature := cloudinarySignature(paramsToSign, apiSecret)

	form := url.Values{}
	form.Set("api_key", apiKey)
	form.Set("timestamp", timestamp)
	form.Set("signature", signature)
	form.Set("folder", folder)
	form.Set("public_id", publicID)

	bodyReader, bodyWriter := io.Pipe()
	writer := multipart.NewWriter(bodyWriter)
	// Write form data and file in a separate goroutine to avoid blocking
	go func() {
		defer bodyWriter.Close()
		defer writer.Close()

		for key, values := range form {
			for _, value := range values {
				_ = writer.WriteField(key, value)
			}
		}

		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			_ = bodyWriter.CloseWithError(err)
			return
		}
		if _, err := io.Copy(part, file); err != nil {
			_ = bodyWriter.CloseWithError(err)
			return
		}
	}()

	endpoint := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cloudName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bodyReader)
	if err != nil {
		return "", "", api.Internal("CLOUDINARY_UPLOAD_FAILED", "Failed to upload media").WithCause(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := ps.client.Do(req)
	if err != nil {
		return "", "", api.Internal("CLOUDINARY_UPLOAD_FAILED", "Failed to upload media").WithCause(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", "", api.BadRequest("CLOUDINARY_UPLOAD_REJECTED", "Cloudinary rejected uploaded media")
	}

	var payload struct {
		PublicID  string `json:"public_id"`
		SecureURL string `json:"secure_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", "", api.Internal("CLOUDINARY_RESPONSE_INVALID", "Invalid upload response").WithCause(err)
	}
	if payload.PublicID == "" || payload.SecureURL == "" {
		return "", "", api.Internal("CLOUDINARY_RESPONSE_INVALID", "Upload response missing required fields")
	}
	return payload.PublicID, payload.SecureURL, nil
}

func cloudinarySignature(params map[string]string, apiSecret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	raw := strings.Join(parts, "&") + apiSecret
	sum := sha1.Sum([]byte(raw))
	return hex.EncodeToString(sum[:])
}

const (
	postFeedScopeAll       = "all"
	postFeedScopeFollowing = "following"
	postFeedScopeMine      = "mine"
	defaultFeedPage        = int32(1)
	defaultFeedLimit       = int32(20)
	maxFeedLimit           = int32(50)
)

func (ps *PostService) GetPostByID(ctx context.Context, postID, viewerUserID int64) (*dto.PostResponseDTO, error) {
	post, err := ps.postRepo.GetPostDetailsByID(ctx, postID, viewerUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.NotFound("POST_NOT_FOUND", "Post not found")
		}
		return nil, api.Internal("POST_LOOKUP_FAILED", "Failed to retrieve post").WithCause(err)
	}

	postIDs := []int64{post.ID}

	imageRows, err := ps.postRepo.GetPostImageURLsByPostIDs(ctx, postIDs)
	if err != nil {
		return nil, api.Internal("POST_IMAGES_LOOKUP_FAILED", "Failed to retrieve post images").WithCause(err)
	}

	taggedUserRows, err := ps.postRepo.GetTaggedUsersByPostIDs(ctx, postIDs)
	if err != nil {
		return nil, api.Internal("TAGGED_USERS_LOOKUP_FAILED", "Failed to retrieve tagged users").WithCause(err)
	}

	imageURLsByPostID := groupImageURLsByPostID(imageRows)
	taggedUsersByPostID := groupTaggedUsersByPostID(taggedUserRows)

	response := buildPostResponseFromDetails(post, imageURLsByPostID[post.ID], taggedUsersByPostID[post.ID])
	return &response, nil
}

func (ps *PostService) GetFeed(ctx context.Context, viewerUserID int64, scope string, page, limit int32) (*dto.PostFeedResponseDTO, error) {
	if scope == "" {
		scope = postFeedScopeAll
	}
	if scope != postFeedScopeAll && scope != postFeedScopeFollowing && scope != postFeedScopeMine {
		return nil, api.BadRequest("INVALID_FEED_SCOPE", "scope must be one of: all, following, mine")
	}
	if page <= 0 {
		page = defaultFeedPage
	}
	if limit <= 0 {
		limit = defaultFeedLimit
	}
	if limit > maxFeedLimit {
		limit = maxFeedLimit
	}

	posts, totalItems, err := ps.postRepo.ListFeedPosts(ctx, viewerUserID, scope, page, limit)
	if err != nil {
		return nil, api.Internal("POST_FEED_LOOKUP_FAILED", "Failed to retrieve post feed").WithCause(err)
	}

	postIDs := make([]int64, 0, len(posts))
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
	}

	imageURLsByPostID := make(map[int64][]string, len(postIDs))
	taggedUsersByPostID := make(map[int64][]dto.PostUserSummaryDTO, len(postIDs))

	if len(postIDs) > 0 {
		imageRows, err := ps.postRepo.GetPostImageURLsByPostIDs(ctx, postIDs)
		if err != nil {
			return nil, api.Internal("POST_IMAGES_LOOKUP_FAILED", "Failed to retrieve post images").WithCause(err)
		}
		imageURLsByPostID = groupImageURLsByPostID(imageRows)

		taggedUserRows, err := ps.postRepo.GetTaggedUsersByPostIDs(ctx, postIDs)
		if err != nil {
			return nil, api.Internal("TAGGED_USERS_LOOKUP_FAILED", "Failed to retrieve tagged users").WithCause(err)
		}
		taggedUsersByPostID = groupTaggedUsersByPostID(taggedUserRows)
	}

	items := make([]dto.PostResponseDTO, 0, len(posts))
	for _, post := range posts {
		items = append(items, buildPostResponseFromFeed(post, imageURLsByPostID[post.ID], taggedUsersByPostID[post.ID]))
	}

	totalPages := int32(0)
	if totalItems > 0 {
		totalPages = int32(math.Ceil(float64(totalItems) / float64(limit)))
	}

	return &dto.PostFeedResponseDTO{
		Items: items,
		Pagination: dto.FeedPaginationDTO{
			Page:       page,
			Limit:      limit,
			TotalItems: totalItems,
			TotalPages: totalPages,
			HasNext:    int64(page)*int64(limit) < totalItems,
		},
	}, nil
}

func groupImageURLsByPostID(rows []db.GetPostImageURLsByPostIDsRow) map[int64][]string {
	result := make(map[int64][]string)
	for _, row := range rows {
		result[row.PostID] = append(result[row.PostID], row.ImageUrl)
	}
	return result
}

func groupTaggedUsersByPostID(rows []db.GetTaggedUsersByPostIDsRow) map[int64][]dto.PostUserSummaryDTO {
	result := make(map[int64][]dto.PostUserSummaryDTO)
	for _, row := range rows {
		result[row.PostID] = append(result[row.PostID], dto.PostUserSummaryDTO{
			ID:         row.ID,
			Username:   row.Username,
			Name:       row.Name.String,
			ProfilePic: row.ProfilePic.String,
		})
	}
	return result
}

func buildTaggedUsersFromIDs(ids []int64) []dto.PostUserSummaryDTO {
	taggedUsers := make([]dto.PostUserSummaryDTO, 0, len(ids))
	for _, id := range ids {
		taggedUsers = append(taggedUsers, dto.PostUserSummaryDTO{ID: id})
	}
	return taggedUsers
}

func buildPostResponseFromDetails(post db.GetPostDetailsByIDRow, imageURLs []string, taggedUsers []dto.PostUserSummaryDTO) dto.PostResponseDTO {
	return dto.PostResponseDTO{
		ID:          post.ID,
		Title:       post.Title.String,
		Body:        post.Body.String,
		Status:      string(post.Status.PostStatus),
		UserID:      post.UserID,
		Author:      buildPostAuthor(post.UserID, post.Username, post.Name, post.ProfilePic),
		TaggedUsers: taggedUsers,
		ImageURLs:   imageURLs,
		Stats: dto.PostStatsDTO{
			LikesCount:    post.LikesCount,
			CommentsCount: post.CommentsCount,
		},
		IsLiked:   post.IsLiked,
		CreatedAt: post.CreatedAt.Time,
		UpdatedAt: post.UpdatedAt.Time,
	}
}

func buildPostResponseFromFeed(post db.ListFeedPostsRow, imageURLs []string, taggedUsers []dto.PostUserSummaryDTO) dto.PostResponseDTO {
	return dto.PostResponseDTO{
		ID:          post.ID,
		Title:       post.Title.String,
		Body:        post.Body.String,
		Status:      string(post.Status.PostStatus),
		UserID:      post.UserID,
		Author:      buildPostAuthor(post.UserID, post.Username, post.Name, post.ProfilePic),
		TaggedUsers: taggedUsers,
		ImageURLs:   imageURLs,
		Stats: dto.PostStatsDTO{
			LikesCount:    post.LikesCount,
			CommentsCount: post.CommentsCount,
		},
		IsLiked:   post.IsLiked,
		CreatedAt: post.CreatedAt.Time,
		UpdatedAt: post.UpdatedAt.Time,
	}
}

func buildPostAuthor(userID int64, username string, name, profilePic pgtype.Text) dto.PostUserSummaryDTO {
	return dto.PostUserSummaryDTO{
		ID:         userID,
		Username:   username,
		Name:       name.String,
		ProfilePic: profilePic.String,
	}
}

func (ps *PostService) AddComment(ctx context.Context, postID, userID int64, commentBody dto.AddCommentRequestDTO) (*dto.CommentResponseDTO, error) {
	fmt.Printf("Adding comment to post %d by user %d: %s\n", postID, userID, commentBody.Message)
	comment, err := ps.postRepo.AddComment(ctx, postID, userID, commentBody.Message, commentBody.ParentCommentID)
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.NotFound("POST_NOT_FOUND", "Post not found")
		}
		if strings.Contains(err.Error(), "invalid parent comment") {
			return nil, api.BadRequest("INVALID_PARENT_COMMENT", "Parent comment does not exist or does not belong to the same post")
		}
		return nil, api.Internal("ADD_COMMENT_FAILED", "Failed to add comment").WithCause(err)
	}

	return &dto.CommentResponseDTO{
		ID:        comment.ID,
		PostID:    comment.PostID,
		UserID:    comment.UserID,
		Message:   comment.Message,
		CreatedAt: comment.CreatedAt.Time,
	}, nil
}
func (ps *PostService) ToggleLikePost(ctx context.Context, userID, postID int64) (bool, error) {

	// Check post exists
	_, err := ps.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, api.NotFound("POST_NOT_FOUND", "Post not found")
		}
		return false, api.Internal("POST_LOOKUP_FAILED", "Failed to retrieve post").WithCause(err)
	}

	// Check already liked
	exists, err := ps.postRepo.CheckPostLikeExists(ctx, postID, userID)
	if err != nil {
		return false, api.Internal("LIKE_CHECK_FAILED", "Failed to check like status").WithCause(err)
	}

	if exists {
		// 👉 UNLIKE
		err = ps.postRepo.UnlikePost(ctx, postID, userID)
		if err != nil {
			return false, api.Internal("UNLIKE_FAILED", "Failed to unlike post").WithCause(err)
		}
		return false, nil // false = now unliked
	}

	// 👉 LIKE
	err = ps.postRepo.LikePost(ctx, postID, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return true, nil // already liked (edge case)
			}
		}
		return false, api.Internal("LIKE_FAILED", "Failed to like post").WithCause(err)
	}

	return true, nil // true = now liked
}

func (ps *PostService) GetComments(ctx context.Context, postID, viewerUserID int64) ([]dto.CommentResponseDTO, error) {
	// Check post exists
	_, err := ps.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.NotFound("POST_NOT_FOUND", "Post not found")
		}
		return nil, api.Internal("POST_LOOKUP_FAILED", "Failed to retrieve post").WithCause(err)
	}

	comments, err := ps.postRepo.GetComments(ctx, postID)			
	if err != nil {
		return nil, api.Internal("COMMENTS_LOOKUP_FAILED", "Failed to retrieve comments").WithCause(err)
	}

	commentDTOs := make([]dto.CommentResponseDTO, 0, len(comments))
	for _, comment := range comments {
		commentDTOs = append(commentDTOs, dto.CommentResponseDTO{
			ID:        comment.ID,
			PostID:    comment.PostID,
			UserID:    comment.UserID,
			Message:   comment.Message,
			CreatedAt: comment.CreatedAt.Time,
			Author: dto.PostUserSummaryDTO{
				ID: comment.UserID,
				Name: comment.Name.String,
				Username: comment.Username.String,
				ProfilePic: comment.ProfilePic.String,
			},
			RepliesCount: comment.RepliesCount,
		})
	}

	return commentDTOs, nil
}	