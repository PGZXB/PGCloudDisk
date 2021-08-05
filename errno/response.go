package errno

const (
	RespSuccess = iota
	RespFailed
	RespInvalidParams
	RespAuthUserNotFound
	RespAuthUserNamePwdNotMatched
	RespTokenCheckFailed
)

var respCodeMsg []string = []string{
	"Response Successfully",
	"RespFailed",
	"InvalidParams",
	"用户不存在",
	"密码错误",
	"Token Checking Failed",
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
