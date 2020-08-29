# Googletrans
[![language](https://img.shields.io/badge/language-Golang-blue)](https://golang.org/)
[![Documentation](https://godoc.org/github.com/mind1949/googletrans?status.svg)](https://godoc.org/github.com/mind1949/googletrans)
[![Go Report Card](https://goreportcard.com/badge/github.com/mind1949/googletrans)](https://goreportcard.com/report/github.com/mind1949/googletrans)

G文⚡️: Concurrency-safe, free and unlimited golang library that implemented Google Translate API.

Inspired by [py-googletrans](https://github.com/ssut/py-googletrans).

# Features
* Out of the box
* Auto language detection
* Bulk translations
* Customizable service URL
 
# Installation

```
go get -u github.com/mind1949/googletrans
```

# Usage

## Detect language
```golang
package main

import (
	"fmt"

	"github.com/mind1949/googletrans"
)

func main() {
	detected, err := googletrans.Detect("hello googletrans")
	if err != nil {
		panic(err)
	}

	format := "language: %q, confidence: %0.2f\n"
	fmt.Printf(format, detected.Lang, detected.Confidence)
}

// output:
// language: "en", cofidence: 1.00
```

## Translate
```golang
package main

import (
	"fmt"

	"github.com/mind1949/googletrans"
	"golang.org/x/text/language"
)

func main() {
	params := googletrans.TranslateParams{
		Src:  "auto",
		Dest: language.SimplifiedChinese.String(),
		Text: "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software. ",
	}
	translated, err := googletrans.Translate(params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("text: %q \npronunciation: %q", translated.Text, translated.Pronunciation)
}

// output:
// text: "Go是一种开放源代码编程语言，可轻松构建简单，可靠且高效的软件。"
// pronunciation: "Go shì yī zhǒng kāifàng yuán dàimǎ biānchéng yǔyán, kě qīngsōng gòujiàn jiǎndān, kěkào qiě gāoxiào de ruǎnjiàn."
```

## Bulk translate
```golang
package main

import (
	"context"
	"fmt"

	"github.com/mind1949/googletrans"
	"golang.org/x/text/language"
)

func main() {
	params := func() <-chan googletrans.TranslateParams {
		params := make(chan googletrans.TranslateParams)
		go func() {
			defer close(params)
			texts := []string{
				"Hello golang",
				"Go is an open source programming language that makes it easy to build simple, reliable, and efficient software.",
				"The Go programming language is an open source project to make programmers more productive.",
			}
			for i := 0; i < len(texts); i++ {
				params <- googletrans.TranslateParams{
					Src:  "auto",
					Dest: language.SimplifiedChinese.String(),
					Text: texts[i],
				}
			}
		}()
		return params
	}()

	for transResult := range googletrans.BulkTranslate(context.Background(), params) {
		if transResult.Err != nil {
			panic(transResult.Err)
		}
		fmt.Printf("text: %q, pronunciation: %q\n", transResult.Text, transResult.Pronunciation)
	}
}

// output:
// text: "你好golang", pronunciation: "Nǐ hǎo golang"
// text: "Go是一种开放源代码编程语言，可轻松构建简单，可靠且高效的软件。", pronunciation: "Go shì yī zhǒng kāifàng yuán dàimǎ biānchéng yǔyán, kě qīngsōng gòujiàn jiǎndān, kěkào qiě gāoxiào de ruǎnjiàn."
// text: "Go编程语言是一个开放源代码项目，旨在提高程序员的生产力。", pronunciation: "Go biānchéng yǔyán shì yīgè kāifàng yuán dàimǎ xiàngmù, zhǐ zài tígāo chéngxù yuán de shēngchǎnlì."

```
