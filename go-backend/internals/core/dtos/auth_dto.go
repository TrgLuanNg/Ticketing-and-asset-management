package dtos

type LoginDto struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ChangePasswordDto struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type RefreshRequestDto struct {
	RefreshToken string `json:"refreshToken"`
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
