package db

import (
	"PGCloudDisk/errno"
	"PGCloudDisk/models"
	"gorm.io/gorm"
)

func UserCheck(username, password string) (s errno.Status) {

	var user models.User
	err := conn.Select("id, password").Where("username = ?", username).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		s.Code = errno.UserNotFound
		return
	}
	if user.Password != password {
		s.Code = errno.UserNamePwdNotMatched
		return
	}

	return
}

func AddUser(username, password string) (s errno.Status) {
	err := conn.Select("id").Where(&models.User{Username: username}).First(&models.User{}).Error
	if err != gorm.ErrRecordNotFound {
		s.Code = errno.UserNameRepeated
		return
	}

	err = conn.Create(&models.User{Username: username, Password: password}).Error
	if err != nil {
		s.Code = errno.UserAddFailed
		return
	}

	return
}
