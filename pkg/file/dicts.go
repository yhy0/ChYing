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
