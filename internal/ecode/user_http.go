package ecode

import (
	"github.com/go-dev-frame/sponge/pkg/errcode"
)

// user business-level http error codes.
// the userNO value range is 1~999, if the same error code is used, it will cause panic.
var (
	userNO       = 78
	userName     = "user"
	userBaseCode = errcode.HCode(userNO)

	ErrCreateUser     = errcode.NewError(userBaseCode+1, "failed to create "+userName)
	ErrDeleteByIDUser = errcode.NewError(userBaseCode+2, "failed to delete "+userName)
	ErrUpdateByIDUser = errcode.NewError(userBaseCode+3, "failed to update "+userName)
	ErrGetByIDUser    = errcode.NewError(userBaseCode+4, "failed to get "+userName+" details")
	ErrListUser       = errcode.NewError(userBaseCode+5, "failed to list of "+userName)

	ErrDeleteByIDsUser    = errcode.NewError(userBaseCode+6, "failed to delete by batch ids "+userName)
	ErrGetByConditionUser = errcode.NewError(userBaseCode+7, "failed to get "+userName+" details by conditions")
	ErrListByIDsUser      = errcode.NewError(userBaseCode+8, "failed to list by batch ids "+userName)
	ErrListByLastIDUser   = errcode.NewError(userBaseCode+9, "failed to list by last id "+userName)

	ErrUsernameAlreadyExists = errcode.NewError(userBaseCode+10, userName+" username already exists")
	ErrUserNotExists         = errcode.NewError(userBaseCode+11, userName+" not exists")
	ErrPassword              = errcode.NewError(userBaseCode+12, userName+" password error")
	ErrNotLogin              = errcode.NewError(userBaseCode+13, userName+" not login")

	// error codes are globally unique, adding 1 to the previous error code
)
