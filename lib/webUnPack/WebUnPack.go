package webUnPack

import (
    "github.com/yhy0/ChYing/lib/webUnPack/pkg"
)

/**
  @author: yhy
  @since: 2023/7/26
  @desc: //TODO
**/

func Run(target string) {
    options := &pkg.Options{
        Output: "./output",
        Target: target,
    }
    
    pkg.Run(options)
}
