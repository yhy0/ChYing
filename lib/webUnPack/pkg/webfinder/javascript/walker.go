package javascript

import (
    "github.com/pkg/errors"
    "github.com/tdewolff/parse/v2/js"
    API "github.com/yhy0/ChYing/lib/webUnPack/pkg/webfinder/api"
    "golang.org/x/exp/slices"
    "strings"
)

type walker struct {
    Apis   []API.API
    Routes []API.RouteMap
    stack  []js.INode
    // Errors 	[]error
}

func (w *walker) push(n js.INode) {
    w.stack = append(w.stack, n)
}

func (w *walker) pop(n js.INode) error {
    size := len(w.stack)
    if size <= 0 {
        return errors.New("pop of empty stack")
    }
    
    toPop := w.stack[size-1]
    if toPop != n {
        return errors.New("pop: nodes do not equal")
    }
    
    w.stack[size-1] = nil
    w.stack = w.stack[:size-1]
    return nil
}

func (w *walker) read(n int) js.INode {
    size := len(w.stack)
    if size <= 0 || size-n <= 0 {
        return nil
    }
    return w.stack[size-n]
}

func (w *walker) Enter(n js.INode) js.IVisitor {
    w.push(n)
    switch n := n.(type) {
    case *js.CallExpr:
        api := w.jqueryWalker(*n)
        if api.Valid() {
            w.Apis = append(w.Apis, api)
        }
        api = w.ajaxWalker(*n)
        if api.Valid() {
            w.Apis = append(w.Apis, api)
        }
        api = w.angularWalker(*n)
        if api.Valid() {
            w.Apis = append(w.Apis, api)
        }
        api = w.fetchWalker(*n)
        if api.Valid() {
            w.Apis = append(w.Apis, api)
        }
        api = w.axiosWalker(*n)
        if api.Valid() {
            w.Apis = append(w.Apis, api)
        }
    case *js.ArrayExpr:
        routes := w.routeWalker(*n)
        w.Routes = append(w.Routes, routes...)
    }
    return w
}

var jqueryMethods = []string{"get", "post", "getJSON", "load", "ajax", "getScript"}

func (w *walker) jqueryWalker(n js.CallExpr) (api API.API) {
    api.Type = "jquery"
    dotExpression, dotOk := n.X.(*js.DotExpr) // callee
    if !dotOk {
        return
    }
    // bug: dotExpression.X string方法不存在的情况.
    if !slices.Contains([]string{"$", "jQuery"}, dotExpression.X.String()) ||
        !slices.Contains(jqueryMethods, trim(dotExpression.Y.String())) {
        return
    }
    
    api.Method = dotExpression.Y.String()
    switch len(n.Args.List) {
    case 1:
        // two cases:
        // 1. jQuery.get({url: "/example"});
        // 2. $.get("test.php");
        obj, isObject := n.Args.List[0].Value.(*js.ObjectExpr)
        if isObject {
            api.Method = "GET"
            for _, value := range obj.List {
                // parse object arg
                switch value.Name.String() {
                case "url":
                    _, ok := value.Value.(*js.LiteralExpr)
                    if ok {
                        api.UriType = 1
                    }
                    api.Uri = value.Value.String()
                case "data":
                    // data: {key:"value"}
                    dataObj, dataOk := value.Value.(*js.ObjectExpr)
                    if !dataOk {
                        break
                    }
                    for _, data := range dataObj.List {
                        api.Params = append(api.Params, API.Param{
                            Key:   data.Name.String(),
                            Value: data.Value.String(),
                        })
                    }
                case "type":
                    _, ok := value.Value.(*js.LiteralExpr)
                    if ok {
                        api.UriType = 1
                    }
                    api.Method = trim(value.Value.String())
                case "username", "password":
                    _, ok := value.Value.(*js.LiteralExpr)
                    if ok {
                        api.UriType = 1
                    }
                    api.Params = append(api.Params, API.Param{
                        Key:   "jquery_arg_" + value.Name.String(),
                        Value: trim(value.Value.String()),
                    })
                }
            }
            break
        }
        // in order
        // url: string or expression
        api.Uri = n.Args.List[0].Value.String()
        _, ok := n.Args.List[0].Value.(*js.LiteralExpr)
        if ok {
            api.UriType = 1
        }
    case 2, 3, 4:
        /*$.get("test.cgi", { name: "John", time: "2pm" },
        function(data){ alert("Data Loaded: " + data); });*/
        api.Uri = n.Args.List[0].Value.String()
        _, ok := n.Args.List[0].Value.(*js.LiteralExpr)
        if ok {
            api.UriType = 1
        }
        for _, arg := range n.Args.List {
            dataObj, isObj := arg.Value.(*js.ObjectExpr)
            if !isObj {
                continue
            }
            for _, data := range dataObj.List {
                api.Params = append(api.Params, API.Param{
                    Key:   data.Name.String(),
                    Value: data.Value.String(),
                })
            }
        }
    }
    return
}

