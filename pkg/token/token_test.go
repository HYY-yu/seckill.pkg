package token

import (
	"net/url"
	"testing"
	"time"
)

const secret = "i1ydX9RtHyuJTrw7frcu"

func TestJwtSignAndParse(t *testing.T) {
	tokenString, err := New(secret).JwtSign(123456789, "test_user", 24*time.Hour)
	if err != nil {
		t.Error("sign error", err)
		return
	}
	t.Log(tokenString)

	user, err := New(secret).JwtParse(tokenString)
	if err != nil {
		t.Error("parse error", err)
		return
	}
	t.Log(user)
}

func TestUrlSign(t *testing.T) {
	urlPath := "/echo"
	method := "post"
	params := url.Values{}
	params.Add("a", "a1")
	params.Add("d", "d1")
	params.Add("c", "c1")

	tokenString, err := New(secret).UrlSign(time.Now().Unix(), urlPath, method, params)
	if err != nil {
		t.Error("sign error", err)
		return
	}
	t.Log(tokenString)
}

func BenchmarkJwtSignAndParse(b *testing.B) {
	b.ResetTimer()
	token := New(secret)
	for i := 0; i < b.N; i++ {
		tokenString, _ := token.JwtSign(123456789, "xinliangnote", 24*time.Hour)
		_, _ = token.JwtParse(tokenString)
	}
}
