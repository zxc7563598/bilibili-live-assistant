// Package cron 提供一个最小可复用的定时任务调度器。
//
// 每个 Job 在独立 goroutine 中以固定间隔执行，执行出错只记录日志、不中断后续调度。
// 供应用启动时装配、退出时统一停止，方便未来接入更多后台定时任务。
package cron

import (
	"context"
	"log"
	"sync"
	"time"
)

// Job 描述一个定时任务。
type Job struct {
	Name     string                          // 任务名，用于日志
	Interval time.Duration                   // 执行间隔
	Run      func(ctx context.Context) error // 任务体
}

// Scheduler 持有多个 Job 并负责其启动与停止。
type Scheduler struct {
	jobs   []Job
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New 创建调度器，可传入若干 Job。
func New(jobs ...Job) *Scheduler {
	return &Scheduler{jobs: jobs}
}

// Start 为每个 Job 启动一个独立的后台 goroutine，按各自间隔执行。
//
// 使用内部 context 控制全部 goroutine 的生命周期，进程退出时通过 Stop 统一取消。
func (s *Scheduler) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	for _, job := range s.jobs {
		s.wg.Add(1)
		go s.run(ctx, job)
	}
}

// run 单个任务的调度循环。
func (s *Scheduler) run(ctx context.Context, job Job) {
	defer s.wg.Done()
	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := job.Run(ctx); err != nil {
				log.Printf("[cron] 任务 %s 执行失败: %v", job.Name, err)
			}
		}
	}
}

// Stop 取消全部任务并等待其 goroutine 退出。
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}
