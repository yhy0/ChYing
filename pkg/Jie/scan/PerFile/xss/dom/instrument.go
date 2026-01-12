package dom

import (
	"bytes"
	"fmt"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
)

// taintTrackerJS is the taint tracking library injected into the JS code.
const taintTrackerJS = `
var __TaintTracker = (function() {
    const taintedObjects = new WeakMap();
    function source(name, value) {
        if (value !== null && (typeof value === 'object' || typeof value === 'function' || typeof value === 'string')) {
            taintedObjects.set(value, { source: name });
        }
        return value;
    }
    function check(sink, value) {
        if (value !== null && (typeof value === 'object' || typeof value === 'function' || typeof value === 'string')) {
            if (taintedObjects.has(value)) {
                const info = taintedObjects.get(value);
                console.warn('[XSS Taint Found] Sink:', sink, 'Source:', info.source, 'Value:', value);
                if (window.reportTaint) {
                    window.reportTaint({ Found: true, Source: info.source, Sink: sink, URL: window.location.href });
                }
            }
        }
        return value;
    }
    return { source: source, check: check };
})();
`

// InstrumentJS instruments the given JavaScript code for taint tracking.
func InstrumentJS(originalJS string) (string, error) {
	ast, err := js.Parse(parse.NewInputString(originalJS), js.Options{})
	if err != nil {
		return "", fmt.Errorf("parsing js failed: %w", err)
	}

	rewriter := &jsRewriter{}
	rewriter.walk(ast)

	var buf bytes.Buffer
	buf.WriteString(taintTrackerJS)
	buf.WriteString(ast.String())
	return buf.String(), nil
}

// jsRewriter walks the AST and rewrites nodes to insert taint tracking hooks.
type jsRewriter struct{}

// walk recursively traverses the AST.
func (r *jsRewriter) walk(node js.INode) {
	if node == nil {
		return
	}
	// A massive switch to handle all node types, inspired by hookparse.go
	// This ensures we traverse the entire AST.
	switch n := node.(type) {
	case *js.AST:
		r.walk(&n.BlockStmt)
	case *js.BlockStmt:
		for i := range n.List {
			r.walk(n.List[i])
		}
	case *js.ExprStmt:
		n.Value = r.rewriteExpr(n.Value)
	case *js.VarDecl:
		for i := range n.List {
			if n.List[i].Default != nil {
				n.List[i].Default = r.rewriteExpr(n.List[i].Default)
			}
		}
	case *js.IfStmt:
		n.Cond = r.rewriteExpr(n.Cond)
		r.walk(n.Body)
		r.walk(n.Else)
	// Add cases for all other statement types to ensure full traversal...
	case *js.DoWhileStmt:
		r.walk(n.Body)
		n.Cond = r.rewriteExpr(n.Cond)
	case *js.WhileStmt:
		n.Cond = r.rewriteExpr(n.Cond)
		r.walk(n.Body)
	case *js.ForStmt:
		r.walk(n.Init)
		n.Cond = r.rewriteExpr(n.Cond)
		n.Post = r.rewriteExpr(n.Post)
		r.walk(n.Body)
	case *js.ForInStmt:
		r.walk(n.Init)
		n.Value = r.rewriteExpr(n.Value)
		r.walk(n.Body)
	case *js.ForOfStmt:
		r.walk(n.Init)
		n.Value = r.rewriteExpr(n.Value)
		r.walk(n.Body)
	case *js.SwitchStmt:
		n.Init = r.rewriteExpr(n.Init)
		for i := range n.List {
			r.walk(&n.List[i])
		}
	case *js.CaseClause:
		n.Cond = r.rewriteExpr(n.Cond)
		for i := range n.List {
			r.walk(n.List[i])
		}
	case *js.TryStmt:
		r.walk(n.Body)
		r.walk(n.Catch)
		r.walk(n.Finally)
	case *js.FuncDecl:
		r.walk(&n.Body)
	// For brevity, only showing a few statement types. A full implementation
	// would have cases for ForStmt, WhileStmt, SwitchStmt, etc.
	case *js.ReturnStmt:
		if n.Value != nil {
			n.Value = r.rewriteExpr(n.Value)
		}
	default:
		// For other node types, we might need to recursively walk them if they
		// contain expressions or statements. This simplified version omits that.
	}
}

