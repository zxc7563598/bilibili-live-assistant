package timeutil

import "time"

// Format 将时间戳格式化为标准格式
func Format(ts int64) string {
	if ts == 0 {
		return ""
	}
	return time.Unix(ts, 0).Format(time.DateTime)
}

// FormatWithLayout 使用自定义格式格式化时间
func FormatWithLayout(ts int64, layout string) string {
	return time.Unix(ts, 0).Format(layout)
}

// Parse 解析标准格式的时间字符串
func Parse(timeStr string) (time.Time, error) {
	return time.Parse(time.DateTime, timeStr)
}

// SecondsPerDay 一天的秒数
const SecondsPerDay = 24 * 60 * 60

// SecondRange 把前端的毫秒级时间区间 [起始, 结束] 换算成秒级查询区间：
// 起始取首项当天 0 点，结束推到末项当天最后一秒（23:59:59）。
//
// 区间不足两项时返回 (nil, nil)，调用方据此跳过时间筛选。
// 原先弹幕 / 礼物 / PK 三个 handler 各抄了一份同样的换算，其中一份还把
// SecondsPerDay - 1 直接写成了 86399，这里收敛为一份。
func SecondRange(ms []int64) (start *int64, end *int64) {
	if len(ms) < 2 {
		return nil, nil
	}
	s := ms[0] / 1000
	e := ms[len(ms)-1]/1000 + SecondsPerDay - 1
	return &s, &e
}
