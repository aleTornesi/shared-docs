package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func makeToken(t *testing.T, userID int64, method jwt.SigningMethod, secret []byte) string {
	t.Helper()
	token := jwt.NewWithClaims(method, Claims{
		UserID:           userID,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	})
	s, err := token.SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestParseToken_Valid(t *testing.T) {
	tok := makeToken(t, 42, jwt.SigningMethodHS256, []byte("secret"))
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+tok)

	claims, err := ParseToken(r)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 42 {
		t.Fatalf("user_id %d, want 42", claims.UserID)
	}
}

func TestParseToken_MissingHeader(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	_, err := ParseToken(r)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseToken_NonBearer(t *testing.T) {
	tok := makeToken(t, 1, jwt.SigningMethodHS256, []byte("secret"))
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Basic "+tok)

	_, err := ParseToken(r)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer invalid.token.here")

	_, err := ParseToken(r)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseToken_ExpiredToken(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:           1,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour))},
	})
	s, _ := token.SignedString([]byte("secret"))

	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+s)

	_, err := ParseToken(r)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestParseToken_WrongSigningMethod(t *testing.T) {
	// Sign with RSA but server expects HMAC
	key, _ := jwt.ParseRSAPrivateKeyFromPEM([]byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF8PbnGy0AHB7MhgHcTz6sE2I2yPB
aNlBaPaWOmCFSYMhxG0RLFXL2VGNrCFmGF0G8Gt8mvVDh0/PH3MVkFMmWuG3gFoN
mC9bZm1MC4RJxEHKBMamZUJ7sRlh0ewPvpw7piSjp6VwQ+a2S0Ij0cjdR/yaA1bZ
blSmJNGlICKEDi0JyNqBCoGa0FMwlVXbEG/Wm50q25UePFGi0exoCQjVS2KQAM3V
tz0UPjNRWGNsIxK+W+M5v8+E8KubRw5JRhvOgNqQIzHYfJJyvdFY3e1lqYdGaHWh
cHjDT9/4bPaUn8vL7JaAxYvKZfWQxylEO+TqxQIDAQABAoIBAC5RgZ+hBx7xHNaM
pPgwGMnCd6GJsBsCvEPSFLqB4G6FblVsbZUbXaV9MoFJ4VYMf3DO5VkacI8PU7v1m
k8G0WMpbVp0TymUBMZXjmVZJ4c6dAL3bGHHhgDME0k9m2fP80oD3H8MNGSM0jJjb
NlQKfKGk6c5IZfhTMGsvkz3NVnzQx3CwFLTJK/C2UECsYFxMFSvKzW+POjPneyLn
sGNzR4NjkW9FJdVQoGs4PyzGQxJyrN5DeC8/wnRX8L2oEgAGHLIwXPJSiHJ6xVqv1
C9d+zOhETJi0PGQJB3GDGcXZ/L3mvM8jJ8DcPVpTFWRWjJqX+9WCBfiIDmeKLp8v1
eUOkWoECgYEA6XJrHM+VUNW2hVCbRQ5JkLt/b/REH0yMY+GqQp/sM+oJ+7bP5NVu
tJ/4gAT3B3SJ+y5MBO0JOVq/GsoTOFx7NzPHJi0lBklGgCaJ0LkxT6hpMrVJ0KB5
hLvKq0S4RMyhGRdvwC/9tK9gM98wDZ/Z8fJbAR0IwG3AXMHBbrGIacCgYEA5SB7v1
+kCJd/VjVyVRLT1qwLAkEUMsHBgB/QLTE5x/NF30H7VGBSS1gFjP3V0cDBqT2Mj3
1V7X7Tnh/VFpB3/gGcU1g3JBwYgdB8MCQF7oG3LOCn2r3R0yGWSMxxPXx1TaRJoVm
XaYM0Fy5GFBPsPq4H3GFx8FKbEXNr5bLexkCgYEAy1T9JC/i0c7qlDcZ/HRhJQIy
hEVhLkS0mRJHNFdmkljwdBmSbXkVuqMV4ftghfYhHlBpwemF0JLdTFqiIvwMOIrJ
pJJ4VJHP4f3ROCfJ0O2TjNFR1G/Q/H86wSFbvJPMJmj+dXi+wK3MWiJE9g8i4S3Z
Z9xLlSnxm3cPO8R7p1cCgYBqJN6SzUfyaNqoJvBJcZl3Y8O7bNj4t0G1OFBUwePq
V7yGJEDj+gF3P1ZmLl1lI5HmBLNbJhOg1t9x42E/3mQanK9q7m9rGWbEPs/pC8Kz
c7T4mz2BYnFMvEJMJ0WA7WJ0fIxJlTC7G3cL8MBLzJf+wl7NRBzILlU7MK3FdfMF
EQKBgQCfr0i6H4H7fE01HA+LPJGN+kBjWlLFwnKJH1jnRCv7Nz6CjGFh0i8BNig3
CLapJsO9M1j5jZp1xr6FQuqgeC1WXGPIIkP2t1V7EeDIxlBP8p6Km2cH50o2JN/L3
V5AqBnW8Hf8Jl+nG5L7x8J7i8NJFLVrOV0sHaeWaPahBFRxg==
-----END RSA PRIVATE KEY-----`))
	if key == nil {
		t.Skip("could not parse test RSA key")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, Claims{
		UserID: 1,
	})
	s, err := token.SignedString(key)
	if err != nil {
		t.Skip("could not sign with RSA")
	}

	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+s)

	_, err = ParseToken(r)
	if err == nil {
		t.Fatal("expected error for wrong signing method")
	}
}
