package models

/*GetAccessRequest
{
 "captcha_token": "",
"client_id": "YNxT9w7GMdWvEOKa",
"client_secret": "dbw2OtmVEeuUvIptb1Coyg",
"password": "xxx",
"username": "xxx"
}
*/
type GetAccessRequest struct {
	CaptchaToken string `json:"captcha_token"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Password     string `json:"password"`
	Username     string `json:"username"`
}

/*GetObjectSRequest
{
	"parent_id": "",
	"thumbnail_size": "SIZE_SMALL",
	"with_audit": "true",
	"limit": "100",
	"filters": "{\"phase\":{\"eq\":\"PHASE_TYPE_COMPLETE\"},\"trashed\":{\"eq\":false}}"
}
*/
type GetObjectSRequest struct {
	ParentId      string `json:"parent_id"`
	ThumbnailSize string `json:"thumbnail_size"` //SIZE_SMALL
	WithAudit     string `json:"with_audit"`
	Limit         string `json:"limit"`
	Filters       string `json:"filters"` //{"phase":{"eq":"PHASE_TYPE_COMPLETE"},"trashed":{"eq":false}} // 应该是显示已完成的
}

/*GetTasksRequest
{
	"type": "offline", //固定
	"page_token": "",
	"thumbnail_size": "SIZE_LARGE", SIZE_SMALL 小size
	"filters": "{}",
    "limit":10000,
	"with": "reference_resource"
}
*/
type GetTasksRequest struct {
	Type          string `json:"type"`
	PageToken     string `json:"page_token"`
	ThumbnailSize string `json:"thumbnail_size"`
	Filters       string `json:"filters"` // {"phase":{"in":"PHASE_TYPE_RUNNING,PHASE_TYPE_ERROR"}} 筛选runnig 和error的
	With          string `json:"with"`
}

/*PutTaskRequest
{
                "kind": "drive#file", //固定
                "name": "",
                "upload_type": "UPLOAD_TYPE_URL", //固定
                "url": {
                    "url": "magnet:?xt=urn:btih:RVS3676INANIYPTELZDLBXI4NQKDO36L"
                },
                "folder_type": "DOWNLOAD" //固定
            }
*/
type PutTaskRequest struct {
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	UploadType string `json:"upload_type"`
	Url        struct {
		Url string `json:"url"`
	} `json:"url"`
	FolderType string `json:"folder_type"`
}

/*RefreshTokenRequest
{
            "client_id": "YNxT9w7GMdWvEOKa",
            "client_secret": "dbw2OtmVEeuUvIptb1Coyg",
            "grant_type": "refresh_token",
            "refresh_token": "os.bStDCnL8QGJiuAccYSMaBj2NVHn27y4qGfg4u1aGfNLe3KIt-Omo7bILPwso"
        }
刷新后 refresh_token会变化 token 不会发生改变
*/
type RefreshTokenRequest struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
}
