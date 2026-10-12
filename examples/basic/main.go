package main

import (
	"fmt"
	"log"

	coccoc "github.com/zangzang1303/coccoc-tokenizer-go"
)

func main() {
	fmt.Println("=== Cốc Cốc Tokenizer Go - Basic Example ===")

	// 1. Khởi tạo Tokenizer với tùy chọn tự động fallback sang Pure-Go nếu môi trường không có CGO
	tok, err := coccoc.New(
		coccoc.WithFallbackToPureGo(true),
		coccoc.WithNormalizeLowercase(true),
	)
	if err != nil {
		log.Fatalf("Khởi tạo tokenizer thất bại: %v", err)
	}
	defer tok.Close()

	engineType := "Pure-Go Engine (Zero-CGO)"
	if tok.IsCGO() {
		engineType = "C++ Native CGO Engine"
	}
	fmt.Printf("Đang hoạt động trên: %s\n\n", engineType)

	sampleText := "Sinh viên trường đại học Bách Khoa Hà Nội đi uống cà phê sữa đá và thảo luận về trí tuệ nhân tạo."
	fmt.Printf("Văn bản gốc:\n  %s\n\n", sampleText)

	// 2. Tách từ ghép (Segment)
	segmented, err := tok.Segment(sampleText)
	if err != nil {
		log.Fatalf("Lỗi phân tách từ: %v", err)
	}
	fmt.Printf("Kết quả Segment (nối từ ghép bằng '_'):\n  %s\n\n", segmented)

	// 3. Lấy danh sách tokens đã chuẩn hóa (Tokenize)
	tokens, err := tok.Tokenize(sampleText)
	if err != nil {
		log.Fatalf("Lỗi tokenize: %v", err)
	}
	fmt.Printf("Danh sách tokens (%d từ):\n  %v\n\n", len(tokens), tokens)

	// 4. Lấy chi tiết thông tin token và byte offset (TokenizeFull)
	infos, err := tok.TokenizeFull(sampleText)
	if err != nil {
		log.Fatalf("Lỗi tokenize full: %v", err)
	}

	fmt.Println("Chi tiết Tokens (Byte Spans & Types):")
	fmt.Printf("%-25s | %-12s | %-12s | %-8s\n", "Token", "Start Byte", "End Byte", "Type")
	fmt.Println("----------------------------------------------------------------------")
	for _, info := range infos {
		if info.Type == coccoc.TokenTypeSpace {
			continue // Bỏ qua dấu cách để bảng gọn gàng
		}
		fmt.Printf("%-25s | %-12d | %-12d | %-8s\n",
			info.Text, info.StartByte, info.EndByte, info.Type.String())
	}
}
