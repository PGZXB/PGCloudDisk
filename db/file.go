package db

import (
	"PGCloudDisk/errno"
	"PGCloudDisk/models"
	"gorm.io/gorm"
)

func AddFile(file *models.File) (s errno.Status) {

	// 不允许[filename, location, user_id]重复
	err := conn.Where(
		"filename = ? AND location = ? AND user_id = ? AND type = ?",
		file.Filename, file.Location, file.UserID, file.Type,
	).First(&models.File{}).Error
	if err != gorm.ErrRecordNotFound {
		s.Code = errno.FileAddRepeated
		return
	}

	if err := conn.Create(file).Error; err != nil {
		s.Code = errno.FileAddFailed
	}
	return
}

func AddFileAtLocation(file *models.File, locId int64) (s errno.Status) { // 根据locId自动填充file.Location
	// 查找路径
	loc := models.File{}
	if err := conn.Where("id = ? AND type = 'DIR'", locId).First(&loc).Error; err != nil {
		s.Code = errno.FileNotFound
		return
	}

	// 请求非法
	if file.UserID != loc.UserID {
		s.Code = errno.FileAddFailed
		return
	}

	// 填充file.Location
	file.Location = loc.Location + loc.Filename + "/"

	// 插入数据
	return AddFile(file)
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
