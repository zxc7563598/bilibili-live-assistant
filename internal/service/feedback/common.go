package feedback

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
)

// toListPageItems 投诉联查结果 → 列表出参
func toListPageItems(list []model.FeedbackListItem) []ListPageItem {
	respList := make([]ListPageItem, 0, len(list))
	for _, v := range list {
		item := ListPageItem{
			ID:        v.ID,
			UserID:    v.UserID,
			UID:       v.UID,
			Uname:     v.Uname,
			Type:      v.Type,
			Contact:   v.Contact,
			CreatedAt: timeutil.Format(v.CreatedAt),
		}
		respList = append(respList, item)
	}
	return respList
}

// toDetailsItem 投诉联查结果 → 详情出参
func toDetailsItem(v model.FeedbackListItem) DetailsItem {
	return DetailsItem{
		ID:        v.ID,
		UserID:    v.UserID,
		UID:       v.UID,
		Uname:     v.Uname,
		Type:      v.Type,
		Content:   v.Content,
		Contact:   v.Contact,
		CreatedAt: timeutil.Format(v.CreatedAt),
	}
}
