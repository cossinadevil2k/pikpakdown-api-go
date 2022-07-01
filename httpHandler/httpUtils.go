package httpHandler

import (
	"github.com/mrxtryagin/pikpakdown-api-go/utils"
	"time"
)

// HttpTimer http相关的一些方法 比如加密等
func HttpTimer(funcName string) func() {
	handler := func(startTime time.Time) {
		utils.Log().Info("Time taken by %s request is %s", funcName, utils.CostTimeInfo(startTime))
	}
	return utils.Timer(handler)
}

//AddQuery 添加query
func AddQuery(prefix string, querys map[string]string) string {
	result := ""
	for k, v := range querys {
		result += k + "=" + v + "&"
	}
	return prefix + "?" + result[:len(result)-1]

}

func GetQueryFromMap(querys map[string]string) string {
	result := ""
	for k, v := range querys {
		result += k + "=" + v + "&"
	}
	return result[:len(result)-1]
}
