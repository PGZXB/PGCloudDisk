package v1

import (
	"PGCloudDisk/config"
	"PGCloudDisk/db"
	"PGCloudDisk/errno"
	"PGCloudDisk/models"
	"PGCloudDisk/utils"
	"PGCloudDisk/utils/fileutils"
	"PGCloudDisk/utils/lg"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// UploadFile upload a file(POST api/v1/files)
// only allow upload small file, content is in request-form "file"
// meta-info least : Filename(filename) LocationID(location_id) Type(type)
// response data : JSON(FileInfoCanBePublicized)
func UploadFile(c *gin.Context) {
	fInfo := struct {
		Filename   string `json:"filename" form:"filename"`
		LocationID int64  `json:"location_id" form:"location_id"`
		Type       string `json:"type" form:"type"`
	}{}
	err := c.ShouldBind(&fInfo)
	if err != nil || fInfo.Filename == "" || fInfo.LocationID == 0 || fInfo.Type == "" {
		utils.Response(c, http.StatusBadRequest, errno.RespCode{Code: errno.RespInvalidParams}, nil)
		return
	}

	// 验证Type的合法性
	if fInfo.Type != models.FileType && fInfo.Type != models.DirType {
		utils.Response(c, http.StatusBadRequest, errno.RespCode{Code: errno.RespRequestFileTypeInvalid}, nil)
		return
	}

	// 验证Filename的合法性, 只许有字母/汉字/下划线(这是展示给用户的虚拟文件名, 并非文件系统实际存储的文件名)
	fInfo.Filename = strings.TrimSpace(fInfo.Filename)
	filenameBytes := len([]byte(fInfo.Filename))
	if filenameBytes <= 0 || filenameBytes > 255 {
		utils.Response(c, http.StatusBadRequest, errno.RespCode{Code: errno.RespRequestFilenameInvalid}, nil)
		return
	}

	// TODO : Check Filename(只许有字母/汉字/下划线) And Namify

	// 如果是File是否有"file"
	var fileHeader *multipart.FileHeader
	if fInfo.Type == models.FileType {
		fileHeader, err = c.FormFile("file")
		if err != nil {
			utils.Response(c, http.StatusBadRequest, errno.RespCode{Code: errno.RespInvalidParams}, nil)
			return
		}
	}

	// 生成文件信息
	//     Size      int64
	//     LocalAddr localSaveRoot + / + UserID + / + LocationID + / + timestamp + _ + base64Encoding(Filename)
	//     UserID    int64
	fileModel := models.File{
		Filename: fInfo.Filename,
		Size:     0,
		Type:     fInfo.Type,
	}

	uid, ok := c.Get("user_id")
	fileModel.UserID, ok = uid.(int64)
	if !ok {
		utils.Response(c, http.StatusInternalServerError, errno.RespCode{Code: errno.RespFailed}, nil)
		return
	}

	var localDir string
	var localFilename string
	if fileModel.Type == models.FileType {
		localDir = filepath.Join(
			config.Cfg.LocalSaveCfg.Root,
			"User_"+strconv.FormatInt(fileModel.UserID, 10),
			"Dir_"+strconv.FormatInt(fInfo.LocationID, 10),
		)

		localFilename = fmt.Sprintf("%s_%s_%s",
			strconv.FormatInt(time.Now().UnixNano(), 16),
			strconv.FormatInt(rand.Int63n(1e8), 16),
			base64.URLEncoding.EncodeToString([]byte(fileModel.Filename)))

		fileModel.LocalAddr = filepath.Join(localDir, localFilename)

		if config.Cfg.RunMode.IsDebug {
			lg.Logger.Printf("The localAddr Of File %#v : %#v\n", fileModel.Filename, fileModel.LocalAddr)
		}
	}

	if fileModel.Type == models.FileType {
		fileModel.Size = fileHeader.Size
	}

	// 开启一个事务
	var allSuccess bool = false
	db.Transaction(func() bool {
		// 存入数据库
		s := db.AddFileAtLocation(&fileModel, fInfo.LocationID)
		if s.Code == errno.FileNotFound {
			utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespVirtualPathNotFound}, nil)
		} else if s.Code == errno.FileAddFailed {
			utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespAddFileFailed}, nil)
		} else if s.Code == errno.FileAddRepeated {
			utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespAddFileRepeated}, nil)
		}
		if !s.Success() {
			return false
		}

		// 读取文件并保存到本地, 如果是目录无需此步
		if fInfo.Type == models.FileType {
			if !fileutils.IsDir(localDir) {
				err = os.MkdirAll(localDir, 0755)
				if err != nil {
					lg.Logger.Printf("Create Dir %s When Saving File %s Failed\n", localDir, fileModel.LocalAddr)
					utils.Response(c, http.StatusInternalServerError, errno.RespCode{Code: errno.RespSaveFileFailed}, nil)
					return false
				}
			}
			err = c.SaveUploadedFile(fileHeader, fileModel.LocalAddr)
			if err != nil {
				lg.Logger.Printf("Save File %s Failed\n", fileModel.LocalAddr)
				utils.Response(c, http.StatusInternalServerError, errno.RespCode{Code: errno.RespSaveFileFailed}, nil)
				return false
			}
		}

		allSuccess = true
		return true
	})

	// 都成功才返回正确的信息
	if allSuccess {
		res, s := db.FindFilesOfUserByName(fileModel.UserID, fileModel.Filename)
		if !s.Success() {
			utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespFailed}, models.FileInfoCanBePublicized{
				ID:        fileModel.ID,
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
				DeletedAt: time.Time{},
				Filename:  fileModel.Filename,
				Size:      fileModel.Size,
				Location:  fileModel.Location,
				Type:      fileModel.Type,
			})
			return
		}
		fileM := res[fileModel.Filename]
		utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespSuccess}, models.FileInfoCanBePublicized{
			ID:        fileM.ID,
			CreatedAt: fileM.CreatedAt.Time,
			UpdatedAt: fileM.UpdatedAt.Time,
			DeletedAt: fileM.DeletedAt.Time,
			Filename:  fileM.Filename,
			Size:      fileM.Size,
			Location:  fileM.Location,
			Type:      fileM.Type,
		})
	}
}

