package domains

import (
	"encoding/json"
	"time"
)

type USERINFO struct {
	UserID              string     `gorm:"column:UserID;primaryKey;type:varchar(50)" json:"userID"`
	Pass                string     `gorm:"column:Pass;type:varchar(255)" json:"-"`
	UserName            string     `gorm:"column:UserName;type:varchar(100)" json:"userName"`
	Description         string     `gorm:"column:Description;type:varchar(150)" json:"description"`
	Area                string     `gorm:"column:Area;type:varchar(50)" json:"area"`
	Role                string     `gorm:"column:Role;type:varchar(50)" json:"role"`
	IsActive            *bool      `gorm:"column:IsActive" json:"isActive"`
	FailedLoginAttempts *int       `gorm:"column:FailedLoginAttempts" json:"failedLoginAttempts,omitempty"`
	LockoutEnd          *time.Time `gorm:"column:LockoutEnd" json:"lockoutEnd,omitempty"`

	// Populated for DTOs
	Roles       []string `gorm:"-" json:"roles,omitempty"`
	Permissions []string `gorm:"-" json:"permissions,omitempty"`
}

func (u USERINFO) MarshalJSON() ([]byte, error) {
	type Alias USERINFO
	return json.Marshal(&struct {
		Alias
		UserIDAlias string `json:"userId"`
	}{
		Alias:       Alias(u),
		UserIDAlias: u.UserID,
	})
}

func (USERINFO) TableName() string {
	return "USERINFO"
}

type PERMISSION struct {
	UserID                      string `gorm:"column:UserID;primaryKey;type:varchar(50)" json:"userID"`
	AssetsModify                *bool  `gorm:"column:AssetsModify" json:"assetsModify"`
	AssetsDel                   *bool  `gorm:"column:AssetsDel" json:"assetsDel"`
	DepartmentModify            *bool  `gorm:"column:DepartmentModify" json:"departmentModify"`
	DepartmentDel               *bool  `gorm:"column:DepartmentDel" json:"departmentDel"`
	TypeModify                  *bool  `gorm:"column:TypeModify" json:"typeModify"`
	TypeDel                     *bool  `gorm:"column:TypeDel" json:"typeDel"`
	UserManager                 *bool  `gorm:"column:UserManager" json:"userManager"`
	Disposal                    *bool  `gorm:"column:Disposal" json:"disposal"`
	PermissionApproval          *bool  `gorm:"column:PermissionApproval" json:"permissionApproval"`
	AllowSeeAllFromAllowedAreas *bool  `gorm:"column:AllowSeeAllFromAllowedAreas" json:"allowSeeAllFromAllowedAreas"`
}

func (p PERMISSION) MarshalJSON() ([]byte, error) {
	type Alias PERMISSION
	return json.Marshal(&struct {
		Alias
		UserIDAlias string `json:"userId"`
	}{
		Alias:       Alias(p),
		UserIDAlias: p.UserID,
	})
}

func (PERMISSION) TableName() string {
	return "PERMISSION"
}

type REFRESH_TOKEN struct {
	ID              int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Token           string     `gorm:"column:token;type:varchar(255)" json:"token"`
	UserID          string     `gorm:"column:user_id;type:varchar(50)" json:"userId"`
	ExpiresAt       *time.Time `gorm:"column:expires_at" json:"expiresAt"`
	IsRevoked       *bool      `gorm:"column:is_revoked" json:"isRevoked"`
	CreatedAt       *time.Time `gorm:"column:created_at" json:"createdAt"`
	ReplacedByToken string     `gorm:"column:replaced_by_token;type:varchar(255)" json:"replacedByToken"`
}

func (REFRESH_TOKEN) TableName() string {
	return "refresh_tokens"
}

type HISTORY struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID     string     `gorm:"column:user_id;type:varchar(50)" json:"userId"`
	Action     string     `gorm:"column:action;type:varchar(250)" json:"action"`
	DateTimeAC *time.Time `gorm:"column:date_time_ac" json:"dateTimeAC"`
}

func (HISTORY) TableName() string {
	return "history"
}
