package users

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/snappy-fix-golang/internal/domain/entities"
	usersservices "github.com/snappy-fix-golang/internal/services/user/users_services"
	jwtpkg "github.com/snappy-fix-golang/pkg/jwt"
	"github.com/snappy-fix-golang/pkg/utils/responses"
)

// RegisterHandler godoc
//
//	@Summary		Register a new user
//	@Description	Registers a new user and returns the result
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	entities.RegisterRequest	true	"Register Request"
//	@Success		201		{object}	map[string]interface{}	"User created successfully"
//	@Failure		400		{object}	map[string]interface{}	"Bad Request"
//	@Failure		422		{object}	map[string]interface{}	"Validation Failed"
//	@Router			/auth/register [post]
//	@example			Register Request Example
//	{
//		"first_name": "John",
//		"last_name": "Doe",
//		"email": "example@example.com",
//		"password": "password",
//		"date_of_birth": "2000-01-01",
//		"phone_number": "+2347061234567",
//	}
func (base *Controller) RegisterHandler(c *gin.Context) {
	var req entities.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Failed to parse request body", err, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}
	if err := base.Validator.Struct(&req); err != nil {
		rd := responses.BuildErrorResponse(http.StatusUnprocessableEntity, "error", "Validation failed", responses.ValidationResponse(err, base.Validator), nil)
		c.JSON(http.StatusUnprocessableEntity, rd)
		return
	}

	if _, err := usersservices.ValidateRegisterUserRequest(req, base.Db.Postgresql.DB()); err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", err.Error(), err, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	respData, code, err := usersservices.RegisterUserService(req, base.Db.Postgresql.DB())
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", err.Error(), err, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	base.Logger.Info("user created successfully")
	rd := responses.BuildSuccessResponse(http.StatusCreated, "user created successfully", respData)
	c.JSON(code, rd)
}

// -------- DTO --------
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	// Optional: if client supplies separate JTI (recommended)
	RefreshJTI string `json:"refresh_jti,omitempty"`
}

// RefreshHandler godoc
//
//	@Summary		Refresh user access token
//	@Description	Exchanges a valid refresh token for a new access & refresh pair (rotation).
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	RefreshRequest	true	"Refresh Request"
//	@Success		200		{object}	map[string]interface{}	"token refreshed"
//	@Failure		400		{object}	map[string]interface{}	"Bad Request"
//	@Failure		401		{object}	map[string]interface{}	"Unauthorized"
//	@Router			/auth/refresh-token [post]
//	@example		Refresh Request
//	{
//	  "refresh_token": "c0rx...",
//	  "refresh_jti": "f6b8c7c3-59ee-4d63-9af0-2f7a1e35d1a0"
//	}
func (base *Controller) RefreshHandler(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Failed to parse request body", err, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	resp, code, err := usersservices.RefreshUserTokens(req.RefreshToken, req.RefreshJTI, base.Db.Postgresql.DB())
	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}
	c.JSON(code, responses.BuildSuccessResponse(code, "token refreshed", resp))
}

