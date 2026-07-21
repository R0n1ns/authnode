package domain

type TokenClaims struct {
	UserID    int64    `json:"userId"`
	Email     string   `json:"email"`
	Nickname  string   `json:"nickname"`
	Roles     []string `json:"roles"`
	ExpiresAt int64    `json:"exp"`
	IssuedAt  int64    `json:"iat"`
	TokenID   string   `json:"jti"`
}
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}
