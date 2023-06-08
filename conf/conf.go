package conf

import (
	"fmt"
	"time"
)

/**
  @author: yhy
  @since: 2023/4/20
  @desc: //TODO
**/

var Proxy string

var Description = fmt.Sprintf("将旦昧爽之交，日夕昏明之际，\n北面而察之，淡淡焉若有物存，莫识其状。\n其所触也，窃窃然有声，经物而物不疾也。\n\n© %d https://github.com/yhy0", time.Now().Year())

const (
	Version        = "v1.1"
	Title          = "承影 " + Version
	VersionNewMsg  = "当前已经是最新版本!"
	VersionOldMsg  = "最新版本: %s, 是否立即更新?"
	BtnConfirmText = "确定"
	BtnCancelText  = "取消"
)
