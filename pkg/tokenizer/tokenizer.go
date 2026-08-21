package tokenizer

import (
	"regexp"
	"unicode"
	"unicode/utf8"

	"github.com/go-ego/gse"
)

var emojiPattern = regexp.MustCompile(`\[[^\]]*\]`)

type Tokenizer struct {
	segmenter gse.Segmenter

	stopWords map[string]struct{}
}

func New() (*Tokenizer, error) {
	t := &Tokenizer{
		stopWords: defaultStopWords,
	}

	if err := t.segmenter.LoadDict(); err != nil {
		return nil, err
	}

	return t, nil
}

// Cut 对文本进行分词。
func (t *Tokenizer) Cut(text string) []string {
	return t.segmenter.Cut(text, true)
}

// CutAndFilter 对文本进行分词并过滤停用词。
func (t *Tokenizer) CutAndFilter(text string) []string {
	text = emojiPattern.ReplaceAllString(text, "")
	words := t.Cut(text)
	result := make([]string, 0, len(words))
	for _, word := range words {
		if utf8.RuneCountInString(word) < 2 {
			continue
		}
		if !isValidWord(word) {
			continue
		}
		if _, ok := t.stopWords[word]; ok {
			continue
		}
		result = append(result, word)
	}
	return result
}

func isValidWord(word string) bool {
	for _, r := range word {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return false
	}
	return true
}

func (t *Tokenizer) CutAll(texts []string) []string {
	var words []string
	for _, text := range texts {
		words = append(words, t.Cut(text)...)
	}
	return words
}

func (t *Tokenizer) CutAndFilterAll(texts []string) []string {
	var words []string
	for _, text := range texts {
		words = append(words, t.CutAndFilter(text)...)
	}
	return words
}
