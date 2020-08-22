// Package tk generates google translate tk
package tk

import (
	"errors"
	"strconv"
	"strings"
)

var (
	// ErrInvalidTkk means function Get’s param tkk is invalid
	ErrInvalidTkk = errors.New("Param tkk is invalid")
)

// Get generates google translate tk
func Get(s string, tkk string) (tk string, err error) {
	_, err = strconv.ParseFloat(tkk, 64)
	if err != nil {
		return "", ErrInvalidTkk
	}

	var a []int
	for _, vRune := range s {
		v := int(vRune)
		if v < 0x10000 {
			a = append(a, v)
		} else {
			a = append(a, (v-0x10000)/0x400+0xD800)
			a = append(a, (v-0x10000)%0x400+0xDC00)
		}
	}

	var e []int
	for g := 0; g < len(a); g++ {
		l := a[g]
		if l < 128 {
			e = append(e, l)
		} else {
			if l < 2048 {
				e = append(e, l>>6|192)
			} else {
				if (l&64512) == 55296 && g+1 < len(a) && a[g+1]&64512 == 56320 {
					g++
					l = 65536 + ((l & 1023) << 10) + (a[g] & 1023)
					e = append(e, l>>18|240)
					e = append(e, l>>12&63|128)
				} else {
					e = append(e, l>>12|224)
				}
				e = append(e, l>>6&63|128)
			}
			e = append(e, l&63|128)
		}
	}

	var (
		tkkl     int
		tkkpaire []int
	)
	for _, str := range strings.Split(tkk, ".") {
		tkkpaire = append(tkkpaire, s2int(str))
	}
	if len(tkkpaire) > 1 {
		tkkl = tkkpaire[0]
	}

	var tkklc = tkkl
	for i := 0; i < len(e); i++ {
		tkklc += e[i]
		tkklc = xr(tkklc, "+-a^+6")
	}
	tkklc = xr(tkklc, "+-3^+b+-f")

	if len(tkkpaire) > 1 {
		tkklc ^= tkkpaire[1]
	} else {
		tkklc ^= 0
	}

	if tkklc < 0 {
		tkklc = (tkklc & 2147483647) + 2147483648
	}

	tkklc %= 1000000

	return strconv.Itoa(tkklc) + "." + strconv.Itoa(tkklc^tkkl), nil
}

func xr(a int, b string) int {
	for c := 0; c < len(b)-2; c += 3 {
		d := string(b[c+2])
		var dd int
		if "a" <= d {
			dd = int(d[0]) - 87
		} else {
			dd = s2int(d)
		}

		if "+" == string(b[c+1]) {
			dd = (a % 0x100000000) >> dd
		} else {
			dd = a << dd
		}

		if "+" == string(b[c]) {
			a = (a + dd) & 4294967295
		} else {
			a = a ^ dd
		}
	}
	return a
}

func s2int(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
