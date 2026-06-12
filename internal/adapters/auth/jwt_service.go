package authadapter

import (
	"github.com/itsLeonB/cashback/internal/domain/service/auth"
	"github.com/itsLeonB/sekure"
)

type jwtServiceAdapter struct {
	inner sekure.JWTService
}

func NewJWTService(inner sekure.JWTService) auth.JWTService {
	return &jwtServiceAdapter{inner}
}

func (a *jwtServiceAdapter) CreateToken(claims map[string]any) (string, error) {
	return a.inner.CreateToken(claims)
}

func (a *jwtServiceAdapter) VerifyToken(token string) (auth.Claims, error) {
	claims, err := a.inner.VerifyToken(token)
	if err != nil {
		return auth.Claims{}, err
	}
	return auth.Claims{Data: claims.Data}, nil
}
