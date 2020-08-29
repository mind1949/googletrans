package googletrans

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/scanner"
	"time"

	"github.com/mind1949/googletrans/tk"
	"github.com/mind1949/googletrans/tkk"
	"github.com/mind1949/googletrans/transcookie"
)

const (
	defaultServiceURL = "https://translate.google.cn"
)

var (
	emptyTranlated     = Translated{}
	emptyDetected      = Detected{}
	emptyRawTranslated = rawTranslated{}

	defaultTranslator = New()
)

// Translate uses defaultTranslator to translate params.text
func Translate(params TranslateParams) (Translated, error) {
	return defaultTranslator.Translate(params)
}

// Detect uses defaultTranslator to detect language
func Detect(text string) (Detected, error) {
	return defaultTranslator.Detect(text)
}

// BulkTranslate uses defaultTranslator to bulk translate
func BulkTranslate(ctx context.Context, params <-chan TranslateParams) <-chan TranslatedResult {
	return defaultTranslator.BulkTranslate(ctx, params)
}

// Append appends serviceURLs to defaultTranslator's serviceURLs
func Append(serviceURLs ...string) {
	defaultTranslator.Append(serviceURLs...)
}

// TranslateParams represents translate params
type TranslateParams struct {
	Src  string `json:"src"`  // source language (default: auto)
	Dest string `json:"dest"` // destination language
	Text string `json:"text"` // text for translating
}

// Translated represents translated result
type Translated struct {
	Params        TranslateParams `json:"params"`
	Text          string          `json:"text"`          // translated text
	Pronunciation string          `json:"pronunciation"` // pronunciation of translated text
}

// TranslatedResult represents a translated result with an error
type TranslatedResult struct {
	Translated
	Err error `json:"err"`
}

// Detected represents language detection result
type Detected struct {
	Lang       string  `json:"lang"`       // detected language
	Confidence float64 `json:"confidence"` // the confidence of detection result (0.00 to 1.00)
}

type rawTranslated struct {
	translated struct {
		text          string
		pronunciation string
	}
	detected struct {
		originalLanguage string
		confidence       float64
	}
}

// Translator is responsible for translation
type Translator struct {
	clt         *http.Client
	serviceURLs []string
	tkkCache    tkk.Cache
}

// New initializes a Translator
func New(serviceURLs ...string) *Translator {
	var has bool
	for i := 0; i < len(serviceURLs); i++ {
		if serviceURLs[i] == defaultServiceURL {
			has = true
			break
		}
	}
	if !has {
		serviceURLs = append(serviceURLs, defaultServiceURL)
	}

	return &Translator{
		clt:         &http.Client{},
		serviceURLs: serviceURLs,
		tkkCache:    tkk.NewCache(random(serviceURLs)),
	}
}

// Translate translates text from src language to dest language
func (t *Translator) Translate(params TranslateParams) (Translated, error) {
	if params.Src == "" {
		params.Src = "auto"
	}

	transData, err := t.do(params)
	if err != nil {
		return emptyTranlated, err
	}

	return Translated{
		Params:        params,
		Text:          transData.translated.text,
		Pronunciation: transData.translated.pronunciation,
	}, nil
}

// BulkTranslate translates texts to dest language
func (t *Translator) BulkTranslate(ctx context.Context, params <-chan TranslateParams) <-chan TranslatedResult {
	stream := make(chan TranslatedResult)

	go func() {
		defer close(stream)

		for param := range params {
			result := TranslatedResult{}
			result.Translated, result.Err = t.Translate(param)

			select {
			case <-ctx.Done():
				result.Err = ctx.Err()
				stream <- result
				return
			case stream <- result:
			}
		}
	}()
	return stream
}

// Detect detects text's language
func (t *Translator) Detect(text string) (Detected, error) {
	transData, err := t.do(TranslateParams{
		Src:  "auto",
		Dest: "en",
		Text: text,
	})
	if err != nil {
		return emptyDetected, err
	}
	return Detected{
		Lang:       transData.detected.originalLanguage,
		Confidence: transData.detected.confidence,
	}, nil
}

