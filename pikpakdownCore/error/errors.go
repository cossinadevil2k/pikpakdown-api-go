package error

import "errors"

//错误类

var (
	AccessTokenNotExistErr = errors.New("access token not exist")

	RefreshTokenNotExistErr = errors.New("refresh token not exist")
)
