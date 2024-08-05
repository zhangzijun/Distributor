package Cron

import (
	"Distributor/model"
	"time"
)

func InitJobs() {
	// 模拟一个定时任务，每隔一段时间执行一次
	// Calculate the duration until next midnight
	now := time.Now()
	nextMidnight := now.Add(time.Minute)
	durationUntilNextMidnight := time.Until(nextMidnight)

	// Start a timer that will fire at the next midnight
	initialTimer := time.NewTimer(durationUntilNextMidnight)

	// Function to handle the task execution and subsequent 24-hour ticker
	go func() {
		<-initialTimer.C
		model.InitTable()

		// Start a ticker that fires every 24 hours
		ticker := time.NewTicker(24 * time.Hour)
		for {
			<-ticker.C
			model.InitTable()
		}
	}()

	// Keep the main function running
	select {}
}
