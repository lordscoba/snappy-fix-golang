package usersservices

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/internal/domain/entities"
	"github.com/snappy-fix-golang/internal/domain/enums"
	"github.com/snappy-fix-golang/internal/inst"
	jwtpkg "github.com/snappy-fix-golang/pkg/jwt"
	"github.com/snappy-fix-golang/pkg/utils"
	validateutil "github.com/snappy-fix-golang/pkg/validate_util"
	"gorm.io/gorm"
)

// ---------- VALIDATION ----------
func ValidateRegisterUserRequest(req entities.RegisterRequest, db *gorm.DB) (entities.RegisterRequest, error) {
	pdb := inst.InitDB(db)
	user := entities.User{}

	// Trim and normalize inputs
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
	if req.Password != nil {
		pw := strings.TrimSpace(*req.Password)
		req.Password = &pw
	}

	// Required + min-length checks
	if len(req.FirstName) < 2 {
		return req, errors.New("first name must be at least 2 characters long")
	}
	if len(req.LastName) < 2 {
		return req, errors.New("last name must be at least 2 characters long")
	}
	if len(req.Email) < 5 {
		return req, errors.New("email must be at least 5 characters long")
	}
	if len(req.PhoneNumber) < 8 {
		return req, errors.New("phone number must be at least 8 digits long")
	}

	// Validate email format
	isValid := validateutil.EmailValid(req.Email)
	if !isValid {
		return req, errors.New("invalid email format")
	}

	// Check if email already exists
	if pdb.CheckExists(&user, "email = ?", req.Email) {
		return req, errors.New("a user with this email already exists")
	}

	// Validate phone number format
	isValidPhone := validateutil.PhoneValid(req.PhoneNumber)
	if !isValidPhone {
		return req, errors.New("invalid phone number format")
	}

	// Check if phone number already exists
	if pdb.CheckExists(&user, "phone_number = ?", req.PhoneNumber) {
		return req, errors.New("a user with this phone number already exists")
	}

	// Password strength validation
	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		if err := validateutil.ValidatePassword(*req.Password); err != nil {
			return req, err
		}
	}

	return req, nil
}

// services/users/register.go
func RegisterUserService(req entities.RegisterRequest, db *gorm.DB) (gin.H, int, error) {
	pdb := inst.InitDB(db)

	email := strings.ToLower(strings.TrimSpace(req.Email))
	firstName := validateutil.ToTitleCase(strings.TrimSpace(req.FirstName))
	lastName := validateutil.ToTitleCase(strings.TrimSpace(req.LastName))
	phoneNumber := strings.TrimSpace(req.PhoneNumber)

	// Parse DOB if provided
	var dob *time.Time
	if s := strings.TrimSpace(req.DateOfBirth); s != "" {
		dt, err := time.Parse("2006-01-02", s)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		dd := dt.UTC()
		dob = &dd
	}

	// Normalize password pointer (allow nil)
	var passwordPtr *string
	if req.Password != nil {
		ps := strings.TrimSpace(*req.Password)
		if ps != "" {
			passwordPtr = &ps
		}
	}

	user := &entities.User{
		ID:          uuid.Must(uuid.NewV4()),
		FirstName:   firstName,
		LastName:    lastName,
		Email:       email,
		PhoneNumber: phoneNumber,
		Password:    passwordPtr, // may be nil; hashing hook must handle nil
		Role:        enums.User,
		DateOfBirth: dob,
	}

	if err := user.CreateUser(pdb); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	response := gin.H{
		"user": gin.H{
			"id":                user.ID.String(),
			"email":             user.Email,
			"first_name":        user.FirstName,
			"last_name":         user.LastName,
			"phone":             user.PhoneNumber,
			"role":              user.Role,
			"date_of_birth":     user.DateOfBirth,
			"created_at":        user.CreatedAt,
			"updated_at":        user.UpdatedAt,
			"is_email_verified": user.IsEmailVerified,
			"is_phone_verified": user.IsPhoneVerified,
		},
	}

	return response, http.StatusCreated, nil
}

