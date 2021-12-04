package main

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
)

type (
	Role       uint8
	Permission uint8
)

const (
	NoRole Role = iota
	SuperUserRole
	ServiceRole
	AdminRole
	UserRole
)

const (
	NoPermission Permission = iota
)

type Claims struct {
	ID          uint         `json:"id"`
	Username    string       `json:"username"`
	Role        Role         `json:"role"`
	Permissions []Permission `json:"permissions"`
	jwt.StandardClaims
}

func MakeJWT(cfg JWT, id uint, username string, role Role, perms ...Permission) (string, error) {
	expiresAt := time.Now().Add(time.Second * time.Duration(cfg.MaxAge))
	claims := Claims{
		ID:          id,
		Username:    username,
		Role:        role,
		Permissions: perms,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
			Issuer:    "andsm",
		},
	}

	token := jwt.NewWithClaims(cfg.Algorithm, claims)
	ss, err := token.SignedString(cfg.Secret)
	if err != nil {
		return "", errors.Wrap(err, "")
	}

	return ss, nil
}

func GetClaimsFromJWT(cfg JWT, token []byte) (*Claims, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return cfg.Secret, nil
	}

	claims := new(Claims)

	tknString := string(token)
	if strings.HasPrefix(tknString, "Bearer ") {
		tknString = strings.Split(tknString, "Bearer ")[1]
	}
	tkn, err := jwt.ParseWithClaims(tknString, claims, keyFunc)
	if err != nil {
		e, ok := err.(*jwt.ValidationError)
		if ok && e.Errors == jwt.ValidationErrorExpired {
			return nil, errors.Wrap(err, "expired token")
		}

		return nil, errors.Wrap(err, "invalid token")
	}

	return tkn.Claims.(*Claims), nil
}

func Authorize(ctx *Context, cfg JWT, r Role, perms ...Permission) error {
	tkn := ctx.Request.Header.Cookie("access_token")
	if r == ServiceRole {
		tkn = ctx.Request.Header.Peek("Authorization")
	}
	claims, err := GetClaimsFromJWT(cfg, tkn)
	if err != nil {
		if r == ServiceRole {
			return ErrUnauthorized(err)
		}

		// TODO: refresh token for users
		// if r != NoRole && err.(ErrorResponseType).IsErrExpiredToken() {
		// }

		return errors.New(err.Error())
	}

	if r != NoRole && claims.Role != r {
		return ErrForbiden()
	}

	if len(perms) > 0 {
		for _, p := range perms {
			if !PermissionExist(p, claims.Permissions) {
				return ErrForbiden()
			}
		}
	}

	if r != ServiceRole {
		ctx = ctx.WithIdentify(claims.ID, claims.Username, claims.Role, claims.Permissions...)
	}

	return nil
}

func PermissionExist(perm Permission, perms []Permission) bool {
	for _, p := range perms {
		if perm == p {
			return true
		}
	}

	return false
}
