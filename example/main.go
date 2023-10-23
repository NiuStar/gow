package main

import "github.com/NiuStar/gow"

type TestRequest struct {
	UserName string `json:"user_name" 备注:"用户名" 默认值:"admin"`
	Password string `json:"password" 备注:"密码" 默认值:"123456"`
}
type TestResponse struct {
	UserId     string `json:"user_id" 备注:"用户id" 默认值:"1"`
	UserName   string `json:"user_name" 备注:"用户名" 默认值:"admin"`
	CreateTime string `json:"create_time" 备注:"创建时间" 默认值:"2020-01-01 00:00:00"`
}

func Login(c *gow.Context) {
	var request TestRequest
	c.DecodeJSONBody(&request)
}
func main() {
	r := gow.Default("测试登录项目2")
	apiGroup := r.Group("/api", nil, nil)
	testGroup := apiGroup.Group("/test", nil, nil)
	testGroup.Handle("GET,POST", "/login", &TestRequest{}, &TestResponse{}, Login)
	r.Run("test", "123456", "8081")
}
