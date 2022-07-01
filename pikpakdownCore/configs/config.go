package configs

import (
	"github.com/jinzhu/copier"
	"github.com/mrxtryagin/pikpakdown-api-go/pikpakdownCore/models"
	"github.com/mrxtryagin/pikpakdown-api-go/utils"
)

var (
	globalConfig = &GlobalConfig{}
)

//GlobalConfig 全局变量
type GlobalConfig struct {
	//access
	CaptchaToken string `json:"captcha_token"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Password     string `json:"password"`
	Username     string `json:"username"`
	TokenType    string `json:"token_type"`   //token认证方式
	AccessToken  string `json:"access_token"` //access_token
	RefreshToken string `json:"refresh_token"`
}

func InitGlobalConfig() *GlobalConfig {

	//初始化accessInfo
	InitAccessInfo(globalConfig)
	return globalConfig

}

func InitGlobalConfigFromMap(source map[string]interface{}) *GlobalConfig {
	//初始化accessInfo
	InitAccessInfo(globalConfig)

	err := copier.Copy(globalConfig, &source)
	if err != nil {
		panic(err)
	}
	return globalConfig

}

func InitGlobalConfigFromStruct(source *GlobalConfig) *GlobalConfig {
	//初始化accessInfo
	InitAccessInfo(globalConfig)

	err := copier.CopyWithOption(globalConfig, source, copier.Option{IgnoreEmpty: true})
	if err != nil {
		panic(err)
	}
	utils.Log().Info("configs:%+v", *globalConfig)
	return globalConfig

}

func InitAccessInfo(config *GlobalConfig) {
	config.ClientId = ClientId
	config.ClientSecret = ClientSecret
	config.CaptchaToken = CaptchaToken
	config.Username = ""
	config.Password = ""

}

func GetGlobalConfig() *GlobalConfig {
	return globalConfig
}

func UpdateToken(accessModel *models.AccessModel) error {
	utils.Log().Debug("udpate access info:\naccessToken:%s\nrefreshToken:%s\n", accessModel.AccessToken, accessModel.RefreshToken)
	//更新
	err := copier.Copy(globalConfig, accessModel)
	if err != nil {
		return err
	}
	return nil
}
