package conf

import (
    wappalyzer "github.com/projectdiscovery/wappalyzergo"
)

/**
  @author: yhy
  @since: 2023/2/1
  @desc: //TODO
**/

var GlobalConfig = &Config{}

var ConfigFile string

var NoProgressBar bool

var Wappalyzer *wappalyzer.Wappalyze
