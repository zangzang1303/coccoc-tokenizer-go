package coccoc_test

import (
	"runtime"
	"testing"

	coccoc "github.com/zangzang1303/coccoc-tokenizer-go"
)

func TestMemoryLeakAudit(t *testing.T) {
	if testing.Short() {
		t.Skip("Bỏ qua leak audit trong chế độ -short")
	}

	tok, err := coccoc.New(coccoc.WithFallbackToPureGo(true))
	if err != nil {
		t.Fatalf("Khởi tạo thất bại: %v", err)
	}
	defer tok.Close()

	// 1. Chạy warm-up để nạp cấu trúc dữ liệu và ổn định GC
	warmupText := "Học sinh sinh viên Việt Nam đi uống cà phê"
	for i := 0; i < 5000; i++ {
		_, _ = tok.Segment(warmupText)
	}

	runtime.GC()
	var mBefore runtime.MemStats
	runtime.ReadMemStats(&mBefore)

	// 2. Chạy stress test 50.000 lượt
	const iterations = 50000
	testText := "Thành phố Hà Nội và Thành phố Hồ Chí Minh là hai trung tâm kinh tế lớn của Việt Nam với nhiều trường đại học hàng đầu đào tạo học sinh sinh viên chất lượng cao."

	for i := 0; i < iterations; i++ {
		_, err := tok.Segment(testText)
		if err != nil {
			t.Fatalf("Lỗi tại iteration %d: %v", i, err)
		}
		if i%10000 == 0 {
			_, _ = tok.TokenizeFull(testText)
		}
	}

	// 3. Thu gom rác và kiểm tra mức tăng bộ nhớ
	runtime.GC()
	var mAfter runtime.MemStats
	runtime.ReadMemStats(&mAfter)

	diffAlloc := int64(mAfter.HeapAlloc) - int64(mBefore.HeapAlloc)
	t.Logf("Chạy %d lần hoàn tất. HeapAlloc trước: %d KB, sau: %d KB, chênh lệch: %d KB",
		iterations, mBefore.HeapAlloc/1024, mAfter.HeapAlloc/1024, diffAlloc/1024)

	// Mức chênh lệch sau GC phải rất nhỏ (< 5MB) khẳng định không có memory leak
	if diffAlloc > 5*1024*1024 {
		t.Errorf("Nghi ngờ rò rỉ bộ nhớ: Heap tăng hơn 5MB (%d bytes)", diffAlloc)
	}
}