func VerifyEmailService(req entities.VerifyEmailRequest, db *gorm.DB, extReq request.ExternalRequest) (gin.H, int, error) {
	pdb := inst.InitDB(db)

	token := strings.TrimSpace(req.Token)
	if token == "" {
		return nil, http.StatusBadRequest, errors.New("invalid token")
	}

	// 1) lookup token
	var ver entities.EmailVerification
	found, err := ver.GetByToken(pdb, token)
	if err != nil {
		return nil, http.StatusBadRequest, errors.New("invalid or expired token")
	}
	// 2) load user by email
	var u entities.User
	user, err := u.GetUserByEmail(pdb, found.Email)
	if err != nil {
		_ = (&found).MarkConsumed(pdb) // consume to prevent reuse
		return nil, http.StatusBadRequest, errors.New("invalid or expired token")
	}

	// 3) mark verified
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"is_email_verified": true,
		"email_verified_at": &now,
	}
	if _, err := (&u).UpdateUserByID(pdb, updates, user.ID.String()); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// 4) consume token
	err = (&found).MarkConsumed(pdb)

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// htmlBody, err := rendertemplate.RenderHTMLTemplate(rendertemplate.RenderHTMLStruct{
	// 	TemplateDir:  constants.TemplateDir,
	// 	TemplateName: "welcome_email",
	// 	Data: map[string]any{
	// 		"email":    user.Email,
	// 		"username": user.FirstName + " " + user.LastName,
	// 	},
	// })

	// if err != nil {
	// 	return nil, http.StatusInternalServerError, err
	// }

	// reqMailgun := mailgun.SendEmailRequest{
	// 	From:    "noreply@" + config.GetConfig().Mailgun.Domain,
	// 	To:      []string{user.Email},
	// 	Subject: "Welcome!",
	// 	Text:    "textBody",
	// 	HTML:    htmlBody,
	// }

	// respAny, err := extReq.SendExternalRequestMailgun(external.MailgunSend, reqMailgun)
	// if err != nil {
	// 	return gin.H{"message": " there was an error sending the welcome email"}, http.StatusOK, fmt.Errorf("email: send mailgun failed: %w", err)
	// }

	// typed, ok := respAny.(mailgun.SendEmailResponse)
	// if !ok {
	// 	return gin.H{"message": " there was an error sending the welcome email."}, http.StatusOK, fmt.Errorf("email: unexpected response type")
	// }

	// fmt.Println("typed", typed)

	return gin.H{
		"message": "Email verified successfully.",
	}, http.StatusOK, nil
}

func ResendVerifyEmailService(req entities.ResendVerifyRequest, db *gorm.DB, extReq request.ExternalRequest) (gin.H, int, error) {
	pdb := inst.InitDB(db)

	// frontendUrl := config.GetConfig().App.Url + "/verify-email"

	// silently OK even if user not found (avoid enumeration)
	user, _ := new(entities.User).FindUserByEmail(pdb, req.Email)
	if user == nil || user.IsEmailVerified {
		return gin.H{"message": "If an account exists, a verification email will arrive shortly."}, http.StatusOK, nil
	}

	token, err := utils.GenerateSecureToken(108)
	if err != nil {
		return gin.H{"message": "If an account exists, a verification email will arrive shortly."}, http.StatusOK, nil
	}

	ver := &entities.EmailVerification{
		ID:        uuid.Must(uuid.NewV4()),
		Email:     user.Email,
		Token:     token,
		ExpiresAt: time.Now().Add(45 * time.Minute).UTC(),
	}
	err = ver.Create(pdb)

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// verifyURL := url.BuildVerifyEmailURL(frontendUrl, user.Email, token)

	// htmlBody, err := rendertemplate.RenderHTMLTemplate(rendertemplate.RenderHTMLStruct{
	// 	TemplateDir:  constants.TemplateDir,
	// 	TemplateName: "verify_email",
	// 	Data: map[string]any{
	// 		"verifyURL": verifyURL,
	// 		"email":     user.Email,
	// 		"username":  user.FirstName + " " + user.LastName,
	// 	},
	// })

	// if err != nil {
	// 	return nil, http.StatusInternalServerError, err
	// }

	// reqMailgun := mailgun.SendEmailRequest{
	// 	From:    "noreply@" + config.GetConfig().Mailgun.Domain,
	// 	To:      []string{user.Email},
	// 	Subject: "Email Verification",
	// 	Text:    "textBody",
	// 	HTML:    htmlBody,
	// }

	// respAny, err := extReq.SendExternalRequestMailgun(external.MailgunSend, reqMailgun)
	// if err != nil {
	// 	return gin.H{"message": " there was an error sending the verification email"}, http.StatusOK, fmt.Errorf("email: send mailgun failed: %w", err)
	// }

	// typed, ok := respAny.(mailgun.SendEmailResponse)
	// if !ok {
	// 	return gin.H{"message": " there was an error sending the verification email."}, http.StatusOK, fmt.Errorf("email: unexpected response type")
	// }

	// fmt.Println("typed", typed)

	return gin.H{"message": "If an account exists, a verification email will arrive shortly."}, http.StatusOK, nil
}

