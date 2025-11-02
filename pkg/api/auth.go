package api

// REQUEST BODIES
type ExchangeRefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RESPOSNE BODIES
type ExchangeRefreshTokenResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}
