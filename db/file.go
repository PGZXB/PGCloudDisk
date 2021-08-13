package db

import (
	"PGCloudDisk/errno"
	"PGCloudDisk/models"
	"gorm.io/gorm"
	"strings"
)

func AddFile(file *models.File) (s errno.Status) {

	// 不允许[filename, location, user_id]重复
	err := conn.Where(
		"filename = ? AND location = ? AND user_id = ?",
		file.Filename, file.Location, file.UserID,
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

func DeleteFileOnly(id int64) (s errno.Status) { // 单纯的删除记录项, 不考虑目录下的文件
	if err := conn.Delete(&models.File{Model: models.Model{ID: id}}).Error; err != nil {
		s.Code = errno.FileDeleteFailed
	}
	return
}

func DeleteFilesOnly(ids []int64) (s errno.Status) {
	if err := conn.Delete(&models.File{}, ids).Error; err != nil {
		s.Code = errno.FileDeleteFailed
	}
	return
}

func FindFilesRecursively(id int64) (res []models.File, s errno.Status) {
	root := models.File{}
	res = make([]models.File, 0, 16)
	if err := conn.Where("id = ?", id).First(&root).Error; err != nil {
		s.Code = errno.FileNotFound
		return nil, s
	}
	err := findFilesRecursively(&res, root)
	if err != nil {
		return nil, errno.Status{Code: errno.FileNotFound}
	}
	return res, s
}

func FindFileOfUserWithFilter(userID int64, args *models.FileInfoQueryArgs) (res []models.FileInfoCanBePublicized, s errno.Status) {
	var queryPatterns = make([]string, 0, 8)
	var queryArgs = make([]interface{}, 0, 8)

	var findAtLocation = false
	var location string
	if args.LocationID != -1 {
		findAtLocation = true
		loc := models.File{}
		if err := conn.Where("id = ? AND user_id = ? AND type = 'DIR'", args.LocationID, userID).First(&loc).Error; err != nil {
			s.Code = errno.FileNotFound
			return
		}
		location = loc.Location + loc.Filename + "/"
	}

	queryPatterns = append(queryPatterns, "user_id = ?")
	queryArgs = append(queryArgs, userID)

	if args.IDRange != nil {
		if args.IDRange[0] == args.IDRange[1] {
			queryPatterns = append(queryPatterns, "id = ?")
			queryArgs = append(queryArgs, args.IDRange[0])
		} else {
			queryPatterns = append(queryPatterns, "id >= ? AND id <= ?")
			queryArgs = append(queryArgs, args.IDRange[0], args.IDRange[1])
		}
	}

	if args.TypeEnum != "" {
		queryPatterns = append(queryPatterns, "type = ?")
		queryArgs = append(queryArgs, args.TypeEnum)
	}

	if args.SizeRange != nil {
		if args.SizeRange[0] == args.SizeRange[1] {
			queryPatterns = append(queryPatterns, "size = ?")
			queryArgs = append(queryArgs, args.SizeRange[0])
		} else {
			queryPatterns = append(queryPatterns, "size >= ? AND size <= ?")
			queryArgs = append(queryArgs, args.SizeRange[0], args.SizeRange[1])
		}
	}

	if args.CreatedAtRange != nil {
		if args.CreatedAtRange[0].Equal(args.CreatedAtRange[1]) {
			queryPatterns = append(queryPatterns, "created_at = ?")
			queryArgs = append(queryArgs, args.CreatedAtRange[0])
		} else {
			queryPatterns = append(queryPatterns, "created_at >= ? AND created_at <= ?")
			queryArgs = append(queryArgs, args.CreatedAtRange[0], args.CreatedAtRange[1])
		}
	}

	if args.UpdatedAtRange != nil {
		if args.UpdatedAtRange[0].Equal(args.UpdatedAtRange[1]) {
			queryPatterns = append(queryPatterns, "updated_at = ?")
			queryArgs = append(queryArgs, args.UpdatedAtRange[0])
		} else {
			queryPatterns = append(queryPatterns, "updated_at >= ? AND updated_at <= ?")
			queryArgs = append(queryArgs, args.UpdatedAtRange[0], args.UpdatedAtRange[1])
		}
	}

	if args.DeletedAtRange != nil {
		if args.DeletedAtRange[0].Equal(args.DeletedAtRange[1]) {
			queryPatterns = append(queryPatterns, "deleted_at = ?")
			queryArgs = append(queryArgs, args.DeletedAtRange[0])
		} else {
			queryPatterns = append(queryPatterns, "deleted_at >= ? AND deleted_at <= ?")
			queryArgs = append(queryArgs, args.DeletedAtRange[0], args.DeletedAtRange[1])
		}
	}

	if findAtLocation {
		queryPatterns = append(queryPatterns, "location = ?")
		queryArgs = append(queryArgs, location)
	} else if args.LocationKeyword != "" { // 探索更好的做法(比如应用层过滤)
		queryPatterns = append(queryPatterns, "location LIKE ?")
		queryArgs = append(queryArgs, "%"+args.LocationKeyword+"%")
	}

	if args.FilenameKeyword != "" { // 探索更好的做法(比如应用层过滤)
		queryPatterns = append(queryPatterns, "filename LIKE ?")
		queryArgs = append(queryArgs, "%"+args.FilenameKeyword+"%")
	}

	queryPattern := strings.Join(queryPatterns, " AND ")
	result := conn.Model(&models.File{}).Where(queryPattern, queryArgs...).Find(&res)
	if result.Error != nil {
		res = nil
		s = errno.Status{Code: errno.FileFindWithFilterFailed}
		return
	}

	return
}

// 禁止修改文件元信息
// func UpdateFileOfUser(userId int64, id int64, to *models.FileInfoCanBeUpdated) (s errno.Status) {
//   data, _ := json.Marshal(to)
//   mp := make(map[string]interface{})
//   _ = json.Unmarshal(data, &mp)
//   if res := conn.Model(&models.File{}).Where("id = ? AND user_id = ?", id, userId).Updates(mp); res.RowsAffected == 0 {
//   	s.Code = errno.FileUpdateFailed
//   }
//   return
// }

func FindFileOfUserByID(userID, fileID int64) (res models.File, s errno.Status) {
	if err := conn.Where("id = ? AND user_id = ?", fileID, userID).First(&res).Error; err != nil {
		s.Code = errno.FileNotFound
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
		s.Code = errno.FileFindOfUserByNameFailed
		return
	}

	res = make(map[string]*models.File)
	for _, item := range temp {
		res[item.Filename] = &item
	}

	return
}

// helper-functions
func findFilesRecursively(res *[]models.File, root models.File) error {
	*res = append(*res, root)
	if root.Type == models.DirType { // 如果是目录则需找到其下的所有 文件/目录
		location := root.Location + root.Filename + "/"
		subFiles := make([]models.File, 0, 16)
		if err := conn.Where("location = ? AND user_id = ?", location, root.UserID).Find(&subFiles).Error; err != nil {
			return err
		}
		// 递归的调用
		for _, f := range subFiles {
			err := findFilesRecursively(res, f)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
