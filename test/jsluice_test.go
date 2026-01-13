//go:build !windows

package test

import (
    "fmt"
    "github.com/yhy0/ChYing/lib/jsluice"
    "github.com/yhy0/logging"
    "strconv"
    "strings"
    "testing"
)

/**
   @author yhy
   @since 2024/9/2
   @desc //TODO
**/

func TestJsluice(t *testing.T) {
    body := `    pt.setRequestInterceptors(function (e) {
      var t = Object(xe["a"])("accessToken"),
        n =
          "123/api/v1/user/pwd/login" == e.url ||
          "service-user222/api/v1/user/sliding/block/query" == e.url ||
          "service-user/api/v1/user/send/verify/code" == e.url ||
          "456/api/v1/user/send/email_verify/code" == e.url ||
          "service-user/api/v1/user/register" == e.url ||
          "service-aggregation/api/v1/verification/code/query" == e.url ||
          "service-user/api/v1/user/reset" == e.url;
      return (
        t &&
          ((e.headers["access_token"] = n ? "" : t),
          (e.headers["group_id"] = Object(xe["a"])("groupId")),
          (e.headers["grant_type"] = "password")),
        (e.headers["application_key"] = Object(xe["a"])("applicationKey")),
        "service-user/api/v1/user/register" === e.url &&
          (e.headers["application_key"] =
            "1331ec8b-3d40-4d9e-9c33-77a8760fe617"),
        (e.headers["Cache-Control"] = "no-cache"),
        (e.headers["Pragma"] = "no-cache"),
        (e.headers["rentId"] = "TRAINING"),
        (e.headers.endpoint = "LOGIN"),
        e
      );
    });
    var dt = lt.axios,
      ft = (n("c975"), "service-user/api/v1/user"),
      ht = "service-user/api/v1/group";
    function mt(e, t) {
      return e({
        url: "".concat(ht, "/pwd/login11"),
        method: "POST",
        data: t,
      }).then(function (e) {
        return e.data;
      });
    }
    function gt(e, t) {
      return e({
        url: "".concat(ft, "/code/login"),
        method: "POST",
        data: t,
      }).then(function (e) {
        return e.data;
      });
    }
    function bt(e, t) {
      return e({
        url: "".concat(ft, "/first-update"),
        method: "POST",
        data: t,
      }).then(function (e) {
        return e.data;
      });
    }`
    logging.Logger = logging.New(true, "", "ChYing", true)
    analyzer := jsluice.NewAnalyzer([]byte(body))
    
    for _, res := range []string{"123", "456"} {
        analyzer.AddURLMatcher(
            // The first value in the jsluice.URLMatcher struct is the type of node to look for.
            // It can be one of "string", "assignment_expression", or "call_expression"
            jsluice.URLMatcher{"string", func(n *jsluice.Node) *jsluice.URL {
                val := n.DecodedString()
                if !strings.HasPrefix(val, res) {
                    return nil
                }
                
                return &jsluice.URL{
                    URL:  val,
                    Type: "mailto",
                }
            }},
        )
    }
    
    analyzer.AddURLMatcher(
        // The first value in the jsluice.URLMatcher struct is the type of node to look for.
        // It can be one of "string", "assignment_expression", or "call_expression"
        jsluice.URLMatcher{"string", func(n *jsluice.Node) *jsluice.URL {
            val := n.DecodedString()
            if !strings.HasPrefix(val, "service-user/api/v1/") {
                return nil
            }
            
            return &jsluice.URL{
                URL:  val,
                Type: "mailto",
            }
        }},
    )
    
    for _, res := range analyzer.GetURLs() {
        fmt.Println(res)
        // j, err := json.MarshalIndent(res, "", "  ")
        // if err != nil {
        //     continue
        // }
        //
        // fmt.Printf("%s\n", j)
    }
}

func TestUrlmatcher(t *testing.T) {
    var asd int64
    asd = 1500
    fmt.Println(strconv.FormatInt(asd, 10))
    analyzer := jsluice.NewAnalyzer([]byte(`
        var fn = () => {
            var meta = {
                contact: "mailto:contact@example.com",
                home: "https://example.com"
            }
            return meta
        }
    `))
    
    analyzer.AddURLMatcher(
        // The first value in the jsluice.URLMatcher struct is the type of node to look for.
        // It can be one of "string", "assignment_expression", or "call_expression"
        jsluice.URLMatcher{Type: "string", Fn: func(n *jsluice.Node) *jsluice.URL {
            val := n.DecodedString()
            if !strings.HasPrefix(val, "mailto:") {
                return nil
            }
            
            return &jsluice.URL{
                URL:  val,
                Type: "mailto",
            }
        }},
    )
    
    for _, match := range analyzer.GetURLs() {
        fmt.Println(match.URL)
    }
}