// DeleteFile delete a file(DELETE api/v1/files/:id)
// soft remove file-info and move the file in trash
func DeleteFile(c *gin.Context) {
	// 获取file_id和user_id
	fileId, userId, ok := getFileIdAndUserId(c)
	if !ok {
		return
	}

	// 获取文件信息
	fileM, status := db.FindFileOfUserByID(userId, fileId)
	if !status.Success() {
		utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespFileNotFound}, nil)
		return
	}

	assert.IsEqual(fileM.ID, fileId)
	if fileM.Type == models.FileType {
		// 在数据库中软删除记录
		status = db.DeleteFileOnly(fileId)
		if !status.Success() {
			utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespFileDeleteFailed}, nil)
			return
		}

		// 如果是文件则将文件移入"回收站", 目录格式 : <回收站>/<精确到天的格式化时间YYYY-MM-DD>/<fileId>
		// 以后会编写工具, 定期删除超时的文件夹
		var srcAddr string = fileM.LocalAddr
		var destAddr string

		destDirAddr := filepath.Join(
			config.Cfg.LocalSaveCfg.TrashPath,
			time.Now().Format("2006-01-02"),
		)
		destFilename := strconv.FormatInt(fileId, 10)
		destAddr = filepath.Join(destDirAddr, destFilename)
		err := os.MkdirAll(destDirAddr, 0755)
		if err != nil {
			lg.Logger.Println("Server Error, Mkdir %s Failed When Calling DeleteFile\n", destDirAddr)
			utils.Response(c, http.StatusInternalServerError, errno.RespCode{Code: errno.RespFileDeleteFailed}, nil)
			return
		}

		err = os.Rename(srcAddr, destAddr) // 移入回收站
		if err != nil {
			lg.Logger.Println("Server Error, Move(Rename) %s Failed When Calling DeleteFile\n", destDirAddr)
			utils.Response(c, http.StatusInternalServerError, errno.RespCode{Code: errno.RespFileDeleteFailed}, nil)
			return
		}

		// 成功应答
		utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespSuccess}, gin.H{
			"count": 1,
		})
	} else {
		// 如果是目录, 则需要递归的删除其下的所有文件和文件夹(软删除其记录, 将所有文件移入回收站)
		//// 递归的找到该目录下的所有文件
		files, status := db.FindFilesRecursively(fileId)
		if !status.Success() {
			utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespFileDeleteFailed}, nil)
			return
		}

		//// 删除记录
		ids := make([]int64, len(files))
		for i, item := range files {
			ids[i] = item.ID
		}
		if s := db.DeleteFilesOnly(ids); !s.Success() {
			utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespFileDeleteFailed}, nil)
			return
		}
		//// 移入回收站, 出错打日志并向前端报告
		destDirAddr := filepath.Join(
			config.Cfg.LocalSaveCfg.TrashPath,
			time.Now().Format("2006-01-02"),
		)
		okNum := mvFilesToTrans(destDirAddr, files)

		// 成功应答, 报告成功删除的个数
		utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespSuccess}, gin.H{
			"count": okNum,
		})
		return
	}
}

