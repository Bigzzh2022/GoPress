package config

import "testing"

func TestValidateSecretRejectsWeakValues(t *testing.T) {
	cases := map[string]string{
		"empty":           "",
		"whitespace":      "   ",
		"too short":       "short-secret",
		"published":       "gopress-development-secret-change-me",
		"compose default": "local-only-change-this-jwt-secret-32-chars",
	}
	for name, secret := range cases {
		t.Run(name, func(t *testing.T) {
			if err := validateSecret(secret); err == nil {
				t.Fatalf("validateSecret(%q) = nil, want error", secret)
			}
		})
	}

	strong := "9f2c5c1d8e4b7a6f0d3e2c1b4a5f6e7d8c9b0a1f2e3d4c5b6a7f8091a2b3c4d5"
	if err := validateSecret(strong); err != nil {
		t.Fatalf("validateSecret(strong) = %v, want nil", err)
	}
	if err := validateSecret(strong[:31]); err == nil {
		t.Fatal("validateSecret() accepted a secret shorter than the minimum length")
	}
}

func TestValidateAdminPassword(t *testing.T) {
	weak := []string{
		"GoPress123!",
		"replace-with-a-strong-password",
		"short",
		"alllowercaseonly",
		"1234567890123",
	}
	for _, password := range weak {
		if err := ValidateAdminPassword(password); err == nil {
			t.Fatalf("ValidateAdminPassword(%q) = nil, want error", password)
		}
	}
	if err := ValidateAdminPassword("Correct-Horse-Battery-9"); err != nil {
		t.Fatalf("ValidateAdminPassword(strong) = %v, want nil", err)
	}
}

func TestValidateAdminEmail(t *testing.T) {
	if err := ValidateAdminEmail(""); err == nil {
		t.Fatal("ValidateAdminEmail(\"\") = nil, want error")
	}
	if err := ValidateAdminEmail("not-an-email"); err == nil {
		t.Fatal("ValidateAdminEmail(invalid) = nil, want error")
	}
	if err := ValidateAdminEmail("Owner@Example.COM"); err != nil {
		t.Fatalf("ValidateAdminEmail(valid) = %v, want nil", err)
	}
}
