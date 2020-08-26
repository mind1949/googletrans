package transcookie

import "testing"

func TestParseCookieStr(t *testing.T) {
	cookieStr := "NID=204=Au7rQwn2eharnT1rtKsoQl32M2ASoamoFj5Rk8LKHZgg7YZfo54k88aqBVcUEYxcLKjpSU5dNgGTrRAu4Uiv7G3fIAeT3L87gsJCdqg_dCJ9tMHTufW8pHIUD1KgCDwUSIH60d4cWVsukZpai43pm9vHr3SLHCQk9ueEpYJ5Cx8; expires=Thu, 25-Feb-2021 15:15:28 GMT; path=/; domain=.google.cn; HttpOnly"
	cookie, err := parseCookieStr(cookieStr)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v\n", cookie)
}

func TestGet(t *testing.T) {
	cookie, err := Get("https://translate.google.cn")
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v\n", cookie)
}
