package coccoc

// Option hàm cấu hình khởi tạo Tokenizer theo Functional Options Pattern
type Option func(*Options)

// Options tập hợp các tham số cấu hình của Cốc Cốc Tokenizer
type Options struct {
	DictPath           string   // Đường dẫn thư mục từ điển Cốc Cốc
	LoadNontone        bool     // Nạp bảng tần suất 200MB (mặc định: false - nhẹ ~45MB RAM)
	NormalizeLowercase bool     // Tự động chuyển token thành chữ thường trong Tokenize (mặc định: true)
	KeepPunctuation    bool     // Giữ dấu câu trong Tokenize (mặc định: false)
	FallbackToPureGo   bool     // Tự động chuyển sang Pure-Go nếu CGO không khả dụng (mặc định: true)
	ForcePureGo        bool     // Bắt buộc dùng Pure-Go engine kể cả khi có CGO (mặc định: false)
	CustomWords        []string // Danh sách từ ghép bổ sung cho Pure-Go
}

func defaultOptions() Options {
	return Options{
		DictPath:           "",
		LoadNontone:        false,
		NormalizeLowercase: true,
		KeepPunctuation:    false,
		FallbackToPureGo:   true,
		ForcePureGo:        false,
		CustomWords:        nil,
	}
}

// WithDictPath cấu hình đường dẫn thư mục chứa các tệp từ điển Cốc Cốc (.dump)
func WithDictPath(path string) Option {
	return func(o *Options) {
		o.DictPath = path
	}
}

// WithLoadNontone cấu hình nạp bảng tần suất nontone 2-gram (~200MB RAM)
// false (khuyên dùng cho search/chat): khởi động siêu nhanh ~150ms, RAM ~45MB
// true: nạp bảng 2-gram hỗ trợ sticky text (URL, tên miền)
func WithLoadNontone(load bool) Option {
	return func(o *Options) {
		o.LoadNontone = load
	}
}

// WithNormalizeLowercase cấu hình tự động lowercase tokens trong hàm Tokenize (mặc định: true)
func WithNormalizeLowercase(normalize bool) Option {
	return func(o *Options) {
		o.NormalizeLowercase = normalize
	}
}

// WithKeepPunctuation cấu hình giữ lại dấu câu khi gọi Tokenize (mặc định: false)
func WithKeepPunctuation(keep bool) Option {
	return func(o *Options) {
		o.KeepPunctuation = keep
	}
}

// WithFallbackToPureGo cấu hình tự động chuyển sang Pure-Go nếu môi trường không có CGO hoặc thiếu tệp C++
func WithFallbackToPureGo(fallback bool) Option {
	return func(o *Options) {
		o.FallbackToPureGo = fallback
	}
}

// WithForcePureGo bắt buộc sử dụng Pure-Go Engine (hữu ích cho testing/benchmarking so sánh)
func WithForcePureGo(force bool) Option {
	return func(o *Options) {
		o.ForcePureGo = force
	}
}

// WithCustomWords bổ sung danh sách từ ghép tùy chỉnh vào từ điển
func WithCustomWords(words ...string) Option {
	return func(o *Options) {
		o.CustomWords = append(o.CustomWords, words...)
	}
}
