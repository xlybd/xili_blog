package service

import (
	"server/global"
	"server/model/database"
	"server/utils"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

type JwtService struct {
}

func (jwtService *JwtService) SetRedisJWT(jwt string, uuid uuid.UUID) error {
	dr, err := utils.ParseDuration(global.Config.Jwt.RefreshTokenExpiryTime)
	if err != nil {
		return err
	}
	return global.Redis.Set(uuid.String(), jwt, dr).Err()
}

func (jwtService *JwtService) GetRedisJWT(uuid uuid.UUID) (string, error) {
	return global.Redis.Get(uuid.String()).Result()
}

func (jwtService *JwtService) JoinInBlacklist(jwtList database.JwtBlacklist) error {
	if err := global.DB.Create(&jwtList).Error; err != nil {
		return err
	}
	global.BlackCache.SetDefault(jwtList.Jwt, struct{}{})
	return nil
}

func (jwtService *JwtService) IsInBlacklist(jwt string) bool {
	_, ok := global.BlackCache.Get(jwt)
	return ok
}

func LoadAll() {
	var data []string
	if err := global.DB.Model(&database.JwtBlacklist{}).Pluck("jwt", &data).Error; err != nil {
		global.Log.Error("Failed to load JWT blacklist from the database", zap.Error(err))
		return
	}
	for i := 0; i < len(data); i++ {
		global.BlackCache.SetDefault(data[i], struct{}{})
	}
}
