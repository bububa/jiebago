package jiebago

import (
	"fmt"
	"github.com/bububa/jiebago/finalseg"
	"regexp"
	"sort"
)

type Jieba struct {
	Dictionary     string
	TT             *TopTrie
	UserWordTagTab map[string]string
}

func NewJieba() *Jieba {
	return &Jieba{UserWordTagTab: make(map[string]string)}
}

type Route struct {
	Freq  float64
	Index int
}

func (route Route) String() string {
	return fmt.Sprintf("(%f, %d)", route.Freq, route.Index)
}

type Routes []*Route

func (routes Routes) Len() int {
	return len(routes)
}

func (routes Routes) Less(i, j int) bool {
	routei := routes[i]
	routej := routes[j]
	if routei.Freq > routej.Freq {
		return true
	} else if routei.Freq == routej.Freq {
		return routei.Index > routej.Index
	}
	return false
}

func (routes Routes) Swap(i, j int) {
	routes[i], routes[j] = routes[j], routes[i]
}

func RegexpSplit(r *regexp.Regexp, sentence string) []string {
	result := make([]string, 0)
	locs := r.FindAllStringIndex(sentence, -1)
	lastLoc := 0
	if len(locs) == 0 {
		return []string{sentence}
	}
	for _, loc := range locs {
		if loc[0] == lastLoc {
			result = append(result, sentence[loc[0]:loc[1]])
		} else {
			result = append(result, sentence[lastLoc:loc[0]])
			result = append(result, sentence[loc[0]:loc[1]])
		}
		lastLoc = loc[1]
	}
	if lastLoc < len(sentence) {
		result = append(result, sentence[lastLoc:])
	}

	return result
}

func RegexpSplitN(re *regexp.Regexp, s string, n int) []string {

	if n == 0 {
		return nil
	}

	if len(re.String()) > 0 && len(s) == 0 {
		return []string{""}
	}

	matches := re.FindAllStringIndex(s, n)
	strings := make([]string, 0, len(matches))

	beg := 0
	end := 0
	for _, match := range matches {
		if n > 0 && len(strings) >= n-1 {
			break
		}

		end = match[0]
		if match[1] != 0 {
			strings = append(strings, s[beg:end])
		}
		beg = match[1]
	}

	if end != len(s) {
		strings = append(strings, s[beg:])
	}

	return strings
}

func (this *Jieba) GetDAG(sentence string) map[int][]int {
	dag := make(map[int][]int)
	runes := []rune(sentence)
	n := len(runes)
	p := this.TT.T
	i, j := 0, 0
	var c rune
	for {
		if i >= n {
			break
		}
		c = runes[j]
		if _, ok := p.Nodes[c]; ok {
			p = p.Nodes[c]
			if p.IsLeaf {
				if _, inDag := dag[i]; !inDag {
					dag[i] = []int{j}
				} else {
					dag[i] = append(dag[i], j)
				}
			}
			j += 1
			if j >= n {
				i += 1
				j = i
				p = this.TT.T
			}
		} else {
			p = this.TT.T
			i += 1
			j = i
		}
	}
	for i := 0; i < n; i++ {
		if _, ok := dag[i]; !ok {
			dag[i] = []int{i}
		}
	}
	return dag
}

func (this *Jieba) Calc(sentence string, dag map[int][]int, idx int) map[int]*Route {
	runes := []rune(sentence)
	number := len(runes)
	routes := make(map[int]*Route)
	routes[number] = &Route{0.0, 0}
	for idx := number - 1; idx >= 0; idx-- {
		candidates := make(Routes, 0)
		for _, i := range dag[idx] {
			var word string
			if i <= idx-1 {
				word = string(runes[i+1 : idx])
			} else {
				word = string(runes[idx : i+1])
			}
			var route *Route
			if _, ok := this.TT.Freq[word]; ok {
				route = &Route{this.TT.Freq[word] + routes[i+1].Freq, i}
			} else {
				route = &Route{this.TT.MinFreq + routes[i+1].Freq, i}
			}
			candidates = append(candidates, route)
		}
		//sort.Sort(sort.Reverse(candidates))
		sort.Sort(candidates)
		routes[idx] = candidates[0]
	}
	return routes
}

type cutAction func(sentence string) []string

