package crypto

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func StoreAllKeys(keySets []*KeySet, partition string) {
	for i := 0; i < len(keySets); i++ {
		StoreKeyToFiles(keySets[i], int64(i), int64(len(keySets)), partition)
	}
}

func StoreKeyToFiles(keySet *KeySet, id int64, total int64, partition string) {
	path := GenPath(int64(id), total, partition)
	if !IsExist(path) {
		err := CreateDir(path)
		if err != nil {
			log.Fatal("[Store keySet Error] create file error!", err)
		}
		fmt.Println("creating " + path)
	}

	var fileName = "keySet"
	keySetBytes, _ := json.Marshal(keySet)
	err := ioutil.WriteFile(path+fileName, keySetBytes, 0644)
	if err != nil {
		log.Fatal("[Store keySet Error] write secretKey error!", err)
		return
	}

}

func GenPath(id int64, total int64, partition string) string {
	return fmt.Sprintf("./crypto/keys/keys_%s_%s%d/%d/", cryptoType, partition, total, id)
}

// Create Dir. Copying from storage to make sdcsm2 library independent
func CreateDir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	os.Chmod(path, os.ModePerm)
	return nil
}

// Determines whether folders exist. Copying from storage to make smcgo library independent
func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

func LoadKeyFromFiles(id int64, total int64, partition string) *KeySet {
	path := GenPath(id, total, partition)

	var fileName = "keySet"
	keySetBytes, err := ioutil.ReadFile(path + fileName)
	if err != nil {
		log.Fatalf("[Load keySet Error] read keySet from file for id %d error: %s", id, err)
		return nil
	}

	keySet := &KeySet{}
	// keySet.SecretKey, keySet.VerifyKeys, keySet.ThresholdPK = bls.UnmarshalStoredData(keySetBytes)
	err = json.Unmarshal(keySetBytes, keySet)
	if err != nil {
		log.Fatal("[Load keySet Error] Unmarshal falied", err)
	}

	return keySet
}

func HashForBls(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func Digest(object interface{}) string {
	msg, err := json.Marshal(object)

	if err != nil {
		return ""
	}

	return HashForBls(msg)
}
