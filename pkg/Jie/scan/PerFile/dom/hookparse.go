package dom

import (
    "fmt"
    "github.com/yhy0/logging"
    "runtime"
    
    "github.com/tdewolff/parse/v2"
    "github.com/tdewolff/parse/v2/js"
)

/**
  @author: yhy
  @since: 2023/3/10
  @desc: https://github.com/ac0d3r/xssfinder
    将 js 转换为 yak/protocols/headless/preload.js 需要的格式，然后交给浏览器进行执行 js，进行污染传播分析
    todo 有待优化，很多无法识别, 覆盖太少了
**/

// HookParse 必须传入<script>xxxx</script> 中的 xxxx
func HookParse(code string) (string, error) {
    defer func() {
        if err := recover(); err != nil {
            logging.Logger.Errorln("recover from:", err)
            debugStack := make([]byte, 1024)
            runtime.Stack(debugStack, false)
            logging.Logger.Errorf("Stack Trace:%v", string(debugStack))
        }
    }()
    
    ast, err := js.Parse(parse.NewInputString(code), js.Options{})
    if err != nil {
        return "", err
    }
    convWalk(ast)
    
    return ast.String(), nil
}

func convExpr(node js.IExpr) js.IExpr {
    switch n := node.(type) {
    case *js.BinaryExpr:
        switch n.Op {
        case js.AddToken: // +
            return &js.CallExpr{
                X: &js.Var{Data: []byte("__xssfinder_plus")},
                Args: js.Args{
                    List: []js.Arg{{Value: convExpr(n.X)}, {Value: convExpr(n.Y)}},
                },
            }
        case js.AddEqToken: // +=
            // // a += b ---> a = __xssfinder_plus(a, b)
            n.Op = js.EqToken
            n.Y = &js.CallExpr{
                X: &js.Var{Data: []byte("__xssfinder_plus")},
                Args: js.Args{
                    List: []js.Arg{{Value: convExpr(n.X)}, {Value: convExpr(n.Y)}},
                },
            }
        case js.EqEqToken: // ==
            return &js.CallExpr{
                X: &js.Var{Data: []byte("__xssfinder_equal")},
                Args: js.Args{
                    List: []js.Arg{{Value: convExpr(n.X)}, {Value: convExpr(n.Y)}},
                },
            }
        case js.NotEqToken: // !=
            return &js.CallExpr{
                X: &js.Var{Data: []byte("__xssfinder_notEqual")},
                Args: js.Args{
                    List: []js.Arg{{Value: convExpr(n.X)}, {Value: convExpr(n.Y)}},
                },
            }
        case js.EqEqEqToken: // ===
            return &js.CallExpr{
                X: &js.Var{Data: []byte("__xssfinder_strictEqual")},
                Args: js.Args{
                    List: []js.Arg{{Value: convExpr(n.X)}, {Value: convExpr(n.Y)}},
                },
            }
        case js.NotEqEqToken: // !==
            return &js.CallExpr{
                X: &js.Var{Data: []byte("__xssfinder_strictNotEqual")},
                Args: js.Args{
                    List: []js.Arg{{Value: convExpr(n.X)}, {Value: convExpr(n.Y)}},
                },
            }
        case js.EqToken: // =
            if obj, ok := n.X.(*js.DotExpr); ok {
                return &js.CallExpr{
                    X: &js.Var{Data: []byte("__xssfinder_put")},
                    Args: js.Args{
                        List: []js.Arg{
                            {Value: obj.X},
                            {Value: &js.LiteralExpr{
                                TokenType: js.StringToken,
                                Data:      []byte(fmt.Sprintf(`"%s"`, obj.Y.String())),
                            }},
                            {Value: convExpr(n.Y)}},
                    },
                }
            } else {
                n.Y = convExpr(n.Y)
            }
        }
    case *js.DotExpr:
        return &js.CallExpr{
            X: &js.Var{Data: []byte("__xssfinder_get")},
            Args: js.Args{
                List: []js.Arg{{Value: convExpr(n.X)}, {Value: &js.LiteralExpr{
                    TokenType: js.StringToken,
                    Data:      []byte(fmt.Sprintf(`"%s"`, n.Y.String())),
                }}},
            },
        }
    case *js.IndexExpr:
        return &js.CallExpr{
            X: &js.Var{Data: []byte("__xssfinder_get")},
            Args: js.Args{
                List: []js.Arg{{Value: convExpr(n.X)}, {Value: convExpr(n.Y)}},
            },
        }
    case *js.CallExpr:
        for i := range n.Args.List {
            n.Args.List[i].Value = convExpr(n.Args.List[i].Value)
        }
        
        if dot, ok := n.X.(*js.DotExpr); ok {
            args := js.Args{
                List: make([]js.Arg, 2+len(n.Args.List)),
            }
            
            args.List[0] = js.Arg{Value: convExpr(dot.X)}
            args.List[1] = js.Arg{Value: &js.LiteralExpr{
                TokenType: js.StringToken,
                Data:      []byte(fmt.Sprintf(`"%s"`, dot.Y.String())),
            }}
            for i := range n.Args.List {
                args.List[i+2] = n.Args.List[i]
            }
            return &js.CallExpr{
                X:    &js.Var{Data: []byte("__xssfinder_property_call")},
                Args: args,
            }
        } else {
            args := js.Args{
                List: make([]js.Arg, 1+len(n.Args.List)),
            }
            args.List[0] = js.Arg{Value: n.X}
            for i := range n.Args.List {
                args.List[i+1] = n.Args.List[i]
            }
            return &js.CallExpr{
                X:    &js.Var{Data: []byte("__xssfinder_call")},
                Args: args,
            }
        }
    case *js.UnaryExpr:
        switch n.Op {
        case js.TypeofToken:
            var idvar js.IExpr
            switch vvar := n.X.(type) {
            case *js.Var:
                idvar = vvar
            case *js.GroupExpr:
                idvar = vvar.X
            default:
                return &js.CallExpr{
                    X: &js.Var{Data: []byte("__xssfinder_typeof")},
                    Args: js.Args{
                        List: []js.Arg{{Value: n.X}},
                    },
                }
            }
            cond := &js.CondExpr{
                Cond: &js.BinaryExpr{
                    Op: js.EqEqEqToken,
                    X: &js.UnaryExpr{
                        Op: js.TypeofToken,
                        X:  idvar,
                    },
                    Y: &js.LiteralExpr{
                        TokenType: js.StringToken,
                        Data:      []byte(`"undefined"`),
                    },
                },
                X: &js.Var{Data: []byte("undefined")},
                Y: idvar,
            }
            return &js.CallExpr{
                X: &js.Var{Data: []byte("__xssfinder_typeof")},
                Args: js.Args{
                    List: []js.Arg{{Value: cond}},
                },
            }
        }
    case *js.NewExpr:
        // if v, ok := n.X.(*js.Var); ok && string(v.Data) == "Function" {
        if _, ok := n.X.(*js.Var); ok && n.Args != nil {
            return &js.CallExpr{
                X:    &js.Var{Data: []byte("__xssfinder_new_Function")},
                Args: *n.Args,
            }
        }
    }
    return node
}

