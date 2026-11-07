package dtos

import (
	"encoding/json"
	"time"
)

type UserInfoDto struct {
	UserID              string     `json:"userID"`
	Pass                string     `json:"pass,omitempty"`
	UserName            string     `json:"userName"`
	Description         string     `json:"description"`
	Area                string     `json:"area"`
	Role                string     `json:"role"`
	IsActive            *bool      `json:"isActive"`
	FailedLoginAttempts *int       `json:"failedLoginAttempts,omitempty"`
	LockoutEnd          *time.Time `json:"lockoutEnd,omitempty"`
	Roles               []string   `json:"roles"`
	Permissions         []string   `json:"permissions"`
}

func (u UserInfoDto) MarshalJSON() ([]byte, error) {
	type Alias UserInfoDto
	return json.Marshal(&struct {
		Alias
		UserIDAlias string `json:"userId"`
	}{
		Alias:       Alias(u),
		UserIDAlias: u.UserID,
	})
}

type PermissionDto struct {
	UserID                      string `json:"userID"`
	AssetsModify                *bool  `json:"assetsModify"`
	AssetsDel                   *bool  `json:"assetsDel"`
	DepartmentModify            *bool  `json:"departmentModify"`
	DepartmentDel               *bool  `json:"departmentDel"`
	TypeModify                  *bool  `json:"typeModify"`
	TypeDel                     *bool  `json:"typeDel"`
	UserManager                 *bool  `json:"userManager"`
	Disposal                    *bool  `json:"disposal"`
	PermissionApproval          *bool  `json:"permissionApproval"`
	AllowSeeAllFromAllowedAreas *bool  `gorm:"column:AllowSeeAllFromAllowedAreas" json:"allowSeeAllFromAllowedAreas"`
}

func (p PermissionDto) MarshalJSON() ([]byte, error) {
	type Alias PermissionDto
	return json.Marshal(&struct {
		Alias
		UserIDAlias string `json:"userId"`
	}{
		Alias:       Alias(p),
		UserIDAlias: p.UserID,
	})
}

type HistoryDto struct {
	ID         int64      `json:"id"`
	UserID     string     `json:"userId"`
	Action     string     `json:"action"`
	DateTimeAC *time.Time `json:"dateTimeAC"`
}

type UserDto struct {
	UserId   string `json:"userId"`
	UserName string `json:"userName"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}
