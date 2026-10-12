package dictutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vntk/coccoc-tokenizer-go/dictutil"
)

func TestDefaultCacheDir(t *testing.T) {
	dir := dictutil.DefaultCacheDir()
	if dir == "" {
		t.Fatal("DefaultCacheDir không được rỗng")
	}
	if !filepath.IsAbs(dir) && dir != filepath.Join(".", "dicts") {
		t.Logf("Cache dir: %s", dir)
	}
}

func TestVerifyDicts(t *testing.T) {
	// Kiểm tra với thư mục không tồn tại
	err := dictutil.VerifyDicts("non_existent_folder_xyz_123")
	if err == nil {
		t.Errorf("Kỳ vọng lỗi khi thư mục không tồn tại")
	}

	// Tạo thư mục tạm kiểm tra thiếu file
	tmpDir, err := os.MkdirTemp("", "test_dict_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	err = dictutil.VerifyDicts(tmpDir)
	if err == nil {
		t.Errorf("Kỳ vọng lỗi khi thiếu dump files")
	}

	// Tạo mock dump files
	_ = os.WriteFile(filepath.Join(tmpDir, dictutil.MultitermDumpFile), []byte("mock_multiterm"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, dictutil.SyllableDumpFile), []byte("mock_syllable"), 0644)

	err = dictutil.VerifyDicts(tmpDir)
	if err != nil {
		t.Errorf("VerifyDicts thất bại với mock dumps: %v", err)
	}
}

func TestFindDictDir(t *testing.T) {
	found := dictutil.FindDictDir()
	t.Logf("FindDictDir tìm thấy: %s", found)
}
