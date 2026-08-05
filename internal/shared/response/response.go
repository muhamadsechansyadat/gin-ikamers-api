package response

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

func NewMeta(page, perPage int, total int64) Meta {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	return Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

type FieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Param   string `json:"param,omitempty"`
	Message string `json:"message"`
}

func JSON(c *gin.Context, status int, res Response) {
	c.JSON(status, res)
}

func Abort(c *gin.Context, status int, res Response) {
	c.AbortWithStatusJSON(status, res)
}

func Success(c *gin.Context, status int, message string, data interface{}) {
	JSON(c, status, Response{Success: true, Message: message, Data: data})
}

func SuccessWithMeta(c *gin.Context, status int, message string, data, meta interface{}) {
	JSON(c, status, Response{Success: true, Message: message, Data: data, Meta: meta})
}

func Error(c *gin.Context, status int, message string, err interface{}) {
	JSON(c, status, Response{Success: false, Message: message, Errors: err})
}

func AbortError(c *gin.Context, status int, message string, err interface{}) {
	Abort(c, status, Response{Success: false, Message: message, Errors: err})
}

func defaultMsg(message, fallback string) string {
	if message == "" {
		return fallback
	}
	return message
}

func OK(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusOK, defaultMsg(message, "Success"), data)
}

func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, defaultMsg(message, "Resource created successfully"), data)
}

func Accepted(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusAccepted, defaultMsg(message, "Request accepted and is being processed"), data)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Paginated(c *gin.Context, message string, data interface{}, page, perPage int, total int64) {
	SuccessWithMeta(c, http.StatusOK, defaultMsg(message, "Success"), data, NewMeta(page, perPage, total))
}

func BadRequest(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusBadRequest, defaultMsg(message, "Invalid request"), err)
}

func Unauthorized(c *gin.Context, message string, err interface{}) {
	AbortError(c, http.StatusUnauthorized, defaultMsg(message, "Authentication required"), err)
}

func Forbidden(c *gin.Context, message string, err interface{}) {
	AbortError(c, http.StatusForbidden, defaultMsg(message, "Access denied"), err)
}

func NotFound(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusNotFound, defaultMsg(message, "Resource not found"), err)
}

func MethodNotAllowed(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusMethodNotAllowed, defaultMsg(message, "Method not allowed"), err)
}

func NotAcceptable(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusNotAcceptable, defaultMsg(message, "Requested response format is not supported"), err)
}

func RequestTimeout(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusRequestTimeout, defaultMsg(message, "Request timed out"), err)
}

func Conflict(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusConflict, defaultMsg(message, "Resource already exists or conflicts with current state"), err)
}

func Gone(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusGone, defaultMsg(message, "Resource is no longer available"), err)
}

func PayloadTooLarge(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusRequestEntityTooLarge, defaultMsg(message, "Payload too large"), err)
}

func UnsupportedMediaType(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusUnsupportedMediaType, defaultMsg(message, "Unsupported media type"), err)
}

func UnprocessableEntity(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusUnprocessableEntity, defaultMsg(message, "Validation failed"), err)
}

func TooManyRequests(c *gin.Context, message string, err interface{}) {
	AbortError(c, http.StatusTooManyRequests, defaultMsg(message, "Too many requests"), err)
}

func InternalServerError(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusInternalServerError, defaultMsg(message, "Internal server error"), err)
}

func NotImplemented(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusNotImplemented, defaultMsg(message, "Not implemented"), err)
}

func BadGateway(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusBadGateway, defaultMsg(message, "Invalid response from upstream service"), err)
}

func ServiceUnavailable(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusServiceUnavailable, defaultMsg(message, "Service temporarily unavailable"), err)
}

func GatewayTimeout(c *gin.Context, message string, err interface{}) {
	Error(c, http.StatusGatewayTimeout, defaultMsg(message, "Upstream service did not respond in time"), err)
}

func ValidationError(c *gin.Context, err error) {
	Error(c, http.StatusUnprocessableEntity, "Validation failed", ParseValidationError(err))
}

func ParseValidationError(err error) interface{} {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make([]FieldError, 0, len(ve))
		for _, fe := range ve {
			out = append(out, FieldError{
				Field:   fe.Field(),
				Tag:     fe.Tag(),
				Param:   fe.Param(),
				Message: validationMessage(fe),
			})
		}
		return out
	}
	return err.Error()
}

func humanizeField(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		r := []rune(p)
		r[0] = unicode.ToUpper(r[0])
		parts[i] = string(r)
	}
	return strings.Join(parts, " ")
}

func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			b.WriteRune('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

func validationMessage(fe validator.FieldError) string {
	field := humanizeField(fe.Field())
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, fe.Param())
	case "numeric":
		return fmt.Sprintf("%s must be a number", field)
	case "alphanum":
		return fmt.Sprintf("%s may only contain letters and numbers", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fe.Param())
	case "eqfield":
		return fmt.Sprintf("%s does not match %s", field, humanizeField(fe.Param()))
	case "nefield":
		return fmt.Sprintf("%s must be different from %s", field, humanizeField(fe.Param()))
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "strong_password":
		return fmt.Sprintf("%s must contain at least 1 uppercase, 1 lowercase, 1 digit, and 1 special character", field)
	default:
		return fmt.Sprintf("%s is invalid (%s)", field, fe.Tag())
	}
}

func HandleBindError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		ValidationError(c, err)
		return
	}
	BadRequest(c, "Malformed request body", err.Error())
}
