package dtos

import "time"

type UserInfoDto struct {
	UserID      string   `json:"userId"`
	UserName    string   `json:"userName"`
	Description string   `json:"description"`
	Area        string   `json:"area"`
	Role        string   `json:"role"`
	IsActive    *bool    `json:"isActive"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

type PermissionDto struct {
	UserID                     string `json:"userId"`
	AssetsModify               *bool  `json:"assetsModify"`
	AssetsDel                  *bool  `json:"assetsDel"`
	DepartmentModify           *bool  `json:"departmentModify"`
	DepartmentDel              *bool  `json:"departmentDel"`
	TypeModify                 *bool  `json:"typeModify"`
	TypeDel                    *bool  `json:"typeDel"`
	UserManager                *bool  `json:"userManager"`
	Disposal                   *bool  `json:"disposal"`
	PermissionApproval         *bool  `json:"permissionApproval"`
	AllowSeeAllFromAllowedAreas *bool  `json:"allowSeeAllFromAllowedAreas"`
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
