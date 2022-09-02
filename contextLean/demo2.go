package main

import (
	"context"
	"github.com/mrxtryagin/pikpakdown-api-go/httpHandler"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client := httpHandler.NewClient(
		httpHandler.WithContext(ctx),
	)

	_, err := client.Get("https://va-trialdist.azureedge.net/stella_trial.zip", nil).CheckHttpStatusOk().GetResponse()
	if err != nil {
		panic(err)
	}
	cancel()
	time.Sleep(5 * time.Second)

}
