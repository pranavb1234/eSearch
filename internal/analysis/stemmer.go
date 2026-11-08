package analysis

import (
	"strings"
)

// StemmerFilter applies the Porter Stemming algorithm to reduce words to their morphological base.
type StemmerFilter struct{}

// NewStemmerFilter instantiates a Porter Stemmer filter.
func NewStemmerFilter() *StemmerFilter {
	return &StemmerFilter{}
}

// Name returns the identifier of this filter.
func (f *StemmerFilter) Name() string {
	return "porter_stemmer"
}

// Filter transforms each token's term to its stemmed form.
func (f *StemmerFilter) Filter(tokens []Token) ([]Token, error) {
	if len(tokens) == 0 {
		return tokens, nil
	}

	result := make([]Token, len(tokens))
	for i, tok := range tokens {
		stemmed := Stem(tok.Term)
		result[i] = Token{
			Term:        stemmed,
			Position:    tok.Position,
			StartOffset: tok.StartOffset,
			EndOffset:   tok.EndOffset,
		}
	}
	return result, nil
}

// Stem returns the Porter-stemmed string for a given word.
func Stem(word string) string {
	lower := strings.ToLower(word)
	runes := []rune(lower)
	if len(runes) <= 2 {
		return lower
	}

	s := &stemmerState{b: runes, k: len(runes) - 1, k0: 0}
	s.step1ab()
	s.step1c()
	s.step2()
	s.step3()
	s.step4()
	s.step5()

	return string(s.b[:s.k+1])
}

type stemmerState struct {
	b  []rune
	k  int // current word end index
	k0 int // start index (usually 0)
	j  int // internal index used in step checks
}

func (s *stemmerState) cons(i int) bool {
	switch s.b[i] {
	case 'a', 'e', 'i', 'o', 'u':
		return false
	case 'y':
		if i == s.k0 {
			return true
		}
		return !s.cons(i - 1)
	default:
		return true
	}
}

// m measures the number of consonant sequences between k0 and j.
func (s *stemmerState) m() int {
	n := 0
	i := s.k0
	for {
		if i > s.j {
			return n
		}
		if !s.cons(i) {
			break
		}
		i++
	}
	i++
	for {
		for {
			if i > s.j {
				return n
			}
			if s.cons(i) {
				break
			}
			i++
		}
		i++
		n++
		for {
			if i > s.j {
				return n
			}
			if !s.cons(i) {
				break
			}
			i++
		}
		i++
	}
}

func (s *stemmerState) vowelInStem() bool {
	for i := s.k0; i <= s.j; i++ {
		if !s.cons(i) {
			return true
		}
	}
	return false
}

func (s *stemmerState) doublec(i int) bool {
	if i < s.k0+1 {
		return false
	}
	if s.b[i] != s.b[i-1] {
		return false
	}
	return s.cons(i)
}

func (s *stemmerState) cvc(i int) bool {
	if i < s.k0+2 || !s.cons(i) || s.cons(i-1) || !s.cons(i-2) {
		return false
	}
	ch := s.b[i]
	if ch == 'w' || ch == 'x' || ch == 'y' {
		return false
	}
	return true
}

func (s *stemmerState) ends(str string) bool {
	r := []rune(str)
	l := len(r)
	o := s.k - l + 1
	if o < s.k0 {
		return false
	}
	for i := 0; i < l; i++ {
		if s.b[o+i] != r[i] {
			return false
		}
	}
	s.j = s.k - l
	return true
}

func (s *stemmerState) setto(str string) {
	r := []rune(str)
	l := len(r)
	o := s.j + 1
	for i := 0; i < l; i++ {
		if o+i < len(s.b) {
			s.b[o+i] = r[i]
		} else {
			s.b = append(s.b, r[i])
		}
	}
	s.k = s.j + l
}

func (s *stemmerState) r(str string) {
	if s.m() > 0 {
		s.setto(str)
	}
}