var ajaxMethods = []string{"GET", "POST", "HEAD", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH"}

func (w *walker) ajaxWalker(n js.CallExpr) (api API.API) {
    // https://developer.mozilla.org/en-US/docs/Web/API/XMLHttpRequest/open
    // open(method, url)
    // open(method, url, async)
    // open(method, url, async, user)
    // open(method, url, async, user, password)
    api.Type = "ajax"
    
    dotExpression, dotOk := n.X.(*js.DotExpr)
    if !dotOk {
        return
    }
    if !strings.EqualFold(dotExpression.Y.String(), "open") {
        return
    }
    
    // 简单做一个过滤,防止过多误报
    if len(n.Args.List) <= 1 {
        return
    } else {
        if !slices.Contains(ajaxMethods, strings.ToUpper(trim(n.Args.List[0].String()))) {
            return
        }
    }
    
    switch len(n.Args.List) {
    case 5:
        api.Params = append(api.Params, API.Param{
            Key:   "ajax_password",
            Value: n.Args.List[4].String(),
        })
        fallthrough
    case 4:
        api.Params = append(api.Params, API.Param{
            Key:   "ajax_user",
            Value: n.Args.List[3].String(),
        })
        fallthrough
    case 2, 3:
        api.Method = n.Args.List[0].String()
        _, isStr := n.Args.List[1].Value.(*js.LiteralExpr)
        if isStr {
            api.UriType = 1
        }
        api.Uri = n.Args.List[1].String()
    }
    return
}

func trim(str string) string {
    str = strings.ReplaceAll(str, "'", "")
    str = strings.ReplaceAll(str, "\"", "")
    return str
}

var angularMethods = []string{"request", "delete", "get", "head", "jsonp", "options", "patch", "post", "put"}

func (w *walker) angularWalker(n js.CallExpr) (api API.API) {
    // https://angular.io/api/common/http/HttpClient
    // request(req)
    // request(method,url,options) // options暂时不解析
    // (delete|get|head|options)(url,options)
    // jsonp(url,callbackParam)
    // (patch|put)(url,body,options)
    
    api.Type = "angular"
    dotExpression, dotOk := n.X.(*js.DotExpr)
    if !dotOk {
        return
    }
    callee := strings.ToLower(dotExpression.X.String())
    // if !strings.Contains(callee, "http") && !strings.Contains(callee, ".") {return} // 肯定有漏报,但是目前为止没有更好的方法了
    if !strings.Contains(callee, "http") {
        return
    }
    if !slices.Contains(angularMethods, dotExpression.Y.String()) || len(n.Args.List) < 1 {
        return
    }
    api.Method = trim(dotExpression.Y.String())
    if strings.EqualFold(api.Method, "request") {
        if len(n.Args.List) == 1 {
            return
        }
        api.Method = trim(n.Args.List[0].Value.String())
        if !slices.Contains(ajaxMethods, api.Method) {
            return
        }
        api.Uri = n.Args.List[1].Value.String()
    } else {
        _, isStr := n.Args.List[0].Value.(*js.LiteralExpr)
        api.Uri = n.Args.List[0].Value.String()
        if isStr {
            api.UriType = 1
        } else {
            return
        }
    }
    return
}

func (w *walker) axiosWalker(n js.CallExpr) (api API.API) {
    // axios(config)
    // axios(url[, config])
    // axios.request(config)
    // axios.get(url[, config])
    // axios.delete(url[, config])
    // axios.head(url[, config])
    // axios.options(url[, config])
    // axios.post(url[, data[, config]])
    // axios.put(url[, data[, config]])
    // axios.patch(url[, data[, config]])
    
    // # instance这种暂时不考虑
    // instance:
    // const instance = axios.create({
    //  baseURL: 'https://some-domain.com/api/',
    //  timeout: 1000,
    //  headers: {'X-Custom-Header': 'foobar'}
    // });
    // https://www.codegrepper.com/code-examples/javascript/axios.create%28%29
    
    // 可能会存在找到了相对路径但是baseurl不对的情况
    // 也可能存在找到了baseURL但是找不到相对路径的情况
    // xx.get xx也不确定是否是否为axios对象
    
    // axios#request(config)
    // axios#get(url[, config])
    // axios#delete(url[, config])
    // axios#head(url[, config])
    // axios#options(url[, config])
    // axios#post(url[, data[, config]])
    // axios#put(url[, data[, config]])
    // axios#patch(url[, data[, config]])
    // axios#getUri([config])
    
    // await axios.postForm('https://httpbin.org/post', {
    //  'myVar' : 'foo',
    //  'file': document.querySelector('#fileInput').files[0]
    // });
    // postForm, putForm, patchForm
    api.Type = "axios"
    dotExpression, dotOk := n.X.(*js.DotExpr)
    _, isStr := n.X.(*js.LiteralExpr)
    if dotOk {
        if !strings.EqualFold(dotExpression.X.String(), "axios") {
            return
        }
        api.Method = dotExpression.Y.String()
    } else if isStr && strings.EqualFold(n.X.String(), "axios") {
        api.Method = "GET" // default
    } else {
        return
    }
    
    switch len(n.Args.List) {
    case 1:
        obj, isObj := n.Args.List[0].Value.(*js.ObjectExpr)
        if isObj {
            // https://axios-http.com/docs/req_config
            // config mode: {method:'post',url:'/user/123',params:{id,'123'},data:{name:'fred'}}
            for _, item := range obj.List {
                switch item.Name.String() {
                case "method":
                    api.Method = item.Value.String()
                case "url":
                    api.Uri = item.Value.String()
                    _, ok := item.Value.(*js.LiteralExpr)
                    if ok {
                        api.UriType = 1
                    }
                case "baseURL":
                    api.Uri = item.Value.String() + "/" + api.Uri
                case "params", "data", "auth", "proxy":
                    params, _ := item.Value.(*js.ObjectExpr)
                    for _, param := range params.List {
                        api.Params = append(api.Params, API.Param{
                            Key:   param.Name.String(),
                            Value: param.Value.String(),
                        })
                    }
                default:
                    break
                }
            }
        } else {
            api.Uri = n.Args.List[0].Value.String()
            _, ok := n.Args.List[0].Value.(*js.LiteralExpr)
            if ok {
                api.UriType = 1
            }
        }
    case 2:
        api.Uri = n.Args.List[0].String()
        obj, isObj := n.Args.List[1].Value.(*js.ObjectExpr)
        if !isObj {
            return API.API{}
        }
        // 这段重复了
        // config
        for _, item := range obj.List {
            switch item.Name.String() {
            case "method":
                api.Method = item.Value.String()
            case "url":
                api.Uri = item.Value.String()
                _, ok := item.Value.(*js.LiteralExpr)
                if ok {
                    api.UriType = 1
                }
            case "baseURL":
                api.Uri = item.Value.String() + "/" + api.Uri
            case "params", "data", "auth", "proxy":
                params, _ := item.Value.(*js.ObjectExpr)
                for _, param := range params.List {
                    api.Params = append(api.Params, API.Param{
                        Key:   param.Name.String(),
                        Value: param.Value.String(),
                    })
                }
            default:
                break
            }
        }
        // data
        for _, item := range obj.List {
            api.Params = append(api.Params, API.Param{
                Key:   item.Name.String(),
                Value: item.Value.String(),
            })
        }
    case 3:
        api.Uri = n.Args.List[0].String()
        _, ok := n.Args.List[0].Value.(*js.LiteralExpr)
        if ok {
            api.UriType = 1
        }
        obj, isObj := n.Args.List[2].Value.(*js.ObjectExpr)
        if !isObj {
            return API.API{}
        }
        // 这段重复了
        // config
        for _, item := range obj.List {
            switch item.Name.String() {
            case "method":
                api.Method = item.Value.String()
            case "url":
                api.Uri = item.Value.String()
                _, ok := item.Value.(*js.LiteralExpr)
                if ok {
                    api.UriType = 1
                }
            case "baseURL":
                api.Uri = item.Value.String() + "/" + api.Uri
            case "params", "data", "auth", "proxy":
                params, _ := item.Value.(*js.ObjectExpr)
                for _, param := range params.List {
                    api.Params = append(api.Params, API.Param{
                        Key:   param.Name.String(),
                        Value: param.Value.String(),
                    })
                }
            default:
                break
            }
        }
        // data
        obj, isObj = n.Args.List[1].Value.(*js.ObjectExpr)
        for _, item := range obj.List {
            api.Params = append(api.Params, API.Param{
                Key:   item.Name.String(),
                Value: item.Value.String(),
            })
        }
    }
    return
}

func (w *walker) fetchWalker(n js.CallExpr) (api API.API) {
    api.Type = "fetch"
    api.Method = "GET"
    // https://github.github.io/fetch/
    _, isVar := n.X.(*js.Var)
    if !isVar {
        return
    }
    if !strings.EqualFold(n.X.String(), "fetch") {
        return
    }
    
    if len(n.Args.List) < 1 {
        return
    }
    api.Uri = n.Args.List[0].String()
    _, ok := n.Args.List[0].Value.(*js.LiteralExpr)
    if ok {
        api.UriType = 1
    }
    
    if len(n.Args.List) == 2 {
        obj, isObj := n.Args.List[1].Value.(*js.ObjectExpr)
        if !isObj {
            return API.API{}
        }
        for _, item := range obj.List {
            if strings.EqualFold(item.Name.String(), "method") {
                api.Method = item.Value.String()
            }
        }
    }
    return
}

// 无法区分提取出来的前端路由是children还是routes,需要获取所有结果之后处理一遍
func (w *walker) routeWalker(n js.ArrayExpr) (routes []API.RouteMap) {
    node := w.read(2) // idx
    if node != nil {
        property, ok := node.(*js.Property)
        if ok {
            if strings.EqualFold(strings.ToLower(property.Name.String()), "children") {
                return
            }
        }
    }
    for _, item := range n.List {
        obj, isObj := item.Value.(*js.ObjectExpr)
        if !isObj {
            continue
        }
        routes = append(routes, w.routerHandler(*obj))
    }
    return
}

func (w *walker) routerHandler(n js.ObjectExpr) (route API.RouteMap) {
    for _, kv := range n.List {
        if kv.Name == nil {
            continue
        }
        switch kv.Name.Literal.String() {
        case "path":
            route.Path = kv.Value.String()
        case "children":
            obj, isObj := kv.Value.(*js.ArrayExpr)
            if !isObj {
                continue
            }
            route.Children = w.routeWalker(*obj)
        default:
            // pass
        }
    }
    return
}

// func ObjToString(n js.PropertyName) string {
// }

func (w *walker) Exit(n js.INode) {
    w.pop(n)
}
