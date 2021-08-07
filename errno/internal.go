package errno

const (
	Success = iota
	UserAddFailed
	UserNotFound
	UserNamePwdNotMatched
	UserNameRepeated
	UserInfoGetFailed
	FileNotFound
	FileAddFailed
	FileDeleteFailed
	FileUpdateFailed
	FileFindOfUserFailed
	FileFindOfUserByNameFailed
	FileAddRepeated
	CreateTokenFailed
	ParseTokenFailed
)

var statusMsg []string = []string{
	"Success",
	"UserAddFailed",
	"UserNotFound",
	"UserNamePwdNotMatched",
	"UserNameRepeated",
	"UserInfoGetFailed",
	"FileNotFound",
	"FileAddFailed",
	"FileDeleteFailed",
	"FileUpdateFailed",
	"FileFindOfUserFailed",
	"FileFindOfUserByNameFailed",
	"FileAddRepeated",
	"CreateTokenFailed",
	"ParseTokenFailed",
}

// Status 表示内部状态(错误)
type Status struct {
	Code uint32
}

func (s *Status) Msg() string {
	return statusMsg[s.Code]
}

func (s *Status) Success() bool {
	return s.Code == Success
}
