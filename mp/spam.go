package mp

import (
	"slices"
	"strings"
)

// Если ложь, то только точное соответствие по словам в строке.
var spamWords = map[string]bool{
	"блядина":   true,
	"бляди":     true,
	"блядь":     true,
	"блядский":  true,
	"блять":     false,
	"гандон":    true,
	"гондон":    true,
	"даун":      true,
	"долбаеб":   true,
	"долбаёб":   true,
	"долбоеб":   true,
	"долбоёб":   true,
	"еб":        false,
	"ебал":      false,
	"ебала":     false,
	"ебанат":    true,
	"ебаная":    true,
	"ебаное":    true,
	"ебаный":    true,
	"ебанная":   true,
	"ебанное":   true,
	"ебанный":   true,
	"ебарь":     true,
	"ебатель":   true,
	"ебеня":     true,
	"ебирь":     true,
	"еблан":     true,
	"ебля":      true,
	"ебу":       false,
	"ебучая":    true,
	"ебучее":    true,
	"ебучие":    true,
	"ебучий":    true,
	"ебырь":     true,
	"ёб":        false,
	"ёбарь":     true,
	"зигомет":   true,
	"зигомёт":   true,
	"маскаль":   true,
	"москаль":   true,
	"нахуй":     true,
	"отсос":     false,
	"отсосу":    false,
	"педик":     true,
	"пенис":     true,
	"пенисы":    true,
	"пидар":     true,
	"пидарас":   true,
	"пидор":     true,
	"пидоры":    true,
	"пидорас":   true,
	"пидорашка": true,
	"пизда":     true,
	"пиздец":    true,
	"писька":    true,
	"писюн":     true,
	"пися":      false,
	"уебан":     true,
	"уебок":     true,
	"хуеглот":   true,
	"хуила":     true,
	"хуеплет":   true,
	"хуесос":    true,
	"хуем":      true,
	"хуй":       true,
	"хуе":       true,
	"хуя":       true,
	"хуйло":     true,
	"хуйня":     true,
	"чурка":     true,
	"шлюшка":    true,
	"шлюха":     true,
	"щлюха":     true,
	"blyat":     true,
	"dolboeb":   true,
	"ebani":     true,
	"eblan":     true,
	"gondon":    true,
	"huesos":    true,
	"hui":       true,
	"huy":       true,
	"nahuy":     true,
	"penis":     true,
	"pidor":     true,
	"xyu":       true,
	"\u0fd6":    true,
	"\u534d":    true,
	"\u5350":    true,
}

// Лат: {Кир} → заменяем все Кир на Лат.
var homoglyphs = map[string][]string{
	"a": {"а", "4"},
	"b": {"в"},
	"c": {"с"},
	"e": {"е", "3"},
	"h": {"н"},
	"i": {"1"},
	"k": {"к"},
	"m": {"м"},
	"o": {"о", "0"},
	"p": {"р"},
	"t": {"т"},
	"x": {"х"},
	"y": {"у"},
}

func latCyr(s string) string {
	for lat, cyrs := range homoglyphs {
		for _, cyr := range cyrs {
			s = strings.ReplaceAll(s, cyr, lat)
		}
	}

	return s
}

func InitTranslitWords() {
	replWords := make(map[string]bool, len(spamWords))

	for word, v := range spamWords {
		repl := latCyr(word)
		replWords[repl] = v
	}

	spamWords = replWords
}

func HasSpam(text string) bool {
	text = latCyr(strings.ToLower(text))
	split := strings.Split(text, " ")

	for curse, strict := range spamWords {
		if strict {
			if strings.Contains(text, curse) {
				return true
			}

			continue
		}

		if slices.Contains(split, curse) {
			return true
		}
	}

	return false
}
