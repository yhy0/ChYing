package output

/**
  @author: yhy
  @since: 2023/8/2
  @desc: //TODO
**/

var ResultChan = make(chan Result)

// Result is a result structure returned by a source
type Result struct {
    Type   string
    Source string
    Value  string
}
