# Googletrans
[![language](https://img.shields.io/badge/language-Golang-blue)](https://golang.org/)
[![Documentation](https://godoc.org/github.com/mind1949/googletrans?status.svg)](https://godoc.org/github.com/mind1949/googletrans)
[![Go Report Card](https://goreportcard.com/badge/github.com/mind1949/googletrans)](https://goreportcard.com/report/github.com/mind1949/googletrans)

G文⚡️: Concurrency-safe, free and unlimited golang library that implemented Google Translate API.

Inspired by [py-googletrans](https://github.com/ssut/py-googletrans).

# Features
* Out of the box
* Auto language detection
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

// Output:
// language: "en", confidence: 1.00
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

// Output:
// text: "Go是一种开放源代码编程语言，可轻松构建简单，可靠且高效的软件。"
// pronunciation: "Go shì yī zhǒng kāifàng yuán dàimǎ biānchéng yǔyán, kě qīngsōng gòujiàn jiǎndān, kěkào qiě gāoxiào de ruǎnjiàn."
```

## Customize service URLs
```golang
package main

import "github.com/mind1949/googletrans"

func main() {
	serviceURLs := []string{
		"https://translate.google.com/",
		"https://translate.google.pl/"}
	googletrans.Append(serviceURLs...)
}
```
