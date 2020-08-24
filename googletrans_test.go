package googletrans

import (
	"testing"
)

func TestDo(t *testing.T) {
	clt := New(DefaultServiceURL).(*translator)
	params := TranslateParams{
		Src:  "auto",
		Dest: "zh-CN",
		Text: "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software. ",
	}
	transData, err := clt.do(params)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v\n", transData)
}

func TestTranslate(t *testing.T) {
	clt := New(DefaultServiceURL)
	params := TranslateParams{
		Src:  "auto",
		Dest: "zh-CN",
		Text: "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software. ",
	}
	translated, err := clt.Translate(params)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v\n", translated)
}

func TestDetect(t *testing.T) {
	clt := New(DefaultServiceURL)
	text := "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software. "
	detected, err := clt.Detect(text)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v\n", detected)
}
