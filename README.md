# Googletrans

[![language](https://img.shields.io/badge/language-Golang-blue)](https://golang.org/)
[![Documentation](https://godoc.org/github.com/mind1949/googletrans?status.svg)](https://godoc.org/github.com/mind1949/googletrans)
[![Go Report Card](https://goreportcard.com/badge/github.com/mind1949/googletrans)](https://goreportcard.com/report/github.com/mind1949/googletrans)

Concurrency-safe, free and unlimited golang library that implemented Google Translate API.

Inspired by [py-googletrans](https://github.com/ssut/py-googletrans).

# Features
* Auto language detection
* Bulk translations
* Customizable service URL
* Auto update cookie
 
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

	fmt.Printf("%+v\n", detected)
}

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
	fmt.Printf("%+v\n", translated)
}
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

	for translatedResult := range googletrans.BulkTranslate(context.Background(), params) {
		fmt.Printf("%+v\n", translatedResult)
	}
}
```
