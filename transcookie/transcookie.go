// Package transcookie just for caching google translation services' cookies
package transcookie

import (
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	defaultCookiesCache = newCache()

	emptyCookie = http.Cookie{}

	// ErrInvalidServiceURL service url is invalid
	ErrInvalidServiceURL = errors.New("invalid translate google service url")
)

// Get gets cookie from defaultCookiesCache
// for example: Get("https://translate.google.com")
func Get(serviceURL string) (http.Cookie, error) {
	return defaultCookiesCache.get(serviceURL)
}

// Update updates defaultCookiesCache's cookie
func Update(serviceURL string, sleep time.Duration) (http.Cookie, error) {
	return defaultCookiesCache.update(serviceURL, sleep)
}

// transCookiesCache caches google tranlation services' cookies
type transCookiesCache struct {
	clt *http.Client

	m       sync.RWMutex
	cookies map[string]http.Cookie
}

func newCache() *transCookiesCache {
	return &transCookiesCache{
		clt:     &http.Client{},
		cookies: make(map[string]http.Cookie),
	}
}

func (c *transCookiesCache) get(serviceURL string) (http.Cookie, error) {
	u, err := url.Parse(serviceURL)
	if err != nil {
		return emptyCookie, ErrInvalidServiceURL
	}
	hostname := u.Hostname()
	if len(hostname) <= len("translate.google") || hostname[:len("translate.google")] != "translate.google" {
		return emptyCookie, ErrInvalidServiceURL
	}

	c.m.RLock()
	cookie, ok := c.cookies[hostname[len("translate"):]]
	c.m.RUnlock()
	if ok && cookie.Expires.After(time.Now()) {
		return cookie, nil
	}

	return c.update(serviceURL, 0)
}

func (c *transCookiesCache) update(serviceURL string, sleep time.Duration) (http.Cookie, error) {
	c.m.Lock()
	defer c.m.Unlock()

	time.Sleep(sleep)
	response, err := c.clt.Get(serviceURL)
	if err != nil {
		return emptyCookie, err
	}
	response.Body.Close()
	cookieStr := response.Header.Get("Set-Cookie")
	cookie, err := c.parseCookieStr(cookieStr)
	if err != nil {
		return emptyCookie, err
	}
	c.cookies[cookie.Domain] = cookie

	return cookie, nil
}

func (*transCookiesCache) parseCookieStr(cookieStr string) (http.Cookie, error) {
	return parseCookieStr(cookieStr)
}

// parseCookieStr
// for example:
// 		cookieStr="NID=204=Au7rQwn2eharnT1rtKsoQl32M2ASoamoFj5Rk8LKHZgg7YZfo54k88aqBVcUEYxcLKjpSU5dNgGTrRAu4Uiv7G3fIAeT3L87gsJCdqg_dCJ9tMHTufW8pHIUD1KgCDwUSIH60d4cWVsukZpai43pm9vHr3SLHCQk9ueEpYJ5Cx8; expires=Thu, 25-Feb-2021 15:15:28 GMT; path=/; domain=.google.cn; HttpOnly"
func parseCookieStr(cookieStr string) (http.Cookie, error) {
	var (
		l, m, r int
		cookie  = http.Cookie{HttpOnly: true}
	)
	for r < len(cookieStr) {
		for m < len(cookieStr) && cookieStr[m] != '=' {
			m++
		}
		if m >= len(cookieStr) {
			break
		}
		k := cookieStr[l:m]

		for r < len(cookieStr) && cookieStr[r] != ';' {
			r++
		}
		v := cookieStr[m+1 : r]

		switch k {
		case "expires":
			var err error
			cookie.Expires, err = time.Parse("Mon, 02-Jan-2006 15:04:05 MST", v)
			if err != nil {
				return emptyCookie, err
			}
		case "path":
			cookie.Path = v
		case "domain":
			cookie.Domain = v
		default:
			cookie.Name = k
			cookie.Value = v
		}

		for r < len(cookieStr) && (cookieStr[r] == ' ' || cookieStr[r] == ';') {
			r++
		}
		l = r
		m = r
	}

	return cookie, nil
}
