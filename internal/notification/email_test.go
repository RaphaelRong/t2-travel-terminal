package notification

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestNewService_MissingConfig(t *testing.T) {
	logger := zap.NewNop()
	cases := []struct {
		name string
		cfg  Config
	}{
		{
			name: "missing host",
			cfg: Config{
				Port:     587,
				Username: "user@example.com",
				Password: "pass",
				From:     "user@example.com",
			},
		},
		{
			name: "missing port",
			cfg: Config{
				Host:     "smtp.example.com",
				Username: "user@example.com",
				Password: "pass",
				From:     "user@example.com",
			},
		},
		{
			name: "missing username",
			cfg: Config{
				Host:     "smtp.example.com",
				Port:     587,
				Password: "pass",
				From:     "user@example.com",
			},
		},
		{
			name: "missing password",
			cfg: Config{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user@example.com",
				From:     "user@example.com",
			},
		},
		{
			name: "missing from",
			cfg: Config{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user@example.com",
				Password: "pass",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewService(tc.cfg, logger)
			if err == nil {
				t.Fatal("expected error for incomplete config")
			}
		})
	}
}

func TestNewService_AuthMethod(t *testing.T) {
	logger := zap.NewNop()
	base := Config{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user@example.com",
		Password: "pass",
		From:     "user@example.com",
	}

	cases := []struct {
		name       string
		authMethod string
		wantErr    bool
	}{
		{"default empty", "", false},
		{"LOGIN", "LOGIN", false},
		{"login lowercase", "login", false},
		{"PLAIN", "PLAIN", false},
		{"invalid", "CRAM-MD5", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := base
			cfg.AuthMethod = tc.authMethod
			svc, err := NewService(cfg, logger)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error for invalid auth method")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if svc.cfg.AuthMethod != "LOGIN" && svc.cfg.AuthMethod != "PLAIN" {
				t.Fatalf("unexpected normalized auth method: %s", svc.cfg.AuthMethod)
			}
		})
	}
}

func TestLoginAuth(t *testing.T) {
	auth := loginAuth{username: "user@example.com", password: "secret"}

	proto, _, err := auth.Start(nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if proto != "LOGIN" {
		t.Fatalf("expected LOGIN, got %s", proto)
	}

	resp, err := auth.Next([]byte("Username:"), true)
	if err != nil {
		t.Fatalf("Next for username failed: %v", err)
	}
	if string(resp) != "user@example.com" {
		t.Fatalf("unexpected username response: %s", resp)
	}

	resp, err = auth.Next([]byte("Password:"), true)
	if err != nil {
		t.Fatalf("Next for password failed: %v", err)
	}
	if string(resp) != "secret" {
		t.Fatalf("unexpected password response: %s", resp)
	}

	_, err = auth.Next([]byte("Unknown:"), true)
	if err == nil {
		t.Fatal("expected error for unknown challenge")
	}
}

func TestBuildVerificationLink(t *testing.T) {
	cases := []struct {
		baseURL string
		token   string
		want    string
	}{
		{
			baseURL: "http://localhost:8080",
			token:   "abc123",
			want:    "http://localhost:8080/api/v1/auth/verify?token=abc123",
		},
		{
			baseURL: "https://example.com/",
			token:   "abc123",
			want:    "https://example.com/api/v1/auth/verify?token=abc123",
		},
		{
			baseURL: "",
			token:   "abc123",
			want:    "http://localhost:8080/api/v1/auth/verify?token=abc123",
		},
	}

	for _, tc := range cases {
		got := buildVerificationLink(tc.baseURL, tc.token)
		if got != tc.want {
			t.Fatalf("buildVerificationLink(%q, %q) = %q, want %q", tc.baseURL, tc.token, got, tc.want)
		}
	}
}

func TestBuildMessage(t *testing.T) {
	svc := &Service{
		cfg: Config{
			Host:     "smtp.example.com",
			From:     "sender@example.com",
			FromName: "T2 Travel Terminal",
		},
		logger: zap.NewNop(),
	}

	msg, err := svc.buildMessage("recipient@example.com", "Verify your email", "plain text", "<html>html body</html>")
	if err != nil {
		t.Fatalf("buildMessage failed: %v", err)
	}

	msgStr := string(msg)
	required := []string{
		"From: T2 Travel Terminal <sender@example.com>",
		"To: recipient@example.com",
		"Subject: Verify your email",
		"Content-Type: multipart/alternative",
		"plain text",
		"<html>html body</html>",
	}

	for _, r := range required {
		if !strings.Contains(msgStr, r) {
			t.Fatalf("message missing expected content %q\n%s", r, msgStr)
		}
	}
}
