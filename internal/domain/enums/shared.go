package enums

type UserType string

const (
	User  UserType = "USER"
	Admin UserType = "ADMIN"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
	UserStatusPending   UserStatus = "PENDING"
	UserStatusInactive  UserStatus = "INACTIVE"
)
