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
	"time"
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
	if Twj == nil {
		return ""
	}

	n := len(file.JwtSecrets)
	for i, s := range file.JwtSecrets {
		Percentage <- fmt.Sprintf("%.2f", float64(i+1)/float64(n)*100)
		hasher := hmac.New(sha256.New, []byte(s))
		hasher.Write(Twj.message)
		msg := hasher.Sum(nil)
		if bytes.Equal(Twj.signature, msg) {
			return s
		}

		time.Sleep(1 * time.Second)

	}
	return ""
}
