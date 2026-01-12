package javascript

import (
    "github.com/pkg/errors"
    "github.com/tdewolff/parse/v2"
    "github.com/tdewolff/parse/v2/js"
)

func UriFromJavascript(source string) (wk walker, err error) {
    ast, err := js.Parse(parse.NewInputString(source), js.Options{})
    if err != nil {
        return wk, errors.Wrap(err, "parse js error")
    }
    if ast == nil {
        return wk, errors.Wrap(err, "nil ast")
    }
    defer func() {
        // 防止出现奇怪的错误
        recover()
    }()
    js.Walk(&wk, ast)
    return
}
