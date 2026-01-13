package jsluice

import (
    tree_sitter "github.com/tree-sitter/go-tree-sitter"
    tree_sitter_javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
    "strconv"
    "testing"
)

func TestCollapsedString(t *testing.T) {
    cases := []struct {
        JS       []byte
        Expected string
    }{
        {[]byte(`"./login.php?redirect="+url`), "./login.php?redirect=EXPR"},
        {[]byte(`'/path/'+['one', 'two', 'three'].join('/')`), "/path/EXPR"},
        {[]byte(`someVar`), "EXPR"},
    }
    
    parser := tree_sitter.NewParser()
    parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_javascript.Language()))
    
    for i, c := range cases {
        t.Run(strconv.Itoa(i), func(t *testing.T) {
            tree := parser.Parse(c.JS, nil)
            root := NewNode(tree.RootNode(), c.JS)
            
            // Example tree:
            //   program
            //     expression_statement
            //       binary_expression
            //         left: string ("./login.php?redirect=")
            //         right: identifier (url)
            //
            // We want the binary_expression to pass to CollapsedString, which is
            // the first Named Child of the first Named Child of the root node.
            actual := root.NamedChild(0).NamedChild(0).CollapsedString()
            
            if actual != c.Expected {
                t.Errorf("want %s for CollapsedString(%s), have: %s", c.Expected, c.JS, actual)
            }
        })
    }
}
