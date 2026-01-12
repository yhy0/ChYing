package utils

import (
    "bufio"
    regexp "github.com/wasilibs/go-re2"
    "github.com/yhy0/logging"
    "strconv"
    "strings"
    "unicode"
)

/**
  @author: yhy
  @since: 2023/5/17
  @desc: //TODO
**/

func Search(pattern string, str string) (error, []string) {
    re, err := regexp.Compile(pattern)
    if err != nil {
        return err, nil
    }
    return nil, re.FindAllString(str, -1)
}

func RegexpStr(patterns []string, str string) bool {
    for _, pattern := range patterns {
        match, err := regexp.MatchString(pattern, str)
        if err != nil {
            continue
        }
        if match {
            return true
        }
    }
    
    return false
}

func RemoveEmptyAndNewlineStrings(arr []string) []string {
    // 定义一个切片用于存储无空白字符和换行符的字符串
    result := make([]string, 0, len(arr))
    
    // 遍历数组并过滤空白字符和换行符
    for _, str := range arr {
        trimmed := strings.TrimSpace(str)
        withoutNewlines := strings.Map(func(r rune) rune {
            if unicode.IsSpace(r) && r != '\n' && r != '\r' {
                return -1
            }
            return r
        }, trimmed)
        
        if len(withoutNewlines) > 0 {
            result = append(result, withoutNewlines)
        }
    }
    
    return result
}

func SplitStringByLines(str string) []string {
    // 创建一个新的 Scanner 对象来读取字符串数据
    scanner := bufio.NewScanner(strings.NewReader(str))
    // 定义一个切片用于存储行数据
    lines := make([]string, 0)
    
    // 使用 Scanner 对象按行读取字符串数据
    for scanner.Scan() {
        // 去除每行的前后空白字符
        line := strings.TrimSpace(scanner.Text())
        if line == "" {
            continue
        }
        // 将处理后的行数据添加到结果切片中
        lines = append(lines, line)
    }
    
    if err := scanner.Err(); err != nil {
        logging.Logger.Errorln(err)
        return nil
    }
    
    return lines
}

// IsInteger 函数检查字符串是否为整数
func IsInteger(s string) bool {
    if s == "" {
        return false // 空字符串不是整数
    }
    _, err := strconv.Atoi(s)
    return err == nil
}
