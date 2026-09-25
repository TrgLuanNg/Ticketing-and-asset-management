package dtos

type LoginDto struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordDto struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword" binding:"required"`
}

type RefreshRequestDto struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LoginResponseDto struct {
	UserName     string   `json:"userName"`
	UserID       string   `json:"userId"`
	Token        string   `json:"token"`
	RefreshToken string   `json:"refreshToken"`
	Roles        []string `json:"roles"`
	Area         string   `json:"area"`
	Permissions  []string `json:"permissions"`
}
