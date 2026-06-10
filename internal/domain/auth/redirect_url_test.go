package auth

import "testing"

func newRedirectTestService(allowlist string) *service {
	return &service{allowedRedirects: parseAllowedRedirects(allowlist)}
}

func TestValidateRedirectURL(t *testing.T) {
	tests := []struct {
		name      string
		allowlist string
		raw       string
		wantOK    bool
	}{
		{"empty url always allowed", "", "", true},
		{"relative path allowed", "", "/post-login", true},
		{"relative path with query allowed", "", "/post-login?next=dashboard", true},
		{"scheme-relative rejected", "", "//evil.com", false},
		{"backslash scheme-relative rejected", "", `/\evil.com`, false},
		{"absolute rejected with empty allowlist", "", "https://evil.com", false},
		{"allowed origin passes", "https://app.example.com", "https://app.example.com", true},
		{"allowed origin with path passes", "https://app.example.com", "https://app.example.com/auth/done", true},
		{"host case-insensitive", "https://app.example.com", "https://APP.EXAMPLE.COM/auth", true},
		{"other host rejected", "https://app.example.com", "https://evil.com", false},
		{"subdomain of allowed host rejected", "https://app.example.com", "https://evil.app.example.com", false},
		{"wrong scheme rejected", "https://app.example.com", "http://app.example.com", false},
		{"wrong port rejected", "https://app.example.com", "https://app.example.com:8443", false},
		{"explicit port must match", "http://localhost:5173", "http://localhost:5173/done", true},
		{"javascript scheme rejected", "https://app.example.com", "javascript:alert(1)", false},
		{"data scheme rejected", "https://app.example.com", "data:text/html,x", false},
		{"garbage rejected", "https://app.example.com", "ht tp://bad url", false},
		{"second allowlist entry matches", "https://a.example.com, https://b.example.com", "https://b.example.com/x", true},
		{"userinfo trick rejected", "https://app.example.com", "https://app.example.com@evil.com/", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newRedirectTestService(tt.allowlist)
			err := s.validateRedirectURL(tt.raw)
			if tt.wantOK && err != nil {
				t.Errorf("validateRedirectURL(%q) = %v, want nil", tt.raw, err)
			}
			if !tt.wantOK && err == nil {
				t.Errorf("validateRedirectURL(%q) = nil, want ErrOAuthRedirectNotAllowed", tt.raw)
			}
		})
	}
}

func TestParseAllowedRedirects(t *testing.T) {
	got := parseAllowedRedirects(" https://App.Example.com/ignored-path , , ftp://bad.com, http://localhost:5173")
	want := []string{"https://app.example.com", "http://localhost:5173"}
	if len(got) != len(want) {
		t.Fatalf("parseAllowedRedirects() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d = %q, want %q", i, got[i], want[i])
		}
	}
}
