package db

import (
	"PGCloudDisk/errno"
	"PGCloudDisk/models"
)

func AddFile(file *models.File) (s errno.Status) {
	if err := conn.Create(file).Error; err != nil {
		s.Code = errno.FileAddFailed
	}
	return
}

func DeleteFile(id int64) (s errno.Status) {
	if err := conn.Delete(&models.File{Model: models.Model{ID: id}}).Error; err != nil {
		s.Code = errno.FileDeleteFailed
	}
	return
}

func UpdateFile(file *models.File) (s errno.Status) {
	if err := conn.Updates(file).Error; err != nil {
		s.Code = errno.FileUpdateFailed
	}
	return
}

func FindFilesOfUser(uid int64) (res []models.File, s errno.Status) {
	if err := conn.Where("user_id = ?", uid).Find(&res).Error; err != nil {
		s.Code = errno.FileFindOfUserFailed
	}
	return
}

func FindFilesOfUserByName(uid int64, infix string) (res map[string]*models.File, s errno.Status) {

	var temp []models.File
	if err := conn.Where("user_id = ? AND filename LIKE ?", uid, "%"+infix+"%").Find(&temp).Error; err != nil {
		s.Code = errno.FileFindOfUserByName
		return
	}

	res = make(map[string]*models.File)
	for _, item := range temp {
		res[item.Filename] = &item
	}

	return
}
