package jwt

import (
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type Claims struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	jwt.RegisteredClaims
}

func CreateJWT(id string, name string) (string, error) {
	jwtExpireHour, err := strconv.Atoi(os.Getenv("AUTH_JWT_EXPIRE_HOUR"))

	if err != nil {
		return "", errors.Wrap(err, "failed to parse session expire hour")
	}

	claims := Claims{
		ID:   id,
		Name: name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(jwtExpireHour) * time.Hour)),
		},
	}

	return jwt.
		NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(os.Getenv("AUTH_JWT_SIGNINGKEY")))
}

func GetJWTClaims(c echo.Context) (*Claims, error) {
	token, ok := c.Get(os.Getenv("AUTH_JWT_CONTEXT_KEY")).(*jwt.Token)

	if !ok {
		return nil, errors.New("failed to get JWT")
	}

	claims, ok := token.Claims.(*Claims)

	if !ok {
		return nil, errors.New("failed to cast JWT")
	}

	return claims, nil
}

func AuthMiddleware() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		ContextKey: os.Getenv("AUTH_JWT_CONTEXT_KEY"),
		SigningKey: []byte(os.Getenv("AUTH_JWT_SIGNING_KEY")),
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(Claims)
		},
		Skipper: func(c echo.Context) bool {
			return true
		},
	})
}
