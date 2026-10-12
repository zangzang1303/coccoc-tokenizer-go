#ifndef COCCOC_BRIDGE_H
#define COCCOC_BRIDGE_H

#ifdef __cplusplus
extern "C" {
#endif

// ==============================================================================
// COCCOC_BRIDGE_H: C Interface cho Cốc Cốc Tokenizer (C++ -> C -> Go CGO)
// Đảm bảo ABI thuần C (không name mangling) và quản lý bộ nhớ heap an toàn
// ==============================================================================

/**
 * Token đại diện cho từng đơn vị từ vựng kèm vị trí byte và loại token
 */
typedef struct {
    char* text;          // Chuỗi UTF-8 của token
    int start_byte;      // Vị trí byte bắt đầu trong văn bản gốc
    int end_byte;        // Vị trí byte kết thúc trong văn bản gốc
    int type;            // 0: WORD, 1: NUMBER, 2: SPACE, 3: PUNCT
    int seg_type;        // Phân loại segment nội bộ Cốc Cốc
} CoccocToken;

/**
 * Kết quả phân tích trả về danh sách các tokens
 */
typedef struct {
    CoccocToken* tokens; // Mảng động các Token
    int count;           // Số lượng token trong mảng
} CoccocTokenResult;

/**
 * Khởi tạo Cốc Cốc Tokenizer từ thư mục chứa các file từ điển đã biên dịch.
 * 
 * @param dict_path: Đường dẫn đến thư mục chứa multiterm_trie.dump và syllable_trie.dump.
 * @param load_nontone: 0 (nhẹ ~45MB RAM), 1 (nạp thêm nontone ~200MB RAM).
 * @return 0 nếu thành công, số âm nếu lỗi.
 */
int coccoc_init(const char* dict_path, int load_nontone);

/**
 * Kiểm tra xem C++ Tokenizer đã được nạp từ điển thành công hay chưa.
 * @return 1 nếu đã khởi tạo, 0 nếu chưa.
 */
int coccoc_is_initialized(void);

/**
 * Phân tách văn bản tiếng Việt và ghép các từ ghép bằng dấu gạch dưới "_".
 * Ví dụ: "Học sinh đi học uống cà phê" -> "Học_sinh đi học uống cà_phê"
 * 
 * Chuỗi trả về được cấp phát trên C heap. Caller PHẢI gọi coccoc_free_string() để giải phóng.
 */
char* coccoc_tokenize_original(const char* text);

/**
 * Giải phóng chuỗi C được cấp phát bởi coccoc_tokenize_original().
 */
void coccoc_free_string(char* str);

/**
 * Phân tích chi tiết trả về danh sách tokens kèm byte offsets và kiểu token.
 * Caller PHẢI gọi coccoc_free_token_result() để giải phóng toàn bộ cấu trúc.
 */
CoccocTokenResult* coccoc_tokenize_full(const char* text);

/**
 * Giải phóng bộ nhớ của CoccocTokenResult.
 */
void coccoc_free_token_result(CoccocTokenResult* result);

/**
 * Reset trạng thái của Cốc Cốc Tokenizer.
 */
void coccoc_destroy(void);

#ifdef __cplusplus
}
#endif

#endif // COCCOC_BRIDGE_H
