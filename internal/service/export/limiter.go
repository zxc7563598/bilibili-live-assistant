package export

// limiter 同时进行的导出数上限
type limiter struct {
	slots chan struct{}
}

// newLimiter 构建并发守卫，n 小于 1 时按 1 处理（不允许把导出完全关死）
func newLimiter(n int) *limiter {
	if n < 1 {
		n = 1
	}
	return &limiter{slots: make(chan struct{}, n)}
}

// tryAcquire 非阻塞占用一个槽位，满了返回 false
func (l *limiter) tryAcquire() bool {
	select {
	case l.slots <- struct{}{}:
		return true
	default:
		return false
	}
}

// hasFree 只查询是否有空位，不占用
func (l *limiter) hasFree() bool {
	return len(l.slots) < cap(l.slots)
}

// release 归还槽位。只在确实占用过时调用，重复释放在此静默忽略。
func (l *limiter) release() {
	select {
	case <-l.slots:
	default:
	}
}
