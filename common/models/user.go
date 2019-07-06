package models

type User struct{
	UserID int `json:"userid"`
	UserPwd string `json:"userpwd"`
	UserName string `json:"username"`
}