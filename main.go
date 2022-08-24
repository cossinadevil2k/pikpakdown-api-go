package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mrxtryagin/pikpakdown-api-go/httpHandler"
	"github.com/mrxtryagin/pikpakdown-api-go/pikpakdownCore"
	"github.com/mrxtryagin/pikpakdown-api-go/pikpakdownCore/configs"
	"github.com/mrxtryagin/pikpakdown-api-go/utils"
	"net/http"
)

func main() {
	configs.InitGlobalConfigFromStruct(&configs.GlobalConfig{
		Username: "mrx1998@126.com",
		Password: "980728qweR!",
	})
	err := pikpakdownCore.FirstLogin()
	if err != nil {
		utils.Log().Error("登陆时发生错误:%s", err.Error())
	}
	for {

	}
}

type SigninModel struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Sub          string `json:"sub"`
}

const (
	loginEndPoint = "https://user.mypikpak.com"
	apiEndPoint   = "https://api-drive.mypikpak.com"
	signin        = "v1/auth/signin"
	getFiles      = "/drive/v1/files"
)

//
func LoginApi() *SigninModel {
	defer utils.FunctionTimer("LoginApi")()

	//query := "client_id=YNxT9w7GMdWvEOKa"

	body := map[string]interface{}{
		"captcha_token": "",
		"client_id":     "YNxT9w7GMdWvEOKa",
		"client_secret": "dbw2OtmVEeuUvIptb1Coyg",
		"password":      "980728qweR!",
		"username":      "mrx1998@126.com",
	}
	bodyByte, _ := json.Marshal(body)
	header := http.Header{

		"User-Agent":   {"protocolversion/200 clientid/YNxT9w7GMdWvEOKa action_type/ networktype/WIFI sessionid/ devicesign/div101.073163586e9858ede866bcc9171ae3dcd067a68cbbee55455ab0b6096ea846a0 sdkversion/1.0.1.101300 datetime/1630669401815 appname/android-com.pikcloud.pikpak session_origin/ grant_type/ clientip/ devicemodel/LG V30 accesstype/ clientversion/ deviceid/073163586e9858ede866bcc9171ae3dc providername/NONE refresh_token/ usrno/null appid/ devicename/Lge_Lg V30 cmd/login osversion/9 platformversion/10 accessmode/"},
		"Content-Type": {"application/json; charset=utf-8"},
		"Host":         {"user.mypikpak.com"},
	}
	client := httpHandler.NewClient(httpHandler.WithHeader(header), httpHandler.WithEndpoint(loginEndPoint))
	query := httpHandler.GetQueryFromMap(map[string]string{"client_id": "YNxT9w7GMdWvEOKa"})
	resp, err := client.Post(
		signin,
		bytes.NewReader(bodyByte),
		httpHandler.WithQueryString(query),
	).CheckHttpStatusOk().GetResponse()
	if err != nil {
		utils.Log().Error("发生错误:%s", err.Error())
	}
	utils.Log().Info("返回:%s", string(resp))
	var result SigninModel
	err = json.Unmarshal(resp, &result)
	if err != nil {
		utils.Log().Error("解析struct出错:%s", err.Error())
	}
	return &result

}

func getFilesApi(tokenType, accessToken string) {
	defer utils.FunctionTimer("getFilesApi")()

	//query := "client_id=YNxT9w7GMdWvEOKa"

	queryMap := map[string]string{
		"parent_id":      "VN-Tex098kAAXJMdT4KBBBGJo1",
		"thumbnail_size": "SIZE_LARGE",
		"with_audit":     "true",
		"limit":          "100",
		"filters":        "{\"phase\":{\"eq\":\"PHASE_TYPE_COMPLETE\"},\"trashed\":{\"eq\":false}}",
	}
	header := http.Header{
		"Authorization": {tokenType + " " + accessToken},
		"Host":          {"api-drive.mypikpak.com"},
	}
	client := httpHandler.NewClient(httpHandler.WithHeader(header), httpHandler.WithEndpoint(apiEndPoint))
	query := httpHandler.GetQueryFromMap(queryMap)
	resp, err := client.Get(
		getFiles,
		nil,
		httpHandler.WithQueryString(query),
	).CheckHttpStatusOk().GetResponse()
	if err != nil {
		utils.Log().Error("发生错误:%s", err.Error())
	} else {
		utils.Log().Info("返回:%s", string(resp))
	}

}

func testGoogle() {
	client := httpHandler.NewClient()
	resp, err := client.Get(
		"https://www.google.com",
		nil,
	).CheckHttpStatusOk().GetResponse()
	if err != nil {
		utils.Log().Error("发生错误:%s", err.Error())
	} else {
		utils.Log().Info("返回:%s", string(resp))
	}
}

func testSukebei() *[]byte {
	client := httpHandler.NewClient()
	resp, err := client.Get(
		"https://sukebei.nyaa.si/",
		nil,
		httpHandler.WithProxy("http://127.0.0.1:7890"),
	).CheckHttpStatusOk().GetResponse()
	if err != nil {
		utils.Log().Error("发生错误:%s", err.Error())
	} else {
		utils.Log().Info("返回:%v", fmt.Sprint(string(resp)))
	}
	return &resp
}

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mrxtryagin/pikpakdown-api-go/httpHandler"
	"github.com/mrxtryagin/pikpakdown-api-go/pikpakdownCore"
	"github.com/mrxtryagin/pikpakdown-api-go/pikpakdownCore/configs"
	"github.com/mrxtryagin/pikpakdown-api-go/utils"
	"net/http"
)