func LoginService(req entities.LoginRequest, db *gorm.DB) (gin.H, int, error) {
	pdb := inst.InitDB(db)

	email := strings.ToLower(strings.TrimSpace(req.Email))
	phone := strings.TrimSpace(req.PhoneNumber)
	deviceToken := strings.TrimSpace(req.DeviceToken)
	deviceType := strings.TrimSpace(req.DeviceType)

	// 1️⃣ Fetch user by email OR phone
	u, err := new(entities.User).FindUserByEmailOrPhone(pdb, email, phone)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("unable to fetch user: %v", err)
	}
	if u == nil {
		return nil, http.StatusBadRequest, errors.New("user not found")
	}

	// 2️⃣ Verify password
	if err := u.CheckPassword(req.Password); err != nil {
		return nil, http.StatusBadRequest, errors.New("invalid password")
	}

	if u.Status == enums.UserStatusSuspended || u.Status == enums.UserStatusInactive {
		return nil, http.StatusForbidden, errors.New("account is suspended or inactive")
	}

	// 5️⃣ Create JWT token pair
	tp, err := jwtpkg.CreateTokenPair(*u)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to create tokens: %v", err)
	}

	// 6️⃣ Save session (Access + Refresh)
	tokens := map[string]string{
		"email":            u.Email,
		"access_token":     tp.AccessToken,
		"exp":              strconv.Itoa(int(tp.AccessExpiresAt.Unix())),
		"access_exp_unix":  strconv.Itoa(int(tp.AccessExpiresAt.Unix())),
		"refresh_token":    tp.RefreshToken,
		"refresh_exp_unix": strconv.Itoa(int(tp.RefreshExpiresAt.Unix())),
		"refresh_jti":      tp.RefreshJTI.String(),
		"device_token":     deviceToken,
		"device_type":      deviceType,
	}
	session := &entities.AccessToken{
		Email:   u.Email,
		ID:      tp.AccessUuid,
		OwnerID: u.ID,
	}
	if err := session.CreateAccessToken(pdb, tokens); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("error saving token: %v", err)
	}

	// 7️⃣ Final response with all requested flags
	resp := gin.H{
		"user": gin.H{
			"id":         u.ID.String(),
			"email":      u.Email,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
			"phone":      u.PhoneNumber,
			"role":       u.Role,
			"created_at": u.CreatedAt,
			"updated_at": u.UpdatedAt,
		},
		"access_token":             tp.AccessToken,
		"access_token_expires_in":  tp.AccessExpiresAt.Unix(),
		"refresh_token":            tp.RefreshToken,
		"refresh_jti":              tp.RefreshJTI.String(),
		"refresh_token_expires_in": tp.RefreshExpiresAt.Unix(),
		"token_type":               "Bearer",
		"device_token":             deviceToken,
		"device_type":              deviceType,
	}

	return resp, http.StatusOK, nil
}
func parseJTI(rawToken string, jtiOverride string) (uuid.UUID, string, error) {
	if jtiOverride != "" {
		j, err := uuid.FromString(jtiOverride)
		if err != nil {
			return uuid.Nil, "", fmt.Errorf("invalid refresh_jti")
		}
		return j, rawToken, nil
	}
	// Optional support for "JTI.raw" format
	if i := strings.IndexByte(rawToken, '.'); i > 0 {
		jtiStr := rawToken[:i]
		secret := rawToken[i+1:]
		j, err := uuid.FromString(jtiStr)
		if err == nil && secret != "" {
			return j, secret, nil
		}
	}
	// If no JTI given, cannot query efficiently
	return uuid.Nil, "", errors.New("refresh_jti required or use 'jti.raw' format")
}

