package twj

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
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

// Sign 使用指定算法和密钥签名 JWT，返回完整的 JWT 字符串
// headerJson: JSON 格式的 header
// payloadJson: JSON 格式的 payload
// secret: 签名密钥
// algorithm: 签名算法 (HS256, HS384, HS512)，如果为空则使用 header 中的 alg
func Sign(headerJson string, payloadJson string, secret string, algorithm string) (string, error) {
	// 验证 JSON 格式
	var headerMap map[string]interface{}
	var payloadMap map[string]interface{}

	if err := json.Unmarshal([]byte(headerJson), &headerMap); err != nil {
		return "", fmt.Errorf("invalid header JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(payloadJson), &payloadMap); err != nil {
		return "", fmt.Errorf("invalid payload JSON: %v", err)
	}

	// 确定使用的算法：优先使用传入的 algorithm，否则使用 header 中的 alg
	alg := algorithm
	if alg == "" {
		if headerAlg, ok := headerMap["alg"].(string); ok && headerAlg != "" {
			alg = headerAlg
		} else {
			alg = "HS256" // 默认算法
		}
	}

	// 更新 header 中的算法（确保一致）
	headerMap["alg"] = alg
	if _, ok := headerMap["typ"]; !ok {
		headerMap["typ"] = "JWT"
	}

	// 重新序列化 header（确保算法一致）
	headerBytes, err := json.Marshal(headerMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %v", err)
	}

	// 序列化 payload
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Base64URL 编码
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// 生成签名
	signingInput := headerEncoded + "." + payloadEncoded
	signature, err := signWithAlgorithm(signingInput, secret, alg)
	if err != nil {
		return "", err
	}

	// 返回完整的 JWT
	return signingInput + "." + signature, nil
}

// signWithAlgorithm 根据算法生成签名
func signWithAlgorithm(signingInput string, secret string, algorithm string) (string, error) {
	var h hash.Hash

	switch algorithm {
	case "HS256":
		h = hmac.New(sha256.New, []byte(secret))
	case "HS384":
		h = hmac.New(sha512.New384, []byte(secret))
	case "HS512":
		h = hmac.New(sha512.New, []byte(secret))
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	h.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
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
