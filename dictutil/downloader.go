package dictutil

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	MultitermDumpFile = "multiterm_trie.dump"
	SyllableDumpFile  = "syllable_trie.dump"
	NontoneDumpFile   = "nontone_pair_freq_map.dump"

	DefaultReleaseURL = "https://github.com/vntk/coccoc-tokenizer-go/releases/download/v1.0.0/coccoc_dicts.tar.gz"
)

// DefaultCacheDir trả về thư mục cache từ điển mặc định theo hệ điều hành
func DefaultCacheDir() string {
	if runtime.GOOS == "windows" {
		localApp := os.Getenv("LOCALAPPDATA")
		if localApp != "" {
			return filepath.Join(localApp, "coccoc", "dicts")
		}
	}
	home, err := os.UserHomeDir()
	if err == nil {
		return filepath.Join(home, ".cache", "coccoc", "dicts")
	}
	return filepath.Join(".", "dicts")
}

// FindDictDir tìm kiếm thư mục chứa từ điển hợp lệ từ các vị trí tiêu chuẩn
func FindDictDir() string {
	// 1. Kiểm tra biến môi trường COCCOC_DICT_PATH
	if envPath := os.Getenv("COCCOC_DICT_PATH"); envPath != "" {
		if VerifyDicts(envPath) == nil {
			return envPath
		}
	}

	// 2. Kiểm tra thư mục cache
	cacheDir := DefaultCacheDir()
	if VerifyDicts(cacheDir) == nil {
		return cacheDir
	}

	// 3. Kiểm tra các đường dẫn tương đối và hệ thống
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
		if VerifyDicts(cand) == nil {
			return cand
		}
	}

	return ""
}

// VerifyDicts kiểm tra sự tồn tại và tính hợp lệ của các tệp từ điển cốt lõi
func VerifyDicts(dir string) error {
	if dir == "" {
		return fmt.Errorf("đường dẫn thư mục từ điển rỗng")
	}
	stat, err := os.Stat(dir)
	if err != nil || !stat.IsDir() {
		return fmt.Errorf("thư mục từ điển không tồn tại: %s", dir)
	}

	coreFiles := []string{MultitermDumpFile, SyllableDumpFile}
	for _, f := range coreFiles {
		fullPath := filepath.Join(dir, f)
		fStat, err := os.Stat(fullPath)
		if err != nil {
			return fmt.Errorf("thiếu tệp từ điển: %s", fullPath)
		}
		if fStat.Size() == 0 {
			return fmt.Errorf("tệp từ điển có dung lượng 0 byte: %s", fullPath)
		}
	}

	return nil
}

// EnsureDicts đảm bảo các tệp từ điển có mặt tại targetDir, tải tự động nếu chưa có
func EnsureDicts(targetDir string, downloadURL string) error {
	if VerifyDicts(targetDir) == nil {
		return nil
	}

	if downloadURL == "" {
		downloadURL = DefaultReleaseURL
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("không thể tạo thư mục đích: %w", err)
	}

	fmt.Printf("[dictutil] Đang tải từ điển Cốc Cốc từ %s ...\n", downloadURL)
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("lỗi khi kết nối tải từ điển: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tải từ điển thất bại, HTTP status: %d", resp.StatusCode)
	}

	if strings.HasSuffix(downloadURL, ".zip") {
		return extractZip(resp.Body, targetDir)
	}
	return extractTarGz(resp.Body, targetDir)
}

func extractTarGz(r io.Reader, targetDir string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(targetDir, filepath.Base(header.Name))
		if header.Typeflag == tar.TypeReg {
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}

func extractZip(r io.Reader, targetDir string) error {
	// Ghi tạm ra file rồi giải nén
	tmpFile, err := os.CreateTemp("", "coccoc_dict_*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, r); err != nil {
		return err
	}

	zr, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		return err
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		target := filepath.Join(targetDir, filepath.Base(f.Name))
		rc, err := f.Open()
		if err != nil {
			return err
		}
		outFile, err := os.Create(target)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
