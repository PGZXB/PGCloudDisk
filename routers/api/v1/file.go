package v1

import (
	"PGCloudDisk/config"
	"PGCloudDisk/db"
	"PGCloudDisk/errno"
	"PGCloudDisk/models"
	"PGCloudDisk/utils"
	"PGCloudDisk/utils/fileutils"
	"PGCloudDisk/utils/lg"
	"fmt"
	"github.com/flytam/filenamify"
	"github.com/gin-gonic/gin"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// UploadFile upload a file(POST api/v1/files)
// only allow upload small file, content is in request-form "file"
// meta-info least : Filename(filename) LocationID(location_id) Type(type)
// response data : JSON(FileInfoCanBePublished)
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

	// 如果是File是否有"file"
	var fileHeader *multipart.FileHeader
	if fInfo.Type == models.FileType {
		fileHeader, err = c.FormFile("file")
		if err != nil {
			utils.Response(c, http.StatusBadRequest, errno.RespCode{Code: errno.RespInvalidParams}, nil)
			return
		}
	}

	// 验证Type的合法性
	if fInfo.Type != models.FileType && fInfo.Type != models.DirType {
		utils.Response(c, http.StatusBadRequest, errno.RespCode{Code: errno.RespRequestFileTypeInvalid}, nil)
		return
	}

	// 验证Filename的合法性, Filename非法则尝试变成合法的
	fInfo.Filename, err = filenamify.Filenamify(fInfo.Filename, filenamify.Options{
		// FIXME 至少在Windows上有BUG(a.txt.* -> a.txt.,
		// FIXME 但是Win上存储会自动转为a.txt 导致数据库与文件系统中不一样)
		Replacement: "_",
		MaxLength:   100,
	})
	if err != nil {
		utils.Response(c, http.StatusBadRequest, errno.RespCode{Code: errno.RespRequestFilenameInvalid}, nil)
		return
	}

	// 生成文件信息
	//     Size      int64
	//     LocalAddr localSaveRoot + / + UserID + / + LocationID + / + timestamp + _ + Filename
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

	if fileModel.Type == models.FileType {
		fileModel.LocalAddr = filepath.Join(
			config.Cfg.LocalSaveCfg.Root,
			strconv.FormatInt(fileModel.UserID, 10),
			strconv.FormatInt(fInfo.LocationID, 10),
			fmt.Sprintf("%s_%s_%s",
				strconv.FormatInt(time.Now().UnixNano(), 16),
				strconv.FormatInt(rand.Int63n(1e8), 16),
				fileModel.Filename),
		)
	}

	if fileModel.Type == models.FileType {
		fileModel.Size = fileHeader.Size
	}

	// 读取文件并保存到本地, 如果是目录无需此步
	if fInfo.Type == models.FileType {
		path, _ := filepath.Split(fileModel.LocalAddr)
		if !fileutils.IsDir(path) {
			err = os.MkdirAll(path, 0755)
			if err != nil {
				lg.Logger.Printf("Create Dir %s When Saving File %s Failed\n", path, fileModel.LocalAddr)
				utils.Response(c, http.StatusInternalServerError, errno.RespCode{Code: errno.RespSaveFileFailed}, nil)
				return
			}
		}
		err = c.SaveUploadedFile(fileHeader, fileModel.LocalAddr)
		if err != nil {
			lg.Logger.Printf("Save File %s Failed\n", fileModel.LocalAddr)
			utils.Response(c, http.StatusInternalServerError, errno.RespCode{Code: errno.RespSaveFileFailed}, nil)
			return
		}
	}

	// 存入数据库
	s := db.AddFileAtLocation(&fileModel, fInfo.LocationID)
	if s.Code == errno.FileNotFound {
		utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespVirtualPathNotFound}, nil)
	} else if s.Code == errno.FileAddFailed {
		utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespAddFileFailed}, nil)
	} else if s.Code == errno.FileAddRepeated {
		utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespAddFileRepeated}, nil)
	}

	// 存入数据库不成功从磁盘中删除
	if fileModel.Type == models.FileType && !s.Success() {
		lg.Logger.Printf("Add File Failed, Deleting File %s\n", fileModel.LocalAddr)
		go func(path string, needDeleteDir bool) {
			err := os.Remove(path)
			if err != nil {
				lg.Logger.Printf("Delete File %s Failed\n", path)
			} else {
				lg.Logger.Printf("Delete File %s Successfully\n", path)
				if needDeleteDir { // 如果以前没有这条路径, 就删除
					parentPath, _ := filepath.Split(path)
					_ = os.Remove(parentPath)
					lg.Logger.Printf("Delete Parent Dir %s Successfully\n", parentPath)
				}
			}
		}(fileModel.LocalAddr, s.Code == errno.FileNotFound)
	} else { // 返回信息
		res, s := db.FindFilesOfUserByName(fileModel.UserID, fileModel.Filename)
		if !s.Success() {
			utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespFailed}, models.FileInfoCanBePublished{
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
		utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespSuccess}, models.FileInfoCanBePublished{
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

// DeleteFiles delete file(DELETE api/v1/files/:id)
// soft remove file-info and move the file in trash
func DeleteFiles(c *gin.Context) {

}

// UpdateFileInfo update file-info(PUT api/v1/file-infos/:id)
// request body has the new file-information. See models.FileInfoCanBeUpdated.
// In fact, only the filename can be updated now.
func UpdateFileInfo(c *gin.Context) {

}

// ReUploadFile upload file again(POST api/v1/files/:id)
// delete the old file and save the new file
func ReUploadFile(c *gin.Context) {

}

// GetFileInfo get a file information(GET api/v1/file-infos/:id)
func GetFileInfo(c *gin.Context) {

}

// GetFileInfosWithFilter get file with filter(GET api/v1/file-infos?)
// filter : min-id max-id id                    : ID
//          cre-at-start cre-at-end cre-at      : CreatedAt
//          u-at-start u-at-end u-at            : UpdatedAt
//          d-at-start u-at-end d-at            : DeletedAt
//          fname-key                           : Filename-Keyword
//          min-size max-size size              : Size
//          loc-key                             : Location-Keyword
//          type                                : Type
func GetFileInfosWithFilter(c *gin.Context) {

}

// DownloadFile download a file(GET api/v1/files/:id)
func DownloadFile(c *gin.Context) {

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
