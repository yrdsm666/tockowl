package kyber_bls

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSign(t *testing.T) {
	n := 128
	k := 86

	sks, pk := SigKeyGen(n, k)

	msg := []byte("4563686f00000001000000008a50a53bbc5414e44e6578455f6ea739860a73bf47a4c205336f8cc1f828f213d6eae439a3a7c30793142689d1c2c9f127b516e8b14e9e7793bc7dbdef90680d")

	sigshare := [][]byte{}
	for i := 0; i < k; i++ {
		sigShare := Sign(sks[i], msg)
		sigshare = append(sigshare, sigShare)
		assert.NoError(t, VerifyShare(pk, msg, sigShare))
	}
	signature, _ := Recover(sigshare, n, k)

	err := VerifySignature(pk, msg, signature)

	assert.NoError(t, err)
}

func TestMarshalPriShare(t *testing.T) {
	n := 4
	k := 3

	sks, _ := SigKeyGen(n, k)
	sk := sks[0]

	skBytes := MarshalPriShare(sk)

	newSk := UnmarshalPriShare(skBytes)

	assert.Equal(t, newSk, sk)
}

func TestMarshalPubPoly(t *testing.T) {
	n := 4
	k := 3

	_, pk := SigKeyGen(n, k)
	base1, commits1 := pk.Info()

	pkBytes := MarshalPubPoly(pk)

	newPk := UnmarshalPubPoly(pkBytes)
	base2, commits2 := newPk.Info()

	assert.Equal(t, base1, base2)
	assert.Equal(t, len(commits1), len(commits2))

	for i := 0; i < len(commits1); i++ {
		assert.Equal(t, true, commits1[i].Equal(commits2[i]))
	}

}

func TestMarshalPoint(t *testing.T) {
	n := 4
	k := 3

	_, pk := SigKeyGen(n, k)
	base, _ := pk.Info()

	baseBytes := MarshalPoint(base)

	newBase := UnmarshalPoint(baseBytes)
	fmt.Println("pk: ", base, "\nBase: ", newBase)

	assert.Equal(t, newBase, base)
}

func TestGetPubPoly(t *testing.T) {
	n := 4
	k := 3

	sks, pk := SigKeyGen(n, k)

	newPk := RecoverPubPolyFromPriShares(sks, n, k)

	assert.Equal(t, newPk, pk)
}
