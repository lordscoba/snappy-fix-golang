package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/internal/adapters/db/postgresql"
	"github.com/snappy-fix-golang/internal/adapters/repository"
	"github.com/snappy-fix-golang/internal/domain/enums"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID               uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	FirstName        string         `gorm:"not null" json:"first_name"`
	LastName         string         `gorm:"not null" json:"last_name"`
	Email            string         `gorm:"uniqueIndex;not null" json:"email"`
	PhoneNumber      string         `gorm:"uniqueIndex;not null" json:"phone_number"`
	Password         *string        `gorm:"" json:"-"`
	Role             enums.UserType `gorm:"type:varchar(20);not null" json:"role"` // "user" | "admin"
	DateOfBirth      *time.Time     `json:"date_of_birth,omitempty"`
	MaritalStatus    *string        `json:"marital_status,omitempty"`
	InsuranceBalance float64        `gorm:"default:0" json:"insurance_balance"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`

	// NOTE: Not polymorphic. Link by OwnerID; filter by owner_type = "PROVIDER_USER".
	// AccessTokens []AccessToken `gorm:"foreignKey:OwnerID;references:ID" json:"-"`
	AccessTokens []AccessToken `gorm:"foreignKey:OwnerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`

	// for image
	ImageKey string `gorm:"index" json:"image_key"`

	// email verification
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"` // null until verified
	IsEmailVerified bool       `gorm:"default:false;not null" json:"is_email_verified"`

	// phone verification
	PhoneVerifiedAt *time.Time `json:"phone_verified_at,omitempty"`
	IsPhoneVerified bool       `gorm:"default:false;not null" json:"is_phone_verified"`

	// for patient location
	Address string `json:"address,omitempty"`
	City    string `json:"city,omitempty"`
	State   string `json:"state,omitempty"`
	Country string `json:"country,omitempty"`

	Status enums.UserStatus `json:"status,omitempty"`

	// check activity status
	LastActiveAt  *time.Time `json:"last_active_at,omitempty"`
	LastIP        *string    `json:"last_ip,omitempty"`
	LastUserAgent *string    `json:"last_user_agent,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	// Normalize email & role early
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))

	if strings.TrimSpace(string(u.Role)) == "" {
		u.Role = enums.User
	}
	u.Role = enums.UserType(strings.ToUpper(strings.TrimSpace(string(u.Role))))

	// If password is nil or blank, SKIP hashing (passwordless/OTP account)
	if u.Password == nil || strings.TrimSpace(*u.Password) == "" {
		u.Password = nil
		return nil
	}

	// Hash password normally
	hashed, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(*u.Password)), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashedStr := string(hashed)
	u.Password = &hashedStr

	return nil
}

func (u *User) CheckPassword(plain string) error {
	// If user registered without a password (OTP login only)
	if u.Password == nil || strings.TrimSpace(*u.Password) == "" {
		return errors.New("no password set for this account")
	}
	return bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(plain))
}

func (u *User) GetUserByID(db repository.DatabaseManager, userID string) (User, error) {
	var user User
	if err := db.DB().Where("id = ?", userID).First(&user).Error; err != nil {
		return user, err
	}
	return user, nil
}

func (u *User) UpdateUserByID(db repository.DatabaseManager, updateData map[string]interface{}, userID string) (*User, error) {
	// Perform the update against the User model, not the receiver pointer
	tx := db.DB().Model(&User{}).Where("id = ?", userID).Updates(updateData)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, errors.New("failed to update user")
	}

	// Read back the fresh row
	var out User
	if err := db.DB().Where("id = ?", userID).First(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (u *User) GetUserByEmail(db repository.DatabaseManager, userEmail string) (User, error) {
	var user User
	if err := db.DB().Where("email = ?", strings.ToLower(strings.TrimSpace(userEmail))).First(&user).Error; err != nil {
		return user, err
	}
	return user, nil
}

func (u *User) UpdateUser(db repository.DatabaseManager) error {
	_, err := db.SaveAllFields(&u)
	return err
}

func (u *User) DeleteAUser(db repository.DatabaseManager) error {
	return db.DeleteRecordFromDb(u)
}

// Safe finder (nil,nil if not found)
func (u *User) FindUserByEmail(db repository.DatabaseManager, email string) (*User, error) {
	var user User
	if err := db.DB().Where("email = ?", strings.ToLower(strings.TrimSpace(email))).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindUserByEmailOrPhone returns (*User, nil) if found, (nil, nil) if not found, or (nil, err) on DB error.
func (u *User) FindUserByEmailOrPhone(db repository.DatabaseManager, email, phone string) (*User, error) {
	var user User
	q := db.DB()

	email = strings.ToLower(strings.TrimSpace(email))
	phone = strings.TrimSpace(phone)

	switch {
	case email != "" && phone != "":
		// If both provided, prefer email match first, then phone.
		if err := q.Where("email = ?", email).First(&user).Error; err == nil {
			return &user, nil
		}
		if err := q.Where("phone_number = ?", phone).First(&user).Error; err == nil {
			return &user, nil
		}
		return nil, nil
	case email != "":
		if err := q.Where("email = ?", email).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			return nil, err
		}
		return &user, nil
	case phone != "":
		if err := q.Where("phone_number = ?", phone).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			return nil, err
		}
		return &user, nil
	default:
		return nil, nil
	}
}

func (pr *PasswordReset) DeletePasswordResetsByEmail(db repository.DatabaseManager, email string) error {
	return db.DB().Where("email = ?", strings.ToLower(strings.TrimSpace(email))).Delete(&PasswordReset{}).Error
}

func (u *User) GetAllUsersFiltered(db repository.DatabaseManager, c *gin.Context) ([]User, repository.PaginationResponse, error) {
	var users []User
	pagination := postgresql.GetPagination(c)

	role := strings.ToUpper(strings.TrimSpace(c.Query("role")))
	status := strings.ToLower(strings.TrimSpace(c.Query("status"))) // "active" | "inactive" -> is_email_verified
	search := strings.TrimSpace(c.Query("search"))
	dateFrom := strings.TrimSpace(c.Query("date_from")) // YYYY-MM-DD
	dateTo := strings.TrimSpace(c.Query("date_to"))     // YYYY-MM-DD

	sort := strings.TrimSpace(c.Query("sort"))
	order := strings.ToLower(strings.TrimSpace(c.Query("order")))
	if sort == "" {
		sort = "created_at"
	}
	if order != "asc" {
		order = "desc"
	}

	// Build WHERE dynamically
	where := "1=1"
	args := []interface{}{}

	if role != "" && (role == "USER" || role == "ADMIN") {
		where += " AND role = ?"
		args = append(args, role)
	}
	if status != "" {
		switch status {
		case "active":
			where += " AND is_email_verified = ?"
			args = append(args, true)
		case "inactive":
			where += " AND (is_email_verified = ? OR is_email_verified = false)"
			args = append(args, false)
		}
	}
	if search != "" {
		where += " AND (LOWER(first_name) ILIKE ? OR LOWER(last_name) ILIKE ? OR LOWER(email) ILIKE ? OR phone_number ILIKE ?)"
		q := "%" + strings.ToLower(search) + "%"
		args = append(args, q, q, q, "%"+search+"%")
	}
	// Date range on created_at
	if dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			where += " AND created_at >= ?"
			args = append(args, t.UTC())
		}
	}
	if dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			// include the full day
			t2 := t.Add(24 * time.Hour).UTC()
			where += " AND created_at < ?"
			args = append(args, t2)
		}
	}
	statusEnum := strings.ToUpper(strings.TrimSpace(c.Query("status_enum"))) // ACTIVE|SUSPENDED|PENDING|INACTIVE
	if statusEnum != "" {
		where += " AND status = ?"
		args = append(args, statusEnum)
	}

	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
		sort,
		order,
		"",
		pagination,
		&users,
		where,
		args...,
	)
	return users, paginationResponse, err
}

type UserImageRequest struct {
	UserID  string `form:"user_id"`
	OldKey  string `form:"old_key"`
	Prefix  string `form:"prefix"`
	Presign bool   `form:"presign"`
}

type UpdateUserLocation struct {
	Address   string  `json:"address,omitempty"`
	City      string  `json:"city,omitempty"`
	State     string  `json:"state,omitempty"`
	Country   string  `json:"country,omitempty"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type AssignRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=USER ADMIN"`
}

