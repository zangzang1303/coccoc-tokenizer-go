package coccoc_test

import (
	"strings"
	"sync"
	"testing"

	coccoc "github.com/vntk/coccoc-tokenizer-go"
)

func TestTokenizerBasicCompounds(t *testing.T) {
	tok, err := coccoc.New(coccoc.WithFallbackToPureGo(true))
	if err != nil {
		t.Fatalf("Không thể khởi tạo Tokenizer: %v", err)
	}
	defer tok.Close()

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Học sinh đi học",
			expected: "Học_sinh đi học",
		},
		{
			input:    "Tôi thích uống cà phê",
			expected: "Tôi thích uống cà_phê",
		},
		{
			input:    "Sinh viên Việt Nam học tập chăm chỉ",
			expected: "Sinh_viên Việt_Nam học_tập chăm_chỉ",
		},
		{
			input:    "Xem thời khóa biểu ngày mai",
			expected: "Xem thời_khóa_biểu ngày_mai",
		},
	}

	for _, tc := range tests {
		got, err := tok.Segment(tc.input)
		if err != nil {
			t.Errorf("Segment(%q) lỗi: %v", tc.input, err)
			continue
		}
		if !strings.EqualFold(got, tc.expected) {
			t.Errorf("Segment(%q) = %q; mong muốn %q", tc.input, got, tc.expected)
		}
	}
}

func TestTokenizerAmbiguityAndOverlaps(t *testing.T) {
	tok, err := coccoc.New(coccoc.WithFallbackToPureGo(true))
	if err != nil {
		t.Fatalf("Khởi tạo thất bại: %v", err)
	}
	defer tok.Close()

	tests := []struct {
		input string
	}{
		{"học sinh học sinh học"},
		{"bàn ghế bàn bạc việc quan trọng"},
		{"trường học và học bổng khuyến học"},
	}

	for _, tc := range tests {
		got, err := tok.Segment(tc.input)
		if err != nil {
			t.Errorf("Segment(%q) lỗi: %v", tc.input, err)
		}
		if got == "" {
			t.Errorf("Segment(%q) không được trả về rỗng", tc.input)
		}
	}
}

func TestTokenizerPunctuationAndSpecialChars(t *testing.T) {
	tok, err := coccoc.New(coccoc.WithFallbackToPureGo(true))
	if err != nil {
		t.Fatalf("Khởi tạo thất bại: %v", err)
	}
	defer tok.Close()

	input := "Hà Nội, thủ đô của Việt Nam! Giá: 50.000đ/ly cà phê. Email: info@coccoc.com"
	seg, err := tok.Segment(input)
	if err != nil {
		t.Fatalf("Segment lỗi: %v", err)
	}

	if !strings.Contains(strings.ToLower(seg), "hà_nội") {
		t.Errorf("Kỳ vọng chứa 'Hà_Nội', nhận được: %s", seg)
	}
	if !strings.Contains(strings.ToLower(seg), "việt_nam") {
		t.Errorf("Kỳ vọng chứa 'Việt_Nam', nhận được: %s", seg)
	}
	if !strings.Contains(strings.ToLower(seg), "cà_phê") {
		t.Errorf("Kỳ vọng chứa 'cà_phê', nhận được: %s", seg)
	}
}

func TestTokenizerMixedLanguages(t *testing.T) {
	tok, err := coccoc.New(coccoc.WithFallbackToPureGo(true))
	if err != nil {
		t.Fatalf("Khởi tạo thất bại: %v", err)
	}
	defer tok.Close()

	input := "Hôm nay tôi deploy code lên OpenSearch cluster với Docker"
	tokens, err := tok.Tokenize(input)
	if err != nil {
		t.Fatalf("Tokenize lỗi: %v", err)
	}

	joined := strings.Join(tokens, " ")
	expectedKeywords := []string{"deploy", "code", "opensearch", "cluster", "docker"}
	for _, kw := range expectedKeywords {
		if !strings.Contains(joined, kw) {
			t.Errorf("Thiếu từ khóa mượn tiếng Anh: %s trong tokens: %v", kw, tokens)
		}
	}
}

func TestTokenizeFullOffsets(t *testing.T) {
	tok, err := coccoc.New(coccoc.WithFallbackToPureGo(true))
	if err != nil {
		t.Fatalf("Khởi tạo thất bại: %v", err)
	}
	defer tok.Close()

	input := "Uống cà phê ngon"
	infos, err := tok.TokenizeFull(input)
	if err != nil {
		t.Fatalf("TokenizeFull lỗi: %v", err)
	}

	if len(infos) == 0 {
		t.Fatalf("Kỳ vọng tokens không rỗng")
	}

	// Xác thực các byte offsets nằm trong giới hạn chuỗi
	inputBytes := []byte(input)
	for _, info := range infos {
		if info.StartByte < 0 || info.EndByte > len(inputBytes) || info.StartByte > info.EndByte {
			t.Errorf("Offset không hợp lệ cho token %q: [%d, %d] trên độ dài %d",
				info.Text, info.StartByte, info.EndByte, len(inputBytes))
		}
	}
}

func TestTokenizerConcurrency(t *testing.T) {
	tok, err := coccoc.New(coccoc.WithFallbackToPureGo(true))
	if err != nil {
		t.Fatalf("Khởi tạo thất bại: %v", err)
	}
	defer tok.Close()

	const numGoroutines = 50
	const iterations = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(id int) {
			defer wg.Done()
			sentences := []string{
				"Học sinh đi học lúc bảy giờ sáng",
				"Uống cà phê sữa đá ở quán quen",
				"Sinh viên Việt Nam khởi nghiệp công nghệ thông tin",
				"Thành phố Hồ Chí Minh là trung tâm kinh tế lớn",
			}
			for i := 0; i < iterations; i++ {
				sent := sentences[i%len(sentences)]
				seg, err := tok.Segment(sent)
				if err != nil || seg == "" {
					t.Errorf("Goroutine %d Segment lỗi: %v", id, err)
					return
				}
				tokens, err := tok.Tokenize(sent)
				if err != nil || len(tokens) == 0 {
					t.Errorf("Goroutine %d Tokenize lỗi: %v", id, err)
					return
				}
			}
		}(g)
	}

	wg.Wait()
}

func TestCustomWordsOption(t *testing.T) {
	tok, err := coccoc.New(
		coccoc.WithForcePureGo(true),
		coccoc.WithCustomWords("máy bay không người lái", "xe điện tự hành"),
	)
	if err != nil {
		t.Fatalf("Khởi tạo thất bại: %v", err)
	}
	defer tok.Close()

	input := "Thử nghiệm máy bay không người lái và xe điện tự hành mới"
	seg, err := tok.Segment(input)
	if err != nil {
		t.Fatalf("Segment lỗi: %v", err)
	}

	if !strings.Contains(seg, "máy_bay_không_người_lái") {
		t.Errorf("Kỳ vọng ghép custom word 'máy_bay_không_người_lái', nhận được: %s", seg)
	}
	if !strings.Contains(seg, "xe_điện_tự_hành") {
		t.Errorf("Kỳ vọng ghép custom word 'xe_điện_tự_hành', nhận được: %s", seg)
	}
}
