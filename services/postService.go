package services

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
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
		ID:            post.ID,
		Title:         post.Title.String,
		Body:          post.Body.String,
		Status:        string(post.Status.PostStatus),
		UserID:        post.UserID,
		TaggedUserIDs: taggedUserIDs,
		ImageURLs:     mediaURLs,
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