func (s *stemmerState) step1ab() {
	if s.b[s.k] == 's' {
		if s.ends("sses") {
			s.k -= 2
		} else if s.ends("ies") {
			s.setto("i")
		} else if s.b[s.k-1] != 's' {
			s.k--
		}
	}

	if s.ends("eed") {
		if s.m() > 0 {
			s.k--
		}
	} else if (s.ends("ed") || s.ends("ing")) && s.vowelInStem() {
		s.k = s.j
		if s.ends("at") {
			s.setto("ate")
		} else if s.ends("bl") {
			s.setto("ble")
		} else if s.ends("iz") {
			s.setto("ize")
		} else if s.doublec(s.k) {
			s.k--
			ch := s.b[s.k]
			if ch == 'l' || ch == 's' || ch == 'z' {
				s.k++
			}
		} else if s.m() == 1 && s.cvc(s.k) {
			s.setto("e")
		}
	}
}

func (s *stemmerState) step1c() {
	if s.ends("y") && s.vowelInStem() {
		s.b[s.k] = 'i'
	}
}

func (s *stemmerState) step2() {
	switch s.b[s.k-1] {
	case 'a':
		if s.ends("ational") {
			s.r("ate")
		} else if s.ends("tional") {
			s.r("tion")
		}
	case 'c':
		if s.ends("enci") {
			s.r("ence")
		} else if s.ends("anci") {
			s.r("ance")
		}
	case 'e':
		if s.ends("izer") {
			s.r("ize")
		}
	case 'l':
		if s.ends("bli") {
			s.r("ble")
		} else if s.ends("alli") {
			s.r("al")
		} else if s.ends("entli") {
			s.r("ent")
		} else if s.ends("eli") {
			s.r("e")
		} else if s.ends("ousli") {
			s.r("ous")
		}
	case 'o':
		if s.ends("ization") {
			s.r("ize")
		} else if s.ends("ation") {
			s.r("ate")
		} else if s.ends("ator") {
			s.r("ate")
		}
	case 's':
		if s.ends("alism") {
			s.r("al")
		} else if s.ends("iveness") {
			s.r("ive")
		} else if s.ends("fulness") {
			s.r("ful")
		} else if s.ends("ousness") {
			s.r("ous")
		}
	case 't':
		if s.ends("aliti") {
			s.r("al")
		} else if s.ends("iviti") {
			s.r("ive")
		} else if s.ends("biliti") {
			s.r("ble")
		}
	}
}

func (s *stemmerState) step3() {
	switch s.b[s.k] {
	case 'e':
		if s.ends("icate") {
			s.r("ic")
		} else if s.ends("ative") {
			s.r("")
		} else if s.ends("alize") {
			s.r("al")
		}
	case 'i':
		if s.ends("iciti") {
			s.r("ic")
		}
	case 'l':
		if s.ends("ical") {
			s.r("ic")
		} else if s.ends("ful") {
			s.r("")
		}
	case 's':
		if s.ends("ness") {
			s.r("")
		}
	}
}

func (s *stemmerState) step4() {
	if s.k-1 < s.k0 {
		return
	}
	switch s.b[s.k-1] {
	case 'a':
		if s.ends("al") {
		} else {
			return
		}
	case 'c':
		if s.ends("ance") || s.ends("ence") {
		} else {
			return
		}
	case 'e':
		if s.ends("er") {
		} else {
			return
		}
	case 'i':
		if s.ends("ic") {
		} else {
			return
		}
	case 'l':
		if s.ends("able") || s.ends("ible") {
		} else {
			return
		}
	case 'n':
		if s.ends("ant") || s.ends("ement") || s.ends("ment") || s.ends("ent") {
		} else {
			return
		}
	case 'o':
		if s.ends("ion") && s.j >= s.k0 && (s.b[s.j] == 's' || s.b[s.j] == 't') {
		} else if s.ends("ou") {
		} else {
			return
		}
	case 's':
		if s.ends("ism") {
		} else {
			return
		}
	case 't':
		if s.ends("ate") || s.ends("iti") {
		} else {
			return
		}
	case 'u':
		if s.ends("ous") {
		} else {
			return
		}
	case 'v':
		if s.ends("ive") {
		} else {
			return
		}
	case 'z':
		if s.ends("ize") {
		} else {
			return
		}
	default:
		return
	}
	if s.m() > 1 {
		s.k = s.j
	}
}

func (s *stemmerState) step5() {
	s.j = s.k
	if s.b[s.k] == 'e' {
		a := s.m()
		if a > 1 || (a == 1 && !s.cvc(s.k-1)) {
			s.k--
		}
	}
	if s.b[s.k] == 'l' && s.doublec(s.k) && s.m() > 1 {
		s.k--
	}
}

