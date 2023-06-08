package test

import (
	"github.com/yhy0/ChYing/tools/nucleiY"
	"github.com/yhy0/logging"
	"testing"
)

/**
   @author yhy
   @since 2023/6/8
   @desc //TODO
**/

func TestNucleiY(t *testing.T) {
	logging.New(true, "", "nucleiY")
	nucleiY.New()
	nucleiY.Scan("", "")
}
