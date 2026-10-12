# Cốc Cốc Tokenizer for Go (`coccoc-tokenizer-go`)

[![Go Reference](https://pkg.go.dev/badge/github.com/vntk/coccoc-tokenizer-go.svg)](https://pkg.go.dev/github.com/vntk/coccoc-tokenizer-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/vntk/coccoc-tokenizer-go)](https://goreportcard.com/report/github.com/vntk/coccoc-tokenizer-go)
[![License: LGPL v2.1](https://img.shields.io/badge/License-LGPL_v2.1-blue.svg)](LICENSE)
[![CI Tests](https://github.com/vntk/coccoc-tokenizer-go/actions/workflows/test.yml/badge.svg)](https://github.com/vntk/coccoc-tokenizer-go/actions/workflows/test.yml)

Thư viện tách từ tiếng Việt chất lượng cao (Production-Grade Vietnamese Word Segmenter) dành cho ngôn ngữ Go, đóng gói độc lập động cơ C++ **Cốc Cốc Tokenizer** kết hợp kiến trúc **Dual-Mode (CGO Engine & Pure-Go Fallback)**.

---

## 🌟 Tính Năng Nổi Bật

* 🚀 **Hiệu Năng Vượt Trội**: Xử lý hàng triệu tokens/giây khi kích hoạt C++ CGO Engine trong môi trường Linux/Docker.
* 🛡️ **Kiến Trúc Dual-Mode Thông Minh**:
  * **CGO Enabled**: Tự động liên kết C++ engine với cấu trúc Trie nén nhị phân (`.dump`) siêu tốc.
  * **Pure-Go Fallback (`CGO_ENABLED=0`)**: Tự động chuyển đổi sang thuật toán Forward Maximum Matching (FMM) trên từ điển nhúng sẵn (`//go:embed`) mà **không cần GCC hay bất kỳ dependency bên ngoài nào**. Chạy mượt mà trên Windows, macOS, Alpine Linux.
* 🔒 **Zero Memory Leak**: Cơ chế giải phóng con trỏ C/C++ tức thời (`defer free`), vượt qua bài kiểm tra stress-test 1.000.000 requests với mức RAM hoàn toàn phẳng.
* 🧵 **Thread-Safe & Lockless**: Hỗ trợ gọi đồng thời từ hàng nghìn Goroutines mà không bị nghẽn khóa (zero contention).
* 🎯 **API Chuẩn Idiomatic Go**: Thiết kế theo Functional Options Pattern, cung cấp cả `Segment`, `Tokenize`, và `TokenizeFull` (kèm byte offsets và loại từ).
* 📦 **Quản Lý Từ Điển Tự Động**: Tích hợp package con `dictutil` tự động tìm kiếm hoặc tải dữ liệu từ điển từ GitHub Release.

---

## 📦 Cài Đặt

```bash
go get -u github.com/vntk/coccoc-tokenizer-go
```

---

## ⚡ Bắt Đầu Nhanh (Quickstart)

Chỉ 3 dòng code để tách từ tiếng Việt:

```go
package main

import (
	"fmt"
	coccoc "github.com/vntk/coccoc-tokenizer-go"
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

* **Zero Memory Leak Guarantee**: Chạy stress-test 50.000 liên tục trong `leak_test.go` khẳng định mức tăng heap sau GC $\le 32\text{ KB}$, không rò rỉ bộ nhớ native C++.

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

## 📁 Cấu Trúc Repository

```
coccoc-tokenizer-go/
├── .github/
│   └── workflows/
│       ├── test.yml            # CI: Chạy test trên Linux (CGO + Non-CGO), macOS, Windows
│       ├── benchmark.yml       # CI: Đo lường regression về hiệu năng và allocations
│       └── release.yml         # CI: Tự động đóng gói release assets
├── csrc/                       # Toàn bộ mã nguồn C/C++ phục vụ biên dịch CGO
│   ├── include/
│   │   ├── coccoc_bridge.h     # Interface extern "C" chuẩn, không mangling
│   │   └── tokenizer/          # Cốc Cốc C++ headers gốc & DATrie templates
│   ├── src/
│   │   └── coccoc_bridge.cpp   # Implementation của C bridge
│   └── CMakeLists.txt          # Build dynamic/static lib rời khi cần
├── dicts/                      # Dữ liệu từ điển nhị phân đã biên dịch sẵn
│   ├── multiterm_trie.dump     # Từ điển âm tiết ghép (~24.5 MB)
│   ├── syllable_trie.dump      # Từ điển âm tiết đơn (~21.4 MB)
│   └── words.txt               # Danh sách 40.000 từ ghép tiếng Việt cho Pure-Go
├── dictutil/                   # Tiện ích tìm kiếm & tải từ điển tự động
│   ├── downloader.go
│   └── downloader_test.go
├── examples/                   # Ví dụ mẫu hướng dẫn tích hợp
│   ├── basic/main.go           # Ví dụ cơ bản: Tách từ một câu đơn
│   ├── http_server/main.go     # Ví dụ REST API Microservice
│   └── opensearch_indexer/     # Ví dụ tiền xử lý nạp OpenSearch / Elasticsearch
├── bridge.cpp                  # CGO compilation unit
├── tokenizer.go                # Public API & Core struct
├── tokenizer_cgo.go            # CGO implementation (//go:build cgo)
├── tokenizer_cgo_stub.go       # Non-CGO stub (//go:build !cgo)
├── tokenizer_pure.go           # Pure Go implementation (//go:embed words.txt)
├── options.go                  # Functional Options Pattern
├── types.go                    # Định nghĩa Token, Segment, Tokenizer Interface
├── tokenizer_test.go           # Unit tests kiểm tra độ chính xác và đa luồng
├── benchmark_test.go           # Benchmark đo ns/op, B/op, allocs/op
├── leak_test.go                # Stress-test kiểm tra rò rỉ bộ nhớ
├── go.mod
├── LICENSE                     # LGPL-2.1 License
└── README.md
```

---

## 📜 Giấy Phép Bản Quyền (License)

Dự án phát hành theo giấy phép **LGPL-2.1 (GNU Lesser General Public License v2.1)** kế thừa từ mã nguồn Cốc Cốc Tokenizer gốc. Bạn có thể sử dụng tự do trong các dự án nguồn đóng hoặc thương mại theo các điều khoản của LGPL-2.1.
