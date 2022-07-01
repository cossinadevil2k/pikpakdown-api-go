package utils

import (
	"fmt"
	"time"
)

func CostTimeInfo(startTime time.Time) string {
	return fmt.Sprintf("%.4f s", float64(time.Since(startTime).Milliseconds())/float64(1000))
}
