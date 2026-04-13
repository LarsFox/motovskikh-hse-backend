package mp

import "github.com/kljensen/snowball"

// Стеммер для выделения корня.
type Stemmer struct{}

func NewStemmer() *Stemmer {
	return &Stemmer{}
}

func (s *Stemmer) Stem(word string) string {
	stem, err := snowball.Stem(word, "russian", true)
	if err != nil {
		return word
	}
	return stem
}