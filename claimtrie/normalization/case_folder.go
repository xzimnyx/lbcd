package normalization

import (
	"bytes"
	_ "embed"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

//go:embed CaseFolding_v11.txt
var v11 string

var foldMap map[rune][]rune

func init() {
	foldMap = map[rune][]rune{}
	r, _ := regexp.Compile(`([[:xdigit:]]+?); (.); ([[:xdigit:] ]+?);`)
	matches := r.FindAllStringSubmatch(v11, 1000000000)
	for i := range matches {
		if matches[i][2] == "C" || matches[i][2] == "F" {
			key, _ := strconv.Unquote(`"\u` + matches[i][1] + `"`)
			splits := strings.Split(matches[i][3], " ")
			var values []rune
			for j := range splits {
				value, _ := strconv.Unquote(`"\u` + splits[j] + `"`)
				values = append(values, []rune(value)[0])
			}
			foldMap[[]rune(key)[0]] = values
		}
	}
}

func CaseFold(name []byte) []byte {
	var b bytes.Buffer
	b.Grow(len(name))
	for i := 0; i < len(name); {
		r, w := utf8.DecodeRune(name[i:])
		if r == utf8.RuneError && w < 2 {
			// HACK: their RuneError is actually a valid character if coming from a width of 2 or more
			return name
		}
		replacements := foldMap[r]
		if len(replacements) > 0 {
			for j := range replacements {
				b.WriteRune(replacements[j])
			}
		} else {
			b.WriteRune(r)
		}
		i += w
	}
	return b.Bytes()
}
