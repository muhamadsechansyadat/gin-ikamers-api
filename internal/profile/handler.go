package profile

import (
	"errors"
	"gin-ikamers-api/internal/auth"
	"gin-ikamers-api/internal/shared/response"
	"gin-ikamers-api/internal/shared/storage"
	"gin-ikamers-api/internal/user"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"time"
)

const maxAvatarSize = 2 * 1024 * 1024 // 2MB

var allowedAvatarMimes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type Handler struct {
	service     *Service
	userService *user.Service
	authService *auth.Service
	storage     *storage.SupabaseClient
}

func NewHandler(service *Service, userService *user.Service, authService *auth.Service, s *storage.SupabaseClient) *Handler {
	return &Handler{
		service: service, userService: userService,
		authService: authService, storage: s,
	}
}

func (h *Handler) Show(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()

	p, err := h.service.GetUserByID(ctx, uid)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch profile", nil)
		return
	}
	if p == nil {
		response.NotFound(c, "Profile not found", nil)
		return
	}

	h.respondWithProfile(c, "Profile fetched", uid, p)
}

func (h *Handler) Update(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleBindError(c, err)
		return
	}

	var birthDate *time.Time
	if req.BirthDate != nil {
		t, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			response.BadRequest(c, "Invalid birth_date format (expected YYYY-MM-DD)", err.Error())
			return
		}
		birthDate = &t
	}

	p, err := h.service.Upsert(ctx, uid, req.FullName, req.Gender, birthDate, req.Bio)
	if err != nil {
		response.InternalServerError(c, "Failed to update profile", nil)
		return
	}

	h.respondWithProfile(c, "Profile updated", uid, p)
}

func (h *Handler) UploadAvatar(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		response.BadRequest(c, "avatar file is required", err.Error())
		return
	}
	if fileHeader.Size > maxAvatarSize {
		response.PayloadTooLarge(c, "avatar must be at most 2MB", nil)
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		response.InternalServerError(c, "Failed to open uploaded file", nil)
		return
	}
	defer src.Close()

	// deteksi MIME dari 512 byte pertama, bukan trust filename/header client
	head := make([]byte, 512)
	n, _ := src.Read(head)
	contentType := http.DetectContentType(head[:n])
	ext, allowed := allowedAvatarMimes[contentType]
	if !allowed {
		response.UnsupportedMediaType(c, "avatar must be JPEG, PNG, or WEBP", nil)
		return
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		response.InternalServerError(c, "Failed to reset file reader", nil)
		return
	}

	p, err := h.service.UploadAvatar(ctx, uid, src, contentType, ext)
	if err != nil {
		if errors.Is(err, ErrProfileNotFound) {
			response.NotFound(c, "Profile not found", nil)
			return
		}
		response.InternalServerError(c, "Failed to upload avatar", err.Error())
		return
	}

	h.respondWithProfile(c, "Avatar uploaded", uid, p)
}

func (h *Handler) DeleteAvatar(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()

	p, err := h.service.DeleteAvatar(ctx, uid)
	if err != nil {
		if errors.Is(err, ErrProfileNotFound) {
			response.NotFound(c, "Profile not found", nil)
			return
		}
		response.InternalServerError(c, "Failed to delete avatar", nil)
		return
	}

	h.respondWithProfile(c, "Avatar deleted", uid, p)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleBindError(c, err)
		return
	}

	err := h.authService.ChangePassword(c.Request.Context(), uid, req.CurrentPassword, req.NewPassword)
	switch {
	case err == nil:
		response.OK(c, "Password updated. Please log in again.", nil)
	case errors.Is(err, auth.ErrCurrentPasswordRequired):
		response.UnprocessableEntity(c, "Validation failed", []response.FieldError{
			{
				Field:   "current_password",
				Tag:     "required",
				Message: "Current password is required",
			},
		})
	case errors.Is(err, auth.ErrInvalidCredentials):
		response.Unauthorized(c, "Current password is incorrect", nil)
	case errors.Is(err, auth.ErrPasswordSameAsCurrent):
		response.BadRequest(c, "New password must differ from current", nil)
	case errors.Is(err, auth.ErrUserNotFound):
		response.NotFound(c, "User not found", nil)
	default:
		response.InternalServerError(c, "Failed to change password", err.Error())
	}
}

func (h *Handler) RequestEmailChange(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	var req RequestEmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleBindError(c, err)
		return
	}

	err := h.authService.RequestEmailChange(c.Request.Context(), uid, req.NewEmail, req.CurrentPassword)
	switch {
	case err == nil:
		response.OK(c, "Verification code sent to the new email", nil)
	case errors.Is(err, auth.ErrCurrentPasswordRequired):
		response.UnprocessableEntity(c, "Validation failed", []response.FieldError{
			{
				Field:   "current_password",
				Tag:     "required",
				Message: "Current password is required",
			},
		})
	case errors.Is(err, auth.ErrInvalidCredentials):
		response.Unauthorized(c, "Current password is incorrect", nil)
	case errors.Is(err, auth.ErrEmailAlreadyUsed):
		response.Conflict(c, "Email is already in use", nil)
	default:
		response.InternalServerError(c, "Failed to request email change", err.Error())
	}
}

func (h *Handler) ConfirmEmailChange(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	var req ConfirmEmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleBindError(c, err)
		return
	}

	err := h.authService.ConfirmEmailChange(c.Request.Context(), uid, req.OTP)
	switch {
	case err == nil:
		response.OK(c, "Email changed. Please login again.", nil)
	case errors.Is(err, auth.ErrInvalidOTP):
		response.BadRequest(c, "Invalid or expired OTP", nil)
	case errors.Is(err, auth.ErrNoActiveVerification):
		response.NotFound(c, "No pending verification. Please request again.", nil)
	case errors.Is(err, auth.ErrEmailAlreadyUsed):
		response.Conflict(c, "Email already in use", nil)
	default:
		response.InternalServerError(c, "Failed to confirm email change", err.Error())
	}
}

func currentUserID(c *gin.Context) (int64, bool) {
	v, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Unauthorized", nil)
		return 0, false
	}
	return v.(int64), true
}

func (h *Handler) respondWithProfile(c *gin.Context, msg string, uid int64, p *Profile) {
	u, err := h.userService.GetByID(c.Request.Context(), uid)
	if err != nil || u == nil {
		response.InternalServerError(c, "Failed to fetch user", nil)
		return
	}
	summary := UserSummaryResponse{
		UUID:  u.UUID,
		Email: u.Email,
		Role:  u.RoleName,
	}
	response.OK(c, msg, ToResponse(p, summary, h.storage))
}