//// UpdateFileInfo update file-info(PUT api/v1/file-infos/:id)
//// request body has the new file-information. See models.FileInfoCanBeUpdated.
//// In fact, only the filename can be updated now.
//func UpdateFileInfo(c *gin.Context) {
//	c.Writer.WriteHeader(http.StatusNotFound)
//}

//// ReUploadFile upload file again(POST api/v1/files/:id)
//// delete the old file and save the new file
//func ReUploadFile(c *gin.Context) {
//	c.Writer.WriteHeader(http.StatusNotFound)
//}

// GetFileInfo get a file information(GET api/v1/file-infos/:id)
func GetFileInfo(c *gin.Context) {
	// 获取file_id和user_id
	fileId, userId, ok := getFileIdAndUserId(c)
	if !ok {
		return
	}

	// 根据两个id查询文件信息并返回
	fileM, status := db.FindFileOfUserByID(userId, fileId)
	if !status.Success() {
		utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespFileNotFound}, nil)
		return
	}

	utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespSuccess}, models.FileInfoCanBePublicized{
		ID:        fileM.ID,
		CreatedAt: fileM.CreatedAt.Time,
		UpdatedAt: fileM.UpdatedAt.Time,
		DeletedAt: fileM.DeletedAt.Time,
		Filename:  fileM.Filename,
		Size:      fileM.Size,
		Location:  fileM.Location,
		Type:      fileM.Type,
	})
}