func (t *Translator) do(params TranslateParams) (rawTranslated, error) {
	req, err := t.buildTransRequest(params)
	if err != nil {
		return emptyRawTranslated, err
	}

	transService := req.URL.Scheme + "://" + req.URL.Hostname()
	var resp *http.Response
	for try := 0; try < 3; try++ {
		cookie, err := transcookie.Get(transService)
		if err != nil {
			return emptyRawTranslated, err
		}
		req.AddCookie(&cookie)
		resp, err = t.clt.Do(req)
		if err != nil {
			return emptyRawTranslated, err
		}

		if resp.StatusCode == http.StatusOK {
			break
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			_, err = transcookie.Update(transService, 3*time.Second)
			if err != nil {
				return emptyRawTranslated, err
			}
		}
	}
	if resp.StatusCode != http.StatusOK {
		return emptyRawTranslated, fmt.Errorf("failed to get translation result, err: %s", resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return emptyRawTranslated, err
	}
	resp.Body.Close()

	result, err := t.parseRawTranslated(data)
	if err != nil {
		return emptyRawTranslated, err
	}

	return result, nil
}

func (t *Translator) buildTransRequest(params TranslateParams) (request *http.Request, err error) {
	tkk, err := t.tkkCache.Get()
	if err != nil {
		return nil, err
	}
	tk, _ := tk.Get(params.Text, tkk)

	u, err := url.Parse(t.randomServiceURL() + "/translate_a/single")
	if err != nil {
		return nil, err
	}

	if params.Src == "" {
		params.Src = "auto"
	}
	queries := url.Values{}
	for k, v := range map[string]string{
		"client": "webapp",
		"sl":     params.Src,
		"tl":     params.Dest,
		"hl":     params.Dest,
		"ie":     "UTF-8",
		"oe":     "UTF-8",
		"otf":    "1",
		"ssel":   "0",
		"tsel":   "0",
		"kc":     "7",
		"tk":     tk,
	} {
		queries.Add(k, v)
	}
	dts := []string{"at", "bd", "ex", "ld", "md", "qca", "rw", "rm", "ss", "t"}
	for i := 0; i < len(dts); i++ {
		queries.Add("dt", dts[i])
	}

	q := url.Values{}
	q.Add("q", params.Text)

	// If the length of the url of the get request exceeds 2000, change to a post request
	if len(u.String()+"?"+queries.Encode()+q.Encode()) >= 2000 {
		u.RawQuery = queries.Encode()
		request, err = http.NewRequest(http.MethodPost, u.String(), strings.NewReader(q.Encode()))
		if err != nil {
			return nil, err
		}
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	} else {
		queries.Add("q", params.Text)
		u.RawQuery = queries.Encode()
		request, err = http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
	}

	return request, nil
}

func (*Translator) parseRawTranslated(data []byte) (result rawTranslated, err error) {
	var s scanner.Scanner
	s.Init(bytes.NewReader(data))
	var (
		coord = []int{-1}
	)
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		tokText := s.TokenText()
		switch tokText {
		case "[":
			coord[len(coord)-1]++
			coord = append(coord, -1)
		case "]":
			coord = coord[:len(coord)-1]
		case ",":
			// no-op
		default:
			coord[len(coord)-1]++

			if len(coord) == 4 && coord[1] == 0 && coord[3] == 0 {
				if tokText != "null" {
					result.translated.text += tokText[1 : len(tokText)-1]
				}
			}
			if len(coord) == 4 && coord[0] == 0 && coord[1] == 0 && coord[2] == 1 && coord[3] == 2 {
				if tokText != "null" {
					result.translated.pronunciation = tokText[1 : len(tokText)-1]
				}
			}
			if len(coord) == 4 && coord[0] == 0 && coord[1] == 0 && coord[3] == 2 {
				if tokText != "null" {
					result.translated.pronunciation = tokText[1 : len(tokText)-1]
				}
			}
			if len(coord) == 2 && coord[0] == 0 && coord[1] == 2 {
				result.detected.originalLanguage = tokText[1 : len(tokText)-1]
			}
			if len(coord) == 2 && coord[0] == 0 && coord[1] == 6 {
				result.detected.confidence, _ = strconv.ParseFloat(s.TokenText(), 64)
			}
		}
	}

	return result, nil
}

// Append appends serviceURLS to  t's serviceURLs
func (t *Translator) Append(serviceURLs ...string) {
	t.serviceURLs = append(t.serviceURLs, serviceURLs...)
}

func (t *Translator) randomServiceURL() (serviceURL string) {
	return random(t.serviceURLs)
}

func random(list []string) string {
	i := rand.Intn(len(list))
	return list[i]
}
