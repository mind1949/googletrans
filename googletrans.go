package googletrans

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"text/scanner"

	"github.com/mind1949/googletrans/tk"
	"github.com/mind1949/googletrans/tkk"
)

const (
	// DefaultServiceURL default google translation service url
	DefaultServiceURL = "https://translate.google.cn"
)

var (
	emptyTranlated     = Translated{}
	emptyDetected      = Detected{}
	emptyRawTranslated = rawTranslated{}
)

// TranslateParams represents translate params
type TranslateParams struct {
	Src  string `json:"src"`  // source language (default: auto)
	Dest string `json:"dest"` // destination language
	Text string `json:"text"` // text for translating
}

// Translated represents translated result
type Translated struct {
	Params TranslateParams `json:"params"`
	Text   string          `json:"text"` // translated text
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
	serviceURL string
	tkkCache   tkk.Cache
}

// New initializes a Translator
func New(serviceURL string) *Translator {
	return &Translator{
		serviceURL: serviceURL,
		tkkCache:   tkk.NewCache(serviceURL),
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
		Params: params,
		Text:   transData.translated.text,
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
	transURL, err := t.buildTransURL(params)
	if err != nil {
		return emptyRawTranslated, err
	}

	clt := http.Client{}
	req, err := http.NewRequest(http.MethodGet, transURL, nil)
	if err != nil {
		return emptyRawTranslated, err
	}

	resp, err := clt.Do(req)
	if err != nil {
		return emptyRawTranslated, err
	}

	if resp.StatusCode != http.StatusOK {
		return emptyRawTranslated, errors.New("request status: " + resp.Status)
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

func (t *Translator) buildTransURL(params TranslateParams) (transURL string, err error) {
	tkk, err := t.tkkCache.Get()
	if err != nil {
		return "", err
	}
	tk, _ := tk.Get(params.Text, tkk)

	u, err := url.Parse(t.serviceURL + "/translate_a/single")
	if err != nil {
		return "", err
	}

	if params.Src == "" {
		params.Src = "auto"
	}
	values := url.Values{}
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
		"q":      params.Text,
		"tk":     tk,
	} {
		values.Add(k, v)
	}
	dts := []string{"at", "bd", "ex", "ld", "md", "qca", "rw", "rm", "ss", "t"}
	for i := 0; i < len(dts); i++ {
		values.Add("dt", dts[i])
	}

	u.RawQuery = values.Encode()

	return u.String(), nil
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
