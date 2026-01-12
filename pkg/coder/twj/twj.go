package twj

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt"
	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/logging"
)

/**
  @author: yhy
  @since: 2023/3/15
  @desc: //TODO
**/

// Claims defines the struct containing the token claims.
type Claims struct {
	jwt.StandardClaims
}

var Percentage chan float64

var Stop = false
var Flag = false

func init() {
	Percentage = make(chan float64, 1)
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

func GenerateSignature(jwtMsg string, jwtPath string) string {
	// 读取配置字典
	file.ReadJwtFile(jwtPath)
	Flag = true

	parts := strings.Split(jwtMsg, ".")
	if len(parts) != 3 {
		Flag = false
		return ""
	}

	// 解码JWT签名部分
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		Flag = false
		return ""
	}

	n := len(file.JwtSecrets)
	logging.Logger.Debugf("JWT brute force %s, load %d secrets", jwtMsg, n)

	var wg sync.WaitGroup
	ch := make(chan struct{}, 20)
	var mu sync.Mutex
	var res string

	for i, s := range file.JwtSecrets {
		if Stop {
			break
		}

		wg.Add(1)
		ch <- struct{}{}

		go func(i int, s string) {
			defer wg.Done()
			defer func() { <-ch }()

			float, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(i+1)/float64(n)*100), 64)
			Percentage <- float

			hasher := hmac.New(sha256.New, []byte(s))
			hasher.Write([]byte(parts[0] + "." + parts[1]))

			if bytes.Equal(signature, hasher.Sum(nil)) {
				mu.Lock()
				res = s
				Stop = true
				mu.Unlock()
				logging.Logger.Infof("[+]JWT brute Success %s, secret %d/%d: %s", jwtMsg, i+1, n, s)
			}
		}(i, s)
	}

	wg.Wait()
	Stop = false
	Flag = false
	return res
}
