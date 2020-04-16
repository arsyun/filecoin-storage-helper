package encypt

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"golang.org/x/xerrors"
)

const (
	AES = "AES"
	RSA = "RSA"
)

type KeyInfo struct {
	Encypt string `json:"encypt"`
	Key    string `json:"key"`
}

//TODO: support for more types of encryption
func GenerateKeyInfo(repo string, typ string, key string) error {
	switch typ {
	case AES:
		k := KeyInfo{
			Encypt: AES,
			Key:    key,
		}
		return GenerateAesKeys(repo, k)
	case RSA:
		return GenerateRsaKeys(repo)
	default:
		return xerrors.Errorf("invalid encypt type %s", typ)
	}
}

func GetKeysbyType(typ string, repo string) (string, error) {
	switch typ {
	case AES:
		keypath := filepath.Join(repo, "keyinfo")

		data, err := ioutil.ReadFile(keypath)
		if err != nil {
			return "", err
		}
		k := KeyInfo{}
		if err = json.Unmarshal(data, &k); err != nil {
			return "", err
		}
		return k.Key, nil
	case RSA:
		return filepath.Join(repo, "private.pem"), nil
	default:
		return "", xerrors.Errorf("invalid encypt type %s", typ)
	}
}

func Encyptdata(typ string, data string, key string) (string, error) {
	switch typ {
	case AES:
		d, err := AesEncrypt([]byte(data), []byte(key))
		if err != nil {
			return "", err
		}

		return base64.StdEncoding.EncodeToString(d), nil
	case RSA:
		return RSA_encrypter("", []byte(data))
	default:
		return data, nil
	}
}

func Decyptdata(typ string, data string, key string) ([]byte, error) {
	switch typ {
	case AES:
		d, _ := base64.StdEncoding.DecodeString(data)
		return AesDecrypt(d, []byte(key))
	case RSA:
		return RSA_decrypter("", []byte(data))
	default:
		return nil, nil
	}
}
