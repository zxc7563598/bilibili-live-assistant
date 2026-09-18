package feedback

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/feedback"
)

// Handler 用户投诉/反馈 HTTP 接口处理器
type Handler struct {
	feedbackSvc *feedback.Service
}

// New 创建 Handler 实例
func New(feedbackSvc *feedback.Service) *Handler {
	return &Handler{
		feedbackSvc: feedbackSvc,
	}
}

// toFeedbackListPageItem 后台投诉列表转换：联查得到的 uid/uname + 投诉表字段，不含投诉正文
func toFeedbackListPageItem(list []feedback.ListPageItem) []resp.FeedbackListPageItem {
	res := make([]resp.FeedbackListPageItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.FeedbackListPageItem{
			ID:        v.ID,
			UserID:    v.UserID,
			UID:       v.UID,
			Uname:     v.Uname,
			Type:      v.Type,
			Contact:   v.Contact,
			CreatedAt: v.CreatedAt,
		})
	}
	return res
}

// toFeedbackDetailsResp 投诉详情转换：投诉全字段 + 用户 uid/uname
func toFeedbackDetailsResp(v feedback.DetailsItem) resp.FeedbackDetailsResp {
	return resp.FeedbackDetailsResp{
		ID:        v.ID,
		UserID:    v.UserID,
		UID:       v.UID,
		Uname:     v.Uname,
		Type:      v.Type,
		Content:   v.Content,
		Contact:   v.Contact,
		CreatedAt: v.CreatedAt,
	}
}
