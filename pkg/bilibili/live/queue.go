package live

import (
	"container/heap"
	"context"
	"log"
	"sync"
	"time"
)

// Sender 定义弹幕发送接口
//
// 使用者需实现此接口以接入实际的弹幕发送逻辑
// 典型的实现是包装 room.Service.SendDanmu：
//
//	type danmuSender struct {
//	    svc    *room.Service
//	    client *bilibili.Client
//	}
//
//	func (s *danmuSender) Send(ctx context.Context, roomID int64, message string) (string, error) {
//	    csrfToken, err := s.client.CSRF()
//	    if err != nil {
//	        return "", err
//	    }
//	    return "", s.svc.SendDanmu(ctx, roomID, message, csrfToken)
//	}
type Sender interface {
	// Send 发送弹幕到指定直播间
	//
	// 若 message 超过单条弹幕允许的长度，实现方可只发送其前缀，并通过
	// 返回值 remaining 返回未发完的剩余部分。Queue 会等到 remaining
	// 为空后才从队列取出下一条消息，因此长消息的各段会连续发出、不会被
	// 更高优先级的其他消息打断，同时每段仍按发送间隔逐个发出，天然遵守
	// 频率限制。若无需拆分，remaining 返回空字符串即可。
	//
	// 返回 err 时整条（含未发的 remaining）视为发送失败被丢弃，不会重试。
	// ctx 在 Queue.Stop() 调用后会被取消
	Send(ctx context.Context, roomID int64, message string) (remaining string, err error)
}

// Queue 是弹幕发送优先级队列
//
// 它在独立的 goroutine 中以固定间隔（默认 4 秒）从队列中取出优先级最高的消息发送。
// 优先级数字越小越优先，同优先级按入队顺序（FIFO）发送。
//
// 设计目标：避免频繁发送触发 B站 限流，同时保证高优先级消息优先发出。
//
// 发送节奏为每个间隔发出一条弹幕。若一条消息超过单条弹幕允许的长度，
// 会按发送间隔拆成多段连续发出（见 Sender.Send 返回值 remaining 的说明）。
//
// 典型用法：
//
//	q := live.NewQueue(roomID, sender)
//	q.Start(ctx)
//
//	q.Enqueue("重要通知！", 0)   // 高优先级，先发送
//	q.Enqueue("欢迎来到直播间~", 5) // 低优先级，后发送
//
//	// ... 队列每隔 interval 自动发送一条
//
//	q.Stop() // 停止队列，等待当前发送完成
type Queue struct {
	roomID   int64
	sender   Sender
	interval time.Duration
	mu       sync.Mutex
	heap     priorityQueue
	nextSeq  int64
	running  bool
	cancel   context.CancelFunc
	done     chan struct{}
	// onError 在发送失败时回调，nil 表示仅 log 输出
	onError func(msg string, err error)
	// pending 当前消息未发完的剩余部分，仅在 worker goroutine 中读写。
	// 非空时下个 tick 继续发送它，而不是从堆中取出下一条消息，
	// 从而保证长消息的各段连续发出。
	pending string
}

// QueueOption 是 Queue 的函数式配置项
type QueueOption func(*Queue)

// WithSendInterval 设置发送间隔（默认 4 秒）
// 根据 B站 弹幕发送频率限制调整，建议不低于 4 秒
func WithSendInterval(d time.Duration) QueueOption {
	return func(q *Queue) {
		q.interval = d
	}
}

// WithOnError 设置发送失败时的回调函数
// 默认行为是 log 打印错误
func WithOnError(fn func(msg string, err error)) QueueOption {
	return func(q *Queue) {
		q.onError = fn
	}
}

// NewQueue 创建一个新的弹幕发送队列
//
// 创建后队列处于未启动状态，需调用 Start 开始工作
//
// 参数：
//   - roomID: 直播间真实房间号（长 ID）
//   - sender: 弹幕发送实现，需满足 Sender 接口
func NewQueue(roomID int64, sender Sender, opts ...QueueOption) *Queue {
	q := &Queue{
		roomID:   roomID,
		sender:   sender,
		interval: 4 * time.Second,
		heap:     make(priorityQueue, 0),
	}
	for _, o := range opts {
		o(q)
	}
	heap.Init(&q.heap)
	return q
}

// Enqueue 将一条弹幕消息加入发送队列
//
// priority 越小越优先发送，同优先级按入队顺序（FIFO）发送
//
// 线程安全，可在任意 goroutine 中调用。
// 即使队列未启动也可以入队，消息会在 Start 后依次发送。
func (q *Queue) Enqueue(msg string, priority int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	item := &queueItem{
		Message:  msg,
		Priority: priority,
		seq:      q.nextSeq,
	}
	q.nextSeq++
	heap.Push(&q.heap, item)
}

