//go:build cgo

package coccoc

/*
#cgo CXXFLAGS: -std=c++11 -O3 -Wno-error -Wno-cast-user-defined -I${SRCDIR}/csrc/include -I${SRCDIR}/csrc/include/tokenizer
#cgo LDFLAGS: -lstdc++
#include "csrc/include/coccoc_bridge.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"unsafe"
)

type cgoTokenizer struct {
	opts     Options
	dictPath string
}

var (
	cgoInstance *cgoTokenizer
	cgoMu       sync.RWMutex
)

func newCGOEngine(opts Options) (Tokenizer, error) {
	dictPath := opts.DictPath
	if dictPath == "" {
		dictPath = os.Getenv("COCCOC_DICT_PATH")
	}
	if dictPath == "" {
		// Tìm kiếm trong các thư mục mặc định thông dụng
		candidates := []string{
			"dicts",
			"../dicts",
			"../../dicts",
			"data/dicts/coccoc",
			"../data/dicts/coccoc",
			"../../data/dicts/coccoc",
			"/usr/local/share/tokenizer/dicts",
			"/etc/coccoc/dicts",
		}
		for _, cand := range candidates {
			if stat, err := os.Stat(cand); err == nil && stat.IsDir() {
				if _, err := os.Stat(filepath.Join(cand, "multiterm_trie.dump")); err == nil {
					dictPath = cand
					break
				}
			}
		}
	}

	if dictPath == "" {
		return nil, errors.New("không tìm thấy thư mục từ điển Cốc Cốc (multiterm_trie.dump). Hãy truyền WithDictPath(...) hoặc thiết lập COCCOC_DICT_PATH")
	}

	cgoMu.Lock()
	defer cgoMu.Unlock()

	if cgoInstance != nil {
		return cgoInstance, nil
	}

	cDictPath := C.CString(dictPath)
	defer C.free(unsafe.Pointer(cDictPath))

	nontoneFlag := C.int(0)
	if opts.LoadNontone {
		nontoneFlag = C.int(1)
	}

	res := C.coccoc_init(cDictPath, nontoneFlag)
	if res < 0 {
		return nil, fmt.Errorf("coccoc_init thất bại từ C++ engine, mã lỗi: %d", int(res))
	}

	cgoInstance = &cgoTokenizer{
		opts:     opts,
		dictPath: dictPath,
	}

	return cgoInstance, nil
}

// Segment phân tách câu và ghép các từ ghép bằng dấu gạch dưới "_"
func (t *cgoTokenizer) Segment(text string) (string, error) {
	if text == "" {
		return "", nil
	}

	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	cResult := C.coccoc_tokenize_original(cText)
	if cResult == nil {
		return "", errors.New("coccoc: lỗi phân tách từ từ C++ engine")
	}
	defer C.coccoc_free_string(cResult)

	return C.GoString(cResult), nil
}

// Tokenize phân đoạn câu và trả về danh sách các tokens đã chuẩn hóa
func (t *cgoTokenizer) Tokenize(text string) ([]string, error) {
	segmented, err := t.Segment(text)
	if err != nil {
		return nil, err
	}
	return extractTokens(segmented, t.opts.NormalizeLowercase, t.opts.KeepPunctuation), nil
}

// TokenizeFull trả về danh sách token chi tiết kèm offset và loại từ
func (t *cgoTokenizer) TokenizeFull(text string) ([]TokenInfo, error) {
	if text == "" {
		return []TokenInfo{}, nil
	}

	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	cResult := C.coccoc_tokenize_full(cText)
	if cResult == nil {
		return nil, errors.New("coccoc: lỗi khi phân tích chi tiết từ C++ engine")
	}
	defer C.coccoc_free_token_result(cResult)

	count := int(cResult.count)
	if count == 0 || cResult.tokens == nil {
		return []TokenInfo{}, nil
	}

	tokensSlice := unsafe.Slice(cResult.tokens, count)
	infos := make([]TokenInfo, count)

	for i := 0; i < count; i++ {
		cTok := tokensSlice[i]
		txt := ""
		if cTok.text != nil {
			txt = C.GoString(cTok.text)
		}
		infos[i] = TokenInfo{
			Text:      txt,
			StartByte: int(cTok.start_byte),
			EndByte:   int(cTok.end_byte),
			Type:      TokenType(cTok.token_type),
			SegType:   int(cTok.seg_type),
		}
	}

	return infos, nil
}

// IsCGO trả về true vì đây là CGO engine
func (t *cgoTokenizer) IsCGO() bool {
	return true
}

// Close giải phóng tài nguyên native C++
func (t *cgoTokenizer) Close() error {
	cgoMu.Lock()
	defer cgoMu.Unlock()
	C.coccoc_destroy()
	cgoInstance = nil
	return nil
}
