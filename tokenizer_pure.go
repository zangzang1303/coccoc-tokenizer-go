package coccoc

import (
	_ "embed"
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
)

//go:embed dicts/words.txt
var embeddedWordsTxt string

var (
	baseDictOnce        sync.Once
	baseCompounds       map[string]bool
	baseUnaccentMap     map[string]string
	baseMaxCompoundLen  int = 6
)

func initBaseDictionary(dictPath string) {
	baseDictOnce.Do(func() {
		baseCompounds = make(map[string]bool, 50000)
		baseUnaccentMap = make(map[string]string, 50000)

		loadedFromFile := false
		if dictPath != "" {
			candidates := []string{
				filepath.Join(dictPath, "words.txt"),
				filepath.Join(dictPath, "vietnamese_compound_words.txt"),
			}
			for _, cand := range candidates {
				if f, err := os.Open(cand); err == nil {
					scanner := bufio.NewScanner(f)
					for scanner.Scan() {
						line := strings.ToLower(strings.TrimSpace(scanner.Text()))
						if line != "" && strings.Contains(line, " ") {
							baseCompounds[line] = true
							unacc := removeDiacritics(line)
							baseUnaccentMap[unacc] = line
							numWords := len(strings.Fields(line))
							if numWords > baseMaxCompoundLen {
								baseMaxCompoundLen = numWords
							}
						}
					}
					f.Close()
					if len(baseCompounds) > 0 {
						loadedFromFile = true
						break
					}
				}
			}
		}

		if !loadedFromFile && embeddedWordsTxt != "" {
			scanner := bufio.NewScanner(strings.NewReader(embeddedWordsTxt))
			for scanner.Scan() {
				line := strings.ToLower(strings.TrimSpace(scanner.Text()))
				if line != "" && strings.Contains(line, " ") {
					baseCompounds[line] = true
					unacc := removeDiacritics(line)
					baseUnaccentMap[unacc] = line
					numWords := len(strings.Fields(line))
					if numWords > baseMaxCompoundLen {
						baseMaxCompoundLen = numWords
					}
				}
			}
		}

		coreCompounds := []string{
			"học sinh", "sinh viên", "cà phê", "phê bình", "phê duyệt", "phê chuẩn",
			"bàn ghế", "bàn bạc", "bàn luận", "bàn thảo", "bàn tán",
			"trà sữa", "trà đá", "sữa chua", "sửa chữa", "sửa sang",
			"trường học", "y sinh", "hy sinh", "sinh nhật", "khuyến học",
			"học bổng", "học phí", "đi làm", "nhập học", "thời sự", "kinh tế",
			"thời khóa biểu", "công nghệ thông tin", "trí tuệ nhân tạo",
			"việt nam", "hà nội", "hồ chí minh", "đà nẵng", "hải phòng",
		}
		for _, w := range coreCompounds {
			baseCompounds[w] = true
			baseUnaccentMap[removeDiacritics(w)] = w
		}
	})
}

// pureTokenizer là engine tách từ thuần Go (Pure Go)
// Áp dụng thuật toán Forward Maximum Matching (FMM) trên từ điển từ ghép tiếng Việt chuẩn
type pureTokenizer struct {
	opts            Options
	customCompounds map[string]bool
	customUnaccent  map[string]string
	maxWords        int
}

func newPureEngine(opts Options) (Tokenizer, error) {
	initBaseDictionary(opts.DictPath)

	maxW := baseMaxCompoundLen
	var customCompounds map[string]bool
	var customUnaccent map[string]string

	if len(opts.CustomWords) > 0 {
		customCompounds = make(map[string]bool, len(opts.CustomWords))
		customUnaccent = make(map[string]string, len(opts.CustomWords))

		for _, w := range opts.CustomWords {
			clean := strings.ToLower(strings.TrimSpace(w))
			if clean != "" && strings.Contains(clean, " ") {
				customCompounds[clean] = true
				customUnaccent[removeDiacritics(clean)] = clean
				nWords := len(strings.Fields(clean))
				if nWords > maxW {
					maxW = nWords
				}
			}
		}
	}

	return &pureTokenizer{
		opts:            opts,
		customCompounds: customCompounds,
		customUnaccent:  customUnaccent,
		maxWords:        maxW,
	}, nil
}

func (t *pureTokenizer) isCompoundMatched(candLower, candUnacc string) bool {
	if t.customCompounds != nil && t.customCompounds[candLower] {
		return true
	}
	if baseCompounds[candLower] {
		return true
	}
	if candLower == candUnacc {
		if t.customUnaccent != nil && t.customUnaccent[candLower] != "" {
			return true
		}
		if baseUnaccentMap[candLower] != "" {
			return true
		}
	}
	return false
}

