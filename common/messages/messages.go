package messages

import (
	"chattingroom/common/models"
)

const (
	LoginMessageType                  = "LoginMessage"
	LoginResponseMessageType          = "LoginResponseMessage"
	RegistryMessageType               = "RegistryMessage"
	RegistryReponseMessageType        = "RegistyResponseMessage"
	DaemonRequestMessageType          = "DaemonRequestMessage"
	DaemonResponseMessageType         = "DaemonResponseMessage"
	GetOnlineUsersMessageType         = "GetOnlineUsersMessage"
	GetOnlineUsersResponseMessageType = "GetOnlineUsersResponseMessage"
	LogoutMessageType                 = "LogoutMessage"
	UserStateChangeMessageType        = "UserStateChangeMessage"
	ShortMessageSenderMessageType     = "ShortMessageSenderMessage"
	ShortMessageReceiverMessageType   = "ShortMessageReceiverMessage"
)

// Message 通用消息类型
type Message struct {
	Type string `json:"type"` // 消息类型
	Data string `json:"data"` // 消息数据
}

// LoginMessage 登录消息
type LoginMessage struct {
	UserID  int    `json:"userid"`  // 用户id
	UserPwd string `json:"userpwd"` // 用户密码
}

// LoginResponseMessage 登录响应消息
type LoginResponseMessage struct {
	Code  int    `json:"code"`  // 状态码
	Error string `json:"error"` // 错误信息
}

// RegistryMessage 注册消息
type RegistryMessage struct {
	models.User
}

// RegistyResponseMessage 注册响应消息
type RegistyResponseMessage struct {
	Code  int    `json:"code"`  // 状态码
	Error string `json:"error"` // 错误信息
}

// DaemonRequestMessage Daemon消息
type DaemonRequestMessage struct {
	UserID int `json:"userid"`
}

// DaemonResponseMessage Daemon响应消息
type DaemonResponseMessage struct {
	Code  int    `json:"code"`  // 状态码
	Error string `json:"error"` // 错误信息
}

// GetOnlineUsersMessage 获取在线用户列表消息
type GetOnlineUsersMessage struct {
	UserID int `json:"userid"`
}

// GetOnlineUsersResponseMessage 获取在线用户列表响应消息
type GetOnlineUsersResponseMessage struct {
	Code        int         `json:"code"`        // 状态码
	Error       string      `json:"error"`       // 错误信息
	UserID      int         `json:"userid"`      // 用户ID
	OnlineUsers map[int]int `json:"onlineusers"` // 在线用户ID
}

// LogoutMessage 登出消息
type LogoutMessage struct {
	UserID int `json:"userid"` // 用户ID
}

const (
	UserOnline = 2 << iota
	UserOffline
)

// UserStateChangeMessage 用户状态改变消息
type UserStateChangeMessage struct {
	UserID    int `json:"userid"`    // 用户ID
	UserState int `json:"userstate"` // 用户状态码
}

const (
	ToAll = 1 << iota
	ToUsers
	ToUserGroups
)

// ShortMessage 短消息发送方
type ShortMessageSenderMessage struct {
	Scope    int           `json:"scope"`    // 收信范围
	SrcUser  models.User   `json:"srcuser"`  // 发送方用户ID
	DstUsers []models.User `json:"dstusers"` // 接收方用户ID
	// DstUserGroups []models.UserGroup `json:"dstusergroups"` // 接收方用户组ID
	Content string `json:"content"` //内容
}

// ShortMessageReceiverMessage 短消息接收方
type ShortMessageReceiverMessage struct {
	Scope    int           `json:"scope"`    // 收信范围
	SrcUser  models.User   `json:"srcuser"`  // 发送方用户ID
	DstUsers []models.User `json:"dstusers"` // 接收方用户ID
	// DstUserGroups []models.UserGroup `json:"dstusergroups"` // 接收方用户组ID
	Content string `json:"content"` //内容
}
