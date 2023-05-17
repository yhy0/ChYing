package utils

import "regexp"

/**
  @author: yhy
  @since: 2023/5/17
  @desc: //TODO
**/

func RegexpStr(patterns []string, str string) bool {
	for _, pattern := range patterns {
		match, err := regexp.MatchString(pattern, str)
		if err != nil {
			continue
		}
		if match {
			return true
		}
	}

	return false
}
