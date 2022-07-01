package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/mrxtryagin/pikpakdown-api-go/httpHandler"
	"github.com/mrxtryagin/pikpakdown-api-go/pikpakdownCore/configs"
	error2 "github.com/mrxtryagin/pikpakdown-api-go/pikpakdownCore/error"
	"github.com/mrxtryagin/pikpakdown-api-go/pikpakdownCore/models"
	"github.com/mrxtryagin/pikpakdown-api-go/utils"
	"net/http"
)

// 用户相关模块

var userHeader = http.Header{

	"User-Agent": {configs.UserAgent1},
	"Host":       {configs.LoginEndPoint},
}

//GetCommonUserRequest 通用用户client
func GetCommonUserRequest() *httpHandler.HTTPClient {
	userEndPoint := fmt.Sprintf("%s://%s", configs.Protocol, configs.LoginEndPoint)
	client := httpHandler.NewClient(httpHandler.WithHeader(userHeader), httpHandler.WithEndpoint(userEndPoint))
	return client
}

//Login 登陆
func Login() (*models.AccessModel, error) {
	// 获得通用的client
	client := GetCommonUserRequest()
	// client
	query := httpHandler.GetQueryFromMap(map[string]string{"client_id": configs.ClientId})
	config := configs.GetGlobalConfig()
	requestBody := &models.GetAccessRequest{}
	err := copier.Copy(requestBody, config)
	if err != nil {
		return nil, err
	}
	bodyByte, _ := json.Marshal(requestBody)
	resp, err := client.Post(
		configs.Signin,
		bytes.NewReader(bodyByte),
		httpHandler.WithQueryString(query),
	).CheckHttpStatusOk().GetResponse()
	//AccessModel 返回
	var result models.AccessModel
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}

	return &result, err

}

//RefreshToken 刷新token
func RefreshToken() (*models.AccessModel, error) {
	config := configs.GetGlobalConfig()
	if config.RefreshToken == "" {
		return nil, error2.RefreshTokenNotExistErr
	}

	// 获得通用的client
	client := GetCommonUserRequest()
	// client
	query := httpHandler.GetQueryFromMap(map[string]string{"client_id": config.ClientId})
	requestBody := &models.RefreshTokenRequest{
		ClientId:     config.ClientId,
		ClientSecret: config.ClientSecret,
		GrantType:    "refresh_token", //
		RefreshToken: config.RefreshToken,
	}

	bodyByte, _ := json.Marshal(requestBody)
	resp, err := client.Post(
		configs.RefreshToken,
		bytes.NewReader(bodyByte),
		httpHandler.WithQueryString(query),
	).CheckHttpStatusOk().GetResponse()
	//AccessModel 返回
	var result models.AccessModel
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}

	return &result, err

}

func UpdateUserToken() bool {
	config := configs.GetGlobalConfig()
	if config.AccessToken == "" {
		utils.Log().Debug("login...")
		//访问
		login, err := Login()
		if err != nil {
			return false
		}
		configs.UpdateToken(login)
	} else {
		//更新
		utils.Log().Debug("refreshToken...")
		token, err := RefreshToken()
		if err != nil {
			return false
		}
		configs.UpdateToken(token)
	}
	return true
}
