package upload

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"sync"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
)

// importTask 一次导入任务的进度快照
type importTask struct {
	id               string
	status           enum.ImportStatus
	adminID          int64
	filename         string
	currentTable     string
	danmuCount       int64
	giftCount        int64
	userCount        int64
	creditLogCount   int64
	skippedUserCount int64
	bytesTotal       int64
	bytesProcessed   int64
	errorCode        int
	startedAt        int64
	finishedAt       int64
}

// isRunning 任务是否还在排队或执行中
func (t *importTask) isRunning() bool {
	return t.status == enum.ImportStatusPending || t.status == enum.ImportStatusRunning
}

// importTaskStore 进程内的导入任务表，同一时间只允许一个任务在跑
type importTaskStore struct {
	mu    sync.Mutex
	tasks map[string]*importTask
}

// newImportTaskStore 创建任务表
func newImportTaskStore() *importTaskStore {
	return &importTaskStore{tasks: make(map[string]*importTask)}
}

// start 登记任务并占用单飞锁，已有任务在跑时返回 false
func (s *importTaskStore) start(adminID int64, filename string, size int64) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeLocked()
	for _, task := range s.tasks {
		if task.isRunning() {
			return "", false
		}
	}
	id := newImportTaskID()
	s.tasks[id] = &importTask{
		id:         id,
		status:     enum.ImportStatusPending,
		adminID:    adminID,
		filename:   filename,
		bytesTotal: size,
		startedAt:  time.Now().Unix(),
	}
	return id, true
}

// cancel 撤销尚未开跑的任务并释放单飞锁
func (s *importTaskStore) cancel(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, id)
}

// update 在锁内改写任务
func (s *importTaskStore) update(id string, fn func(task *importTask)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if task, ok := s.tasks[id]; ok {
		fn(task)
	}
}

// get 读取任务快照，不存在返回 nil
func (s *importTaskStore) get(id string) *importTask {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return nil
	}
	snapshot := *task
	return &snapshot
}

// purgeLocked 淘汰已结束且超过保留时长的任务
func (s *importTaskStore) purgeLocked() {
	now := time.Now().Unix()
	for id, task := range s.tasks {
		expired := now-task.finishedAt > int64(importTaskRetention/time.Second)
		if !task.isRunning() && expired {
			delete(s.tasks, id)
		}
	}
}

// newImportTaskID 生成 32 位十六进制任务 ID
func newImportTaskID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(buf)
}
