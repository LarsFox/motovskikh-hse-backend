package mp

// Структура Blacklist - черный список слов.
type Blacklist struct {
	words      map[string]bool
	normalizer Normalizer
}

// Инициализация черного списка.
func NewBlacklist(normalizer Normalizer) *Blacklist {
	return &Blacklist{
		words:      make(map[string]bool),
		normalizer: normalizer,
	}
}

// Загрузка.
func (b *Blacklist) Load(words map[string]bool) {

	b.words = make(map[string]bool)

	for word, strict := range words {
		norm := b.normalizer.Normalize(word)
		if norm == "" {
			continue
		}
		b.words[norm] = strict
	}
}

// Белый список.
type Whitelist struct {
	words map[string]bool
}

// Инициализация.
func NewWhitelist() *Whitelist {
	return &Whitelist{
		words: make(map[string]bool),
	}
}

func (w *Whitelist) Load(words []string, normalizer Normalizer) {
	w.words = make(map[string]bool)
	for _, word := range words {
		norm := normalizer.Normalize(word)
		if norm != "" {
			w.words[norm] = true
		}
	}
}

func (w *Whitelist) Contains(word string) bool {
	return w.words[word]
}