package token

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

// JwtSign 生成一个 expireDuration 后过期的 JWT 令牌
func (t *token) JwtSign(userId int64, userName string, expireDuration time.Duration) (tokenString string, err error) {
	// The token content.
	// iss: （Issuer）签发者
	// iat: （Issued At）签发时间，用Unix时间戳表示
	// exp: （Expiration Time）过期时间，用Unix时间戳表示
	// aud: （Audience）接收该JWT的一方
	// sub: （Subject）该JWT的主题
	// nbf: （Not Before）不要早于这个时间
	// jti: （JWT ID）用于标识JWT的唯一ID
	claims := claims{
		userId,
		userName,
		jwt.StandardClaims{
			NotBefore: time.Now().Unix(),
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(expireDuration).Unix(),
		},
	}
	tokenString, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(t.secret))
	return
}

func (t *token) JwtParseUnsafe(tokenString string) (*claims, error) {
	tokenClaims, _, err := new(jwt.Parser).ParseUnverified(tokenString, &claims{})
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*claims); ok {
			return claims, nil
		}
	}
	return nil, err
}

var (
	ErrorTokenCannotParse        = errors.New("cannot parse the token. ")
	ErrorTokenExpiredOrNotActive = errors.New("token expired or not active. ")
)

// JwtParse 从 JWT 中解密数据
func (t *token) JwtParse(tokenString string) (*claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.secret), nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*claims); ok && tokenClaims.Valid {
			return claims, nil
		} else {
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorMalformed != 0 {
					return nil, ErrorTokenCannotParse
				} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
					// Token is either expired or not active yet
					return nil, ErrorTokenExpiredOrNotActive
				} else {
					return nil, ErrorTokenCannotParse
				}
			} else {
				return nil, ErrorTokenCannotParse
			}
		}
	}
	return nil, ErrorTokenCannotParse
}

func (t *token) JwtParseFromAuthorizationHeader(tokenString string) (*claims, error) {
	tokenString = stripBearerPrefixFromTokenString(tokenString)
	return t.JwtParse(tokenString)
}

// Strips 'Bearer ' prefix from bearer token string
func stripBearerPrefixFromTokenString(tok string) string {
	// Should be a bearer token
	if len(tok) > 6 && strings.ToUpper(tok[0:7]) == "BEARER " {
		return tok[7:]
	}
	return tok
}
