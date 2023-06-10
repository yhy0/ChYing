package test

import (
	_ "embed"
	"fmt"
	"github.com/yhy0/ChYing/tools/gadget"
	"testing"
)

/**
   @author yhy
   @since 2023/6/9
   @desc //TODO
**/

func TestShiro(t *testing.T) {
	shiro, err := gadget.DecryptShiro("", "")
	if err != nil {
		return
	}

	fmt.Printf("%v", shiro)
}
