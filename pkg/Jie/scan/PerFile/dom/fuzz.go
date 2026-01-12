package dom

var (
    fuzzPrefixes = []string{
        `javascript://alert({{RAND}})//`,
    }
    fuzzSuffixes = []string{
        `'-alert({{RAND}})-'`,
        `"-alert({{RAND}})-"`,
        `-alert({{RAND}})-`,
        `'"><img src=x onerror=alert({{RAND}})>`,
        `alert({{RAND}})`,
    }
)

//func genRand(s string) (string, string) {
//    r := random.RandomDigitString(5)
//    return strings.ReplaceAll(s, "{{RAND}}", r), r
//}

type FuzzUrl struct {
    Url  string `json:"url"`
    Rand string `json:"rand"`
}

//// GenPocUrls generates fuzz urls with payload
//func GenPocUrls(point VulPoint) ([]FuzzUrl, error) {
//    payloads := make([]FuzzUrl, 0)
//
//    u, err := url.Parse(point.Url)
//    if err != nil {
//        return nil, err
//    }
//
//    prefixURLs := mix.Payloads(*u, fuzzPrefixes, []mix.Rule{mix.RuleAppendPrefix, mix.RuleReplace}, mix.DefaultScopes)
//    for _, u := range prefixURLs {
//        furl, rand := genRand(u.String())
//        payloads = append(payloads, FuzzUrl{Url: furl, Rand: rand})
//    }
//
//    suffixURLs := mix.Payloads(*u, fuzzSuffixes, []mix.Rule{mix.RuleAppendSuffix, mix.RuleReplace}, mix.DefaultScopes)
//    for _, u := range suffixURLs {
//        furl, rand := genRand(u.String())
//        payloads = append(payloads, FuzzUrl{Url: furl, Rand: rand})
//    }
//
//    // TODO referrer
//    // if strings.Contains(point.Source.Label, "referrer") {
//    //     for _, suf := range fuzzSuffixes {
//    //         u, rand := genRand(fmt.Sprintf("%s&%s", point.Url, suf))
//    //         payloads = append(payloads, FuzzUrl{
//    //             Url:  u,
//    //             Rand: rand,
//    //         })
//    //     }
//    // }
//
//    return payloads, nil
//}
