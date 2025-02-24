package utils

import (
	"errors"
	"server/global"
	"server/model/request"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWT struct {
	AccessTokenSecret  []byte
	RefreshTokenSecret []byte
}

var (
	TokenExpired     = errors.New("token is expired")
	TokenNotValidYet = errors.New("token not active yet")
	TokenMalformed   = errors.New("that's not even a token")
	TokenInvalid     = errors.New("couldn't handle this token")
)

func NewJWT() *JWT {
	return &JWT{
		AccessTokenSecret:  []byte(global.Config.Jwt.AccessTokenSecret),
		RefreshTokenSecret: []byte(global.Config.Jwt.RefreshTokenSecret),
	}
}

func (j *JWT) CreateAccessClaims(baseClaims request.BaseClaims) request.JwtCustomClaims {
	ep, _ := ParseDuration(global.Config.Jwt.AccessTokenExpiryTime)
	claims := request.JwtCustomClaims{
		BaseClaims: baseClaims,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"Go Blog"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ep)),
			Issuer:    global.Config.Jwt.Issuer,
		},
	}
	return claims
}

func (j *JWT) CreateAccessToken(claims request.JwtCustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(j.AccessTokenSecret)
}

func (j *JWT) CreateRefreshClaims(baseClaims request.BaseClaims) request.JwtCustomRefreshClaims {
	ep, _ := ParseDuration(global.Config.Jwt.RefreshTokenExpiryTime)
	claims := request.JwtCustomRefreshClaims{
		UserID: baseClaims.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"Go Blog"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ep)),
			Issuer:    global.Config.Jwt.Issuer,
		},
	}
	return claims
}

func (j *JWT) CreateRefreshToken(claims request.JwtCustomRefreshClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(j.RefreshTokenSecret)
}

func (j *JWT) ParseAccessToken(tokenString string) (*request.JwtCustomClaims, error) {
	claims, err := j.parseToken(tokenString, &request.JwtCustomClaims{}, j.AccessTokenSecret)
	if err != nil {
		return nil, err
	}
	if customClaims, ok := claims.(*request.JwtCustomClaims); ok {
		return customClaims, nil
	}
	return nil, TokenInvalid
}

func (j *JWT) ParseRefreshToken(tokenString string) (*request.JwtCustomRefreshClaims, error) {
	claims, err := j.parseToken(tokenString, &request.JwtCustomRefreshClaims{}, j.RefreshTokenSecret)
	if err != nil {
		return nil, err
	}
	if refreshClaims, ok := claims.(*request.JwtCustomRefreshClaims); ok {
		return refreshClaims, nil
	}
	return nil, TokenInvalid
}

func (j *JWT) parseToken(tokenString string, claims jwt.Claims, secretKey interface{}) (interface{}, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			switch {
			case ve.Errors&jwt.ValidationErrorMalformed != 0:
				return nil, TokenMalformed
			case ve.Errors&jwt.ValidationErrorExpired != 0:
				return nil, TokenExpired
			case ve.Errors&jwt.ValidationErrorNotValidYet != 0:
				return nil, TokenNotValidYet
			default:
				return nil, TokenInvalid
			}
		}
		return nil, TokenInvalid
	}
	if token.Valid {
		return token.Claims, nil
	}
	return nil, TokenInvalid
}
