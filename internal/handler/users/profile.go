package users

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/external/thirdparty/s3store"
	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/domain/entities"
	"github.com/snappy-fix-golang/internal/domain/enums"
	profileservices "github.com/snappy-fix-golang/internal/services/user/profile_services"
	"github.com/snappy-fix-golang/pkg/activity"
	jwtpkg "github.com/snappy-fix-golang/pkg/jwt"
	logutil "github.com/snappy-fix-golang/pkg/logger"
	"github.com/snappy-fix-golang/pkg/storage/s3images"
	"github.com/snappy-fix-golang/pkg/utils/responses"
)

type Controller struct {
	Db        *db.Database
	Validator *validator.Validate
	Logger    *logutil.Logger
	ExtReq    request.ExternalRequest
}

// @Summary      Get a user profile
// @Description  Get a user profile by ID
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Insert your access token"
// @Success      200            {object}  map[string]interface{}	"User profile retrieved successfully"
// @Failure      400            {object}  map[string]interface{}	"Bad Request"
// @Failure      404            {object}  map[string]interface{}	"User profile not found"
// @Router       /profile  [get]
// @securityDefinitions.apikey  ApiKeyAuth
func (base *Controller) GetProfile(c *gin.Context) {

	claims, exists := c.Get("userClaims")
	if !exists {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "unable to get user claims", "Bad Request", nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	userClaims := claims.(jwt.MapClaims)
	userID := userClaims["user_id"].(string)

	data, status, err := profileservices.GetProfileService(userID, base.Db.Postgresql.DB())
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusNotFound, "error", "failed to fetch profile", err, nil)
		c.JSON(http.StatusNotFound, rd)
		return
	}

	base.Logger.Info("Profile retrieved successfully.")
	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, status)
	activity.UserActivityLog(base.Db.Postgresql, uuid.Must(uuid.FromString(userID)), enums.User, "User Profile", "Profile User Profile")
	c.JSON(http.StatusOK, rd)

}

// UpdateProfile godoc
//
//	@Summary		Update user profile
//	@Description	Update the authenticated user's profile (fields provided will be updated; others remain unchanged)
//	@Tags			profile
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer access token"
//	@Param			request			body		entities.User	true	"Partial user profile update"
//	@Success		200				{object}	map[string]interface{}	"User profile updated successfully"
//	@Failure		400				{object}	map[string]interface{}	"Bad Request"
//	@Failure		404				{object}	map[string]interface{}	"User profile not found"
//	@Router			/profile [patch]
//	@securityDefinitions.apikey	ApiKeyAuth
//	@example		Example Request
//	{
//	  "first_name": "John",
//	  "last_name": "Doe",
//	  "email": "example@example.com",
//	  "phone_number": "+2347061234567"
//	}
func (base *Controller) UpdateProfile(c *gin.Context) {
	var req entities.User

	claims, exists := c.Get("userClaims")
	if !exists {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "unable to get user claims", "Bad Request", nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	userClaims := claims.(jwt.MapClaims)
	userID := userClaims["user_id"].(string)

	err := c.ShouldBind(&req)
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Failed to parse request body", err, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	err = base.Validator.Struct(&req)
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Validation failed", responses.ValidationResponse(err, base.Validator), nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	data, status, err := profileservices.UpdateProfileService(userID, req, base.Db.Postgresql.DB())
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusNotFound, "error", "failed to fetch profile", err, nil)
		c.JSON(http.StatusNotFound, rd)
		return
	}

	base.Logger.Info("Profile updated successfully.")
	activity.UserActivityLog(base.Db.Postgresql, uuid.Must(uuid.FromString(userID)), enums.User, "User Profile", "Update User Profile")
	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, status)
	c.JSON(http.StatusOK, rd)

}

