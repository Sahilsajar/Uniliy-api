package dto

import "time"

type CreatePostRequestDTO struct {
	Title         string  `json:"title" binding:"required"`
	Body          string  `json:"body" binding:"required"`
	TaggedUserIDs []int64 `json:"tagged_user_ids"`
	MediaIDs      []int64 `json:"media_ids"`
}

type TagUsersRequestDTO struct {
	UserIDs []int64 `json:"user_ids" binding:"required,min=1"`
}

type PostUserSummaryDTO struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	Name       string `json:"name,omitempty"`
	ProfilePic string `json:"profile_pic,omitempty"`
}

type PostStatsDTO struct {
	LikesCount    int64 `json:"likes_count"`
	CommentsCount int64 `json:"comments_count"`
}

type PostResponseDTO struct {
	ID          int64                `json:"id"`
	Title       string               `json:"title"`
	Body        string               `json:"body"`
	Status      string               `json:"status"`
	UserID      int64                `json:"user_id"`
	Author      PostUserSummaryDTO   `json:"author"`
	TaggedUsers []PostUserSummaryDTO `json:"tagged_users"`
	ImageURLs   []string             `json:"image_urls"`
	Stats       PostStatsDTO         `json:"stats"`
	IsLiked     bool                 `json:"is_liked"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type FeedPaginationDTO struct {
	Page       int32 `json:"page"`
	Limit      int32 `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int32 `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
}

type PostFeedResponseDTO struct {
	Items      []PostResponseDTO `json:"items"`
	Pagination FeedPaginationDTO `json:"pagination"`
}

type UploadPostMediaResponseDTO struct {
	MediaID  int64  `json:"media_id"`
	PublicID string `json:"public_id"`
	URL      string `json:"url"`
}

type AddCommentRequestDTO struct {
	Message string `json:"message" binding:"required"`
	ParentCommentID *int64 `json:"parent_comment_id,omitempty"`
}

type CommentResponseDTO struct {
	ID        int64     `json:"id"`
	PostID    int64     `json:"post_id"`
	UserID    int64     `json:"user_id"`
	Message   string    `json:"message"`
	Author    PostUserSummaryDTO `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	RepliesCount int64     `json:"replies_count"`
}