package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	coccoc "github.com/zangzang1303/coccoc-tokenizer-go"
)

// OpenSearchMessageDocument cấu trúc tài liệu đồng bộ chuẩn Task 1
type OpenSearchMessageDocument struct {
	ID        int64  `json:"id"`
	ThreadID  string `json:"thread_id"`
	UserID    string `json:"user_id"`
	Text      string `json:"text"`      // Văn bản thô (Standard Analyzer)
	TextVi    string `json:"text_vi"`   // Văn bản ghép từ Cốc Cốc
	TextRaw   string `json:"text_raw"`  // Văn bản ghép từ + bỏ dấu
	TextEdge  string `json:"text_edge"` // Văn bản phục vụ Edge N-gram search tiền tố
	CreatedAt int64  `json:"create_at"`
}

func main() {
	fmt.Println("=== OpenSearch / Elasticsearch Index Pre-processing Pipeline ===")

	tok, err := coccoc.New(coccoc.WithFallbackToPureGo(true))
	if err != nil {
		log.Fatalf("Khởi tạo tokenizer thất bại: %v", err)
	}
	defer tok.Close()

	rawMessages := []string{
		"Học sinh trường chuyên uống trà sữa và cà phê sữa đá.",
		"Tôi yêu Việt Nam và thích đi du lịch Đà Nẵng.",
		"Bàn ghế mới mua phục vụ học sinh ngồi học bài.",
	}

	for i, raw := range rawMessages {
		// 1. Sinh text_vi: Phân đoạn từ ghép bằng Cốc Cốc Tokenizer
		seg, err := tok.Segment(raw)
		if err != nil {
			log.Fatalf("Lỗi segment: %v", err)
		}

		// 2. Sinh text_raw: Bỏ dấu từ text_vi
		textRaw := removeDiacritics(seg)

		// 3. Chuẩn bị tài liệu JSON nạp OpenSearch Bulk API
		doc := OpenSearchMessageDocument{
			ID:        int64(i + 1),
			ThreadID:  "thread_general_1",
			UserID:    fmt.Sprintf("user_%d", i+10),
			Text:      raw,
			TextVi:    seg,
			TextRaw:   textRaw,
			TextEdge:  textRaw,
			CreatedAt: 1728000000 + int64(i*60),
		}

		docJSON, _ := json.MarshalIndent(doc, "", "  ")
		fmt.Printf("\n--- Document #%d ---\n%s\n", doc.ID, string(docJSON))
	}

	fmt.Println("\n✅ Đã tiền xử lý thành công. Tài liệu sẵn sàng gửi qua OpenSearch Bulk API!")
}

func removeDiacritics(s string) string {
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
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if mapped, ok := accents[r]; ok {
			b.WriteRune(mapped)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
