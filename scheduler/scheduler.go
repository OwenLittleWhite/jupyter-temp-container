package scheduler

import (
	"manager/services"

	"github.com/jasonlvhit/gocron"
)

func Init() {
	cron := gocron.NewScheduler()
	cron.Every(1).Minute().Do(services.StartJupyterHubIfNotAvailable)
	cron.Every(1).Minute().Do(services.ProcessHub)
	cron.Every(5).Second().Do(services.SequProcessUserSession)
	cron.Every(1).Minute().Do(services.DestroyUserSession)
	// 启动定时任务
	go cron.Start()
}
