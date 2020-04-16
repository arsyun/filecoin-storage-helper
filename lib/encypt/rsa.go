package encypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"os"
	"path/filepath"
)

func GenerateRsaKeys(path string) error {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	xPrivatekey := x509.MarshalPKCS1PrivateKey(privateKey)
	// fp, _ := os.Create("private.pem")
	fp, _ := os.Create(filepath.Join(path, "private.pem"))
	defer fp.Close()
	pemblock := pem.Block{
		Type:  "privateKey",
		Bytes: xPrivatekey,
	}
	pem.Encode(fp, &pemblock)

	publickKey := privateKey.PublicKey
	xPublicKey, _ := x509.MarshalPKIXPublicKey(&publickKey)
	pemPublickKey := pem.Block{
		Type:  "PublicKey",
		Bytes: xPublicKey,
	}
	// file, _ := os.Create("PublicKey.pem")
	file, _ := os.Create(filepath.Join(path, "PublicKey.pem"))
	defer file.Close()
	if err := pem.Encode(file, &pemPublickKey); err != nil {
		return err
	}
	return nil
}

func RSA_encrypter(path string, msg []byte) (string, error) {
	fp, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer fp.Close()

	fileinfo, _ := fp.Stat()
	buf := make([]byte, fileinfo.Size())
	fp.Read(buf)

	block, _ := pem.Decode(buf)
	pub, _ := x509.ParsePKIXPublicKey(block.Bytes)
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, pub.(*rsa.PublicKey), msg)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(cipherText), nil
}

func RSA_decrypter(path string, cipherText []byte) ([]byte, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	fileinfo, _ := fp.Stat()
	buf := make([]byte, fileinfo.Size())
	fp.Read(buf)

	block, _ := pem.Decode(buf)
	PrivateKey, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
	afterDecrypter, err := rsa.DecryptPKCS1v15(rand.Reader, PrivateKey, cipherText)
	if err != nil {
		return nil, err
	}

	return afterDecrypter, nil
}
