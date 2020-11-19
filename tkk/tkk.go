// Package tkk gets google translate tkk
package tkk

import (
	"errors"
	"fmt"
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

	tkkRegexp = regexp.MustCompile(`tkk:'(\d+\.\d+)'`)
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

	token := make(chan struct{}, 1)
	token <- struct{}{}
	cache := &tkkCache{
		v: "0",
		u: serviceURL,

		m:     &sync.RWMutex{},
		cond:  sync.NewCond(&sync.Mutex{}),
		token: token,
	}

	return cache
}

type tkkCache struct {
	v string // google translate tkk
	u string // google translation url

	m     *sync.RWMutex
	cond  *sync.Cond
	token chan struct{} // update token
}

// Set sets google translation url
func (t *tkkCache) Set(googleTransURL string) {
	t.u = googleTransURL
}

// Get gets tkk
func (t *tkkCache) Get() (tkk string, err error) {
	t.m.RLock()
	isvalid := t.isvalid()
	t.m.RUnlock()
	if isvalid {
		return t.v, nil
	}

	return t.update()
}

func (t *tkkCache) isvalid() bool {
	now := math.Floor(float64(
		time.Now().Unix() * 1000 / 3600000),
	)
	ttkf64, err := strconv.ParseFloat(t.v, 64)
	if err != nil {
		return false
	}
	if now != math.Floor(ttkf64) {
		return false
	}

	return true
}

// update gets tkk from t.u
func (t *tkkCache) update() (string, error) {
	// only one goroutine is allowed to obtain the update token at the same time
	// other goroutines can only wait until the update of this goroutine ends
	select {
	case <-t.token:
		// no-op
	default:
		t.cond.L.Lock()
		defer t.cond.L.Unlock()
		t.cond.Wait()
		if t.isvalid() {
			return t.v, nil
		}
		<-t.token
	}
	t.m.Lock()
	defer func() {
		t.m.Unlock()
		t.token <- struct{}{}
	}()

	// try to get tkk within timeout
	var (
		start   = time.Now()
		sleep   = 1 * time.Second
		timeout = 1 * time.Minute

		err error
	)
	for time.Now().Sub(start) < timeout {
		t.v, err = func() (string, error) {
			req, err := http.NewRequest(http.MethodGet, t.u, nil)
			if err != nil {
				return "", err
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 400 {
				format := "couldn't found tkk from google translation url, status code: %d"
				err = fmt.Errorf(format, resp.StatusCode)
				return "", err
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}
			data := string(body)
			if !tkkRegexp.MatchString(data) {
				return "", ErrNotFound
			}

			tkk := tkkRegexp.FindStringSubmatch(data)[1]
			return tkk, nil
		}()
		if err == nil {
			// if the update is successful,
			// notify all goroutines waiting for the update
			t.cond.Broadcast()
			return t.v, nil
		}

		time.Sleep(sleep)
	}
	if err != nil {
		// if the update fails,
		// notify one goroutine that is waiting to perform the update
		t.cond.Signal()
		return "", err
	}

	return t.v, nil
}
