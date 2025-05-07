package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateUserRequest request params
type CreateUserRequest struct {
	Username string `json:"username" binding:""`
	Nickname string `json:"nickname" binding:""`
	Password string `json:"password" binding:""`
}

// LoginRequest request params
type LoginRequest struct {
	Username string `json:"username" binding:""`
	Password string `json:"password" binding:""`
}

// UpdateUserByIDRequest request params
type UpdateUserByIDRequest struct {
	Username string `json:"username" binding:""`
	Nickname string `json:"nickname" binding:""`
}

// UpdatePasswordByIDRequest request params
type UpdatePasswordByIDRequest struct {
	Password string `json:"password" binding:""`
}

// UserObjDetail detail
type UserObjDetail struct {
	ID uint64 `json:"id,string"` // convert to uint64 id

	Username  string     `json:"username"`
	Nickname  string     `json:"nickname"`
	LoginAt   *time.Time `json:"loginAt"`
	LoginIP   string     `json:"loginIP"`
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

type TokenObjDetail struct {
	ID    uint64 `json:"id,string"`
	Token string `json:"token"`
}

// CreateUserReply only for api docs
type CreateUserReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		User UserObjDetail `json:"user"`
	} `json:"data"` // return data
}

type LoginReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		TokenObjDetail
	} `json:"data"` // return data
}

// UpdateUserByIDReply only for api docs
type UpdateUserByIDReply struct {
	Result
}

// UpdatePasswordByIDReply only for api docs
type UpdatePasswordByIDReply struct {
	Result
}

// GetUserByIDReply only for api docs
type GetUserByIDReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		User UserObjDetail `json:"user"`
	} `json:"data"` // return data
}

// DeleteUserByIDReply only for api docs
type DeleteUserByIDReply struct {
	Result
}

// DeleteUsersByIDsReply only for api docs
type DeleteUsersByIDsReply struct {
	Result
}

// ListUsersRequest request params
type ListUsersRequest struct {
	query.Params
}

// ListUsersReply only for api docs
type ListUsersReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		Users []UserObjDetail `json:"users"`
	} `json:"data"` // return data
}

// DeleteUsersByIDsRequest request params
type DeleteUsersByIDsRequest struct {
	IDs []uint64 `json:"ids" binding:"min=1"` // id list
}

// GetUserByConditionRequest request params
type GetUserByConditionRequest struct {
	query.Conditions
}

// GetUserByConditionReply only for api docs
type GetUserByConditionReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		User UserObjDetail `json:"user"`
	} `json:"data"` // return data
}

// ListUsersByIDsRequest request params
type ListUsersByIDsRequest struct {
	IDs []uint64 `json:"ids" binding:"min=1"` // id list
}

// ListUsersByIDsReply only for api docs
type ListUsersByIDsReply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		Users []UserObjDetail `json:"users"`
	} `json:"data"` // return data
}
