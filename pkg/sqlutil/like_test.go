package sqlutil

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// 用真实的 SQLite 跑一遍模糊搜索，把「转义之外还必须写 ESCAPE 子句」钉住。
//
// SQLite 没有默认的 LIKE 转义字符（MySQL / PostgreSQL 默认是反斜杠），所以漏写 ESCAPE
// 时 SQLite 会把转义符当成普通字符去匹配，搜索词里含 _ % \ 就一条都查不到。
func TestEscapeLike(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "like.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开 sqlite 失败: %v", err)
	}
	if err := db.Exec("CREATE TABLE t (name TEXT NOT NULL)").Error; err != nil {
		t.Fatalf("建表失败: %v", err)
	}
	// userX123 / axxb 是「通配符误当字面量」的反例，用来证明 _ % 没有被当成通配符
	names := []string{"user_123", "userX123", "a%b", "axxb", `a\b`, "wow!", "哎呀又胖辣"}
	for _, n := range names {
		if err := db.Exec("INSERT INTO t (name) VALUES (?)", n).Error; err != nil {
			t.Fatalf("插入 %q 失败: %v", n, err)
		}
	}

	tests := []struct {
		name string
		term string
		want []string
	}{
		{"下划线按字面匹配，不连带命中 userX123", "user_123", []string{"user_123"}},
		{"百分号按字面匹配，不连带命中 axxb", "a%b", []string{"a%b"}},
		{"反斜杠按字面匹配", `a\b`, []string{`a\b`}},
		{"转义字符自身按字面匹配", "wow!", []string{"wow!"}},
		{"不含特殊字符时行为不变", "胖", []string{"哎呀又胖辣"}},
		{"只输入通配符时按字面匹配", "%", []string{"a%b"}},
		{"匹配不到时返回空", "zzz", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			err := db.Raw("SELECT name FROM t WHERE name LIKE ? ESCAPE '!' ORDER BY name",
				"%"+EscapeLike(tt.term)+"%").Scan(&got).Error
			if err != nil {
				t.Fatalf("查询失败: %v", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Fatalf("搜索 %q 得到 %v，期望 %v", tt.term, got, tt.want)
			}
		})
	}
}