type textSpan struct {
	startByte int
	endByte   int
	isWord    bool
	text      string
}

// splitIntoSpans chia nhỏ văn bản thành các spans từ vựng và khoảng trống/dấu câu
func splitIntoSpans(text string) []textSpan {
	var spans []textSpan
	n := len(text)
	if n == 0 {
		return spans
	}

	runes := []rune(text)
	rLen := len(runes)
	byteIdx := 0

	i := 0
	for i < rLen {
		r := runes[i]
		isW := unicode.IsLetter(r) || unicode.IsDigit(r)

		startByte := byteIdx
		startRune := i

		for i < rLen {
			currR := runes[i]
			currIsW := unicode.IsLetter(currR) || unicode.IsDigit(currR)
			if currIsW != isW {
				break
			}
			byteIdx += len(string(currR))
			i++
		}

		spanText := string(runes[startRune:i])
		spans = append(spans, textSpan{
			startByte: startByte,
			endByte:   byteIdx,
			isWord:    isW,
			text:      spanText,
		})
	}

	return spans
}

// Segment phân tách câu và ghép các từ ghép bằng dấu gạch dưới "_"
func (t *pureTokenizer) Segment(text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return text, nil
	}

	spans := splitIntoSpans(text)
	numSpans := len(spans)

	type wordRef struct {
		spanIdx int
		text    string
	}
	var wordRefs []wordRef
	for i, sp := range spans {
		if sp.isWord {
			wordRefs = append(wordRefs, wordRef{spanIdx: i, text: sp.text})
		}
	}

	numWords := len(wordRefs)
	matchedUntilWord := 0

	var out strings.Builder
	out.Grow(len(text) + 16)

	wIdx := 0
	spanIdx := 0

	for spanIdx < numSpans {
		sp := spans[spanIdx]
		if !sp.isWord {
			out.WriteString(sp.text)
			spanIdx++
			continue
		}

		matched := false
		if wIdx >= matchedUntilWord {
			// Thử Maximum Matching từ window maxWords xuống 2 từ
			for win := t.maxWords; win >= 2; win-- {
				if wIdx+win <= numWords {
					validSpacing := true
					for k := 0; k < win-1; k++ {
						currSpan := wordRefs[wIdx+k].spanIdx
						nextSpan := wordRefs[wIdx+k+1].spanIdx
						if nextSpan != currSpan+2 {
							validSpacing = false
							break
						}
						midSpan := spans[currSpan+1]
						if strings.TrimSpace(midSpan.text) != "" {
							validSpacing = false
							break
						}
					}

					if !validSpacing {
						continue
					}

					var candParts []string
					for k := 0; k < win; k++ {
						candParts = append(candParts, strings.ToLower(wordRefs[wIdx+k].text))
					}
					candLower := strings.Join(candParts, " ")
					candUnacc := removeDiacritics(candLower)

					if t.isCompoundMatched(candLower, candUnacc) {
						var compoundJoined []string
						for k := 0; k < win; k++ {
							compoundJoined = append(compoundJoined, wordRefs[wIdx+k].text)
						}
						out.WriteString(strings.Join(compoundJoined, "_"))

						lastSpan := wordRefs[wIdx+win-1].spanIdx
						spanIdx = lastSpan + 1
						wIdx += win
						matchedUntilWord = wIdx
						matched = true
						break
					}
				}
			}
		}

		if !matched {
			out.WriteString(sp.text)
			spanIdx++
			wIdx++
		}
	}

	return out.String(), nil
}

// Tokenize phân đoạn câu và trả về danh sách các tokens đã chuẩn hóa
func (t *pureTokenizer) Tokenize(text string) ([]string, error) {
	segmented, err := t.Segment(text)
	if err != nil {
		return nil, err
	}
	return extractTokens(segmented, t.opts.NormalizeLowercase, t.opts.KeepPunctuation), nil
}

