package sensitive

/**
  @author: yhy
  @since: 2023/10/18
  @desc: //TODO
**/

// Detection 页面敏感信息检测
func Detection(url, req, body string) {
    go KeyDetection(url, body)
    go PageErrorMessageCheck(url, req, body)
    go Wih(url, req, body)
}
