package models

import "time"

//访问用结构体
type AccessModel struct {
	TokenType    string `json:"token_type"`    //token认证方式
	AccessToken  string `json:"access_token"`  //access_token
	RefreshToken string `json:"refresh_token"` //refresh_token
	ExpiresIn    int    `json:"expires_in"`    //失效时间
	Sub          string `json:"sub"`           //截取
}

// 用户模型
type UserModel struct {
	Sub               string    `json:"sub"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	Password          string    `json:"password"`
	CreatedAt         time.Time `json:"created_at"`
	PasswordUpdatedAt time.Time `json:"password_updated_at"`
}

// 文件或者文件夹模型
type ObjectModel struct {
	Kind           string    `json:"kind"`
	Id             string    `json:"id"`        // id
	ParentId       string    `json:"parent_id"` //父id
	Name           string    `json:"name"`      //名称
	UserId         string    `json:"user_id"`
	Size           string    `json:"size"`
	Revision       string    `json:"revision"`
	FileExtension  string    `json:"file_extension"`
	MimeType       string    `json:"mime_type"`
	Starred        bool      `json:"starred"`
	WebContentLink string    `json:"web_content_link"` // link
	CreatedTime    time.Time `json:"created_time"`
	ModifiedTime   time.Time `json:"modified_time"`
	IconLink       string    `json:"icon_link"`
	ThumbnailLink  string    `json:"thumbnail_link"`
	Md5Checksum    string    `json:"md5_checksum"`
	Hash           string    `json:"hash"`
	Links          struct {
		ApplicationOctetStream struct {
			Url    string    `json:"url"`
			Token  string    `json:"token"`
			Expire time.Time `json:"expire"`
		} `json:"application/octet-stream"`
	} `json:"links"`
	Phase string `json:"phase"`
	Audit struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Title   string `json:"title"`
	} `json:"audit"`
	//Medias []struct {
	//	MediaId   string `json:"media_id"`
	//	MediaName string `json:"media_name"`
	//	Video     struct {
	//		Height     int    `json:"height"`
	//		Width      int    `json:"width"`
	//		Duration   int    `json:"duration"`
	//		BitRate    int    `json:"bit_rate"`
	//		FrameRate  int    `json:"frame_rate"`
	//		VideoCodec string `json:"video_codec"`
	//		AudioCodec string `json:"audio_codec"`
	//		VideoType  string `json:"video_type"`
	//	} `json:"video"`
	//	Link struct {
	//		Url    string    `json:"url"`
	//		Token  string    `json:"token"`
	//		Expire time.Time `json:"expire"`
	//	} `json:"link"`
	//	NeedMoreQuota  bool          `json:"need_more_quota"`
	//	VipTypes       []interface{} `json:"vip_types"`
	//	RedirectLink   string        `json:"redirect_link"`
	//	IconLink       string        `json:"icon_link"`
	//	IsDefault      bool          `json:"is_default"`
	//	Priority       int           `json:"priority"`
	//	IsOrigin       bool          `json:"is_origin"`
	//	ResolutionName string        `json:"resolution_name"`
	//	IsVisible      bool          `json:"is_visible"`
	//	Category       string        `json:"category"`
	//} `json:"medias"`
	Trashed     bool   `json:"trashed"`
	DeleteTime  string `json:"delete_time"`
	OriginalUrl string `json:"original_url"`
	Params      struct {
		Duration     string `json:"duration"`
		Height       string `json:"height"`
		PlatformIcon string `json:"platform_icon"`
		Url          string `json:"url"`
		Width        string `json:"width"`
	} `json:"params"`
	OriginalFileIndex int           `json:"original_file_index"`
	Space             string        `json:"space"`
	Apps              []interface{} `json:"apps"`
	Writable          bool          `json:"writable"`
	FolderType        string        `json:"folder_type"`
	Collection        interface{}   `json:"collection"`
}

//目录模型
type ObjectsModel struct {
	Kind            string        `json:"kind"`
	NextPageToken   string        `json:"next_page_token"`
	Files           []ObjectModel `json:"files"`
	Version         string        `json:"version"`
	VersionOutdated bool          `json:"version_outdated"`
}

//任务模型
type TaskModel struct {
	Kind       string        `json:"kind"`
	Id         string        `json:"id"`
	Name       string        `json:"name"`
	Type       string        `json:"type"` //task类型 离线任务还是?
	UserId     string        `json:"user_id"`
	Statuses   []interface{} `json:"statuses"`
	StatusSize int           `json:"status_size"`
	Params     struct {
		MimeType            string `json:"mime_type,omitempty"`
		PredictType         string `json:"predict_type,omitempty"`
		Url                 string `json:"url"`
		ErrorDetail         string `json:"error_detail,omitempty"`
		FailedStatusCount   string `json:"failed_status_count,omitempty"`
		FilteredStatusCount string `json:"filtered_status_count,omitempty"`
		PredictSpeed        string `json:"predict_speed,omitempty"`
	} `json:"params"`
	FileId            string      `json:"file_id"`
	FileName          string      `json:"file_name"`
	FileSize          string      `json:"file_size"`    //文件总大小
	Message           string      `json:"message"`      //信息
	CreatedTime       time.Time   `json:"created_time"` //创建时间
	UpdatedTime       time.Time   `json:"updated_time"` //更新时间
	ThirdTaskId       string      `json:"third_task_id"`
	Phase             string      `json:"phase"`
	Progress          int         `json:"progress"`
	IconLink          string      `json:"icon_link"`
	Callback          string      `json:"callback"`
	ReferenceResource ObjectModel `json:"reference_resource"` //对应的文件信息就是 ObjectModel信息
	Space             string      `json:"space"`
}

//任务列表模型
type TasksModel struct {
	Tasks         []TaskModel `json:"tasks"`
	NextPageToken string      `json:"next_page_token"`
	ExpiresIn     int         `json:"expires_in"`
}

// 上传信息 比如 提交离线文件的返回值
type UploadInfoModel struct {
	UploadType string `json:"upload_type"`
	Url        struct {
		Kind string `json:"kind"`
	} `json:"url"`
	File interface{} `json:"file"`
	Task TaskModel   `json:"task"`
}