func RefreshUserTokens(rawRefresh, refreshJTI string, db *gorm.DB) (gin.H, int, error) {
	pdb := inst.InitDB(db)

	jti, secret, err := parseJTI(rawRefresh, refreshJTI)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	// Find live session by JTI
	var session entities.AccessToken
	if _, err := session.GetByRefreshJTI(pdb, jti); err != nil {
		return nil, http.StatusUnauthorized, errors.New("invalid refresh session")
	}
	if !session.IsLive || time.Now().After(session.RefreshTokenExpiresAt) {
		return nil, http.StatusUnauthorized, errors.New("refresh expired")
	}
	// Compare secrets
	if err := bcrypt.CompareHashAndPassword([]byte(session.RefreshTokenHash), []byte(secret)); err != nil {
		return nil, http.StatusUnauthorized, errors.New("invalid refresh token")
	}

	// Load user
	var user entities.User
	user, err = user.GetUserByID(pdb, session.OwnerID.String())
	if err != nil {
		return nil, http.StatusUnauthorized, errors.New("user not found")
	}

	if user.Status == enums.UserStatusSuspended || user.Status == enums.UserStatusInactive {
		return nil, http.StatusForbidden, errors.New("account is suspended or inactive")
	}

	// Mint new pair and rotate
	tp, err := jwtpkg.CreateTokenPair(user)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to create tokens: %v", err)
	}

	// Revoke old session
	if err := session.RevokeAccessToken(pdb); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to revoke old session: %v", err)
	}

	// Persist new session
	tokens := map[string]string{
		"access_token":     tp.AccessToken,
		"exp":              strconv.Itoa(int(tp.AccessExpiresAt.Unix())),
		"access_exp_unix":  strconv.Itoa(int(tp.AccessExpiresAt.Unix())),
		"refresh_token":    tp.RefreshToken,
		"refresh_exp_unix": strconv.Itoa(int(tp.RefreshExpiresAt.Unix())),
		"refresh_jti":      tp.RefreshJTI.String(),
	}
	newSession := &entities.AccessToken{
		ID:      tp.AccessUuid,
		OwnerID: user.ID,
	}
	if err := newSession.CreateAccessToken(pdb, tokens); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to save new session: %v", err)
	}

	resp := gin.H{
		"access_token":             tp.AccessToken,
		"access_token_expires_in":  tp.AccessExpiresAt.Unix(),
		"refresh_token":            tp.RefreshToken,
		"refresh_jti":              tp.RefreshJTI.String(),
		"refresh_token_expires_in": tp.RefreshExpiresAt.Unix(),
	}
	return resp, http.StatusOK, nil
}

