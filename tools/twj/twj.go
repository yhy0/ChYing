package twj

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/yhy0/ChYing/pkg/file"
	"strings"
)

/**
  @author: yhy
  @since: 2023/3/15
  @desc: //TODO
**/

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

var Percentage chan string
var Twj *Jwt
var Stop = false
var Flag = false

func init() {
	Percentage = make(chan string, 1)
}

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

func GenerateSignature() string {
	// 每次扫描前都重新读取一遍配置字典
	file.ReadFiles()
	if Twj == nil {
		return ""
	}
	// 代表有进程在爆破
	Flag = true
	var res = ""
	ch := make(chan struct{}, 20)

	n := len(file.JwtSecrets)
	for i, s := range file.JwtSecrets {
		if Stop {
			Stop = false
			Flag = false
			return res
		}

		ch <- struct{}{}
		go func(i int, s string) {
			Percentage <- fmt.Sprintf("%.2f", float64(i+1)/float64(n)*100)
			hasher := hmac.New(sha256.New, []byte(s))
			hasher.Write(Twj.message)
			msg := hasher.Sum(nil)
			if bytes.Equal(Twj.signature, msg) {
				res = s
				Stop = true
			}
			<-ch
		}(i, s)

	}
	close(ch)
	Stop = false
	Flag = false
	return res
}
