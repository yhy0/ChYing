package file

import regexp "github.com/wasilibs/go-re2"

/**
   @author yhy
   @since 2024/12/16
   @desc //TODO
**/

type BBscanRule struct {
    Tag          string   // 文本内容
    Status       string   // 状态码
    Type         string   // 返回的 ContentType
    TypeNo       string   // 不可能返回的 ContentType
    FingerPrints []string // 指纹，只有匹配到该指纹，才会进行目录扫描
    Root         bool     // 是否为一级目录
}

type MitmRule struct {
    Rule              string `json:"rule"`
    RegexCompiled     *regexp.Regexp
    NoReplace         bool     `json:"noReplace"`
    Color             string   `json:"color"`
    EnableForRequest  bool     `json:"enableForRequest"`
    EnableForResponse bool     `json:"enableForResponse"`
    EnableForHeader   bool     `json:"enableForHeader"`
    EnableForBody     bool     `json:"enableForBody"`
    Index             int      `json:"index"`
    ExtraTag          []string `json:"extraTag"`
    VerboseName       string   `json:"verboseName"`
}
