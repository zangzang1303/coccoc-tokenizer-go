package coccoc

// TokenType đại diện cho loại token được phân loại
type TokenType int

const (
	TokenTypeWord   TokenType = 0 // 0: Từ vựng (từ đơn hoặc từ ghép tiếng Việt)
	TokenTypeNumber TokenType = 1 // 1: Ký tự số
	TokenTypeSpace  TokenType = 2 // 2: Dấu cách / khoảng trắng
	TokenTypePunct  TokenType = 3 // 3: Dấu câu
	TokenTypeOther  TokenType = 4 // 4: Ký tự hoặc loại khác
)

// String trả về tên biểu diễn chuỗi của TokenType
func (t TokenType) String() string {
	switch t {
	case TokenTypeWord:
		return "WORD"
	case TokenTypeNumber:
		return "NUMBER"
	case TokenTypeSpace:
		return "SPACE"
	case TokenTypePunct:
		return "PUNCT"
	default:
		return "OTHER"
	}
}

// TokenInfo chứa thông tin chi tiết của một token kèm vị trí byte trong văn bản gốc
type TokenInfo struct {
	Text      string    `json:"text"`       // Chuỗi token
	StartByte int       `json:"start_byte"` // Vị trí byte bắt đầu trong văn bản gốc
	EndByte   int       `json:"end_byte"`   // Vị trí byte kết thúc trong văn bản gốc
	Type      TokenType `json:"type"`       // Phân loại token (từ, số, dấu cách, dấu câu)
	SegType   int       `json:"seg_type"`   // Segment type nội bộ Cốc Cốc
}

// Tokenizer đại diện cho bộ tách từ tiếng Việt chuẩn Idiomatic Go
type Tokenizer interface {
	// Segment phân tách câu và ghép các từ ghép bằng dấu gạch dưới ("_")
	// Ví dụ: "Tôi yêu Việt Nam" -> "Tôi yêu Việt_Nam"
	Segment(text string) (string, error)

	// Tokenize phân đoạn câu và trả về danh sách các tokens đã chuẩn hóa (mặc định lowercase, bỏ dấu câu)
	// Ví dụ: "Tôi yêu Việt Nam!" -> ["tôi", "yêu", "việt_nam"]
	Tokenize(text string) ([]string, error)

	// TokenizeFull trả về danh sách token chi tiết kèm offset và loại từ
	TokenizeFull(text string) ([]TokenInfo, error)

	// IsCGO trả về true nếu đang chạy trên C++ native engine, false nếu chạy Pure-Go
	IsCGO() bool

	// Close giải phóng tài nguyên native C++ (nếu có)
	Close() error
}
