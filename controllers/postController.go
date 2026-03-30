package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/unilly-api/api"
	"github.com/unilly-api/dto"
	"github.com/unilly-api/services"
)

type PostController struct {
	postService *services.PostService
}

func NewPostController(postService *services.PostService) *PostController {
	return &PostController{postService: postService}
}

func (pc *PostController) CreatePost(ctx *gin.Context) error {
	var req dto.CreatePostRequestDTO
	if err := api.BindJSON(ctx, &req); err != nil {
		return err
	}

	userIDTemp, _ := ctx.Get("user_id")
	userID := userIDTemp.(int64)

	post, err := pc.postService.CreatePost(ctx.Request.Context(), userID, req)
	if err != nil {
		return err
	}

	api.Success(ctx, http.StatusCreated, "Post created successfully", post)
	return nil
}

// func (pc *PostController) TagUsers(ctx *gin.Context) error {
// 	postID, err := strconv.ParseInt(ctx.Param("postID"), 10, 64)
// 	if err != nil || postID <= 0 {
// 		return api.BadRequest("INVALID_POST_ID", "Invalid post ID")
// 	}

// 	var req dto.TagUsersRequestDTO
// 	if err := api.BindJSON(ctx, &req); err != nil {
// 		return err
// 	}

// 	userID, err := strconv.ParseInt(ctx.GetString("user_id"), 10, 64)
// 	if err != nil {
// 		return api.BadRequest("INVALID_USER_ID", "Invalid user ID")
// 	}

// 	if err := pc.postService.TagUsersOnPost(ctx.Request.Context(), userID, postID, req); err != nil {
// 		return err
// 	}
// 	api.Success(ctx, http.StatusOK, "Users tagged successfully", nil)
// 	return nil
// }

func (pc *PostController) UploadTempMedia(ctx *gin.Context) error {
	userIDtemp, _ := ctx.Get("user_id")
	userID, _ := userIDtemp.(int64)

	file, err := ctx.FormFile("file")
	if err != nil {
		return api.BadRequest("FILE_REQUIRED", "file is required")
	}

	media, err := pc.postService.UploadTempPostMedia(ctx.Request.Context(), userID, file)
	if err != nil {
		return err
	}

	api.Success(ctx, http.StatusCreated, "Media uploaded successfully", media)
	return nil
}
