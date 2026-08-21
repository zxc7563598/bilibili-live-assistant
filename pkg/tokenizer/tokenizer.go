package tokenizer

import (
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/go-ego/gse"
)

var emojiPattern = regexp.MustCompile(`\[[^\]]*\]`)

type Tokenizer struct {
	segmenter gse.Segmenter
	stopWords map[string]struct{}
}

type WordFrequency struct {
	Word  string
	Count int64
}

type MessageFrequency struct {
	Message string
	Count   int64
}

type PhraseFrequency struct {
	Phrase string
	Count  int64
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

var (
	once      sync.Once
	shared    *Tokenizer
	sharedErr error
)

// Get 返回全局共享的分词器实例。词典只在首次调用时加载一次，
// 后续复用同一实例，避免每次请求重复加载词典的开销。
// 返回的实例加载完成后只读，可安全地并发调用分词相关方法。
func Get() (*Tokenizer, error) {
	once.Do(func() {
		shared, sharedErr = New()
	})
	return shared, sharedErr
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

// CutAndFilterAll 对多个文本分词并排序
func (t *Tokenizer) CutAndFilterAll(texts []string) []WordFrequency {
	var words []string
	for _, text := range texts {
		words = append(words, t.CutAndFilter(text)...)
	}
	result := make(map[string]int64)
	for _, word := range words {
		result[word]++
	}
	frequencies := make([]WordFrequency, 0, len(result))
	for word, count := range result {
		frequencies = append(frequencies, WordFrequency{
			Word:  word,
			Count: count,
		})
	}
	sort.Slice(frequencies, func(i, j int) bool {
		return frequencies[i].Count > frequencies[j].Count
	})
	return frequencies
}

// NGram 对文本进行分词过滤后，生成指定长度的连续 N-gram
func (t *Tokenizer) NGram(text string, n int) []string {
	if n <= 0 {
		return nil
	}
	words := t.CutAndFilter(text)
	if len(words) < n {
		return nil
	}
	result := make([]string, 0, len(words)-n+1)
	for i := 0; i <= len(words)-n; i++ {
		result = append(result, strings.Join(words[i:i+n], ""))
	}
	return result
}

// CountNGram 对多条文本生成指定长度的 N-gram 并进行排序
func (t *Tokenizer) CountNGram(texts []string, n int) []PhraseFrequency {
	counts := make(map[string]int64)
	for _, text := range texts {
		text = emojiPattern.ReplaceAllString(text, "")
		for _, phrase := range t.NGram(text, n) {
			counts[phrase]++
		}
	}
	result := make([]PhraseFrequency, 0, len(counts))
	for phrase, count := range counts {
		result = append(result, PhraseFrequency{
			Phrase: phrase,
			Count:  count,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	return result
}

// CountMessages 统计多条文本中相同消息的出现次数并进行排序
func (t *Tokenizer) CountMessages(texts []string) []MessageFrequency {
	counts := make(map[string]int64)
	for _, text := range texts {
		text = emojiPattern.ReplaceAllString(text, "")
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		counts[text]++
	}
	result := make([]MessageFrequency, 0, len(counts))
	for message, count := range counts {
		result = append(result, MessageFrequency{
			Message: message,
			Count:   count,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	return result
}

// isValidWord 验证是否是有效字
func isValidWord(word string) bool {
	for _, r := range word {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return false
	}
	return true
}
