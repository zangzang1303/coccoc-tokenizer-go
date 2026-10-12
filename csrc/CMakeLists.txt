cmake_minimum_required(VERSION 3.10)
project(coccoc_tokenizer_native CXX)

set(CMAKE_CXX_STANDARD 11)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

include_directories(include)
include_directories(include/tokenizer)

set(SOURCES
    src/coccoc_bridge.cpp
)

# Static library
add_library(coccoc_tokenizer_static STATIC ${SOURCES})
set_target_properties(coccoc_tokenizer_static PROPERTIES OUTPUT_NAME "coccoc_tokenizer")

# Shared library
add_library(coccoc_tokenizer SHARED ${SOURCES})

# Installation rules
install(TARGETS coccoc_tokenizer coccoc_tokenizer_static
    ARCHIVE DESTINATION lib
    LIBRARY DESTINATION lib
    RUNTIME DESTINATION bin
)

install(FILES include/coccoc_bridge.h DESTINATION include)
install(DIRECTORY include/tokenizer DESTINATION include)
