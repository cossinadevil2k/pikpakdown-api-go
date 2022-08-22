package pikpakdownCore

import (
	"github.com/mrxtryagin/pikpakdown-api-go/pikpakdownCore/configs"
	"github.com/mrxtryagin/pikpakdown-api-go/pikpakdownCore/service"
)

//第一次登陆
func FirstLogin() error {
	login, err := service.Login()
	if err != nil {
		return err
	}
	//fmt.Printf("login_obj: %v\n", login)
	//updateToken
	err = configs.UpdateToken(login)
	if err != nil {
		return err
	}
	//开启激活
	go Active()

	return nil
}
