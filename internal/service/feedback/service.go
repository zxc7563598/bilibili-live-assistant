package feedback

import (
	"context"
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