// TokenizeFull trả về danh sách token chi tiết kèm offset và loại từ
func (t *pureTokenizer) TokenizeFull(text string) ([]TokenInfo, error) {
	if strings.TrimSpace(text) == "" {
		return []TokenInfo{}, nil
	}

	spans := splitIntoSpans(text)
	numSpans := len(spans)

	type wordRef struct {
		spanIdx int
		text    string
	}
	var wordRefs []wordRef
	for i, sp := range spans {
		if sp.isWord {
			wordRefs = append(wordRefs, wordRef{spanIdx: i, text: sp.text})
		}
	}

	numWords := len(wordRefs)
	var infos []TokenInfo

	wIdx := 0
	spanIdx := 0

	for spanIdx < numSpans {
		sp := spans[spanIdx]
		if !sp.isWord {
			tType := TokenTypePunct
			if strings.TrimSpace(sp.text) == "" {
				tType = TokenTypeSpace
			}
			infos = append(infos, TokenInfo{
				Text:      sp.text,
				StartByte: sp.startByte,
				EndByte:   sp.endByte,
				Type:      tType,
				SegType:   0,
			})
			spanIdx++
			continue
		}

		matched := false
		for win := t.maxWords; win >= 2; win-- {
			if wIdx+win <= numWords {
				validSpacing := true
				for k := 0; k < win-1; k++ {
					currSpan := wordRefs[wIdx+k].spanIdx
					nextSpan := wordRefs[wIdx+k+1].spanIdx
					if nextSpan != currSpan+2 {
						validSpacing = false
						break
					}
					midSpan := spans[currSpan+1]
					if strings.TrimSpace(midSpan.text) != "" {
						validSpacing = false
						break
					}
				}

				if !validSpacing {
					continue
				}

				var candParts []string
				for k := 0; k < win; k++ {
					candParts = append(candParts, strings.ToLower(wordRefs[wIdx+k].text))
				}
				candLower := strings.Join(candParts, " ")
				candUnacc := removeDiacritics(candLower)

				if t.isCompoundMatched(candLower, candUnacc) {
					var compoundJoined []string
					for k := 0; k < win; k++ {
						compoundJoined = append(compoundJoined, wordRefs[wIdx+k].text)
					}
					firstSpan := spans[wordRefs[wIdx].spanIdx]
					lastSpan := spans[wordRefs[wIdx+win-1].spanIdx]

					infos = append(infos, TokenInfo{
						Text:      strings.Join(compoundJoined, "_"),
						StartByte: firstSpan.startByte,
						EndByte:   lastSpan.endByte,
						Type:      TokenTypeWord,
						SegType:   0,
					})

					spanIdx = wordRefs[wIdx+win-1].spanIdx + 1
					wIdx += win
					matched = true
					break
				}
			}
		}

		if !matched {
			tokType := TokenTypeWord
			if isAllDigits(sp.text) {
				tokType = TokenTypeNumber
			}
			infos = append(infos, TokenInfo{
				Text:      sp.text,
				StartByte: sp.startByte,
				EndByte:   sp.endByte,
				Type:      tokType,
				SegType:   0,
			})
			spanIdx++
			wIdx++
		}
	}

	return infos, nil
}

func (t *pureTokenizer) IsCGO() bool {
	return false
}

func (t *pureTokenizer) Close() error {
	return nil
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func removeDiacritics(s string) string {
	var b strings.Builder
	accents := map[rune]rune{
		'à': 'a', 'á': 'a', 'ả': 'a', 'ã': 'a', 'ạ': 'a',
		'ă': 'a', 'ằ': 'a', 'ắ': 'a', 'ẳ': 'a', 'ẵ': 'a', 'ặ': 'a',
		'â': 'a', 'ầ': 'a', 'ấ': 'a', 'ẩ': 'a', 'ẫ': 'a', 'ậ': 'a',
		'đ': 'd',
		'è': 'e', 'é': 'e', 'ẻ': 'e', 'ẽ': 'e', 'ẹ': 'e',
		'ê': 'e', 'ề': 'e', 'ế': 'e', 'ể': 'e', 'ễ': 'e', 'ệ': 'e',
		'ì': 'i', 'í': 'i', 'ỉ': 'i', 'ĩ': 'i', 'ị': 'i',
		'ò': 'o', 'ó': 'o', 'ỏ': 'o', 'õ': 'o', 'ọ': 'o',
		'ô': 'o', 'ồ': 'o', 'ố': 'o', 'ổ': 'o', 'ỗ': 'o', 'ộ': 'o',
		'ơ': 'o', 'ờ': 'o', 'ớ': 'o', 'ở': 'o', 'ỡ': 'o', 'ợ': 'o',
		'ù': 'u', 'ú': 'u', 'ủ': 'u', 'ũ': 'u', 'ụ': 'u',
		'ư': 'u', 'ừ': 'u', 'ứ': 'u', 'ử': 'u', 'ữ': 'u', 'ự': 'u',
		'ỳ': 'y', 'ý': 'y', 'ỷ': 'y', 'ỹ': 'y', 'ỵ': 'y',
	}
	for _, r := range strings.ToLower(s) {
		if mapped, ok := accents[r]; ok {
			b.WriteRune(mapped)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
