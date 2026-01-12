package file

import (
    "embed"
)

// https://github.com/lijiejie/BBScan/tree/master/rules
//
//go:embed bbscan
var bbscanRules embed.FS

// https://github.com/maurosoria/dirsearch/blob/master/db/dicc.txt
//
//go:embed dict.txt
var fileDict embed.FS

// https://github.com/devploit/dontgo403/tree/main/payloads
//
//go:embed 403bypass
var bypass403 embed.FS

//go:embed jwt.txt
var jwtSecrets embed.FS

//go:embed default_mitm_rule.json
var mitmRules embed.FS
