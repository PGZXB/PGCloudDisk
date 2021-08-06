package utils

import (
	"PGCloudDisk/config"
	"PGCloudDisk/db"
	"PGCloudDisk/errno"
	"PGCloudDisk/utils/lg"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type claims struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

func GetToken(username string) (string, errno.Status) {
	nowTime := time.Now()
	expireTime := nowTime.Add(2 * time.Hour)

	user, status := db.GetUserInfo(username)
	if !status.Success() {
		return "", errno.Status{Code: errno.CreateTokenFailed}
	}
	c := claims{
		user.ID,
		username,
		jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    "PGCloudDisk",
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	token, err := tokenClaims.SignedString([]byte(config.Cfg.JwtCfg.JwtSecret))

	if err != nil {
		return "", errno.Status{Code: errno.CreateTokenFailed}
	}

	return token, errno.Status{}
}

func ParseToken(token string) (*claims, errno.Status) {
	tok, err := jwt.ParseWithClaims(token, &claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.Cfg.JwtCfg.JwtSecret), nil
	})

	if err != nil {
		lg.Logger.Println("ParseTokenFailed")
		return nil, errno.Status{Code: errno.ParseTokenFailed}
	}

	if claims, ok := tok.Claims.(*claims); ok && tok.Valid {
		return claims, errno.Status{Code: errno.Success}
	}

	lg.Logger.Println("ParseTokenFailed")
	return nil, errno.Status{Code: errno.ParseTokenFailed}
}
