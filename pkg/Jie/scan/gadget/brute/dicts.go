package brute

import (
    _ "embed"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/util"
)

var (
    tomcatuserpass   = []util.UserPass{}
    jbossuserpass    = []util.UserPass{}
    top100pass       = []string{}
    weblogicuserpass = []util.UserPass{}
)

//go:embed dicts/tomcatuserpass.txt
var szTomcatuserpass string

//go:embed dicts/jbossuserpass.txt
var szJbossuserpass string

//go:embed dicts/weblogicuserpass.txt
var szWeblogicuserpass string

//go:embed dicts/top100pass.txt
var szTop100pass string

func init() {
    tomcatuserpass = util.CvtUps(szTomcatuserpass)
    jbossuserpass = util.CvtUps(szJbossuserpass)
    weblogicuserpass = util.CvtUps(szWeblogicuserpass)
    top100pass = append(top100pass, util.CvtLines(szTop100pass)...)
}
