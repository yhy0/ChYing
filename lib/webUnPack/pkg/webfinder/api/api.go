package api

import (
    "fmt"
    "strings"
)

type SiteMap struct {
    APIMap   []API
    RouteMap []RouteMap
    // 引用的外部链接或者api
    LinkAPI []API
    Links   []string
}

type Param struct {
    Key   string
    Value string
}

type API struct {
    Type    string
    Method  string
    Uri     string
    UriType byte // 1->is string, 0-> 非string,可能是误报
    Params  []Param
}

func (api *API) Valid() bool {
    return !(api.Method == "" || api.Uri == "")
}

type RouteMap struct {
    Path     string
    Children []RouteMap
}

func (r *RouteMap) Valid() bool {
    return !strings.EqualFold(r.Path, "") || len(r.Children) > 0
}

func (r *RouteMap) Out(s int) {
    if len(r.Path) <= 0 {
        return
    }
    for i := 0; i < s; i++ {
        fmt.Printf("--")
    }
    r.Path = strings.ReplaceAll(r.Path, "'", "")
    r.Path = strings.ReplaceAll(r.Path, "\"", "")
    fmt.Printf("%s\n", r.Path)
    for _, child := range r.Children {
        child.Out(s + 1)
    }
}

func FilterApis(apis []API) (results []API) {
    for _, api := range apis {
        if filterUri(api.Uri) {
            results = append(results, api)
        }
    }
    return
}

func filterUri(uri string) bool {
    if strings.HasPrefix(uri, "#") {
        return false
    }
    blacklist := []string{".png", ".ico", ".gif", ".svg", ".jpg", ".css", ".js", ".eot", ".woff", ".ttf", ".mp4", ".mp3"}
    for _, item := range blacklist {
        if strings.HasSuffix(uri, item) {
            return false
        }
    }
    return true
}
