package mp

import (
	"strings"
	"unicode"
)

// Normalizer интерфейс.
type Normalizer interface {
	Normalize(text string) string
}

// TextNormalizer - здесь мапы для нормализации.
type TextNormalizer struct {
	charReplacements map[string]string
	homoglyphs       map[rune]rune
}

// Инициализация нормализатора.
func NewTextNormalizer(config *Config) *TextNormalizer {
	return &TextNormalizer{
		charReplacements: config.CharReplacements,
		homoglyphs: map[rune]rune{
			'a': 'а',
			'e': 'е',
			'o': 'о',
			'p': 'р',
			'c': 'с',
			'x': 'х',
			'y': 'у',
			'k': 'к',
			'b': 'б',
			'm': 'м',
			't': 'т',
			'h': 'н',
			'i': 'и',
			'l': 'л',
		},
	}
}

// Normalize - нормализация ника.
func (n *TextNormalizer) Normalize(text string) string {
	if text == "" {
		return ""
	}
	text = strings.ToLower(text)

	// Замена цифр и спецсимволов.
	for from, to := range n.charReplacements {
		text = strings.ReplaceAll(text, from, to)
	}

	// Нормализация по homoglyph.
	var builder strings.Builder
	for _, r := range text {
		if repl, ok := n.homoglyphs[r]; ok {
			builder.WriteRune(repl)
		} else {
			builder.WriteRune(r)
		}
	}

	// Оставляем только буквы.
	clean := make([]rune, 0)
	for _, r := range builder.String() {
		if unicode.IsLetter(r) {
			clean = append(clean, r)
		}
	}

	// Убираем повторения.
	return collapseRepeats(string(clean))
}

// Убираем повторения слов.
func collapseRepeats(s string) string {
	var result []rune
	var prev rune
	count := 0

	for _, r := range s {
		if r == prev {
			count++
			if count < 2 {
				result = append(result, r)
			}
		} else {
			prev = r
			count = 1
			result = append(result, r)
		}
	}
	return string(result)
}
