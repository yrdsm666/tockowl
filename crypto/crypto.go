package crypto

import (
	"fmt"

	"github.com/yrdsm666/tockowl/crypto/kyber_bls"
	"github.com/yrdsm666/tockowl/crypto/tpke"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/share"
)

//	type KeySet struct {
//		SecretKey   *math.Zr
//		VerifyKeys  map[int]*math.G2
//		ThresholdPK *math.G2
//	}
type KeySet struct {
	SecretKey   []byte
	VerifyKeys  map[int][]byte
	ThresholdPK []byte

	EncPK []byte   //tse pk
	EncVK [][]byte //tse vk
	EncSK []byte   //tse sk
}

var cryptoType string

var N int
var T int

// const cryptoType1 string = "baseBls"
// const cryptoType2 string = "cBls"
const cryptoType3 string = "tbls"

// const cryptoType3 string = "kyberBls"

func CryptoInit(nowCryptoType string, n int, t int) {
	cryptoType = nowCryptoType
	if cryptoType == cryptoType3 {
		N = n
		T = t
	}
}

func KeyGen(n, k int) []*KeySet {
	resKeys := make([]*KeySet, n)
	if cryptoType == cryptoType3 {
		sk, pk := kyber_bls.SigKeyGen(n, k)
		verifyKeys := make(map[int][]byte)
		for i := 0; i < len(sk); i++ {
			verifyKeys[i] = kyber_bls.MarshalPubPoly(pk)
		}
		for i := 0; i < len(sk); i++ {
			keySet := &KeySet{}
			keySet.SecretKey = kyber_bls.MarshalPriShare(sk[i])
			keySet.VerifyKeys = verifyKeys
			keySet.ThresholdPK = kyber_bls.MarshalPubPoly(pk)
			resKeys[i] = keySet
		}
	} else {
		panic("unsupport cryptoType")
	}

	epk, evk, esks := tpke.EncKeyGen(uint32(n), uint32(n-k+1))
	encPk, err := tpke.MarshalEncPK(epk)
	if err != nil {
		panic(err)
	}

	evks := make([][]byte, n)
	for i := 0; i < n; i++ {
		evks[i], err = tpke.MarshalEncVK(evk[i])
		if err != nil {
			panic(err)
		}
	}
	for i := 0; i < n; i++ {
		resKeys[i].EncPK = encPk
		resKeys[i].EncVK = evks
		resKeys[i].EncSK, err = tpke.MarshalEncSK(esks[i])
		if err != nil {
			panic(err)
		}
	}

	return resKeys
}

func ThresholdSign(sk []byte, msg []byte) []byte {
	if cryptoType == cryptoType3 {
		priShare := kyber_bls.UnmarshalPriShare(sk)
		return kyber_bls.Sign(priShare, msg)
	} else {
		panic("unsupport cryptoType")
	}
}

func VerifyShare(pk []byte, msg []byte, sig []byte) error {
	if cryptoType == cryptoType3 {
		pubPoly := kyber_bls.UnmarshalPubPoly(pk)
		return kyber_bls.VerifyShare(pubPoly, msg, sig)
	} else {
		panic("unsupport cryptoType")
	}
}

func VerifySignature(pk []byte, msg []byte, sig []byte) error {
	if cryptoType == cryptoType3 {
		pubPoly := kyber_bls.UnmarshalPubPoly(pk)
		return kyber_bls.VerifySignature(pubPoly, msg, sig)
		// return nil
	} else {
		panic("unsupport cryptoType")
	}
}

func CombineSignatures(signatures [][]byte, evaluationPoints []int64) ([]byte, error) {
	if cryptoType == cryptoType3 {
		return kyber_bls.Recover(signatures, N, T)
	} else {
		panic("unsupport cryptoType")
	}
}

func ThresholdEncrypt(pk []byte, msg []byte) ([]byte, error) {
	encPk, err := tpke.UnmarshalEncPK(pk)
	if err != nil {
		panic(err)
	}
	return tpke.Encrypt(pairing.NewSuiteBn256(), encPk, msg)
}

func ThresholdDecShare(sk []byte, ciphertext []byte) ([]byte, error) {
	encSk, err := tpke.UnmarshalEncSK(sk)
	if err != nil {
		panic(err)
	}
	return tpke.DecShare(pairing.NewSuiteBn256(), encSk, ciphertext)
}

func ThresholdVerifyDecShare(vk [][]byte, ciphertext []byte, decShare []byte) error {
	encVk := make([]*share.PubShare, len(vk))
	var err error
	for i := 0; i < len(vk); i++ {
		encVk[i], err = tpke.UnmarshalEncVK(vk[i])
		if err != nil {
			panic(err)
		}
	}
	if tpke.VerifyDecShare(pairing.NewSuiteBn256(), encVk, ciphertext, decShare) {
		return nil
	} else {
		return fmt.Errorf("DecShare is invalid")
	}
}

func ThresholdDecrypt(vk [][]byte, ciphertext []byte, decShares [][]byte, t int, n int) ([]byte, error) {
	encVk := make([]*share.PubShare, n)
	var err error
	for i := 0; i < n; i++ {
		encVk[i], err = tpke.UnmarshalEncVK(vk[i])
		if err != nil {
			panic(err)
		}
	}
	return tpke.Decrypt(pairing.NewSuiteBn256(), encVk, ciphertext, decShares, t, n)
}