// LoginHandler godoc
//
//	@Summary		Login a user
//	@Description	Login a user with email OR phone and return access & refresh tokens
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	entities.LoginRequest	true	"Login Request"
//	@Success		200		{object}	map[string]interface{}	"User logged in successfully"
//	@Failure		400		{object}	map[string]interface{}	"Bad Request"
//	@Failure		422		{object}	map[string]interface{}	"Validation Failed"
//	@Router			/auth/login [post]
//	@example		Login with Email
//	{
//	  "email": "example@example.com",
//	  "password": "StrongPass1!",
//	}
func (base *Controller) LoginHandler(c *gin.Context) {
	var req entities.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Failed to parse request body", err, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	// Basic struct validation (password required, email format if provided)
	if err := base.Validator.Struct(&req); err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Validation failed", responses.ValidationResponse(err, base.Validator), nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	// Custom rule: either email OR phone_number must be provided
	if strings.TrimSpace(req.Email) == "" && strings.TrimSpace(req.PhoneNumber) == "" {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Either email or phone_number is required", nil, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	respData, code, err := usersservices.LoginService(req, base.Db.Postgresql.DB())
	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	base.Logger.Info("user login successfully")
	rd := responses.BuildSuccessResponse(http.StatusOK, "user login successfully", respData)
	c.JSON(http.StatusOK, rd)
}

// ForgotPasswordHandler godoc
//
//	@Summary		Reset a user password
//	@Description	Reset a user password and sends a reset link to the user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	entities.ForgotPasswordRequest	true	"Forgot password request"
//	@Success		200		{object}	map[string]interface{}	"Password reset email sent"
//	@Failure		400		{object}	map[string]interface{}	"Bad Request"
//	@Failure		422		{object}	map[string]interface{}	"Validation Failed"
//	@Router			/auth/forgot-password [post]
//	@securityDefinitions.apikey  ApiKeyAuth
//	@example			Forgot Password Request Example
//	{
//		"email": "example@example.com"
//	}
func (base *Controller) ForgotPasswordHandler(c *gin.Context) {
	var req entities.ForgotPasswordRequest

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

	respData, code, err := usersservices.ForgotPasswordService(req, base.Db.Postgresql.DB(), base.ExtReq)
	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	base.Logger.Info("password reset email sent")

	rd := responses.BuildSuccessResponse(http.StatusOK, "Password reset email sent", respData)
	c.JSON(http.StatusOK, rd)
}

// ResetPasswordHandler godoc
//
// @Summary      Reset user password
// @Description  This endpoint allows a user to reset their password using a valid token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body entities.ResetPasswordRequest true "Reset Password Request Example"
// @Success      200 {object} map[string]interface{} "Password has been reset successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      422 {object} map[string]interface{} "Validation Failed"
// @Router       /auth/reset-password [post]
// @securityDefinitions.apikey  ApiKeyAuth
// @example      Reset Password Request Example
//
//	{
//	  "token": "exampletoken123",
//	  "new_password": "newSecurePass123",
//	  "confirm_password": "newSecurePass123"
//	}
func (base *Controller) ResetPasswordHandler(c *gin.Context) {
	var req entities.ResetPasswordRequest

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

	respData, code, err := usersservices.ResetPasswordService(req, base.Db.Postgresql.DB())
	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	base.Logger.Info("password has been reset successfully")

	rd := responses.BuildSuccessResponse(http.StatusOK, "Password has been reset successfully", respData)
	c.JSON(http.StatusOK, rd)
}

// @Summary      Logout User
// @Description  Logout user and remove access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Insert your access token"
// @Success      200            {object}  map[string]interface{}
// @Failure      400            {object}  map[string]interface{}
// @Failure      422            {object}  map[string]interface{}
// @Router       /auth/logout   [post]
// @securityDefinitions.apikey  ApiKeyAuth
// @example      example request
//
//	{
//	  "Authorization": "Bearer ..
//	}
func (base *Controller) LogoutHandler(c *gin.Context) {

	claims, exists := c.Get("userClaims")
	if !exists {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "unable to get user claims", nil, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	userClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "invalid claims type", nil, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	accessUUID, err := jwtpkg.GetUUIDClaim(userClaims, "access_uuid")
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "unable to get access id: "+err.Error(), nil, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	ownerID, err := jwtpkg.GetUUIDClaim(userClaims, "user_id")
	if err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "unable to get user id: "+err.Error(), nil, nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	respData, code, err := usersservices.LogoutUser(accessUUID, ownerID, base.Db.Postgresql.DB())
	if err != nil {
		c.JSON(code, responses.BuildErrorResponse(code, "error", err.Error(), err, nil))
		return
	}

	base.Logger.Info("user logout successfully")

	c.JSON(http.StatusOK, responses.BuildSuccessResponse(http.StatusOK, "user logout successfully", respData))
}

// VerifyEmailHandler godoc
//
//		@Summary		Verify user email
//		@Description	Verify a user's email address
//		@Tags			auth
//		@Accept			json
//		@Produce		json
//		@Param			request	body	entities.VerifyEmailRequest	true	"Verify Email Request"
//		@Success		200		{object}	map[string]interface{}	"email verified"
//		@Failure		400		{object}	map[string]interface{}	"Bad Request"
//		@Failure		422		{object}	map[string]interface{}	"Validation Failed"
//		@Router			/auth/verify-email [post]
//		@example			Verify Email Request Example
//		{
//			"token": "exampletoken123"
//	    "email": "example@example.com"
//		}
func (base *Controller) VerifyEmailHandler(c *gin.Context) {
	var req entities.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.BuildErrorResponse(http.StatusBadRequest, "error", "Failed to parse request body", err, nil))
		return
	}
	resp, code, err := usersservices.VerifyEmailService(req, base.Db.Postgresql.DB(), base.ExtReq)
	if err != nil {
		c.JSON(code, responses.BuildErrorResponse(code, "error", err.Error(), err, nil))
		return
	}
	c.JSON(code, responses.BuildSuccessResponse(code, "email verified", resp))
}

// ResendVerifyEmailHandler godoc
//
//		@Summary		Resend verification email
//		@Description	Resend a verification email to a user
//		@Tags			auth
//		@Accept			json
//		@Produce		json
//		@Param			request	body	entities.ResendVerifyRequest	true	"Resend Verify Email Request"
//		@Success		200		{object}	map[string]interface{}	"ok"
//		@Failure		400		{object}	map[string]interface{}	"Bad Request"
//		@Failure		422		{object}	map[string]interface{}	"Validation Failed"
//		@Router			/auth/resend-verify-email [post]
//		@example			Resend Verify Email Request Example
//		{
//			"email": "example@example.com"
//	    }
func (base *Controller) ResendVerifyEmailHandler(c *gin.Context) {
	var req entities.ResendVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.BuildErrorResponse(http.StatusBadRequest, "error", "Failed to parse request body", err, nil))
		return
	}
	resp, code, err := usersservices.ResendVerifyEmailService(req, base.Db.Postgresql.DB(), base.ExtReq)
	if err != nil {
		c.JSON(code, responses.BuildErrorResponse(code, "error", err.Error(), err, nil))
		return
	}
	c.JSON(code, responses.BuildSuccessResponse(code, "ok", resp))
}
