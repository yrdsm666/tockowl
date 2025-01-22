package crypto

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestThresholdSignatureForTbls(t *testing.T) {
	var executeTime time.Time
	var executeDuration time.Duration
	var sig []byte
	var signatures [][]byte
	var idArr []int64

	n := 4
	k := 3

	CryptoInit("tbls", n, k)

	signatures = make([][]byte, 0)
	keys := KeyGen(n, k)
	StoreAllKeys(keys, "")
	message := []byte("4563686f00000001000000008a50a53bbc5414e44e6578455f6ea739860a73bf47a4c205336f8cc1f828f213d6eae439a3a7c30793142689d1c2c9f127b516e8b14e9e7793bc7dbdef90680d")
	keySets := make([]*KeySet, n)
	for id := 0; id < n; id++ {
		keySets[id] = LoadKeyFromFiles(int64(id), int64(n), "")
	}

	// -------------------------------------
	for id := 1; id <= n; id++ {
		// Each node signs the message and obtains a share
		executeTime = time.Now()
		sig = ThresholdSign(keySets[id-1].SecretKey, message)
		signatures = append(signatures, sig)
		executeDuration = time.Since(executeTime)
		fmt.Println("ThresholdSign:", executeDuration)

		// Verify the validity of the share
		executeTime = time.Now()
		assert.NoError(t, VerifyShare(keySets[0].VerifyKeys[id-1], message, signatures[id-1]))
		executeDuration = time.Since(executeTime)
		fmt.Println("VerifySignature:", executeDuration)
	}

	// -------------------------------------
	// Combine shares to get the signature: sig1
	executeTime = time.Now()
	partSigs := make([][]byte, 0)
	idArr = []int64{3, 2, 1}
	for i := 0; i < len(idArr); i++ {
		partSigs = append(partSigs, signatures[idArr[i]-1])
	}
	fullSig1, err := CombineSignatures(partSigs, idArr)
	assert.NoError(t, err)
	executeDuration = time.Since(executeTime)
	fmt.Println("CombineSignatures:", executeDuration)

	// Verify the validity of the signature
	executeTime = time.Now()
	assert.NoError(t, VerifySignature(keySets[0].ThresholdPK, message, fullSig1))
	executeDuration = time.Since(executeTime)
	fmt.Println("SignatureVerify:", executeDuration)

	// -------------------------------------
	// Combine shares to get the signature: sig2
	executeTime = time.Now()
	partSigs = make([][]byte, 0)
	idArr = []int64{2, 4, 1}
	for i := 0; i < len(idArr); i++ {
		partSigs = append(partSigs, signatures[idArr[i]-1])
	}
	fullSig2, err := CombineSignatures(partSigs, idArr)
	assert.NoError(t, err)
	executeDuration = time.Since(executeTime)
	fmt.Println("CombineSignatures:", executeDuration)

	// Verify the validity of the signature
	executeTime = time.Now()
	assert.NoError(t, VerifySignature(keySets[0].ThresholdPK, message, fullSig2))
	executeDuration = time.Since(executeTime)
	fmt.Println("SignatureVerify:", executeDuration)

	// -------------------------------------
	// Verify whether two signatures are equal
	assert.Equal(t, fullSig1, fullSig2)
	fmt.Println("fullSig:", hex.EncodeToString(fullSig1))
}

func TestThresholdEncrypt(t *testing.T) {
	n := 4
	k := 3

	CryptoInit("tbls", n, k)
	keys := KeyGen(n, k)
	StoreAllKeys(keys, "")
	message := []byte("the little fox jumps over the lazy dog")
	keySets := make([]*KeySet, n)
	for id := 0; id < n; id++ {
		keySets[id] = LoadKeyFromFiles(int64(id), int64(n), "")
	}

	ciphertext, err := ThresholdEncrypt(keySets[0].EncPK, message)
	assert.NoError(t, err)

	decShare0, _ := ThresholdDecShare(keySets[0].EncSK, ciphertext)
	decShare1, _ := ThresholdDecShare(keySets[1].EncSK, ciphertext)
	decShare2, _ := ThresholdDecShare(keySets[2].EncSK, ciphertext)
	decShare3, _ := ThresholdDecShare(keySets[3].EncSK, ciphertext)

	assert.NoError(t, ThresholdVerifyDecShare(keySets[0].EncVK, ciphertext, decShare0))
	assert.NoError(t, ThresholdVerifyDecShare(keySets[2].EncVK, ciphertext, decShare1))
	assert.NoError(t, ThresholdVerifyDecShare(keySets[1].EncVK, ciphertext, decShare2))
	assert.NoError(t, ThresholdVerifyDecShare(keySets[0].EncVK, ciphertext, decShare3))

	decShares1 := [][]byte{}
	decShares1 = append(decShares1, decShare0)
	decShares1 = append(decShares1, decShare1)
	decShares1 = append(decShares1, decShare2)
	msg1, err := ThresholdDecrypt(keySets[0].EncVK, ciphertext, decShares1, k, n)
	assert.NoError(t, err)

	decShares2 := [][]byte{}
	decShares2 = append(decShares2, decShare1)
	decShares2 = append(decShares2, decShare2)
	decShares2 = append(decShares2, decShare3)
	msg2, err := ThresholdDecrypt(keySets[0].EncVK, ciphertext, decShares2, k, n)
	assert.NoError(t, err)

	assert.Equal(t, message, msg2)
	assert.Equal(t, msg1, msg2)
}