// GetFileInfosWithFilter get file with filter(GET api/v1/file-infos?)
//
// filters : | arg-name | meaning     | struct-attribute  | limit
//           | --------------------------------------------------------------
//           | id       | range       | ID                | int64-int64
//           | creat    | range       | CreatedAt         | YYYYMMDDhhmmss-YYYYMMDDhhmmss
//           | updat    | range       | UpdatedAt         | YYYYMMDDhhmmss-YYYYMMDDhhmmss
//           | delat    | range       | DeletedAt         | YYYYMMDDhhmmss-YYYYMMDDhhmmss
//           | fnamekey | keyword     | Filename(Keyword) | `^([._a-z0-9A-Z\p{Han}])*$`(bytes-size:[1, 255])
//           | size     | range       | Size              | int64-int64
//           | lockey   | range       | Location(Keyword) | `^([._a-z0-9A-Z\p{Han}])*$`(bytes-size:[1, 1024])
//           | type     | enum(f, d)  | Type              | `[fd]`
//           | locid    |   -------   |   -------         | int64
// range :  string like '<a>-<b>', bounded interval, the validity of a and b will be checked.
// keyword : string, keyword for searching, will be checked.
// All string will be trimmed.
// If query-args is invalid, the invalid argument will be ignored and the event will be logged.
func GetFileInfosWithFilter(c *gin.Context) {
	// 获取user-id
	uid, ok := c.Get("user_id")
	userId, ok := uid.(int64)
	if !ok {
		lg.Logger.Println("Get user_id Failed")
		utils.Response(c, http.StatusInternalServerError, errno.RespCode{Code: errno.RespFailed}, nil)
		return
	}

	queryArgs := struct {
		IDRange         string `form:"id"`
		CreatedAtRange  string `form:"creat"`
		UpdatedAtRange  string `form:"updat"`
		DeletedAtRange  string `form:"delat"`
		FilenameKeyword string `form:"fnamekey"`
		SizeRange       string `form:"size"`
		LocationKeyword string `form:"lockey"`
		TypeEnum        string `form:"type"`
		LocationID      int64  `form:"locid,default=-1"`
	}{}

	err := c.ShouldBindQuery(&queryArgs)
	if err != nil {
		utils.Response(c, http.StatusBadRequest, errno.RespCode{Code: errno.RespRequestFileTypeInvalid}, nil)
		return
	}

	//jsonStr, _ := json.Marshal(queryArgs)
	//lg.Logger.Println(string(jsonStr))

	// Don't Trim, Frontend Should Do It
	//queryArgs.IDRange = strings.TrimSpace(queryArgs.IDRange)
	//queryArgs.CreatedAtRange = strings.TrimSpace(queryArgs.CreatedAtRange)
	//queryArgs.UpdatedAtRange = strings.TrimSpace(queryArgs.UpdatedAtRange)
	//queryArgs.DeletedAtRange = strings.TrimSpace(queryArgs.DeletedAtRange)
	//queryArgs.FilenameKeyword = strings.TrimSpace(queryArgs.FilenameKeyword)
	//queryArgs.SizeRange = strings.TrimSpace(queryArgs.SizeRange)
	//queryArgs.LocationKeyword = strings.TrimSpace(queryArgs.LocationKeyword)
	//queryArgs.TypeEnum = strings.TrimSpace(queryArgs.TypeEnum)

	checkInt64Positive := func(i64 int64) bool { return i64 >= 0 }
	checkTime := func(t time.Time) bool { return true }
	splitter := "-"

	var queryArgsParsed models.FileInfoQueryArgs
	queryArgsParsed.IDRange = utils.GetInt64Range(queryArgs.IDRange, splitter, checkInt64Positive)
	queryArgsParsed.CreatedAtRange = utils.GetTimeRange(queryArgs.CreatedAtRange, splitter, checkTime)
	queryArgsParsed.UpdatedAtRange = utils.GetTimeRange(queryArgs.UpdatedAtRange, splitter, checkTime)
	queryArgsParsed.DeletedAtRange = utils.GetTimeRange(queryArgs.DeletedAtRange, splitter, checkTime)
	queryArgsParsed.SizeRange = utils.GetInt64Range(queryArgs.SizeRange, splitter, checkInt64Positive)

	// Check Validity Of TypeEnum
	if queryArgs.TypeEnum == "FILE" || queryArgs.TypeEnum == "DIR" {
		queryArgsParsed.TypeEnum = queryArgs.TypeEnum
	}

	// Check Validity Of FilenameKeyword
	regExpStr := `^([._a-z0-9A-Z\p{Han}])*$` // 文件_a1.jpg
	reg, err := regexp.Compile(regExpStr)
	if err != nil {
		utils.Response(c, http.StatusInternalServerError, errno.RespCode{Code: errno.RespFailed}, nil)
		lg.Logger.Printf("Compile Regexp %s Failed\n", regExpStr)
		return
	}
	byteSize := len([]byte(queryArgs.FilenameKeyword))
	if byteSize >= models.FilenameMinByteSize &&
		byteSize <= models.FilenameMaxByteSize &&
		reg.MatchString(queryArgs.FilenameKeyword) {

		queryArgsParsed.FilenameKeyword = queryArgs.FilenameKeyword
	}

	// Check Validity Of LocationKeyword
	regExpStr = `^([/._a-z0-9A-Z\p{Han}])*$` // 文件_a1.jpg
	reg, err = regexp.Compile(regExpStr)
	if err != nil {
		utils.Response(c, http.StatusInternalServerError, errno.RespCode{Code: errno.RespFailed}, nil)
		lg.Logger.Printf("Compile Regexp %s Failed\n", regExpStr)
		return
	}
	byteSize = len([]byte(queryArgs.LocationKeyword))
	if byteSize >= models.LocationMinByteSize &&
		byteSize <= models.LocationMaxByteSize &&
		reg.MatchString(queryArgs.LocationKeyword) {

		queryArgsParsed.LocationKeyword = queryArgs.LocationKeyword
	}

	// LocationID
	queryArgsParsed.LocationID = queryArgs.LocationID

	//jsonStr, _ = json.Marshal(queryArgsParsed)
	//lg.Logger.Println(string(jsonStr))

	fInfos, status := db.FindFileOfUserWithFilter(userId, &queryArgsParsed)
	if !status.Success() {
		utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespFailed}, nil)
		return
	}

	utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespSuccess}, fInfos)
}