// rewriteExpr rewrites an expression node to wrap sources and sinks.
func (r *jsRewriter) rewriteExpr(node js.IExpr) js.IExpr {
	if node == nil {
		return nil
	}

	// First, recursively rewrite sub-expressions.
	switch n := node.(type) {
	case *js.BinaryExpr:
		n.X = r.rewriteExpr(n.X)
		n.Y = r.rewriteExpr(n.Y)
	case *js.UnaryExpr:
		n.X = r.rewriteExpr(n.X)
	case *js.CondExpr:
		n.Cond = r.rewriteExpr(n.Cond)
		n.X = r.rewriteExpr(n.X)
		n.Y = r.rewriteExpr(n.Y)
	case *js.DotExpr:
		n.X = r.rewriteExpr(n.X)
	case *js.IndexExpr:
		n.X = r.rewriteExpr(n.X)
		n.Y = r.rewriteExpr(n.Y)
	case *js.CallExpr:
		n.X = r.rewriteExpr(n.X)
		for i := range n.Args.List {
			n.Args.List[i].Value = r.rewriteExpr(n.Args.List[i].Value)
		}
	}

	// Then, check if the current node itself is a source or sink.
	if dotExpr, ok := node.(*js.DotExpr); ok {
		if ident, ok := dotExpr.X.(*js.Var); ok && string(ident.Data) == "location" {
			sourceName := string(dotExpr.Y.Data)
			if isSource(sourceName) {
				// location.hash -> __TaintTracker.source('location.hash', location.hash)
				return &js.CallExpr{
					X: &js.Var{Data: []byte("__TaintTracker.source")},
					Args: js.Args{List: []js.Arg{
						{Value: &js.LiteralExpr{TokenType: js.StringToken, Data: []byte(fmt.Sprintf(`'%s'`, "location."+sourceName))}},
						{Value: node},
					}},
				}
			}
		}
	}

	// Check for sinks (e.g., element.innerHTML = value)
	// This happens in assignment expressions, which are handled in the walk function for VarDecl or as BinaryExpr with '=' token.
	if assign, ok := node.(*js.BinaryExpr); ok && assign.Op == js.EqToken {
		if prop, ok := assign.X.(*js.DotExpr); ok {
			sinkName := string(prop.Y.Data)
			if isSink(sinkName) {
				// .innerHTML = x -> .innerHTML = __TaintTracker.check('innerHTML', x)
				assign.Y = &js.CallExpr{
					X: &js.Var{Data: []byte("__TaintTracker.check")},
					Args: js.Args{List: []js.Arg{
						{Value: &js.LiteralExpr{TokenType: js.StringToken, Data: []byte(fmt.Sprintf(`'%s'`, sinkName))}},
						{Value: r.rewriteExpr(assign.Y)},
					}},
				}
				return assign
			}
		}
	}

	return node
}

// isSource 判断一个标识符是否是已知的DOM XSS Source。
func isSource(s string) bool {
	// 为了简化，我们只检查几个关键的source
	// 在实际应用中，这里应该是一个更完整的列表
	switch s {
	case "location", "href", "search", "hash", "pathname", "cookie", "referrer", "name":
		return true
	}
	return false
}

// isSink 判断一个标识符是否是已知的DOM XSS Sink。
func isSink(s string) bool {
	// 为了简化，我们只检查几个关键的sink
	switch s {
	case "innerHTML", "outerHTML", "insertAdjacentHTML", "write", "writeln":
		return true
	case "eval", "setTimeout", "setInterval", "execScript":
		return true
	}
	return false
}
