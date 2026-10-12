package coccoc

import (
	"fmt"
	"strings"
	"unicode"
)

// New khởi tạo một instance Tokenizer tiếng Việt theo Functional Options Pattern
func New(opts ...Option) (Tokenizer, error) {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(&cfg)
	}

	// Nếu người dùng yêu cầu bắt buộc dùng Pure-Go
	if cfg.ForcePureGo {
		return newPureEngine(cfg)
	}

	// Thử khởi tạo CGO Engine nếu CGO được kích hoạt lúc build
	tok, err := newCGOEngine(cfg)
	if err == nil {
		return tok, nil
	}

	// Nếu CGO lỗi hoặc không khả dụng nhưng bật FallbackToPureGo -> tự động chuyển sang Pure-Go
	if cfg.FallbackToPureGo {
		return newPureEngine(cfg)
	}

	return nil, fmt.Errorf("coccoc: khởi tạo CGO tokenizer thất bại: %w (đặt WithFallbackToPureGo(true) để tự động fallback)", err)
}

// NewPure khởi tạo trực tiếp Pure-Go Tokenizer độc lập với CGO
func NewPure(opts ...Option) (Tokenizer, error) {
	opts = append(opts, WithForcePureGo(true))
	return New(opts...)
}

// extractTokens chuẩn hóa chuỗi đã phân đoạn thành danh sách tokens
func extractTokens(segmented string, lowercase bool, keepPunct bool) []string {
	if strings.TrimSpace(segmented) == "" {
		return []string{}
	}

	text := segmented
	if lowercase {
		text = strings.ToLower(text)
	}

	if keepPunct {
		return strings.Fields(text)
	}

	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_'
	})
}
