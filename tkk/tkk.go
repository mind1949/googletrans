// Package tkk gets google translate tkk
package tkk

import (
	"errors"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"
)

// Get gets tkk
func Get() (string, error) {
	return defaultCache.Get()
}

// Set sets google translation url
func Set(googleTransURL string) {
	defaultCache.Set(googleTransURL)
}

const defaultServiceURL = "https://translate.google.cn"

var (
	defaultCache = NewCache(defaultServiceURL)

	// ErrNotFound couldn't found tkk
	ErrNotFound = errors.New("couldn't found tkk from google translation url")
)

// Cache is responsible for getting google translte tkk
type Cache interface {
	Set(googleTransURL string)
	Get() (tkk string, err error)
}

// NewCache initializes a cache
func NewCache(serviceURL string) Cache {
	if serviceURL == "" {
		serviceURL = defaultServiceURL
	}
	return &tkkCache{v: "0", u: serviceURL}
}

type tkkCache struct {
	v string
	m sync.RWMutex
	u string // google translation url
}

// Set sets google translation url
func (t *tkkCache) Set(googleTransURL string) {
	t.u = googleTransURL
}

// Get gets tkk
func (t *tkkCache) Get() (string, error) {
	now := math.Floor(float64(
		time.Now().UnixNano() / 3600000),
	)
	ttkf64, err := strconv.ParseFloat(t.read(), 64)
	if err != nil {
		return "", err
	}
	if now == ttkf64 {
		return t.read(), nil
	}

	resp, err := http.Get(t.u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	data := string(body)
	r := regexp.MustCompile(`tkk:'(\d+\.\d+)'`)
	if r.MatchString(data) {
		v := r.FindStringSubmatch(data)[1]
		return t.update(v), nil
	}

	return "", ErrNotFound
}

func (t *tkkCache) read() string {
	t.m.RLock()
	t.m.RUnlock()

	return t.v
}

func (t *tkkCache) update(v string) string {
	t.m.Lock()
	t.v = v
	t.m.Unlock()

	return t.v
}