// ForgotPasswordService requests a reset link. (No account enumeration.)
func ForgotPasswordService(req entities.ForgotPasswordRequest, db *gorm.DB, extReq request.ExternalRequest) (gin.H, int, error) {
	pdb := inst.InitDB(db)

	// frontendUrl := config.GetConfig().App.FrontendUrl + "/provider/reset-password"

	// Normalize input
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if !validateutil.EmailValid(email) {
		// respond generic; don't leak which emails are valid
		return gin.H{"message": "If an account exists, a reset link will be sent shortly."}, http.StatusOK, nil
	}

	user, err := new(entities.User).FindUserByEmail(pdb, email)
	if err != nil {

		// log internally if you want, but return generic to avoid enumeration
		return gin.H{"message": "If an account exists, a reset link will be sent shortly."}, http.StatusOK, nil
	}

	if user == nil {
		// not found -> generic OK
		return gin.H{"message": "If an account exists, a reset link will be sent shortly."}, http.StatusOK, nil
	}

	// Upsert a PasswordReset entry (invalidate previous)
	// _ = pdb.DB().Where("email = ?", email).Delete(&entities.PasswordReset{}).Error
	// Upsert a PasswordReset entry (invalidate previous)
	// Upsert a PasswordReset entry (invalidate previous)
	err = new(entities.PasswordReset).DeletePasswordResetsByEmail(pdb, email)

	if err != nil {
		// log internally if you want, but return generic to avoid enumeration
		return gin.H{"message": "If an account exists, a reset link will be sent shortly."}, http.StatusOK, nil
	}

	token, err := utils.GenerateSecureToken(108) // url-safe, 48 bytes ~ 64 chars
	if err != nil {
		// Still return generic to user; but log/return internal error for API caller
		return gin.H{"message": "If an account exists, a reset link will be sent shortly."}, http.StatusOK, fmt.Errorf("failed generating reset token: %w", err)
	}

	pr := &entities.PasswordReset{
		ID:        uuid.Must(uuid.NewV4()),
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	if err := pr.CreatePasswordReset(pdb); err != nil {
		return gin.H{"message": "If an account exists, a reset link will be sent shortly."}, http.StatusOK, fmt.Errorf("failed saving reset token: %w", err)
	}

	// resetURL := url.BuildPasswordResetURL(frontendUrl, email, token)

	// htmlBody, err := rendertemplate.RenderHTMLTemplate(rendertemplate.RenderHTMLStruct{
	// 	TemplateDir:  constants.TemplateDir,
	// 	TemplateName: "reset_password",
	// 	Data: map[string]any{
	// 		"resetURL": resetURL,
	// 		"email":    email,
	// 		"username": user.FirstName + " " + user.LastName,
	// 	},
	// })

	// if err != nil {
	// 	return gin.H{"message": "If an account exists, a reset link will be sent shortly."}, http.StatusOK, fmt.Errorf("failed rendering reset email: %w", err)
	// }

	// reqMailgun := mailgun.SendEmailRequest{
	// 	From:    "noreply@" + config.GetConfig().Mailgun.Domain,
	// 	To:      []string{email},
	// 	Subject: "Password Reset",
	// 	Text:    "textBody",
	// 	HTML:    htmlBody,
	// }

	// respAny, err := extReq.SendExternalRequestMailgun(external.MailgunSend, reqMailgun)
	// if err != nil {
	// 	return gin.H{"message": "If an account exists, a reset link will be sent shortly."}, http.StatusOK, fmt.Errorf("email: send mailgun failed: %w", err)
	// }

	// typed, ok := respAny.(mailgun.SendEmailResponse)
	// if !ok {
	// 	return gin.H{"message": "If an account exists, a reset link will be sent shortly."}, http.StatusOK, fmt.Errorf("email: unexpected response type")
	// }

	// fmt.Println("typed", typed)

	return gin.H{"message": "If an account exists, a reset link will be sent shortly."}, http.StatusOK, nil
}

// ResetPasswordService consumes the token and sets a new password.
func ResetPasswordService(req entities.ResetPasswordRequest, db *gorm.DB) (gin.H, int, error) {
	pdb := inst.InitDB(db)

	token := strings.TrimSpace(req.Token)
	newPassword := strings.TrimSpace(req.NewPassword)
	confirmPassword := strings.TrimSpace(req.ConfirmPassword)

	if token == "" {
		return nil, http.StatusBadRequest, errors.New("invalid token")
	}
	if newPassword != confirmPassword {
		return nil, http.StatusBadRequest, errors.New("passwords do not match")
	}
	if err := validateutil.ValidatePassword(newPassword); err != nil {
		return nil, http.StatusBadRequest, err
	}

	// 1) Lookup token
	var pr entities.PasswordReset
	found, err := pr.GetPasswordResetByToken(pdb, token)
	if err != nil {
		return nil, http.StatusBadRequest, errors.New("invalid or expired token")
	}

	// 2) Load user by email
	var u entities.User
	userData, err := u.GetUserByEmail(pdb, found.Email)
	if err != nil {
		tmp := found
		_ = (&tmp).DeletePasswordReset(pdb)
		return nil, http.StatusBadRequest, errors.New("invalid or expired token")
	}

	// 3) Hash and set new password (as *string)
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to hash password: %w", err)
	}
	hashedStr := string(hashed)
	userData.Password = &hashedStr

	// 4) Persist
	if err := (&userData).UpdateUser(pdb); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to update password: %w", err)
	}

	// 5) Invalidate token
	tmp := found
	_ = (&tmp).DeletePasswordReset(pdb)

	// (Optional) Invalidate existing sessions
	// _ = new(entities.AccessToken).DeleteByOwnerID(pdb, userData.ID.String())

	return gin.H{"message": "Password has been reset successfully. You can now log in."}, http.StatusOK, nil
}

func LogoutUser(access_uuid, owner_id uuid.UUID, db *gorm.DB) (gin.H, int, error) {

	// instance of Postgresql db
	pdb := inst.InitDB(db)
	var (
		responseData gin.H
	)

	access_token := entities.AccessToken{ID: access_uuid, OwnerID: owner_id}

	// revoke user access_token to invalidate session
	err := access_token.RevokeAccessToken(pdb)

	if err != nil {
		return responseData, http.StatusInternalServerError, fmt.Errorf("failed to revoke access token: %v", err)
	}

	responseData = gin.H{}

	return responseData, http.StatusOK, nil
}
