package errno

const (
	RespSuccess = iota
	RespFailed
	RespInvalidParams
	RespAuthUserNotFound
	RespAuthUserNamePwdNotMatched
	RespTokenCheckFailed
	RespGetUserInfoFailed
	RespRequestFileTypeInvalid
	RespRequestFilenameInvalid
	RespRequestLocationInvalid
	RespVirtualPathNotFound
	RespAddFileFailed
	RespSaveFileFailed
	RespAddFileRepeated
)

var respCodeMsg []string = []string{
	"Response Successfully",
	"RespFailed",
	"Invalid Request Params",
	"用户不存在",
	"密码错误",
	"Token Checking Failed",
	"获取用户信息失败",
	"文件类型错误",
	"文件名不合法",
	"文件路径不合法",
	"虚拟路径不存在",
	"文件添加失败",
	"文件存储失败",
	"添加同名文件或路径",
}

type RespCode struct {
	Code uint32
}

func (s *RespCode) Msg() string {
	return respCodeMsg[s.Code]
}

func (s *RespCode) Success() bool {
	return s.Code == RespSuccess
}