func (this *Jieba) cut_DAG(sentence string) []string {
	dag := this.GetDAG(sentence)
	routes := this.Calc(sentence, dag, 0)
	x := 0
	var y int
	runes := []rune(sentence)
	length := len(runes)
	result := make([]string, 0)
	buf := make([]rune, 0)
	for {
		if x >= length {
			break
		}
		y = routes[x].Index + 1
		l_word := runes[x:y]
		if y-x == 1 {
			buf = append(buf, l_word...)
		} else {
			if len(buf) > 0 {
				if len(buf) == 1 {
					result = append(result, string(buf))
					buf = make([]rune, 0)
				} else {
					bufString := string(buf)
					if _, ok := this.TT.Freq[bufString]; !ok {
						recognized := finalseg.Cut(bufString)
						for _, t := range recognized {
							result = append(result, t)
						}
					} else {
						for _, elem := range buf {
							result = append(result, string(elem)) // TODO: I don't get this?
						}
					}
					buf = make([]rune, 0)
				}
			}
			result = append(result, string(l_word))
		}
		x = y
	}

	if len(buf) > 0 {
		if len(buf) == 1 {
			result = append(result, string(buf))
		} else {
			bufString := string(buf)
			if _, ok := this.TT.Freq[bufString]; !ok {
				recognized := finalseg.Cut(bufString)
				for _, t := range recognized {
					result = append(result, t)
				}
			} else {
				for _, elem := range buf {
					result = append(result, string(elem)) // TODO: I don't get this?
				}
			}
		}
	}
	return result
}

func (this *Jieba) cut_DAG_NO_HMM(sentence string) []string {
	result := make([]string, 0)
	re_eng := regexp.MustCompile(`[[:alnum:]]`)
	dag := this.GetDAG(sentence)
	routes := this.Calc(sentence, dag, 0)
	x := 0
	var y int
	runes := []rune(sentence)
	length := len(runes)
	buf := make([]rune, 0)
	for {
		if x >= length {
			break
		}
		y = routes[x].Index + 1
		l_word := runes[x:y]
		if re_eng.MatchString(string(l_word)) && len(l_word) == 1 {
			buf = append(buf, l_word...)
			x = y
		} else {
			if len(buf) > 0 {
				result = append(result, string(buf))
				buf = make([]rune, 0)
			}
			result = append(result, string(l_word))
			x = y
		}
	}
	if len(buf) > 0 {
		result = append(result, string(buf))
		buf = make([]rune, 0)
	}
	return result
}

func (this *Jieba) cut_All(sentence string) []string {
	result := make([]string, 0)
	runes := []rune(sentence)
	dag := this.GetDAG(sentence)
	old_j := -1
	ks := make([]int, 0)
	for k := range dag {
		ks = append(ks, k)
	}
	sort.Ints(ks)
	for k := range ks {
		l := dag[k]
		if len(l) == 1 && k > old_j {
			result = append(result, string(runes[k:l[0]+1]))
			old_j = l[0]
		} else {
			for _, j := range l {
				if j > k {
					result = append(result, string(runes[k:j+1]))
					old_j = j
				}
			}
		}
	}
	return result
}

func (this *Jieba) Cut(sentence string, cut_all bool, HMM bool) []string {
	result := make([]string, 0)
	var re_han, re_skip *regexp.Regexp
	if cut_all {
		re_han = regexp.MustCompile(`\p{Han}+`)
		re_skip = regexp.MustCompile(`[^[:alnum:]+#\n]`)
	} else {
		re_han = regexp.MustCompile(`([\p{Han}+[:alnum:]+#&\._]+)`)
		re_skip = regexp.MustCompile(`(\r\n|\s)`)
	}
	blocks := RegexpSplit(re_han, sentence)
	var cut_block []string

	for _, blk := range blocks {
		if len(blk) == 0 {
			continue
		}
		if re_han.MatchString(blk) {
			if HMM {
				cut_block = this.cut_DAG(blk)
			} else {
				cut_block = this.cut_DAG_NO_HMM(blk)
			}
			if cut_all {
				cut_block = this.cut_All(blk)
			}
			for _, word := range cut_block {
				result = append(result, word)
			}
		} else {
			type skipSplitFunc func(sentence string) []string
			var ssf skipSplitFunc
			if cut_all {
				ssf = func(sentence string) []string {
					//return re_skip.Split(sentence, -1)
					return RegexpSplitN(re_skip, sentence, -1)
				}
			} else {
				ssf = func(sentence string) []string {
					return RegexpSplit(re_skip, sentence)
				}
			}

			for _, x := range ssf(blk) {
				if re_skip.MatchString(x) {
					result = append(result, x)
				} else if !cut_all {
					for _, xx := range x {
						result = append(result, string(xx))
					}
				} else {
					result = append(result, x)
				}
			}
		}
	}
	return result
}

func (this *Jieba) CutForSearch(sentence string, hmm bool) []string {
	result := make([]string, 0)
	words := this.Cut(sentence, false, hmm)
	for _, word := range words {
		runes := []rune(word)
		for _, increment := range []int{2, 3} {
			if len(runes) > increment {
				var gram2 string
				for i := 0; i < len(runes)-increment+1; i++ {
					gram2 = string(runes[i : i+increment])
					if _, ok := this.TT.Freq[gram2]; ok {
						result = append(result, gram2)
					}
				}
			}
		}
		result = append(result, word)
	}
	return result
}

func (this *Jieba) SetDictionary(dict_path string) (TT *TopTrie, err error) {
	TT, err = newTopTrie(dict_path)
	this.TT = TT
	this.Dictionary = dict_path
	return
}
