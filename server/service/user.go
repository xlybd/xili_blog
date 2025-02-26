package service

import (
	"errors"
	"server/global"
	"server/model/appTypes"
	"server/model/database"
	"server/model/other"
	"server/utils"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type UserService struct {
}

func (userService *UserService) Register(u database.User) (database.User, error) {
	if !errors.Is(global.DB.Where("email = ?", u.Email).First(&database.User{}).Error, gorm.ErrRecordNotFound) {
		return database.User{}, errors.New("this eamil address is already registered, please check the ifformation you filled in, or retrieve your password")
	}

	u.Password = utils.BcryptHash(u.Password)
	u.UUID = uuid.Must(uuid.NewV4())
	u.Avatar = "/image/avatar.jpg"
	u.RoleID = appTypes.User
	u.Register = appTypes.Email

	if err := global.DB.Create(&u).Error; err != nil {
		return database.User{}, err
	}

	return u, nil
}

func (userService *UserService) EmailLogin(u database.User) (database.User, error) {
	var user database.User
	err := global.DB.Where("email = ?", u.Email).First(&user).Error
	if err == nil {
		if ok := utils.BcryptCheck(u.Password, user.Password); !ok {
			return database.User{}, errors.New("incorrect email or password")
		}

		return user, nil
	}

	return database.User{}, err
}

func (userService *UserService) QQLogin(accessTokenResponse other.AccessTokenResponse) (database.User, error) {
	var user database.User

	err := global.DB.Where("openid = ?", accessTokenResponse.Openid).First(&user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return database.User{}, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		userInfoResponse, err := ServiceGroupApp.QQService.GetUserInfoByAccessTokenAndOpenid(accessTokenResponse.AccessToken, accessTokenResponse.Openid)
		if err != nil {
			return database.User{}, err
		}

		user.UUID = uuid.Must(uuid.NewV4())
		user.Username = userInfoResponse.Nickname
		user.Openid = accessTokenResponse.Openid
		user.Avatar = userInfoResponse.FigureurlQQ2
		user.RoleID = appTypes.User
		user.Register = apptypes.QQ

		if err := global.DB.Create(&user).Error; err != nil {
			return database.User{}, err
		}
	}

	return user, nil
}

func (userService *UserService) ForgotPassword() {}

func (userService *UserService) UserCard() {}

func (userService *UserService) Logout() {}

func (userService *UserService) UserResetPassword() {}

func (userService *UserService) UserInfo() {}

func (userService *UserService) UserChangeInfo() {}

func (userService *UserService) UserWeather() {}

func (userService *UserService) UserChart() {}

func (userService *UserService) UserList() {}

func (userService *UserService) UserFreeze() {}

func (userService *UserService) UserUnfreeze() {}

func (userService *UserService) UserLoginList() {}
