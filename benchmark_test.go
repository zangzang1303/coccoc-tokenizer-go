package coccoc_test

import (
	"strings"
	"testing"

	coccoc "github.com/vntk/coccoc-tokenizer-go"
)

var (
	shortText  = "Hôm nay tôi đi học uống cà phê sữa đá."
	mediumText = "Học sinh sinh viên Việt Nam tích cực tham gia các hoạt động nghiên cứu khoa học, ứng dụng trí tuệ nhân tạo và công nghệ thông tin vào việc giải quyết các bài toán thực tiễn của xã hội và nền kinh tế số."
	longText   = strings.Repeat("Học sinh sinh viên Việt Nam đi uống cà phê sữa đá thảo luận về thời khóa biểu và kinh tế. ", 5)
)

func BenchmarkSegmentShort(b *testing.B) {
	tok, err := coccoc.New(coccoc.WithFallbackToPureGo(true))
	if err != nil {
		b.Fatalf("Khởi tạo thất bại: %v", err)
	}
	defer tok.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = tok.Segment(shortText)
	}
}

func BenchmarkSegmentMedium(b *testing.B) {
	tok, err := coccoc.New(coccoc.WithFallbackToPureGo(true))
	if err != nil {
		b.Fatalf("Khởi tạo thất bại: %v", err)
	}
	defer tok.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = tok.Segment(mediumText)
	}
}

func BenchmarkSegmentLong(b *testing.B) {
	tok, err := coccoc.New(coccoc.WithFallbackToPureGo(true))
	if err != nil {
		b.Fatalf("Khởi tạo thất bại: %v", err)
	}
	defer tok.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = tok.Segment(longText)
	}
}

func BenchmarkTokenize(b *testing.B) {
	tok, err := coccoc.New(coccoc.WithFallbackToPureGo(true))
	if err != nil {
		b.Fatalf("Khởi tạo thất bại: %v", err)
	}
	defer tok.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = tok.Tokenize(mediumText)
	}
}

func BenchmarkTokenizeFull(b *testing.B) {
	tok, err := coccoc.New(coccoc.WithFallbackToPureGo(true))
	if err != nil {
		b.Fatalf("Khởi tạo thất bại: %v", err)
	}
	defer tok.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = tok.TokenizeFull(mediumText)
	}
}

func BenchmarkPureGoSegment(b *testing.B) {
	tok, err := coccoc.New(coccoc.WithForcePureGo(true))
	if err != nil {
		b.Fatalf("Khởi tạo thất bại: %v", err)
	}
	defer tok.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = tok.Segment(mediumText)
	}
}