// convWalk traverses an AST in depth-first order
func convWalk(n js.INode) {
    if n == nil {
        return
    }
    
    switch n := n.(type) {
    case *js.AST:
        convWalk(&n.BlockStmt)
    case *js.Var:
        return
    case *js.BlockStmt:
        if n.List != nil {
            for i := 0; i < len(n.List); i++ {
                convWalk(n.List[i])
            }
        }
    case *js.EmptyStmt:
        return
    case *js.ExprStmt:
        n.Value = convExpr(n.Value)
    case *js.VarDecl:
        if n.List != nil {
            for i := range n.List {
                if n.List[i].Default != nil {
                    n.List[i].Default = convExpr(n.List[i].Default)
                }
            }
        }
    case *js.IfStmt:
        convWalk(n.Body)
        convWalk(n.Else)
        n.Cond = convExpr(n.Cond)
    case *js.DoWhileStmt:
        convWalk(n.Body)
        n.Cond = convExpr(n.Cond)
    case *js.WhileStmt:
        convWalk(n.Body)
        n.Cond = convExpr(n.Cond)
    case *js.ForStmt:
        if n.Body != nil {
            convWalk(n.Body)
        }
        
        n.Init = convExpr(n.Init)
        n.Cond = convExpr(n.Cond)
        n.Post = convExpr(n.Post)
    case *js.ForInStmt:
        if n.Body != nil {
            convWalk(n.Body)
        }
        
        n.Init = convExpr(n.Init)
        n.Value = convExpr(n.Value)
    case *js.ForOfStmt:
        if n.Body != nil {
            convWalk(n.Body)
        }
        
        n.Init = convExpr(n.Init)
        n.Value = convExpr(n.Value)
    case *js.CaseClause:
        if n.List != nil {
            for i := 0; i < len(n.List); i++ {
                convWalk(n.List[i])
            }
        }
        
        n.Cond = convExpr(n.Cond)
    case *js.SwitchStmt:
        if n.List != nil {
            for i := 0; i < len(n.List); i++ {
                convWalk(&n.List[i])
            }
        }
        
        n.Init = convExpr(n.Init)
    case *js.BranchStmt:
        return
    case *js.ReturnStmt:
        n.Value = convExpr(n.Value)
    case *js.WithStmt:
        convWalk(n.Body)
        n.Cond = convExpr(n.Cond)
    case *js.LabelledStmt:
        convWalk(n.Value)
    case *js.ThrowStmt:
        n.Value = convExpr(n.Value)
    case *js.TryStmt:
        if n.Body != nil {
            convWalk(n.Body)
        }
        
        if n.Catch != nil {
            convWalk(n.Catch)
        }
        
        if n.Finally != nil {
            convWalk(n.Finally)
        }
        
        convWalk(n.Binding)
    case *js.DebuggerStmt:
        return
    case *js.Alias:
        return
    case *js.ImportStmt:
        // TODO
        // if n.List != nil {
        //     for i := 0; i < len(n.List); i++ {
        //         convWalk(&n.List[i])
        //     }
        // }
        return
    case *js.ExportStmt:
        // TODO
        // if n.List != nil {
        //     for i := 0; i < len(n.List); i++ {
        //         convWalk(&n.List[i])
        //     }
        // }
        
        // n.Decl = convExpr(n.Decl)
        return
    case *js.DirectivePrologueStmt:
        return
    case *js.PropertyName:
        // TODO
        // convWalk(&n.Literal)
        // n.Computed = convExpr(n.Computed)
        return
    case *js.BindingArray:
        // TODO
        // if n.List != nil {
        //     for i := 0; i < len(n.List); i++ {
        //         convWalk(&n.List[i])
        //     }
        // }
        
        // convWalk(n.Rest)
        return
    case *js.BindingObjectItem:
        // TODO
        // if n.Key != nil {
        //     convWalk(n.Key)
        // }
        
        // convWalk(&n.Value)
        return
    case *js.BindingObject:
        // TODO
        // if n.List != nil {
        //     for i := 0; i < len(n.List); i++ {
        //         convWalk(&n.List[i])
        //     }
        // }
        
        // if n.Rest != nil {
        //     convWalk(n.Rest)
        // }
        return
    case *js.BindingElement:
        // convWalk(n.Binding)
        n.Default = convExpr(n.Default)
    case *js.Params:
        if n.List != nil {
            for i := range n.List {
                n.List[i].Default = convExpr(n.List[i].Default)
            }
        }
        
        // convWalk(n.Rest)
    case *js.FuncDecl:
        convWalk(&n.Body)
        convWalk(&n.Params)
        
        // if n.Name != nil {
        //     convWalk(n.Name)
        // }
    case *js.MethodDecl:
        convWalk(&n.Body)
        convWalk(&n.Params)
        // convWalk(&n.Name)
    case *js.Field:
        // convWalk(&n.Name)
        n.Init = convExpr(n.Init)
    case *js.ClassDecl:
        // if n.Name != nil {
        //     convWalk(n.Name)
        // }
        
        n.Extends = convExpr(n.Extends)
        
        if n.List != nil {
            for i := 0; i < len(n.List); i++ {
                convWalk(n.List[i].Field)
                convWalk(n.List[i].Method)
            }
        }
    case *js.LiteralExpr:
        return
    case *js.Element:
        n.Value = convExpr(n.Value)
    case *js.ArrayExpr:
        if n.List != nil {
            for i := 0; i < len(n.List); i++ {
                n.List[i].Value = convExpr(n.List[i].Value)
            }
        }
    case *js.Property:
        if n.Name != nil {
            convWalk(n.Name)
        }
        
        n.Value = convExpr(n.Value)
        n.Init = convExpr(n.Init)
    case *js.ObjectExpr:
        if n.List != nil {
            for i := 0; i < len(n.List); i++ {
                convWalk(&n.List[i])
            }
        }
    case *js.TemplatePart:
        n.Expr = convExpr(n.Expr)
    case *js.TemplateExpr:
        if n.List != nil {
            for i := 0; i < len(n.List); i++ {
                convWalk(&n.List[i])
            }
        }
        
        n.Tag = convExpr(n.Tag)
    case *js.GroupExpr:
        return
    case *js.IndexExpr:
        return
    case *js.DotExpr:
        return
    case *js.NewTargetExpr:
        return
    case *js.ImportMetaExpr:
        return
    case *js.Arg:
        return
    case *js.Args:
    case *js.NewExpr:
    case *js.CallExpr:
    
    case *js.UnaryExpr:
    
    case *js.BinaryExpr:
    
    case *js.CondExpr:
    
    case *js.YieldExpr:
        n.X = convExpr(n.X)
    case *js.ArrowFunc:
        convWalk(&n.Body)
        convWalk(&n.Params)
    case *js.CommaExpr:
        for i := range n.List {
            n.List[i] = convExpr(n.List[i])
        }
    default:
        return
    }
}
