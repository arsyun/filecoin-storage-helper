package encypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"go-filecoin-storage-helper/utils"
	"os"
	"path/filepath"
)

func GenerateAesKeys(keypath string, k KeyInfo) error {
	kpath := filepath.Join(keypath, "keyinfo")
	fmt.Println("kpath:", kpath)
	if utils.Exists(kpath) {
		return nil
	}
	s, err := json.Marshal(k)
	fmt.Println("s:", string(s))
	if err != nil {
		return err
	}

	f, err := os.Create(kpath)
	if err != nil {
		return err
	}

	defer f.Close()
	if _, err = f.Write(s); err != nil {
		return err
	}
	return nil
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

//AES Encrypt
func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS7Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

//AES Decrypt
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS7UnPadding(origData)
	return origData, nil
}
