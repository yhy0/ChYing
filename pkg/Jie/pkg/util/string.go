package util

import (
    "encoding/json"
    "math/rand"
    "net/http"
    "strings"
    "time"
    "unicode/utf8"
    "unsafe"
)

/**
  @author: yhy
  @since: 2023/3/8
  @desc: //TODO
**/

func init() {
    rand.Seed(time.Now().Unix())
}

func StringToBytes(s string) []byte {
    return *(*[]byte)(unsafe.Pointer(
        &struct {
            string
            Cap int
        }{s, len(s)},
    ))
}

/*

第一个函数 BytesToString 使用了 unsafe.Pointer 将 []byte 类型的底层表示强制转换为 string 类型的底层表示，因此在底层内存布局上，它们是一样的。
这种方式可以避免数据的拷贝，因此在性能上可能更高效。但是，由于使用了 unsafe 包，这种方式可能会存在安全风险，并且可能会导致代码不可移植。

第二个函数 string(b) 则是通过将 []byte 类型的数据复制到新的内存地址中来创建一个新的 string 类型的对象。这种方式更加安全，并且在代码可移植性上更好。但是，由于涉及到数据的拷贝，因此在性能上可能不如第一个函数。

因此，如果对性能要求比较高，而且能够保证代码在特定平台上的可移植性，可以使用第一个函数。否则，应该使用第二个函数。

防止 内存覆盖, https://github.com/yhy0/Jie/issues/2 使用 string(b)
*/

func BytesToString(b []byte) string {
    if !utf8.Valid(b) {
        return ""
    }
    return string(b[:])
    // return *(*string)(unsafe.Pointer(&b))
}

// StructToJsonString 将结构体输出位 json 字符
func StructToJsonString(item interface{}) string {
    jsonBytes, err := json.MarshalIndent(item, "", "  ")
    if err != nil {
        panic(err)
    }

    return string(jsonBytes)
}

// ReverseString 反转字符串 style --> elyts
func ReverseString(s string) string {
    runes := []rune(s)
    for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
        runes[from], runes[to] = runes[to], runes[from]
    }
    return string(runes)
}

func Trim(s string) string {
    s = strings.ReplaceAll(s, " ", "")
    s = strings.ReplaceAll(s, "\t", "")
    s = strings.ReplaceAll(s, "\r", "")
    s = strings.ReplaceAll(s, "\n", "")
    return s
}

// RemoveDuplicateElement  数组去重
func RemoveDuplicateElement(strs []string) []string {
    var result []string
    if len(strs) == 0 {
        return result
    }

    for _, item := range strs {
        item = strings.TrimSpace(item)
        if item != "" && !SliceInCaseFold(item, result) {
            item = strings.TrimSpace(item)
            result = append(result, item)
        }
    }
    return result
}

func MapToJson(param http.Header) string {
    dataType, _ := json.Marshal(param)
    dataString := string(dataType)
    return dataString
}

func RemoveQuotationMarks(str string) string {
    str = strings.ReplaceAll(str, "\"", "")
    str = strings.ReplaceAll(str, "'", "")
    return str
}