// UpdateUserLocation godoc
//
//	@Summary		Update user location
//	@Description	Update the authenticated user's location/address data
//	@Tags			profile
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer access token"
//	@Param			request			body		entities.UpdateUserLocation	true	"Location payload"
//	@Success		200				{object}	map[string]interface{}	"Location updated successfully"
//	@Failure		400				{object}	map[string]interface{}	"Bad Request"
//	@Failure		404				{object}	map[string]interface{}	"User not found"
//	@Failure		500				{object}	map[string]interface{}	"Internal Server Error"
//	@Router			/profile/location [patch]
//	@securityDefinitions.apikey	ApiKeyAuth
//	@example		Example Request
//	{
//	  "address": "51st",
//	  "city": "Uyo",
//	  "state": "Akwa Ibom",
//	  "country": "Nigeria",
//	  "longitude": 8.0,
//	  "latitude": 10.8
//	}
func (base *Controller) UpdateUserLocation(c *gin.Context) {
	var req entities.UpdateUserLocation

	claims, exists := c.Get("userClaims")
	if !exists {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "unable to get user claims", "Bad Request", nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	userClaims := claims.(jwt.MapClaims)
	userID := userClaims["user_id"].(string)

	err := c.ShouldBind(&req)
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Failed to parse request body", err, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	err = base.Validator.Struct(&req)
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Validation failed", responses.ValidationResponse(err, base.Validator), nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	data, status, err := profileservices.UpdateUserLocationService(userID, req, base.Db.Postgresql.DB())
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusNotFound, "error", "failed to update location", err, nil)
		c.JSON(http.StatusNotFound, rd)
		return
	}

	base.Logger.Info("Location updated successfully.")
	activity.UserActivityLog(base.Db.Postgresql, uuid.Must(uuid.FromString(userID)), enums.User, "User Location", "Update User Location")
	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, status)
	c.JSON(http.StatusOK, rd)

}

// ChangePassword godoc
//
//	@Summary		Change user password
//	@Description	Change the authenticated user's password
//	@Tags			profile
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer access token"
//	@Param			request			body		entities.ChangePasswordRequest	true	"Change Password Request"
//	@Success		200				{object}	map[string]interface{}	"Password changed successfully"
//	@Failure		400				{object}	map[string]interface{}	"Bad Request"
//	@Failure		422				{object}	map[string]interface{}	"Validation Failed"
//	@Router			/profile/change-password [patch]
//	@securityDefinitions.apikey	ApiKeyAuth
//	@example		Change Password Request Example
//	{
//	  "old_password": "oldPass123",
//	  "new_password": "newSecurePass123",
//	  "confirm_password": "newSecurePass123"
//	}
func (base *Controller) ChangePassword(c *gin.Context) {
	var req entities.ChangePasswordRequest

	claims, exists := c.Get("userClaims")
	if !exists {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "unable to get user claims", "Bad Request", nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	userClaims := claims.(jwt.MapClaims)
	userID := userClaims["user_id"].(string)

	err := c.ShouldBind(&req)
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Failed to parse request body", err, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	err = base.Validator.Struct(&req)
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Validation failed", responses.ValidationResponse(err, base.Validator), nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	data, status, err := profileservices.ChangePasswordService(userID, req, base.Db.Postgresql.DB())
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusNotFound, "error", "failed to fetch profile", err, nil)
		c.JSON(http.StatusNotFound, rd)
		return
	}

	base.Logger.Info("Password changed successfully.")
	activity.UserActivityLog(base.Db.Postgresql, uuid.Must(uuid.FromString(userID)), enums.User, "User Password", "Change User Password")
	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, status)
	c.JSON(http.StatusOK, rd)
}

