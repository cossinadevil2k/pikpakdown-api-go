package pikpakdownCore

import (
	"github.com/mrxtryagin/pikpakdown-api-go/pikpakdownCore/service"
	"github.com/mrxtryagin/pikpakdown-api-go/utils"
	"time"
)

const (
	// 10s
	defaultKeepAliveCircle = 2
)

//定时器
func Keep(fun func() bool, keepAliveCircle int64) {
	utils.Log().Debug("开启保活周期... 每 %d s执行一次", keepAliveCircle)
	//首次循环立即执行
	interval := 50 * time.Millisecond
	//只创造一次
	timer := time.NewTimer(interval)
	//退出时保证停止
	defer timer.Stop()
	for {
		//每次使用前先重置
		timer.Reset(interval)
		select {
		case <-timer.C:
			interval = time.Duration(keepAliveCircle) * time.Second
			utils.Log().Debug("开始执行下一次任务...")
			if fun() {
				utils.Log().Debug("退出keep...")
				return
			}

		}
	}
}

func DefaultKeep(fun func() bool) {
	Keep(fun, defaultKeepAliveCircle)
}

func Active() {

	temp := func() bool {
		service.UpdateUserToken()
		return false
	}
	Keep(temp, 10)

}
