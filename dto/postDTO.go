package dto
type CreatePostRequestDTO struct {
	Title         string  `json:"title" binding:"required"`
	Body          string  `json:"body" binding:"required"`
	TaggedUserIDs []int64 `json:"tagged_user_ids"`
	MediaIDs      []int64 `json:"media_ids"`
}

type TagUsersRequestDTO struct {
	UserIDs []int64 `json:"user_ids" binding:"required,min=1"`
}

type PostResponseDTO struct {
	ID            int64    `json:"id"`
	Title         string   `json:"title"`
	Body          string   `json:"body"`
	Status        string   `json:"status"`
	UserID        int64    `json:"user_id"`
	TaggedUserIDs []int64  `json:"tagged_user_ids"`
	ImageURLs     []string `json:"image_urls"`
}

type UploadPostMediaResponseDTO struct {
	MediaID  int64  `json:"media_id"`
	PublicID string `json:"public_id"`
	URL      string `json:"url"`
}
