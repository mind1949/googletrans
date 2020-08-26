package googletrans

import (
	"testing"

	"golang.org/x/text/language"
)

func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		params := TranslateParams{
			Src:  "auto",
			Dest: language.SimplifiedChinese.String(),
			Text: "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software. ",
		}
		Translate(params)
	}
}
