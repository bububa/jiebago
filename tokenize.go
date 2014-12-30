package jiebago

type Token struct {
	Word  string
	Start int
	End   int
}

func (this *Jieba) Tokenize(sentence string, mode string, HMM bool) []Token {
	tokens := make([]Token, 0)
	start := 0
	var width int
	if mode == "default" {
		for _, word := range this.Cut(sentence, false, HMM) {
			width = len([]rune(word))
			tokens = append(tokens, Token{word, start, start + width})
			start += width
		}
	} else {
		for _, word := range this.Cut(sentence, false, HMM) {
			runes := []rune(word)
			width = len(runes)
			for _, step := range []int{2, 3} {
				if width > step {
					for i := 0; i < width-step+1; i++ {
						gram := string(runes[i : i+step])
						if _, ok := this.TT.Freq[gram]; ok {
							tokens = append(tokens, Token{gram, start + i, start + i + step})
						}
					}
				}
			}
			tokens = append(tokens, Token{word, start, start + width})
		}
	}
	return tokens
}
