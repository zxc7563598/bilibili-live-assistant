package feedback

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/feedback"
)

type Service struct {
	feedbackRepo feedback.Repository
}

func New(feedbackRepo feedback.Repository) *Service {
	return &Service{feedbackRepo: feedbackRepo}
}

// Add 新增一条用户投诉/反馈记录
func (s *Service) Add(ctx context.Context, userID int64, req CreateReq) (int, error) {
	t := strings.TrimSpace(req.Type)
	content := strings.TrimSpace(req.Content)
	contact := strings.TrimSpace(req.Contact)
	if t == "" {
		return 11201, nil
	}
	if utf8.RuneCountInString(t) > 100 {
		return 11202, nil
	}
	if content == "" {
		return 11203, nil
	}
	if utf8.RuneCountInString(content) > 5000 {
		return 11204, nil
	}
	if contact == "" {
		return 11205, nil
	}
	if utf8.RuneCountInString(contact) > 100 {
		return 11206, nil
	}
	if _, err := s.feedbackRepo.Create(ctx, nil, &model.Feedback{
		UserID:  userID,
		Type:    t,
		Content: content,
		Contact: contact,
	}); err != nil {
		return 61201, err
	}
	return 0, nil
}

// ListPage 后台分页查询投诉，联查 live_users 返回 uid/uname
func (s *Service) ListPage(ctx context.Context, req ListPageReq) (ListPageResp, int, error) {
	// 获取列表数据
	offset, limit, sortField, sortOrder := req.OffsetLimit()
	list, total, err := s.feedbackRepo.ListPage(ctx, nil, model.FeedbackListPageQuery{
		UID:       req.UID,
		Uname:     req.Uname,
		SortField: sortField,
		SortOrder: sortOrder,
		Offset:    offset,
		Limit:     limit,
	})
	if err != nil {
		return ListPageResp{}, 61201, err
	}
	// 返回数据
	return ListPageResp{
		Total:    total,
		PageData: toListPageItems(list),
	}, 0, nil
}

// Details 后台查询投诉详情，返回投诉全部字段与用户 uid/uname
func (s *Service) Details(ctx context.Context, id int64) (DetailsItem, int, error) {
	item, err := s.feedbackRepo.GetDetailByID(ctx, nil, id)
	if err != nil {
		return DetailsItem{}, 61201, err
	}
	if item == nil {
		return DetailsItem{}, 51201, errors.New("投诉记录不存在")
	}
	return toDetailsItem(*item), 0, nil
}
