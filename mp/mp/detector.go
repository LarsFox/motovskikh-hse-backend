package mp

import (
	"strings"
)

// DetectionResult - результат проверки.
type DetectionResult struct {
	IsAllowed bool
	Reason    string
	Word      string
}

// Detector - основной детектор.
type Detector struct {
	config     *Config
	normalizer Normalizer
	stemmer    *Stemmer
	blacklist  *Blacklist
	whitelist  *Whitelist
	rootBlacklist *RootBlacklist
}

// Структура Match - для вхождений. 
type Match struct {
	Start int
	End   int
	Word  string
}

// Инициализация детектора.
func NewDetector(config *Config) *Detector {
	normalizer := NewTextNormalizer(config)

	detector := &Detector{
		config:     config,
		normalizer: normalizer,
		stemmer:    NewStemmer(),
		blacklist:  NewBlacklist(normalizer),
		whitelist:  NewWhitelist(),
		rootBlacklist: NewRootBlacklist(),
	}

	detector.blacklist.Load(config.Blacklist)
	detector.whitelist.Load(config.Whitelist, normalizer)
	detector.rootBlacklist.Load(config.RootBlacklist, normalizer)

	return detector
}

// Check проверяет никнейм.
func (d *Detector) Check(nickname string) *DetectionResult {
	normalized := d.normalizer.Normalize(nickname)
	// Здесь мы находим все вхождения плохих слов в наш ник.
	badMatches := findAllMatches(normalized, d.blacklist.words)
	if len(badMatches) > 0 {
		// Здесь находим вхождение слов из белого списка.
		whiteMatches := findAllMatches(normalized, d.whitelist.words)
		// если таких слов нет — сразу бан (нашли плохие).
		if len(whiteMatches) == 0 {
			return &DetectionResult{
				IsAllowed: false,
				Reason:    "substring_match",
				Word:      badMatches[0].Word,
			}
		}
		// Если есть, то нам нужно смотреть, входит ли плохое в состав хорошего.
		// Потому что человек может одновременно дать нам и плохое, и хорошее слово.
		// И если наше плохое покрывается хорошим, то мы пропускаем.
		// Если нет (то есть нашли плохое вне хорошего), то бан.
		for _, bad := range badMatches {
			if !isCovered(bad, whiteMatches) {
				return &DetectionResult{
					IsAllowed: false,
					Reason:    "substring_match",
					Word:      bad.Word,
				}
			}
		}
		// Если слово прошло через эту адскую проверку, то дальше не проверяем, approved.
		return &DetectionResult{IsAllowed: true}
	}
	// Стемминг.
	stemmed := d.stemmer.Stem(normalized)
	if stemmed != normalized {
		if res := d.fastCheck(stemmed); res != nil {
			res.Reason = "stem_match"
			return res
		}
	}
	// Root-проверка.
	if d.rootBlacklist.Contains(stemmed) {
		return &DetectionResult{
			IsAllowed: false,
			Reason:    "forbidden_root",
			Word:      stemmed,
		}
	}

	return &DetectionResult{IsAllowed: true}
}
// Быстрая проверка по подстроке.
func (d *Detector) fastCheck(text string) *DetectionResult {
	// Пробегаемся по черному списку.
	for word := range d.blacklist.words {
		// Проверяем подстроки.
		if strings.Contains(text, word) {
			return &DetectionResult{
				IsAllowed: false,
				Reason:    "substring_match",
				Word:      word,
			}
		}
	}
	return nil
}

// Здесь находим вхождения плохих или хороших слов в ник.
func findAllMatches(text string, words map[string]bool) []Match {
	var matches []Match
	// Проходимся по списку и проверяем. Так как наши списки не очень большие будут, то по времени
	// проблем быть не должно.
	for word := range words {
		start := 0
		for {
			idx := strings.Index(text[start:], word)
			// Если вхождений нет больше, то выходим из цикла.
			if idx == -1 {
				break
			}
			realStart := start + idx
			realEnd := realStart + len(word)

			matches = append(matches, Match{
				Start: realStart,
				End:   realEnd,
				Word:  word,
			})

			start = realStart + 1
		}
	}
	return matches
}

// Проверка на покрытие плохого слова хорошим. (Примеры: сЕБАстьян, хлЕБАть, оскорБЛЯТЬ).
func isCovered(bad Match, whites []Match) bool {
	for _, w := range whites {
		if bad.Start >= w.Start && bad.End <= w.End {
			return true
		}
	}
	return false
}