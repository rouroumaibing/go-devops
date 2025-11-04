package jwtauth

import (
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rouroumaibing/go-devops/pkg/apis/auths"
	"github.com/rouroumaibing/go-devops/pkg/utils/errcode"
	"github.com/rouroumaibing/go-devops/pkg/utils/response"
)

func Init() {
	beego.InsertFilter("/api/*", beego.BeforeRouter, AuthMiddleware)
}

func GenerateTokens(userID string) (accessToken, refreshToken string, err error) {
	accessTokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, auths.JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(auths.AccessTokenExpire)),
			Issuer:    auths.JWTIssuer,
		},
	})
	accessToken, err = accessTokenClaims.SignedString(auths.AccessSecret)
	if err != nil {
		return
	}

	refreshTokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, auths.JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(auths.RefreshTokenExpire)),
			Issuer:    auths.JWTIssuer,
		},
	})
	refreshToken, err = refreshTokenClaims.SignedString(auths.RefreshSecret)
	return
}

// 用于普通接口认证
func ParseAccessToken(tokenString string) (*auths.JWTClaims, error) {
	var claims auths.JWTClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return auths.AccessSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return &claims, nil
}

// 仅用于刷新token认证使用
func ParseRefreshToken(tokenString string) (*auths.JWTClaims, error) {
	var claims auths.JWTClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return auths.RefreshSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return &claims, nil
}

func RefreshUserToken(refreshToken string) (newAccessToken, newRefreshToken string, err error) {
	claims, err := ParseRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}
	return GenerateTokens(claims.UserID)
}

func AuthMiddleware(context *context.Context) {
	for _, uri := range auths.RequestURIWhiteList {
		if context.Request.RequestURI == uri {
			return
		}
	}
	authorizationHeader := context.Request.Header.Get("Authorization")
	if authorizationHeader == "" {
		response.SetErrorResponse(context, errcode.E.Base.Token.WithMessage("missing token"))
		return
	}
	tokenParts := strings.Split(authorizationHeader, " ")
	tokenString := authorizationHeader
	if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
		tokenString = tokenParts[1]
	}
	claims, err := ParseAccessToken(tokenString)
	if err != nil {
		response.SetErrorResponse(context, errcode.E.Base.Token.WithMessage("invalid token"))
		return
	}
	context.Input.SetData("useruuid", claims.UserID)
}
