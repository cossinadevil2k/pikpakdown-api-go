package configs

const (
	//deviceId
	ClientId     = "YNxT9w7GMdWvEOKa"
	ClientSecret = "dbw2OtmVEeuUvIptb1Coyg"

	//captchaToken
	CaptchaToken = ""

	//userAgent 安卓格式
	UserAgent1 = "protocolversion/200 clientid/YNxT9w7GMdWvEOKa action_type/ networktype/WIFI sessionid/ devicesign/div101.073163586e9858ede866bcc9171ae3dcd067a68cbbee55455ab0b6096ea846a0 sdkversion/1.0.1.101300 datetime/1630669401815 appname/android-com.pikcloud.pikpak session_origin/ grant_type/ clientip/ devicemodel/LG V30 accesstype/ clientversion/ deviceid/073163586e9858ede866bcc9171ae3dc providername/NONE refresh_token/ usrno/null appid/ devicename/Lge_Lg V30 cmd/login osversion/9 platformversion/10 accessmode/"
	UserAgent2 = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36 Edg/100.0.1185.44"
	// url
	LoginEndPoint = "user.mypikpak.com"
	Protocol      = "https"
	ApiEndPoint   = "api-drive.mypikpak.com"

	// api
	Signin       = "/v1/auth/signin" //登陆
	RefreshToken = "/v1/auth/token"  //刷新
	GetFiles     = "/drive/v1/files" // 获得单个文件 或者目录
	Me           = "/v1/user/me"     //获得个人信息
	GetTasks     = "/drive/v1/tasks" // 获得任务列表 或者单个任务情况

	//context-type
	ContentType = "application/json; charset=utf-8"
)
