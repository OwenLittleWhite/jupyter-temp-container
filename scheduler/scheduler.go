package scheduler

import (
	"manager/services"

	"github.com/jasonlvhit/gocron"
)

func Init() {
	cron := gocron.NewScheduler()
	cron.Every(1).Minute().Do(services.StartJupyterHubIfNotAvailable())
	// 启动定时任务
	go cron.Start()
}
