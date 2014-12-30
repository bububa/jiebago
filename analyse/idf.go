package analyse

import (
	"github.com/bububa/bufio"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func (this *Analyzer) SetIdf(idfFilePath string) error {
	if !filepath.IsAbs(idfFilePath) {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		idfFilePath = filepath.Clean(filepath.Join(pwd, idfFilePath))
	}
	idfFile, err := os.Open(idfFilePath)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(idfFile)
	freqs := make([]float64, 0)
	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Split(line, " ")
		word, freqStr := words[0], words[1]
		freq, err := strconv.ParseFloat(freqStr, 64)
		if err != nil {
			continue
		}
		this.idfFreq[word] = freq
		freqs = append(freqs, freq)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	sort.Float64s(freqs)
	this.medianIdf = freqs[len(freqs)/2]
	return nil
}

func (this *Analyzer) SetStopWords(stopWordsFilePath string) error {
	if !filepath.IsAbs(stopWordsFilePath) {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		stopWordsFilePath = filepath.Clean(filepath.Join(pwd, stopWordsFilePath))
	}
	stopWordsFile, err := os.Open(stopWordsFilePath)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(stopWordsFile)
	for scanner.Scan() {
		stopWord := scanner.Text()
		stopWord = strings.TrimSpace(stopWord)
		this.stopWords[stopWord] = stopWord
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
