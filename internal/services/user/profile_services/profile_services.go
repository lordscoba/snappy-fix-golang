package profileservices

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/external/thirdparty/s3store"
	"github.com/snappy-fix-golang/internal/domain/entities"
	"github.com/snappy-fix-golang/internal/inst"
	validateutil "github.com/snappy-fix-golang/pkg/validate_util"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func ProfileGetImageService(userID uuid.UUID, db *gorm.DB) (entities.User, int, error) {
	pdb := inst.InitDB(db)

	var mh entities.User

	existing, err := mh.GetUserByID(pdb, userID.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entities.User{}, http.StatusNotFound, errors.New("user not found")
		}
		return entities.User{}, http.StatusBadRequest, err
	}

	return existing, http.StatusOK, nil

}

func ProfileUpdateImageService(userID uuid.UUID, db *gorm.DB, up s3store.UploadResponse) (gin.H, int, error) {
	pdb := inst.InitDB(db)

	var mh entities.User

	existing, err := mh.GetUserByID(pdb, userID.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gin.H{}, http.StatusNotFound, errors.New("user not found")
		}
		return gin.H{}, http.StatusBadRequest, err
	}

	updates := map[string]interface{}{}
	if up.Key != "" {
		updates["image_key"] = up.Key
	}

	updated, err := existing.UpdateUserByID(pdb, updates, userID.String())

	if err != nil {
		return gin.H{}, http.StatusInternalServerError, err
	}

	resp := gin.H{
		"user": updated,
		"upload": gin.H{
			"key":          up.Key,
			"url":          up.URL,
			"etag":         up.ETag,
			"size":         up.Size,
			"content_type": up.ContentType,
		},
	}

	return resp, http.StatusOK, nil

}

func GetProfileService(id string, db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)
	var user entities.User
	user, err := user.GetUserByID(pdb, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gin.H{}, http.StatusNotFound, errors.New("user not found")
		}
		return gin.H{}, http.StatusBadRequest, err
	}

	return gin.H{"user": user}, http.StatusOK, nil
}

func UpdateProfileService(id string, req entities.User, db *gorm.DB) (entities.User, int, error) {

	pdb := inst.InitDB(db)
	user := entities.User{}

	// Fetch existing record (you'll need to pass the user ID in a real request)
	existing, err := user.GetUserByID(pdb, id)
	if err != nil {
		return entities.User{}, http.StatusNotFound, err
	}

	// Update logic — replace with actual fields from the request
	if req.FirstName != "" {
		existing.FirstName = req.FirstName
	}
	if req.LastName != "" {
		existing.LastName = req.LastName
	}
	if req.Email != "" {
		existing.Email = req.Email
	}
	if req.PhoneNumber != "" {
		existing.PhoneNumber = req.PhoneNumber

	}

	err = existing.UpdateUser(pdb)
	if err != nil {
		return entities.User{}, http.StatusInternalServerError, err
	}

	return existing, http.StatusOK, nil
}

func UpdateUserLocationService(id string, req entities.UpdateUserLocation, db *gorm.DB) (entities.User, int, error) {

	pdb := inst.InitDB(db)
	user := entities.User{}

	// Fetch existing record (you'll need to pass the user ID in a real request)
	existing, err := user.GetUserByID(pdb, id)
	if err != nil {
		return entities.User{}, http.StatusNotFound, err
	}

	// Update logic — replace with actual fields from the request
	if req.Address != "" {
		existing.Address = req.Address
	}
	if req.City != "" {
		existing.City = req.City
	}
	if req.State != "" {
		existing.State = req.State
	}
	if req.Country != "" {
		existing.Country = req.Country
	}

	fmt.Println(existing)

	err = existing.UpdateUser(pdb)
	if err != nil {
		return entities.User{}, http.StatusInternalServerError, err
	}
	return existing, http.StatusOK, nil
}

func ChangePasswordService(id string, req entities.ChangePasswordRequest, db *gorm.DB) (gin.H, int, error) {
	pdb := inst.InitDB(db)

	oldPwd := strings.TrimSpace(req.OldPassword)
	newPwd := strings.TrimSpace(req.NewPassword)
	confirm := strings.TrimSpace(req.ConfirmPassword)

	// Validate new/confirm
	if newPwd == "" {
		return nil, http.StatusBadRequest, errors.New("new_password is required")
	}
	if newPwd != confirm {
		return nil, http.StatusBadRequest, errors.New("confirm_password must match new_password")
	}
	if err := validateutil.ValidatePassword(newPwd); err != nil {
		return nil, http.StatusBadRequest, err
	}

	// Fetch user
	var u entities.User
	user, err := u.GetUserByID(pdb, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("user not found")
		}
		return nil, http.StatusBadRequest, err
	}

	// Determine if this is a first-time set (no existing password)
	hasExistingPassword := !(user.Password == nil || strings.TrimSpace(*user.Password) == "")

	if hasExistingPassword {
		// Require and verify old password
		if oldPwd == "" {
			return nil, http.StatusBadRequest, errors.New("old_password is required")
		}
		if err := user.CheckPassword(oldPwd); err != nil {
			return nil, http.StatusBadRequest, errors.New("old password is incorrect")
		}
		// Prevent reusing the same password
		if bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(newPwd)) == nil {
			return nil, http.StatusBadRequest, errors.New("new password must be different from old password")
		}
	} else {
		// First-time set: skip old password check
		// (User is already authenticated to call this endpoint.)
	}

	// Hash and set new password (as *string)
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to hash password: %w", err)
	}
	hashedStr := string(hashed)
	user.Password = &hashedStr

	// Persist update
	if err := (&user).UpdateUser(pdb); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to update password: %w", err)
	}

	// (Optional) Invalidate existing tokens/sessions for safety
	// _ = new(entities.AccessToken).DeleteByOwnerID(pdb, user.ID.String())

	return gin.H{"message": "Password updated successfully."}, http.StatusOK, nil
}

func DeleteMyAccountService(userID string, email string, db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)
	var u entities.User

	existing, err := u.GetUserByID(pdb, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gin.H{}, http.StatusNotFound, errors.New("user not found")
		}
		return gin.H{}, http.StatusBadRequest, err
	}

	// Extra safety: cross-check email from DB
	if strings.ToLower(strings.TrimSpace(existing.Email)) != email {
		return gin.H{}, http.StatusUnauthorized, errors.New("email mismatch")
	}

	tx := db.Begin()

	// ✅ 1. Delete access tokens
	if err := tx.Where("owner_id = ?", existing.ID).
		Delete(&entities.AccessToken{}).Error; err != nil {
		tx.Rollback()
		return gin.H{}, http.StatusInternalServerError, err
	}

	// ✅ 2. Delete user
	if err := tx.Delete(&existing).Error; err != nil {
		tx.Rollback()
		return gin.H{}, http.StatusInternalServerError, err
	}

	tx.Commit()

	return gin.H{
		"message": "Account deleted successfully",
		"email":   email,
	}, http.StatusOK, nil
}