// GetUserImage godoc
//
//	@Summary		Get user image (presigned URL)
//	@Description	Returns a short-lived presigned URL for the authenticated user's image
//	@Tags			profile
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer access token"
//	@Success		200				{object}	map[string]interface{}	"Presigned URL response"
//	@Failure		400				{object}	map[string]interface{}	"Bad Request"
//	@Failure		404				{object}	map[string]interface{}	"User or image not found"
//	@Failure		502				{object}	map[string]interface{}	"Presign failed"
//	@Router			/profile/get-image [get]
//	@securityDefinitions.apikey	ApiKeyAuth
//	@example		Example Success Response
//	{
//	  "message": "presigned",
//	  "data": {
//	    "key": "users/123/avatars/20250826T090102Z-a1b2c3.jpg",
//	    "url": "https://s3.amazonaws.com/...signed-url...",
//	    "ttl_min": "15m0s"
//	  },
//	  "status": 200
//	}
func (base *Controller) GetUserImage(c *gin.Context) {

	claims, _ := c.Get("userClaims")
	userClaims := claims.(jwt.MapClaims)

	UserID, err := jwtpkg.GetUUIDClaim(userClaims, "user_id")

	if err != nil {
		c.JSON(http.StatusBadRequest, responses.BuildErrorResponse(http.StatusBadRequest, "error", "unable to get user id: "+err.Error(), nil, nil))
		return
	}

	data, status, err := profileservices.ProfileGetImageService(UserID, base.Db.Postgresql.DB())
	if err != nil {
		fmt.Println(err)
		rd := responses.BuildErrorResponse(status, "error", "failed to get image key", err, nil)
		c.JSON(status, rd)
		return
	}

	svc := s3images.Service{ExtReq: base.ExtReq}

	ttl := time.Duration(15) * time.Minute

	url, err := svc.Presign(c.Request.Context(), data.ImageKey, ttl)
	if err != nil {
		base.Logger.Error("image presign failed", "err", err.Error())
		rd := responses.BuildErrorResponse(http.StatusBadGateway, "error", "presign failed", err, nil)
		c.JSON(http.StatusBadGateway, rd)
		return
	}
	rd := responses.BuildSuccessResponse(http.StatusOK, "presigned", map[string]any{
		"key": data.ImageKey, "url": url, "ttl_min": ttl,
	})
	activity.UserActivityLog(base.Db.Postgresql, UserID, enums.User, "User Image", "Get User Image")
	c.JSON(http.StatusOK, rd)

}