func main() {
	configs.InitGlobalConfigFromStruct(&configs.GlobalConfig{
		Username: "mrx1998@126.com",
		Password: "980728qweR!",
	})
	err := pikpakdownCore.FirstLogin()
	if err != nil {
		utils.Log().Error("登陆时发生错误:%s", err.Error())
	}
	for {

	}
}

type SigninModel struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Sub          string `json:"sub"`
}

const (
	loginEndPoint = "https://user.mypikpak.com"
	apiEndPoint   = "https://api-drive.mypikpak.com"
	signin        = "v1/auth/signin"
	getFiles      = "/drive/v1/files"
)

//
func LoginApi() *SigninModel {
	defer utils.FunctionTimer("LoginApi")()

	//query := "client_id=YNxT9w7GMdWvEOKa"

	body := map[string]interface{}{
		"captcha_token": "",
		"client_id":     "YNxT9w7GMdWvEOKa",
		"client_secret": "dbw2OtmVEeuUvIptb1Coyg",
		"password":      "980728qweR!",
		"username":      "mrx1998@126.com",
	}
	bodyByte, _ := json.Marshal(body)
	header := http.Header{

		"User-Agent":   {"protocolversion/200 clientid/YNxT9w7GMdWvEOKa action_type/ networktype/WIFI sessionid/ devicesign/div101.073163586e9858ede866bcc9171ae3dcd067a68cbbee55455ab0b6096ea846a0 sdkversion/1.0.1.101300 datetime/1630669401815 appname/android-com.pikcloud.pikpak session_origin/ grant_type/ clientip/ devicemodel/LG V30 accesstype/ clientversion/ deviceid/073163586e9858ede866bcc9171ae3dc providername/NONE refresh_token/ usrno/null appid/ devicename/Lge_Lg V30 cmd/login osversion/9 platformversion/10 accessmode/"},
		"Content-Type": {"application/json; charset=utf-8"},
		"Host":         {"user.mypikpak.com"},
	}
	client := httpHandler.NewClient(httpHandler.WithHeader(header), httpHandler.WithEndpoint(loginEndPoint))
	query := httpHandler.GetQueryFromMap(map[string]string{"client_id": "YNxT9w7GMdWvEOKa"})
	resp, err := client.Post(
		signin,
		bytes.NewReader(bodyByte),
		httpHandler.WithQueryString(query),
	).CheckHttpStatusOk().GetResponse()
	if err != nil {
		utils.Log().Error("发生错误:%s", err.Error())
	}
	utils.Log().Info("返回:%s", string(resp))
	var result SigninModel
	err = json.Unmarshal(resp, &result)
	if err != nil {
		utils.Log().Error("解析struct出错:%s", err.Error())
	}
	return &result

}

func getFilesApi(tokenType, accessToken string) {
	defer utils.FunctionTimer("getFilesApi")()

	//query := "client_id=YNxT9w7GMdWvEOKa"

	queryMap := map[string]string{
		"parent_id":      "VN-Tex098kAAXJMdT4KBBBGJo1",
		"thumbnail_size": "SIZE_LARGE",
		"with_audit":     "true",
		"limit":          "100",
		"filters":        "{\"phase\":{\"eq\":\"PHASE_TYPE_COMPLETE\"},\"trashed\":{\"eq\":false}}",
	}
	header := http.Header{
		"Authorization": {tokenType + " " + accessToken},
		"Host":          {"api-drive.mypikpak.com"},
	}
	client := httpHandler.NewClient(httpHandler.WithHeader(header), httpHandler.WithEndpoint(apiEndPoint))
	query := httpHandler.GetQueryFromMap(queryMap)
	resp, err := client.Get(
		getFiles,
		nil,
		httpHandler.WithQueryString(query),
	).CheckHttpStatusOk().GetResponse()
	if err != nil {
		utils.Log().Error("发生错误:%s", err.Error())
	} else {
		utils.Log().Info("返回:%s", string(resp))
	}

}

func testGoogle() {
	client := httpHandler.NewClient()
	resp, err := client.Get(
		"https://www.google.com",
		nil,
	).CheckHttpStatusOk().GetResponse()
	if err != nil {
		utils.Log().Error("发生错误:%s", err.Error())
	} else {
		utils.Log().Info("返回:%s", string(resp))
	}
}

func testSukebei() *[]byte {
	client := httpHandler.NewClient()
	resp, err := client.Get(
		"https://sukebei.nyaa.si/",
		nil,
		httpHandler.WithProxy("http://127.0.0.1:7890"),
	).CheckHttpStatusOk().GetResponse()
	if err != nil {
		utils.Log().Error("发生错误:%s", err.Error())
	} else {
		utils.Log().Info("返回:%v", fmt.Sprint(string(resp)))
	}
	return &resp
}
