package dto

import "linke/internal/shared/pool"

// Object pools for frequently used DTO types to reduce GC pressure
var (
	tokenResponsePool = pool.NewPool(func() *TokenResponse {
		return &TokenResponse{}
	}, func(tr *TokenResponse) {
		*tr = TokenResponse{}
	})

	claimsPool = pool.NewPool(func() *Claims {
		return &Claims{}
	}, func(c *Claims) {
		*c = Claims{}
	})

	authResponsePool = pool.NewPool(func() *AuthResponse {
		return &AuthResponse{}
	}, func(ar *AuthResponse) {
		if ar.User != nil {
			userResponsePool.Put(ar.User)
		}
		if ar.Token != nil {
			tokenResponsePool.Put(ar.Token)
		}
		ar.User = nil
		ar.Token = nil
	})

	userResponsePool = pool.NewPool(func() *UserResponse {
		return &UserResponse{}
	}, func(ur *UserResponse) {
		*ur = UserResponse{}
	})
)

// GetTokenResponse gets a TokenResponse from the pool
func GetTokenResponse() *TokenResponse {
	return tokenResponsePool.Get()
}

// PutTokenResponse returns a TokenResponse to the pool after resetting it
func PutTokenResponse(tr *TokenResponse) {
	if tr == nil {
		return
	}
	tokenResponsePool.Put(tr)
}

// GetClaims gets a Claims from the pool
func GetClaims() *Claims {
	return claimsPool.Get()
}

// PutClaims returns a Claims to the pool after resetting it
func PutClaims(c *Claims) {
	if c == nil {
		return
	}
	claimsPool.Put(c)
}

// GetAuthResponse gets an AuthResponse from the pool
func GetAuthResponse() *AuthResponse {
	return authResponsePool.Get()
}

// PutAuthResponse returns an AuthResponse to the pool after resetting it
func PutAuthResponse(ar *AuthResponse) {
	if ar == nil {
		return
	}
	authResponsePool.Put(ar)
}

// GetUserResponse gets a UserResponse from the pool
func GetUserResponse() *UserResponse {
	return userResponsePool.Get()
}

// PutUserResponse returns a UserResponse to the pool after resetting it
func PutUserResponse(ur *UserResponse) {
	if ur == nil {
		return
	}
	userResponsePool.Put(ur)
}

// NewTokenResponseFromPool creates a new TokenResponse from pool and populates it
func NewTokenResponseFromPool(accessToken, tokenType string, expiresIn int, expiresAt interface{}, refreshToken string) *TokenResponse {
	tr := GetTokenResponse()
	tr.AccessToken = accessToken
	tr.TokenType = tokenType
	tr.ExpiresIn = expiresIn
	if expiresAtTime, ok := expiresAt.(interface {
		Time() interface{}
	}); ok {
		if timeVal, ok := expiresAtTime.Time().(interface {
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
