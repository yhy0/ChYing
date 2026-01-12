package jwt

import (
    "bytes"
    "crypto/hmac"
    "crypto/sha256"
    "embed"
    "encoding/base64"
    "encoding/hex"
    "errors"
    "github.com/golang-jwt/jwt"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/util"
    "github.com/yhy0/sizedwaitgroup"
    "strings"
)

/**
  @author: yhy
  @since: 2023/3/15
  @desc: //TODO
**/

//go:embed secrets.txt
var jwtSecrets embed.FS

var Secrets []string

// Jwts 用来判断是否执行过爆破
var Jwts = make(map[string]bool)

func init() {
    f, err := jwtSecrets.ReadFile("secrets.txt")
    if err != nil {
        panic(err)
    }
    
    Secrets = strings.Split(string(f), "\n")
}

type Jwt struct {
    Header             string `json:"header"`
    Payload            string `json:"payload"`
    Message            string `json:"message"`
    signature, message []byte
    SignatureStr       string `json:"signature"`
}

// Claims defines the struct containing the token claims.
type Claims struct {
    jwt.StandardClaims
}

var Twj *Jwt

func ParseJWT(input string) (*Jwt, error) {
    parts := strings.Split(input, ".")
    decodedParts := make([][]byte, len(parts))
    if len(parts) != 3 {
        return nil, errors.New("invalid jwt: does not contain 3 parts (header, payload, signature)")
    }
    for i := range parts {
        decodedParts[i] = make([]byte, base64.RawURLEncoding.DecodedLen(len(parts[i])))
        if _, err := base64.RawURLEncoding.Decode(decodedParts[i], []byte(parts[i])); err != nil {
            return nil, err
        }
    }
    
    Twj = &Jwt{
        Header:       string(decodedParts[0]),
        Payload:      string(decodedParts[1]),
        signature:    decodedParts[2],
        message:      []byte(parts[0] + "." + parts[1]),
        SignatureStr: hex.EncodeToString(decodedParts[2]),
    }
    
    return Twj, nil
}

func Verify(jwtString string, secret string) (*Claims, error) {
    tokenClaims, err := jwt.ParseWithClaims(jwtString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })
    
    if tokenClaims != nil {
        if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
            return claims, nil
        }
    }
    
    return nil, err
}

// GenerateSignature 使用字典爆破 ，默认字典加上根据域名生成的字典
func GenerateSignature(pwds ...string) string {
    var res = ""
    wg := sizedwaitgroup.New(20)
    secrets := util.RemoveDuplicateElement(append(Secrets, pwds...))
    
    var stop = false
    for _, s := range secrets {
        if stop {
            return res
        }
        wg.Add()
        go func(s string) {
            defer wg.Done()
            hasher := hmac.New(sha256.New, []byte(s))
            hasher.Write(Twj.message)
            msg := hasher.Sum(nil)
            if bytes.Equal(Twj.signature, msg) {
                res = s
                stop = true
            }
            
        }(s)
    }
    
    wg.Wait()
    return res
}