type AdminUpdateUserRequest struct {
	FirstName   *string    `json:"first_name"`
	LastName    *string    `json:"last_name"`
	Email       *string    `json:"email"`
	PhoneNumber *string    `json:"phone_number"`
	Password    *string    `json:"password"`
	DateOfBirth *time.Time `json:"date_of_birth"`
}

type UpdateStatusRequest struct {
	// Optional toggles; timestamps will auto-populate if set to true (or clear if false)
	IsEmailVerified *bool `json:"is_email_verified,omitempty" validate:"omitempty"`
	IsPhoneVerified *bool `json:"is_phone_verified,omitempty" validate:"omitempty"`

	// Optional explicit timestamps (UTC); if provided alone, booleans are assumed true
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty" validate:"omitempty"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at,omitempty" validate:"omitempty"`

	// Optional overall status enum (separate from is_*_verified)
	Status *enums.UserStatus `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE SUSPENDED PENDING INACTIVE"`
}

type VerifyNINRequest struct {
	NINID               string `json:"nin_id" binding:"required"`
	VerificationConsent bool   `json:"verification_consent" binding:"required"`
	Selfie              string `json:"selfie,omitempty"` // optional: data:image/jpeg;base64,...
}

type VerifyBVNRequest struct {
	BVNID               string `json:"bvn_id" binding:"required"`
	VerificationConsent bool   `json:"verification_consent" binding:"required"`
	Selfie              string `json:"selfie,omitempty"` // optional: data:image/jpeg;base64,...
}

type DeleteUserByEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}
