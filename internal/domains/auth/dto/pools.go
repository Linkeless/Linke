package dto

import "sync"

// Object pools for frequently used DTO types to reduce GC pressure
var (
	tokenResponsePool = sync.Pool{
		New: func() interface{} {
			return &TokenResponse{}
		},
	}

	claimsPool = sync.Pool{
		New: func() interface{} {
			return &Claims{}
		},
	}

	authResponsePool = sync.Pool{
		New: func() interface{} {
			return &AuthResponse{}
		},
	}

	userResponsePool = sync.Pool{
		New: func() interface{} {
			return &UserResponse{}
		},
	}
)

// GetTokenResponse gets a TokenResponse from the pool
func GetTokenResponse() *TokenResponse {
	return tokenResponsePool.Get().(*TokenResponse)
}

// PutTokenResponse returns a TokenResponse to the pool after resetting it
func PutTokenResponse(tr *TokenResponse) {
	if tr == nil {
		return
	}
	// Reset fields to avoid data leakage
	*tr = TokenResponse{}
	tokenResponsePool.Put(tr)
}

// GetClaims gets a Claims from the pool
func GetClaims() *Claims {
	return claimsPool.Get().(*Claims)
}

// PutClaims returns a Claims to the pool after resetting it
func PutClaims(c *Claims) {
	if c == nil {
		return
	}
	// Reset fields to avoid data leakage
	*c = Claims{}
	claimsPool.Put(c)
}

// GetAuthResponse gets an AuthResponse from the pool
func GetAuthResponse() *AuthResponse {
	return authResponsePool.Get().(*AuthResponse)
}

// PutAuthResponse returns an AuthResponse to the pool after resetting it
func PutAuthResponse(ar *AuthResponse) {
	if ar == nil {
		return
	}
	// Reset fields carefully to avoid data leakage
	if ar.User != nil {
		PutUserResponse(ar.User)
	}
	if ar.Token != nil {
		PutTokenResponse(ar.Token)
	}
	ar.User = nil
	ar.Token = nil
	authResponsePool.Put(ar)
}

// GetUserResponse gets a UserResponse from the pool
func GetUserResponse() *UserResponse {
	return userResponsePool.Get().(*UserResponse)
}

// PutUserResponse returns a UserResponse to the pool after resetting it
func PutUserResponse(ur *UserResponse) {
	if ur == nil {
		return
	}
	// Reset fields to avoid data leakage
	*ur = UserResponse{}
	userResponsePool.Put(ur)
}

// NewTokenResponseFromPool creates a new TokenResponse from pool and populates it
func NewTokenResponseFromPool(accessToken, tokenType string, expiresIn int, expiresAt interface{}, refreshToken string) *TokenResponse {
	tr := GetTokenResponse()
	tr.AccessToken = accessToken
	tr.TokenType = tokenType
	tr.ExpiresIn = expiresIn
	if expiresAtTime, ok := expiresAt.(interface{ 
		Time() interface{} 
	}); ok {
		if timeVal, ok := expiresAtTime.Time().(interface{ 
			Unix() int64 
		}); ok {
			// Handle time conversion if needed
			_ = timeVal
		}
	}
	tr.RefreshToken = refreshToken
	return tr
}

// NewAuthResponseFromPool creates a new AuthResponse from pool with provided user and token
func NewAuthResponseFromPool(user *UserResponse, token *TokenResponse) *AuthResponse {
	ar := GetAuthResponse()
	ar.User = user
	ar.Token = token
	return ar
}