package fileutil

import "testing"

func TestResolveLocalPath(t *testing.T) {
	ok := []struct{ in, want string }{
		{"/uploads/login_bg/1724_x.png", "uploads/login_bg/1724_x.png"},
		{"uploads/login_bg/1724_x.png", "uploads/login_bg/1724_x.png"},
		{"/uploads/site_icon/a.png", "uploads/site_icon/a.png"},
		{"/uploads/../uploads/logo/b.png", "uploads/logo/b.png"},
	}
	for _, c := range ok {
		got, err := ResolveLocalPath(c.in)
		if err != nil {
			t.Errorf("ResolveLocalPath(%q) unexpected err: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ResolveLocalPath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	bad := []string{
		"", "/uploads", "/uploads/", "/etc/passwd",
		"/uploads/../../etc/passwd", "C:\\uploads\\a.png", "/uploads/..",
	}
	for _, in := range bad {
		if got, err := ResolveLocalPath(in); err == nil {
			t.Errorf("ResolveLocalPath(%q) = %q, want error", in, got)
		}
	}
}