// GetFilesInfoAtLocation get info of files at location(GET api/v1/file-infos/)
//
func GetFilesInfoAtLocation(c *gin.Context) {

}

// DownloadFile download a file(GET api/v1/files/:id)
func DownloadFile(c *gin.Context) {
	// 获取file-id和user-id
	fileId, userId, ok := getFileIdAndUserId(c)
	if !ok {
		return
	}

	// 根据file-id和use-id查询获取Filename和local_addr
	file, status := db.FindFileOfUserByID(userId, fileId)
	if !status.Success() || file.Type != models.FileType {
		utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespFileNotFound}, nil)
		return
	}
	filename := file.Filename
	localAddr := file.LocalAddr

	// 根据local_addr打开打开文件
	fHandle, err := os.Open(localAddr)
	if err != nil {
		utils.Response(c, http.StatusInternalServerError, errno.RespCode{Code: errno.RespFileNotFound}, nil)
		return
	}

	// 设置状态码
	c.Writer.WriteHeader(http.StatusOK)

	// Content-Disposition : 设置文件名字
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename = \"%s\"", url.QueryEscape(filename))) // FIXME 中文乱码

	// Content-Type : 下载文件, 二进制流即可
	c.Header("Content-Type", fileutils.GetHttpContentTypeByFilename(filename))

	// Accept-Length : 获取并设置文件长度
	c.Header("Accept-Length", fmt.Sprintf("%v", file.Size))

	// 读取文件并分批写入数据
	for {
		buffer := make([]byte, 4096)
		n, err := fHandle.Read(buffer)
		if err != nil || n == 0 {
			break
		}
		_, err = c.Writer.Write(buffer[0:n])
		if err != nil {
			break
		}
	}
	_ = fHandle.Close()
}

// For Big File

// StartUploadFile for upload big file
// init a file uploading
func StartUploadFile(c *gin.Context) {

}

// UploadFilePart upload part of file
func UploadFilePart(c *gin.Context) {

}

// EndUploadFile finish upload file
func EndUploadFile(c *gin.Context) {

}

// helper-functions
func getFileIdAndUserId(c *gin.Context) (int64, int64, bool) {
	// 获取file-id
	fileId := struct {
		ID int64 `uri:"id" binding:"required"`
	}{}
	err := c.ShouldBindUri(&fileId)
	if err != nil {
		utils.Response(c, http.StatusBadRequest, errno.RespCode{Code: errno.RespInvalidParams}, nil)
		return 0, 0, false
	}

	// 获取user-id
	uid, ok := c.Get("user_id")
	userId, ok := uid.(int64)
	if !ok {
		lg.Logger.Println("Get user_id Failed")
		utils.Response(c, http.StatusInternalServerError, errno.RespCode{Code: errno.RespFailed}, nil)
		return 0, 0, false
	}

	return fileId.ID, userId, true
}

func mvFilesToTrans(destDirAddr string, files []models.File) (cnt int64) {
	// 返回未成功删除的个数
	cnt = 0
	for _, f := range files {
		if f.Type == models.FileType {
			var srcAddr string = f.LocalAddr
			var destAddr string

			destFilename := strconv.FormatInt(f.ID, 10)
			destAddr = filepath.Join(destDirAddr, destFilename)
			err := os.MkdirAll(destDirAddr, 0755)
			if err != nil {
				lg.Logger.Printf("IMPORTANT : Move %s To Trans Failed, Please Check\n", f)
				lg.Logger.Printf("Server Error, Mkdir %s Failed When Calling DeleteFile\n", destDirAddr)
				continue
			}

			err = os.Rename(srcAddr, destAddr) // 移入回收站
			if err != nil {
				lg.Logger.Printf("IMPORTANT : Move %s To Trans Failed, Please Check\n", f)
				lg.Logger.Printf("Server Error, Move(Rename) %s Failed When Calling DeleteFile\n", destDirAddr)
				continue
			}
			cnt++
		}
	}
	return cnt
}
