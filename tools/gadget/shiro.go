package gadget

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	_ "embed"
	"encoding/base64"
	"errors"
	"github.com/phith0n/zkar/serz"
	"github.com/yhy0/logging"
	"strings"
)

/**
   @author yhy
   @since 2023/6/9
   @desc https://github.com/chuanjiesun/shiro_rememberMe_decrypt shiro 解密
	https://github.com/phith0n/zkar
**/

//go:embed keys.txt
var keys1 string

type Shiro struct {
	Key           string `json:"key"`
	IV            string `json:"iv"`
	Type          string `json:"type"`
	Decrypt       string `json:"decrypt"`
	DecryptB64    string `json:"decryptB64"`
	Serialization string `json:"serialization"`
}

var keys []string

func init() {
	if "" != keys1 {
		keys1 = strings.TrimSpace(keys1)
		keys = strings.Split(keys1, "\n")
	} else {
		logging.Logger.Errorln("Warning, unable to load into dicts/keys.txt")
	}
}

func DecryptShiro(b64key, b64encrypt_str string) (*Shiro, error) {
	var _keys []string
	if b64key == "" {
		_keys = keys
	} else {
		_keys = append(_keys, b64key)
	}

	for _, key := range _keys {
		cipher_key, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			continue
		}
		cipher_data, err := base64.StdEncoding.DecodeString(b64encrypt_str)
		if err != nil {
			continue
		}

		err, decrypt_data := AesCBCDecrypt(cipher_data, cipher_key)
		if err == nil {
			iv_data_b64 := base64.StdEncoding.EncodeToString(decrypt_data[:16])
			decrypt_data_b64 := base64.StdEncoding.EncodeToString(decrypt_data[16:])
			shiro := &Shiro{
				Key:        key,
				IV:         iv_data_b64,
				Type:       "CBC",
				Decrypt:    string(decrypt_data),
				DecryptB64: decrypt_data_b64,
			}
			serialization, err := serz.FromBytes(decrypt_data[16 : len(decrypt_data)-2])
			if err == nil {
				shiro.Serialization = serialization.ToString()
			}
			return shiro, nil
		}

		err, decrypt_data = AesGCMDecrypt(cipher_data, cipher_key)

		if err == nil {
			iv_data_b64 := base64.StdEncoding.EncodeToString(decrypt_data[:16])
			decrypt_data_b64 := base64.StdEncoding.EncodeToString(decrypt_data[16:])

			shiro := &Shiro{
				Key:        key,
				IV:         iv_data_b64,
				Type:       "GCM",
				Decrypt:    string(decrypt_data),
				DecryptB64: decrypt_data_b64,
			}
			serialization, err := serz.FromBytes(decrypt_data[16 : len(decrypt_data)-2])
			if err == nil {
				shiro.Serialization = serialization.ToString()
			}
			return shiro, nil
		}
	}

	return nil, errors.New("not find")
}

func AesCBCDecrypt(encrypted []byte, key []byte) (err error, decrypted_data []byte) {
	block, err := aes.NewCipher(key) // 分组秘钥
	if err != nil {
		return err, []byte("")
	}
	blockSize := block.BlockSize() // 获取秘钥块的长度
	if len(encrypted)%blockSize != 0 {
		return errors.New("input not full blocks"), []byte("")
	}
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize]) // 加密模式
	decrypted_data = make([]byte, len(encrypted))               // 创建数组
	blockMode.CryptBlocks(decrypted_data, encrypted)            // 解密
	decrypted_data = pkcs5UnPadding(decrypted_data)             // 去除补全码

	//java 序列化开头 aced00057372
	if len(decrypted_data) < 22 || bytes.Compare(decrypted_data[16:22], []byte{0xac, 0xed, 0x00, 0x05, 0x73, 0x72}) != 0 {
		return errors.New("maybe not valid java serilized data"), []byte("")
	}
	return nil, decrypted_data
}

func AesGCMDecrypt(encrypted []byte, key []byte) (err error, decrypted_data []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err, nil
	}

	nonceSize := 16
	if len(encrypted) < nonceSize {
		return errors.New("ciphertext too short"), nil
	}

	nonce := encrypted[:nonceSize]
	ciphertext := encrypted[nonceSize:]

	aesgcm, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return err, nil
	}

	decrypted, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err, nil
	}

	decrypted_data = pkcs5UnPadding(decrypted) // 去除补全码

	//java 序列化开头 aced00057372
	if len(decrypted_data) < 22 || bytes.Compare(decrypted_data[16:22], []byte{0xac, 0xed, 0x00, 0x05, 0x73, 0x72}) != 0 {
		return errors.New("maybe not valid java serilized data"), []byte("")
	}
	return nil, decrypted_data
}

func pkcs5UnPadding(origData []byte) []byte {
	if len(origData) == 0 {
		return origData
	}
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
