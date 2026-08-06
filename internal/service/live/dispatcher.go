package live

import (
	"context"
	"log"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
)

// MessageProcessor 消息处理器接口
//
// 每个 MessageProcessor 对应一种业务场景，可以处理一个或多个 Cmd。
// 实现者通过 Cmds() 声明自己处理哪些命令字，Process 方法中根据 cmd 做类型断言。
type MessageProcessor interface {
	// Cmds 返回本处理器要处理的命令字列表
	Cmds() []live.Cmd
	// Process 处理一条已解密的消息，返回 error 表示处理失败
	// roomID 是当前监听的直播间房间号
	Process(ctx context.Context, cmd live.Cmd, data any, roomID int64) error
}

// messageDispatcher 消息分发器，将 WebSocket 消息路由到对应的 MessageProcessor
type messageDispatcher struct {
	processors map[live.Cmd]MessageProcessor
}

// newMessageDispatcher 创建消息分发器，注册所有 MessageProcessor
func newMessageDispatcher(processors ...MessageProcessor) *messageDispatcher {
	d := &messageDispatcher{
		processors: make(map[live.Cmd]MessageProcessor),
	}
	for _, p := range processors {
		for _, cmd := range p.Cmds() {
			d.processors[cmd] = p
		}
	}
	return d
}

// dispatch 将消息分发给对应的处理器，未注册的命令字静默跳过
func (d *messageDispatcher) dispatch(ctx context.Context, cmd live.Cmd, data any, roomID int64) {
	p, ok := d.processors[cmd]
	if !ok {
		return
	}
	if err := p.Process(ctx, cmd, data, roomID); err != nil {
		log.Printf("[live.Dispatcher] [%s] 业务处理失败: %v", cmd, err)
	}
}
