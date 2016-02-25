package stats

import (
	"fmt"
	"github.com/emirpasic/gods/maps/treemap"
	"strings"
)

type Stats struct {
	wordsCount map[string]int
	wordsTree  *treemap.Map

	appendCh chan string
	askCh    chan ask
}

func NewStats() *Stats {
	tmp := &Stats{
		wordsCount: make(map[string]int),
		appendCh:   make(chan string, 1000),
		askCh:      make(chan ask),
		wordsTree:  treemap.NewWith(comparator),
	}

	go tmp.listen()
	return tmp
}

func (s *Stats) AppendWord(w string) {
	s.appendCh <- w
}

func (s *Stats) GetStats(responseCh chan []string, limit int) {
	s.askCh <- ask{responseCh: responseCh, limit: limit}
}

type ask struct {
	responseCh chan []string
	limit      int
}

type wordKey struct {
	count int
	word  string
}

func comparator(a, b interface{}) int {
	awk := a.(wordKey)
	bwk := b.(wordKey)

	switch {
	case awk.count < bwk.count:
		return 1
	case awk.count > bwk.count:
		return -1
	default:
		return strings.Compare(awk.word, bwk.word)
	}
}

func (s *Stats) listen() {
	for {
		select {
		case word := <-s.appendCh:
			s.doAppend(word)
		case ask := <-s.askCh:
			s.doAsk(ask)
		}
	}
}

func (s *Stats) doAppend(word string) {
	oldCount := s.wordsCount[word]
	newCount := oldCount + 1
	s.wordsCount[word] = newCount

	oldKey := wordKey{oldCount, word}
	newKey := wordKey{newCount, word}

	s.wordsTree.Remove(oldKey)
	s.wordsTree.Put(newKey, nil)
}

type limitWriter struct {
	result []string
	limit  int
}

func (w *limitWriter) append(word string) bool {
	n := len(w.result)

	w.result = w.result[0 : n+1]
	w.result[n] = word
	return n != w.limit-1
}

func (s *Stats) doAsk(ask ask) {
	writer := limitWriter{limit: ask.limit, result: make([]string, 0, ask.limit)}

OuterLoop:
	for _, wk := range s.wordsTree.Keys() {
		if !writer.append(wk.(wordKey).word) {
			break OuterLoop
		}
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Channel is already closed", r)
		}
	}()
	ask.responseCh <- writer.result
}
