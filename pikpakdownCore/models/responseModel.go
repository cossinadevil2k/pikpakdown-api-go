package models

type ErrorInfo struct {
	Error            string `json:"error"`
	ErrorCode        int    `json:"error_code"`
	ErrorUrl         string `json:"error_url"`
	ErrorDescription string `json:"error_description"`
	ErrorDetails     []struct {
		Type         string        `json:"@type"`
		StackEntries []interface{} `json:"stack_entries"`
		Detail       string        `json:"detail"`
	} `json:"error_details"`
}
