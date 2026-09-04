package region

import "testing"

func TestResolve(t *testing.T) {
	tests := []struct {
		name  string
		codes []string
		want  string
		ok    bool
	}{
		{"常规省市区三段", []string{"370000", "370100", "370116"}, "山东省 济南市 莱芜区", true},
		{"直辖市三段", []string{"110000", "110100", "110101"}, "北京市 市辖区 东城区", true},
		{"无区县地级市两段", []string{"440000", "441900"}, "广东省 东莞市", true},
		{"乱序code", []string{"370100", "370000", "370116"}, "", false},
		{"code不存在", []string{"999999", "000000"}, "", false},
		{"跳级(省直接到区)", []string{"370000", "370116"}, "", false},
		{"仅省一段", []string{"370000"}, "", false},
		{"空数组", []string{}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Resolve(tt.codes)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("Resolve(%v) = (%q, %v), want (%q, %v)", tt.codes, got, ok, tt.want, tt.ok)
			}
		})
	}
}
