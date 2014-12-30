package analyse

import (
	"github.com/bububa/jiebago"
	"sort"
	"strings"
	"unicode/utf8"
)

type TfIdf struct {
	word string
	freq float64
}

type TfIdfs []TfIdf

func (tis TfIdfs) Len() int {
	return len(tis)
}

func (tis TfIdfs) Less(i, j int) bool {
	if tis[i].freq == tis[j].freq {
		return tis[i].word > tis[j].word
	}
	return tis[i].freq > tis[j].freq
}

func (tis TfIdfs) Swap(i, j int) {
	tis[i], tis[j] = tis[j], tis[i]
}

type Analyzer struct {
	jieba     *jiebago.Jieba
	stopWords map[string]string
	idfFreq   map[string]float64
	medianIdf float64
}

func NewAnalyzer(jieba *jiebago.Jieba) *Analyzer {
	return &Analyzer{
		jieba:   jieba,
		idfFreq: make(map[string]float64),
		stopWords: map[string]string{
			"the": "the", "of": "of", "is": "is", "and": "and", "to": "to", "in": "in", "that": "that", "we": "we", "for": "for", "an": "an", "are": "are", "by": "bye", "be": "be", "as": "as", "on": "on", "with": "with", "can": "can", "if": "of", "from": "from", "which": "which", "you": "you", "it": "it", "this": "this", "then": "then", "at": "at", "have": "have", "all": "all", "not": "not", "one": "one", "has": "has", "or": "or",
		},
	}
}

func (this *Analyzer) ExtractTags(sentence string, topK int) []string {
	words := this.jieba.Cut(sentence, false, true)
	freq := make(map[string]float64)

	for _, w := range words {
		w = strings.TrimSpace(w)
		if utf8.RuneCountInString(w) < 2 {
			continue
		}
		if _, ok := this.stopWords[w]; ok {
			continue
		}
		if f, ok := freq[w]; ok {
			freq[w] = f + 1.0
		} else {
			freq[w] = 1.0
		}
	}
	total := 0.0
	for _, f := range freq {
		total += f
	}
	for k, v := range freq {
		freq[k] = v / total
	}
	tis := make(TfIdfs, 0)
	for k, v := range freq {
		var ti TfIdf
		if freq_, ok := this.idfFreq[k]; ok {
			ti = TfIdf{word: k, freq: freq_ * v}
		} else {
			ti = TfIdf{word: k, freq: this.medianIdf * v}
		}
		tis = append(tis, ti)
	}
	//sort.Sort(sort.Reverse(tis))
	sort.Sort(tis)
	var topTfIdfs TfIdfs
	if len(tis) > topK {
		topTfIdfs = tis[:topK]
	} else {
		topTfIdfs = tis
	}
	tags := make([]string, len(topTfIdfs))
	for index, ti := range topTfIdfs {
		tags[index] = ti.word
	}
	return tags
}