// Start 启动队列工作器
//
// 在独立的 goroutine 中以固定间隔从队列取出消息并调用 Sender.Send。
// 调用后立即返回。重复调用（已启动状态）是幂等的。
//
// 支持 Stop 后重新 Start（会创建新的 worker goroutine）。
//
// 参数 ctx 用于控制工作器的生命周期：
//   - ctx 被取消时，工作器自动停止
//   - 也可主动调用 Stop() 来停止
func (q *Queue) Start(ctx context.Context) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.running {
		return
	}
	ctx, q.cancel = context.WithCancel(ctx)
	q.done = make(chan struct{})
	q.running = true
	go q.worker(ctx)
}

// Stop 停止队列工作器
//
// 取消内部 context，等待当前正在发送的消息完成后返回。
// 未启动时 Stop 是幂等的（无操作）。
//
// 注意：Stop 不会清空队列中未发送的消息，
// 再次 Start 后会从剩余消息继续发送。如需清空，请调用 Clear。
func (q *Queue) Stop() {
	q.mu.Lock()
	if !q.running {
		q.mu.Unlock()
		return
	}
	if q.cancel != nil {
		q.cancel()
	}
	done := q.done
	q.mu.Unlock()

	<-done

	q.mu.Lock()
	q.running = false
	q.mu.Unlock()
}

// Len 返回当前队列中待发送的消息数量
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.heap.Len()
}

// IsRunning 返回队列工作器是否正在运行
func (q *Queue) IsRunning() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.running
}

// Clear 清空队列中所有未发送的消息
//
// 建议在 Stop 之后调用（worker 已退出），此时连带清掉未发完的
// pending 剩余部分。若队列正在运行中调用，则只清堆中的消息，
// 正在跨 tick 续传的消息不受影响。
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.heap = q.heap[:0]
	// 重置 seq 避免无限增长（可选）
	q.nextSeq = 0
	if !q.running {
		q.pending = ""
	}
}

// =========================================================================
// 优先队列实现（container/heap）
// =========================================================================

// queueItem 是优先队列中的一条弹幕消息
type queueItem struct {
	Message  string // 弹幕内容
	Priority int    // 优先级（越小越优先）
	seq      int64  // 全局递增序号（同优先级 FIFO）
	index    int    // heap.Interface 维护的内部索引
}

// priorityQueue 实现 heap.Interface
type priorityQueue []*queueItem

// Len 返回队列中元素的数量
func (pq priorityQueue) Len() int { return len(pq) }

// Less 定义堆中的排序规则
func (pq priorityQueue) Less(i, j int) bool {
	// 优先级数字小的优先
	if pq[i].Priority != pq[j].Priority {
		return pq[i].Priority < pq[j].Priority
	}
	// 同优先级按入队顺序（FIFO）
	return pq[i].seq < pq[j].seq
}

// Swap 交换堆中两个元素的位置，并更新它们的 index 字段
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push 向队列中添加一个新元素
func (pq *priorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*queueItem)
	item.index = n
	*pq = append(*pq, item)
}

// Pop 从队列中移除并返回优先级最高的元素
func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // 避免内存泄漏
	item.index = -1
	*pq = old[:n-1]
	return item
}

// =========================================================================
// 内部实现
// =========================================================================

// worker 是队列工作循环，运行在独立的 goroutine 中
func (q *Queue) worker(ctx context.Context) {
	defer close(q.done)
	ticker := time.NewTicker(q.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			q.sendNext(ctx)
		}
	}
}

// sendNext 发送一条弹幕。
//
// 若上一条消息还有未发完的剩余（pending 非空），优先继续发送它；
// 否则从队列取出下一条优先级最高的消息。发送后把剩余部分写回 pending，
// 由下个 tick 继续处理，直到整条消息发完才取下一条。
func (q *Queue) sendNext(ctx context.Context) {
	if q.pending == "" {
		msg := q.pop()
		if msg == "" {
			return
		}
		q.pending = msg
	}
	rest, err := q.sender.Send(ctx, q.roomID, q.pending)
	if err != nil {
		msg := q.pending
		q.pending = "" // 发送失败整条（含未发剩余）放弃，避免每个 tick 无限重试同一条消息
		if q.onError != nil {
			q.onError(msg, err)
		} else {
			log.Printf("[live.Queue] 推送弹幕到房间 %d 失败: %v (弹幕内容: %s)", q.roomID, err, msg)
		}
		return
	}
	q.pending = rest
}

// pop 从优先队列中取出优先级最高的消息
func (q *Queue) pop() string {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.heap.Len() == 0 {
		return ""
	}
	item := heap.Pop(&q.heap).(*queueItem)
	return item.Message
}
