# Cốc Cốc Tokenizer for Go (`coccoc-tokenizer-go`)

[![Go Reference](https://pkg.go.dev/badge/github.com/zangzang1303/coccoc-tokenizer-go.svg)](https://pkg.go.dev/github.com/zangzang1303/coccoc-tokenizer-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/zangzang1303/coccoc-tokenizer-go)](https://goreportcard.com/report/github.com/zangzang1303/coccoc-tokenizer-go)
[![License: LGPL v2.1](https://img.shields.io/badge/License-LGPL_v2.1-blue.svg)](LICENSE)
[![CI Tests](https://github.com/zangzang1303/coccoc-tokenizer-go/actions/workflows/test.yml/badge.svg)](https://github.com/zangzang1303/coccoc-tokenizer-go/actions/workflows/test.yml)

Thư viện tách từ tiếng Việt chất lượng cao (Production-Grade Vietnamese Word Segmenter) dành cho ngôn ngữ Go, đóng gói độc lập động cơ C++ **Cốc Cốc Tokenizer** kết hợp kiến trúc **Dual-Mode (CGO Engine & Pure-Go Fallback)**.

---

## 🌟 Tính Năng Nổi Bật

* 🚀 **Hiệu Năng Vượt Trội**: Xử lý hàng triệu tokens/giây khi kích hoạt C++ CGO Engine trong môi trường Linux/Docker.
* 🛡️ **Kiến Trúc Dual-Mode Thông Minh**:
  * **CGO Enabled**: Tự động liên kết C++ engine với cấu trúc Trie nén nhị phân (`.dump`) siêu tốc.
  * **Pure-Go Fallback (`CGO_ENABLED=0`)**: Tự động chuyển đổi sang thuật toán Forward Maximum Matching (FMM) trên từ điển nhúng sẵn (`//go:embed`) mà **không cần GCC hay bất kỳ dependency bên ngoài nào**. Chạy mượt mà trên Windows, macOS, Alpine Linux.
* 🔒 **Zero Memory Leak**: Cơ chế giải phóng con trỏ C/C++ tức thời (`defer free`), vượt qua bài kiểm tra stress-test 50.000 requests với mức RAM hoàn toàn phẳng.
* 🧵 **Thread-Safe & Lockless**: Hỗ trợ gọi đồng thời từ hàng nghìn Goroutines mà không bị nghẽn khóa (zero contention).
* 🎯 **API Chuẩn Idiomatic Go**: Thiết kế theo Functional Options Pattern, cung cấp cả `Segment`, `Tokenize`, và `TokenizeFull` (kèm byte offsets và loại từ).
* 📦 **Quản Lý Từ Điển Tự Động**: Tích hợp package con `dictutil` tự động tìm kiếm hoặc tải dữ liệu từ điển từ GitHub Release.

---

## 📦 HƯỚNG DẪN CÀI ĐẶT CHI TIẾT (INSTALLATION GUIDE)

Thư viện hỗ trợ **hai chế độ cài đặt và vận hành** tùy theo nhu cầu môi trường của bạn:

| Chế độ | Môi trường mục tiêu | Yêu cầu Compiler | Tốc độ xử lý | Từ điển cần nạp |
| :--- | :--- | :---: | :---: | :--- |
| **Pure-Go Mode** | Local Dev, Windows, macOS, Alpine Scratch | **Không cần C++ (Chỉ cần Go)** | ~22.000 câu/s | Đã nhúng sẵn (`words.txt` 500KB) |
| **CGO Native Mode** | Linux Production Server, Docker Kubernetes | Cần `gcc`/`g++` $\ge 4.8$ | ~830.000 câu/s | Tệp nhị phân `.dump` (~45MB) |

---

### Bước 1: Tải module vào dự án Go

Mở terminal trong thư mục dự án Go của bạn và chạy lệnh:

```bash
go get -u github.com/zangzang1303/coccoc-tokenizer-go
```

File `go.mod` của bạn sẽ tự động thêm dependency:
```go
require github.com/zangzang1303/coccoc-tokenizer-go v1.0.0
```

---

### Bước 2: Lựa chọn chế độ cài đặt môi trường

#### 👉 Lựa chọn A: Cài đặt chế độ Pure-Go (Khuyên dùng cho Local Dev / Windows / Mac)
* **Ưu điểm**: Không cần cài GCC, không cần cài CMake, không cần tải bất kỳ file ngoài nào.
* **Cách sử dụng**:
  Chỉ cần gõ lệnh chạy như thông thường, Go sẽ tự động dùng Pure-Go engine:
  ```bash
  go run main.go
  ```
  Hoặc ép kiểu build tĩnh không phụ thuộc CGO:
  ```bash
  # Linux / macOS
  CGO_ENABLED=0 go build -o myapp main.go

  # Windows PowerShell
  $env:CGO_ENABLED=0; go build -o myapp.exe main.go
  ```

---

#### 👉 Lựa chọn B: Cài đặt chế độ C++ CGO Native (Khuyên dùng cho Linux Production Server)
Để kích hoạt tốc độ tối đa của Cốc Cốc C++ DATrie engine (~1.2 µs/câu), bạn cần chuẩn bị trình biên dịch C++ trên hệ điều hành:

##### 1. Cài đặt C++ Build Tools:
* **Ubuntu / Debian**:
  ```bash
  sudo apt-get update
  sudo apt-get install -y build-essential
  ```
* **CentOS / RHEL / Amazon Linux**:
  ```bash
  sudo yum groupinstall -y "Development Tools"
  ```
* **Alpine Linux**:
  ```bash
  apk add --no-cache build-base g++ libstdc++
  ```
* **macOS**:
  ```bash
  xcode-select --install
  ```
* **Windows (nếu muốn CGO)**:
  Cài đặt [MinGW-w64](https://www.mingw-w64.org/) hoặc [MSYS2](https://www.msys2.org/) và thêm thư mục `bin` chứa `gcc.exe`, `g++.exe` vào biến môi trường `PATH`.

##### 2. Cài đặt dữ liệu từ điển nhị phân C++ (`.dump` files):
Động cơ C++ CGO cần 2 tệp nhị phân nén:
* `multiterm_trie.dump` (~24.5 MB)
* `syllable_trie.dump` (~21.4 MB)

Bạn có thể chuẩn bị từ điển bằng một trong các cách sau:

* **Cách 1: Sử dụng thư mục `dicts/` có sẵn trong package**:
  Nếu bạn clone repo hoặc copy thư mục `dicts/` vào ứng dụng, chỉ cần chỉ định:
  ```go
  tok, err := coccoc.New(coccoc.WithDictPath("./dicts"))
  ```

* **Cách 2: Tự động tải qua code (`dictutil`)**:
  Sử dụng sub-package `dictutil` để tự động tải về thư mục cache hệ thống:
  ```go
  package main

  import (
      coccoc "github.com/zangzang1303/coccoc-tokenizer-go"
      "github.com/zangzang1303/coccoc-tokenizer-go/dictutil"
  )

  func main() {
      // Tự động tải từ điển về ~/.cache/coccoc/dicts nếu chưa có
      cacheDir := dictutil.DefaultCacheDir()
      _ = dictutil.EnsureDicts(cacheDir, "")

      tok, err := coccoc.New(coccoc.WithDictPath(cacheDir))
      // ...
  }
  ```

* **Cách 3: Cài đặt vào thư mục hệ thống chuẩn Linux (`/usr/local/share/tokenizer/dicts`)**:
  ```bash
  # Tạo thư mục từ điển hệ thống
  sudo mkdir -p /usr/local/share/tokenizer/dicts

  # Tải bộ từ điển từ GitHub Release
  curl -L -O https://github.com/zangzang1303/coccoc-tokenizer-go/releases/download/v1.0.0/coccoc_dicts.tar.gz

  # Giải nén vào thư mục chuẩn
  sudo tar -xzvf coccoc_dicts.tar.gz -C /usr/local/share/tokenizer/dicts/
  ```

* **Cách 4: Thiết lập biến môi trường `COCCOC_DICT_PATH`**:
  ```bash
  # Linux / macOS
  export COCCOC_DICT_PATH="/usr/local/share/tokenizer/dicts"

  # Windows PowerShell
  $env:COCCOC_DICT_PATH="C:\coccoc\dicts"
  ```
  *(Khi có biến môi trường này, hàm `coccoc.New()` sẽ tự động tìm thấy mà không cần truyền tham số `WithDictPath`).*

---

### Bước 3: Xác thực cài đặt thành công (Verify Installation)

Tạo file `test_install.go`:

```go
package main

import (
	"fmt"
	coccoc "github.com/zangzang1303/coccoc-tokenizer-go"
)

func main() {
	tok, err := coccoc.New(coccoc.WithFallbackToPureGo(true))
	if err != nil {
		panic(err)
	}
	defer tok.Close()

	mode := "Pure-Go Engine (Zero-CGO)"
	if tok.IsCGO() {
		mode = "C++ Native CGO Engine"
	}
	fmt.Printf("✅ Cài đặt thành công! Đang kích hoạt: %s\n", mode)

	result, _ := tok.Segment("Học sinh sinh viên Việt Nam đi uống cà phê sữa đá")
	fmt.Printf("Kết quả tách từ: %s\n", result)
}
```

Chạy thử nghiệm:
```bash
go run test_install.go
```

Output kỳ vọng:
```
✅ Cài đặt thành công! Đang kích hoạt: Pure-Go Engine (Zero-CGO)
Kết quả tách từ: Học_sinh sinh_viên Việt_Nam đi uống cà_phê sữa_đá
```

---

## ⚡ Bắt Đầu Nhanh (Quickstart)

```go
package main

import (
	"fmt"
	coccoc "github.com/zangzang1303/coccoc-tokenizer-go"
)

func main() {
	tok, _ := coccoc.New()
	defer tok.Close()

	// 1. Tách từ ghép nối bằng dấu gạch dưới "_"
	seg, _ := tok.Segment("Học sinh đi uống cà phê sữa đá ở Hà Nội")
	fmt.Println(seg)
	// Output: "Học_sinh đi uống cà_phê sữa_đá ở Hà_Nội"

	// 2. Lấy danh sách tokens đã chuẩn hóa (lowercase, bỏ dấu câu)
	tokens, _ := tok.Tokenize("Học sinh đi uống cà phê sữa đá ở Hà Nội!")
	fmt.Println(tokens)
	// Output: ["học_sinh", "đi", "uống", "cà_phê", "sữa_đá", "ở", "hà_nội"]
}
```

---

## 🏗️ Kiến Trúc Hệ Thống (Dual-Mode Architecture)

```
                       +-----------------------------+
                       |    coccoc.New(opts...)      |
                       +--------------+--------------+
                                      |
                       +--------------v--------------+
                       |   Kiểm tra môi trường CGO    |
                       +-------+--------------+------+
                               |              |
                (CGO Khả dụng) |              | (Không có CGO hoặc
                               |              |  WithForcePureGo=true)
              +----------------v--+        +--v------------------+
              |   C++ CGO Engine  |        |    Pure-Go Engine   |
              | (coccoc_bridge.h) |        | (Forward Max Match) |
              +--------+----------+        +----------+----------+
                       |                              |
            +----------v----------+        +----------v----------+
            | Multiterm & Syllable|        | Embedded words.txt  |
            |   DATrie Dumps      |        |   (~40.000 từ ghép) |
            +---------------------+        +---------------------+
```

---

## ⚙️ Cấu Hình Nâng Cao (Functional Options Pattern)

```go
tok, err := coccoc.New(
    // Chỉ định đường dẫn thư mục từ điển C++ (.dump)
    coccoc.WithDictPath("/usr/local/share/tokenizer/dicts"),

    // Nạp bảng tần suất nontone 200MB (mặc định false: tiết kiệm RAM ~45MB)
    coccoc.WithLoadNontone(false),

    // Tự động chuyển tokens thành chữ thường trong hàm Tokenize (mặc định true)
    coccoc.WithNormalizeLowercase(true),

    // Giữ lại dấu câu khi gọi Tokenize (mặc định false)
    coccoc.WithKeepPunctuation(false),

    // Tự động fallback sang Pure-Go nếu môi trường không có CGO (mặc định true)
    coccoc.WithFallbackToPureGo(true),

    // Bổ sung các từ ghép tùy chỉnh cho ứng dụng
    coccoc.WithCustomWords("máy bay không người lái", "xe điện tự hành"),
)
defer tok.Close()
```

---

## 🔍 Phân Tích Chi Tiết Kèm Offset (`TokenizeFull`)

Phù hợp cho các bài toán highlight tìm kiếm, tính toán span vị trí từ hoặc trích xuất thông tin:

```go
infos, err := tok.TokenizeFull("Uống cà phê ngon")
for _, info := range infos {
    fmt.Printf("Token: %-10s | Spans: [%d, %d] | Type: %s\n",
        info.Text, info.StartByte, info.EndByte, info.Type)
}
// Token: Uống       | Spans: [0, 5]   | Type: WORD
// Token: cà_phê     | Spans: [6, 14]  | Type: WORD
// Token: ngon       | Spans: [15, 19] | Type: WORD
```

---

## 📊 Kết Quả Benchmark Hiệu Năng

Thực hiện kiểm thử trên CPU Intel Core i7-12700H (14 Cores / 20 Threads):

| Chế độ vận hành | Tác vụ | Độ trễ (Latency) | Throughput | Go Heap Alloc |
| :--- | :--- | :---: | :---: | :---: |
| **C++ CGO Engine** | Câu ngắn (10 từ) | **1.2 µs** | **830.000 câu/s** | 48 B/op (1 alloc) |
| **C++ CGO Engine** | Đoạn văn (50 từ) | **5.8 µs** | **172.000 đoạn/s** | 192 B/op (1 alloc) |
| **Pure-Go Engine** | Câu ngắn (10 từ) | **45.1 µs** | **22.000 câu/s** | 42 KB/op |
| **Pure-Go Engine** | Đoạn văn (50 từ) | **180.0 µs** | **5.500 đoạn/s** | 120 KB/op |

* **Zero Memory Leak Guarantee**: Chạy stress-test 50.000 lượt liên tục trong `leak_test.go` khẳng định mức tăng heap sau GC $\le 32\text{ KB}$, không rò rỉ bộ nhớ native C++.

---

## 🐳 Hướng Dẫn Docker Multi-stage Tối Ưu

### Dockerfile với C++ CGO Engine (Hiệu năng cực đại):

```dockerfile
# Stage 1: Build
FROM golang:1.23-bookworm AS builder
RUN apt-get update && apt-get install -y build-essential cmake

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=1
RUN go build -ldflags="-s -w" -o server ./examples/http_server

# Stage 2: Minimal Runtime
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y libstdc++6 && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/dicts /app/dicts

ENV COCCOC_DICT_PATH=/app/dicts
EXPOSE 8080
CMD ["./server"]
```

### Dockerfile Scratch thuần Go (`CGO_ENABLED=0` - Image siêu nhẹ ~15MB):

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o server ./examples/http_server

FROM scratch
COPY --from=builder /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
```

---

## ❓ Xử Lý Sự Cố Thường Gặp (Troubleshooting)

1. **Lỗi `gcc: not found` hoặc `exec: "gcc": executable file not found in %PATH%`**:
   * **Nguyên nhân**: Môi trường chưa cài GCC hoặc `CGO_ENABLED` đang bật trên Windows/Mac không có compiler.
   * **Khắc phục**: 
     - Chạy với `CGO_ENABLED=0 go run main.go` để dùng Pure-Go engine (không cần GCC).
     - Hoặc thêm `coccoc.WithFallbackToPureGo(true)` vào code.

2. **Lỗi `thư mục từ điển không tồn tại` trong CGO Mode**:
   * **Khắc phục**: Thiết lập biến môi trường `export COCCOC_DICT_PATH=/path/to/dicts` hoặc dùng `coccoc.WithDictPath("./dicts")`.

3. **Lỗi thiếu thư viện runtime trên Docker Alpine**:
   * **Khắc phục**: Thêm lệnh `apk add --no-cache libstdc++` vào runtime image.

---

## 📜 Giấy Phép Bản Quyền (License)

Dự án phát hành theo giấy phép **LGPL-2.1 (GNU Lesser General Public License v2.1)** kế thừa từ mã nguồn Cốc Cốc Tokenizer gốc. Bạn có thể sử dụng tự do trong các dự án nguồn đóng hoặc thương mại theo các điều khoản của LGPL-2.1.
