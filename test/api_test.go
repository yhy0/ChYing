package test

import (
	"fmt"
	"github.com/yhy0/ChYing/conf"
	"github.com/yhy0/ChYing/pkg/Jie/scan/gadget/collection"
	"github.com/yhy0/ChYing/pkg/utils"
	"github.com/yhy0/logging"
	"strings"
	"testing"
)

/**
   @author yhy
   @since 2024/9/2
   @desc //TODO
**/

func TestApi(t *testing.T) {
	logging.Logger = logging.New(true, "", "ChYing", true)

	// Jie 配置由 ChYing 统一管理，不再单独初始化
	// 如需独立 Jie，取消下行注释:
	// JieConf.Init(file.ChyingDir)

	// 使用 ChYing 配置初始化
	conf.InitClickHouseConfig()
	conf.SyncJieConfig()
	body := `self.__BUILD_MANIFEST = function(s, e, c, a, t, o) {
    return {
        __rewrites: {
            beforeFiles: [],
            afterFiles: [],
            fallback: []
        },
        "/": [a, t, o, "static/chunks/pages/index-0b985ea04a49b255.js"],
        "/404": ["static/chunks/pages/404-bd2636e3f9050644.js"],
        "/_error": ["static/chunks/pages/_error-8353112a01355ec2.js"],
        "/about": ["static/chunks/pages/about-1943bf125f2e41c1.js"],
        "/blog": [c, s, "static/chunks/pages/blog-e2a4cf2a88129c95.js"],
        "/blog/categories/experience": [c, s, "static/chunks/pages/blog/categories/experience-3ea34fb480c0e1b6.js"],
        "/blog/categories/others": [c, s, "static/chunks/pages/blog/categories/others-bf1cf6762fe3f32e.js"],
        "/blog/categories/sharing": [c, s, "static/chunks/pages/blog/categories/sharing-4d52b79cb44ede72.js"],
        "/blog/categories/software": [c, s, "static/chunks/pages/blog/categories/software-74a15158e8306922.js"],
        "/blog/categories/translate": [c, s, "static/chunks/pages/blog/categories/translate-bca65e7a24187dcd.js"],
        "/blog/categories/typographic": [c, s, "static/chunks/pages/blog/categories/typographic-91780676b711258f.js"],
        "/blog/components/2023-03-12/Dragable": ["static/chunks/446-1efbd238eeafdfb8.js", "static/chunks/pages/blog/components/2023-03-12/Dragable-05abecd7643b8bf7.js"],
        "/blog/components/common/DownloadFile": ["static/chunks/pages/blog/components/common/DownloadFile-080ca4ab50258ebc.js"],
        "/blog/[slug]": [s, a, "static/chunks/pages/blog/[slug]-26f486b6a992e821.js"],
        "/me": [a, t, o, "static/chunks/pages/me-2dab1be0b037e601.js"],
        "/project/analytics": [e, "static/chunks/pages/project/analytics-3c4c61616a8328bd.js"],
        "/project/analytics-guide": [e, "static/chunks/pages/project/analytics-guide-25e36869b1a86867.js"],
        "/project/ones/app-subscription": [e, "static/css/599a71fc145bcf35.css", "static/chunks/pages/project/ones/app-subscription-c652a2b365848804.js"],
        "/project/qlchat": [e, "static/chunks/pages/project/qlchat-51c410cfe80454d4.js"],
        "/project/sl-appstore": [e, "static/chunks/pages/project/sl-appstore-18899c53e678c301.js"],
        "/project/sl-appstore_用组件前的备份": ["static/chunks/pages/project/sl-appstore_用组件前的备份-f144a50b68c54c0d.js"],
        "/project/slapp-components": [e, "static/chunks/pages/project/slapp-components-08718764b38e8a1c.js"],
        "/project/slapp-components_用组件前的备份": ["static/chunks/pages/project/slapp-components_用组件前的备份-4ff6c701c8e9b679.js"],
        "/project/watsons": [e, "static/chunks/pages/project/watsons-0593a621d1ef6934.js"],
        "/project/ytscrm": [e, "static/chunks/pages/project/ytscrm-a68ce34f76560c9b.js"],
        "/updates": ["static/chunks/pages/updates-928e2451566e1db1.js"],
        sortedPages: ["/", "/404", "/_app", "/_error", "/about", "/blog", "/blog/categories/experience", "/blog/categories/others", "/blog/categories/sharing", "/blog/categories/software", "/blog/categories/translate", "/blog/categories/typographic", "/blog/components/2023-03-12/Dragable", "/blog/components/common/DownloadFile", "/blog/[slug]", "/me", "/project/analytics", "/project/analytics-guide", "/project/ones/app-subscription", "/project/qlchat", "/project/sl-appstore", "/project/sl-appstore_用组件前的备份", "/project/slapp-components", "/project/slapp-components_用组件前的备份", "/project/watsons", "/project/ytscrm", "/updates"]
    }
}("static/chunks/358-14c60bb94e9d94b9.js", "static/chunks/651-bf564aa041711377.js", "static/chunks/d64684d8-97a0d32f0fd4f32b.js", "static/chunks/893-51f6199e919234fa.js", "static/chunks/616-9b82b5014ed2cb4d.js", "static/css/691a4ce23c4238d3.css"),
self.__BUILD_MANIFEST_CB && self.__BUILD_MANIFEST_CB();
`
	fmt.Println(collection.Info("www.baidu.com", "www.baidu.com", body, ""))

	fmt.Println(utils.IsPortOccupied(conf.ProxyPort))

	fmt.Println(strings.Split("unregexp<SEP>\"proc_chan\":.*(categraf_upgrade_bak|_sre-flow-analyzer)", "<SEP>")[1])
}