// UpdateUserImage godoc
//
//	@Summary		Update user image
//	@Description	Upload a new image for the authenticated user. Optionally replaces an old S3 object and returns a presigned URL.
//	@Tags			profile
//	@Accept			mpfd
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer access token"
//	@Param			user_id			formData	string	true	"User ID (UUID)"
//	@Param			old_key			formData	string	false	"Old S3 object key to replace"
//	@Param			prefix			formData	string	false	"S3 folder prefix e.g. users/123/avatars"
//	@Param			presign			formData	boolean	false	"Return presigned GET URL for the new image"
//	@Param			file			formData	file	true	"Image file (.png, .jpg, .webp, .gif)"
//	@Success		200				{object}	map[string]interface{}	"Image has been updated successfully"
//	@Failure		400				{object}	map[string]interface{}	"Bad Request"
//	@Failure		422				{object}	map[string]interface{}	"Validation Failed"
//	@Failure		502				{object}	map[string]interface{}	"S3 error / Presign error"
//	@Router			/profile/update-image [put]
//	@securityDefinitions.apikey	ApiKeyAuth
//
//	# cURL examples:
//
//	## Replace an old key with a new image
//	curl -X PUT "http://localhost:8019/api/v1/profile/update-image" \
//	  -H "Authorization: Bearer <token>" \
//	  -F "user_id=1c5e3d4a-1234-5678-9abc-ffeeddccbbaa" \
//	  -F "file=@/path/to/new-avatar.jpg" \
//	  -F "old_key=users/123/avatars/20250101T120000Z-abc123.jpg" \
//	  -F "prefix=users/123/avatars"
//
//	## Replace and get a presigned link for the new one
//	curl -X PUT "http://localhost:8019/api/v1/profile/update-image" \
//	  -H "Authorization: Bearer <token>" \
//	  -F "user_id=1c5e3d4a-1234-5678-9abc-ffeeddccbbaa" \
//	  -F "file=@/path/to/new-avatar.jpg" \
//	  -F "old_key=users/123/avatars/old.jpg" \
//	  -F "prefix=users/123/avatars" \
//	  -F "presign=true"
func (base *Controller) UpdateUserImage(c *gin.Context) {
	var q entities.UserImageRequest

	if err := c.ShouldBind(&q); err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Failed to parse request body", err, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	userID, err := uuid.FromString(strings.Trim(q.UserID, `[]"`))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "file is required", err, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "failed to read file", err, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	ct := header.Header.Get("Content-Type")
	filename := header.Filename

	imageDetails := s3store.UploadInput{
		Bytes:       buf.Bytes(),
		ContentType: ct,
		Filename:    filename,
		Prefix:      q.Prefix,
		Presign:     q.Presign,
		PresignTTL:  time.Duration(15) * time.Minute,
	}

	svc := s3images.Service{ExtReq: base.ExtReq}
	up, err := svc.UpdateImage(c.Request.Context(), q.OldKey, imageDetails)
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Failed to update image", err, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	data, status, err := profileservices.ProfileUpdateImageService(userID, base.Db.Postgresql.DB(), up)
	if err != nil {
		rd := responses.BuildErrorResponse(status, "error", "failed to update image", err, nil)
		c.JSON(status, rd)
		return
	}

	base.Logger.Info("Image has been updated successfully.")
	activity.UserActivityLog(base.Db.Postgresql, userID, enums.User, "User Image", "Update User Image")
	rd := responses.BuildSuccessResponse(http.StatusOK, "Image has been updated successfully", data, nil)
	c.JSON(http.StatusOK, rd)
}

// DeleteMyAccount godoc
//
//	@Summary		Delete my account
//	@Description	Delete authenticated user's account (email must match session)
//	@Tags			profile
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header	string	true	"Bearer access token"
//	@Param			request			body	entities.DeleteMyAccountRequest	true	"Email confirmation payload"
//	@Success		200	{object}	map[string]interface{}	"Account deleted successfully"
//	@Failure		400	{object}	map[string]interface{}	"Bad Request"
//	@Failure		401	{object}	map[string]interface{}	"Unauthorized"
//	@Failure		404	{object}	map[string]interface{}	"User not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal Server Error"
//	@Router			/profile/delete-account [delete]
//	@securityDefinitions.apikey ApiKeyAuth
func (base *Controller) DeleteMyAccount(c *gin.Context) {

	var req entities.DeleteUserByEmailRequest

	claims, exists := c.Get("userClaims")
	if !exists {
		rd := responses.BuildErrorResponse(
			http.StatusUnauthorized,
			"error",
			"unable to get user claims",
			"Unauthorized",
			nil,
		)
		c.JSON(http.StatusUnauthorized, rd)
		return
	}

	userClaims := claims.(jwt.MapClaims)
	fmt.Println(userClaims)
	sessionUserID := userClaims["user_id"].(string)

	// Bind request
	if err := c.ShouldBindJSON(&req); err != nil {
		rd := responses.BuildErrorResponse(
			http.StatusBadRequest,
			"error",
			"invalid request body",
			err.Error(),
			nil,
		)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	data, status, err := profileservices.DeleteMyAccountService(
		sessionUserID,
		req.Email,
		base.Db.Postgresql.DB(),
	)

	if err != nil {
		rd := responses.BuildErrorResponse(
			status,
			"error",
			"failed to delete account",
			err.Error(),
			nil,
		)
		c.JSON(status, rd)
		return
	}

	base.Logger.Info("User account deleted successfully.")

	activity.UserActivityLog(
		base.Db.Postgresql,
		uuid.Must(uuid.FromString(sessionUserID)),
		enums.User,
		"User Account",
		"Delete Account",
	)

	rd := responses.BuildSuccessResponse(
		http.StatusOK,
		"success",
		data,
		nil,
		status,
	)

	c.JSON(http.StatusOK, rd)
}
