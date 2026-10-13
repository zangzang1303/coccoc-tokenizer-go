#include "coccoc_bridge.h"
#include <tokenizer/tokenizer.hpp>
#include <cstdlib>
#include <cstring>
#include <string>
#include <vector>
#include <mutex>

static bool g_initialized = false;
static std::mutex g_init_mutex;

int coccoc_init(const char* dict_path, int load_nontone) {
    if (!dict_path) return -1;
    std::lock_guard<std::mutex> lock(g_init_mutex);
    bool nontone = (load_nontone != 0);
    int res = Tokenizer::instance().initialize(std::string(dict_path), nontone);
    if (res >= 0) {
        g_initialized = true;
    }
    return res;
}

int coccoc_is_initialized(void) {
    return g_initialized ? 1 : 0;
}

char* coccoc_tokenize_original(const char* text) {
    if (!text || text[0] == '\0') {
        char* empty = (char*)malloc(1);
        if (empty) empty[0] = '\0';
        return empty;
    }

    std::string input(text);
    std::vector<FullToken> res = Tokenizer::instance().segment_original(input, Tokenizer::TOKENIZE_NORMAL);

    if (res.empty()) {
        char* result = (char*)malloc(input.size() + 1);
        if (!result) return NULL;
        std::memcpy(result, input.c_str(), input.size() + 1);
        return result;
    }

    std::string output;
    output.reserve(input.size() + 16);

    for (size_t i = 0; i < res.size(); ++i) {
        size_t punct_start = (i > 0) ? res[i - 1].original_end : 0;
        size_t punct_len = (res[i].original_start > punct_start) ? (res[i].original_start - punct_start) : 0;

        if (punct_len > 0) {
            output += input.substr(punct_start, punct_len);
        } else if (i > 0) {
            output += ' ';
        }

        output += res[i].text;
    }

    if (!res.empty()) {
        size_t last_end = res.back().original_end;
        if (last_end < input.size()) {
            output += input.substr(last_end);
        }
    }

    char* result = (char*)malloc(output.size() + 1);
    if (!result) return NULL;
    std::memcpy(result, output.c_str(), output.size() + 1);
    return result;
}

void coccoc_free_string(char* str) {
    if (str) {
        free(str);
    }
}

CoccocTokenResult* coccoc_tokenize_full(const char* text) {
    CoccocTokenResult* result = (CoccocTokenResult*)malloc(sizeof(CoccocTokenResult));
    if (!result) return NULL;
    result->tokens = NULL;
    result->count = 0;

    if (!text || text[0] == '\0') {
        return result;
    }

    std::string input(text);
    std::vector<FullToken> res = Tokenizer::instance().segment_original(input, Tokenizer::TOKENIZE_NORMAL);
    if (res.empty()) {
        return result;
    }

    result->count = (int)res.size();
    result->tokens = (CoccocToken*)calloc(result->count, sizeof(CoccocToken));
    if (!result->tokens) {
        free(result);
        return NULL;
    }

    for (size_t i = 0; i < res.size(); ++i) {
        result->tokens[i].start_byte = res[i].original_start;
        result->tokens[i].end_byte = res[i].original_end;
        result->tokens[i].token_type = res[i].type;
        result->tokens[i].seg_type = res[i].seg_type;

        size_t len = res[i].text.size();
        result->tokens[i].text = (char*)malloc(len + 1);
        if (result->tokens[i].text) {
            std::memcpy(result->tokens[i].text, res[i].text.c_str(), len + 1);
        }
    }

    return result;
}

void coccoc_free_token_result(CoccocTokenResult* result) {
    if (!result) return;
    if (result->tokens) {
        for (int i = 0; i < result->count; ++i) {
            if (result->tokens[i].text) {
                free(result->tokens[i].text);
            }
        }
        free(result->tokens);
    }
    free(result);
}

void coccoc_destroy(void) {
    std::lock_guard<std::mutex> lock(g_init_mutex);
    g_initialized = false;
}
