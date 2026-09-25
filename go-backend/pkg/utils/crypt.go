package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

// CryptLib implements AES-256-CBC encryption compatible with C# CryptLib
type CryptLib struct{}

func NewCryptLib() *CryptLib {
	return &CryptLib{}
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("empty data for unpadding")
	}
	padding := int(data[length-1])
	if padding > length || padding == 0 {
		return nil, errors.New("invalid padding")
	}
	for i := length - padding; i < length; i++ {
		if data[i] != byte(padding) {
			return nil, errors.New("invalid padding byte")
		}
	}
	return data[:length-padding], nil
}

func (c *CryptLib) prepareKeyAndIV(keyStr, ivStr string) ([]byte, []byte) {
	key := make([]byte, 32)
	iv := make([]byte, 16)

	pwd := []byte(keyStr)
	ivBytes := []byte(ivStr)

	copyLen := len(pwd)
	if copyLen > 32 {
		copyLen = 32
	}
	copy(key, pwd[:copyLen])

	ivLen := len(ivBytes)
	if ivLen > 16 {
		ivLen = 16
	}
	copy(iv, ivBytes[:ivLen])

	return key, iv
}

// Encrypt encrypts plaintext using AES-256-CBC matching C# CryptLib
func (c *CryptLib) Encrypt(plainText, keyStr, ivStr string) (string, error) {
	key, iv := c.prepareKeyAndIV(keyStr, ivStr)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	padded := pkcs7Pad([]byte(plainText), block.BlockSize())
	ciphertext := make([]byte, len(padded))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64 encoded ciphertext using AES-256-CBC matching C# CryptLib
func (c *CryptLib) Decrypt(cipherText, keyStr, ivStr string) (string, error) {
	key, iv := c.prepareKeyAndIV(keyStr, ivStr)

	ciphertext, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(ciphertext)%block.BlockSize() != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	unpadded, err := pkcs7Unpad(plaintext)
	if err != nil {
		return "", err
	}

	return string(unpadded), nil
}
