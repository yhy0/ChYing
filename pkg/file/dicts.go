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
//go:embed dicc.txt
var fileDicc embed.FS

//go:embed twj.txt
var jwtSecrets embed.FS

// https://github.com/devploit/dontgo403/tree/main/payloads
//
//go:embed 403bypass
var bypass403 embed.FS

// AvFile https://github.com/r00tSe7en/get_AV/blob/master/av.json
//
//go:embed av.json
var AvFile embed.FS
