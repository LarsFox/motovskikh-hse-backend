package mp

import (
	"strings"
	"fmt"
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
	}

	detector.blacklist.Load(config.Blacklist)
	detector.whitelist.Load(config.Whitelist, normalizer)

	return detector
}

// Check проверяет никнейм.
func (d *Detector) Check(nickname string) *DetectionResult {
	if nickname == "" {
		return &DetectionResult{IsAllowed: false} // Пустой ник нельзя.
	}
	// Нормализуем.
	normalized := d.normalizer.Normalize(nickname)

	// Проверка по белому списку.
	if d.whitelist.Contains(normalized) {
		return &DetectionResult{IsAllowed: true}
	}

	// Быстрая проверка на подстроку.
	if res := d.fastCheck(normalized); res != nil {
		return res
	}

	// Попытка выделить корень стеммером и найти в подстроке.
	stemmed := d.stemmer.Stem(normalized)
	if stemmed != normalized {
		if res := d.fastCheck(stemmed); res != nil {
			res.Reason = "stem_match"
			return res
		}
	}
	fmt.Println(stemmed)

	// Проверка через расстояние Левенштейна.
	if res := d.levCheck(normalized); res != nil {
		return res
	}

	return &DetectionResult{IsAllowed: true}
}

// Быстрая проверка по подстроке.
func (d *Detector) fastCheck(text string) *DetectionResult {
	// Пробегаемся по черному списку.
	for word := range d.blacklist.words {
		// Если короткое слово (из 2 букв), то пропускаем.
		if len([]rune(word)) < 3 {
			continue
		}
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

// Levenshtein-проверка.
func (d *Detector) levCheck(text string) *DetectionResult {
	// Проходимся по черному списку.
	for word := range d.blacklist.words {
		// если короткое слово, то пропускаем (отловили бы до этого).
		if len([]rune(word)) <= 3 {
			continue
		}
		if levContains(text, word) {
			return &DetectionResult{
				IsAllowed: false,
				Reason:    "fuzzy_match",
				Word:      word,
			}
		}
	}
	return nil
}

// Проверяем по расстоянию Левенштейна.
func levContains(text, word string) bool {
	tr := []rune(text)
	wr := []rune(word)

	for i := 0; i <= len(tr)-len(wr); i++ {
		sub := string(tr[i : i+len(wr)])
		if len([]rune(sub)) < 3 {
			continue
		}

		if levenshtein(sub, word) <= 1 {
			return true
		}
	}
	return false
}

func levenshtein(a, b string) int {
	da := make([][]int, len(a)+1)
	for i := range da {
		da[i] = make([]int, len(b)+1)
	}

	for i := 0; i <= len(a); i++ {
		da[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		da[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}

			da[i][j] = min(
				da[i-1][j]+1,
				da[i][j-1]+1,
				da[i-1][j-1]+cost,
			)
		}
	}

	return da[len(a)][len(b)]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}
