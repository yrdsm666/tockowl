package tpke

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/share"
)

// EncKeyGen return tpkes
func EncKeyGen(n uint32, t uint32) (kyber.Point, []*share.PubShare, []*share.PriShare) {
	suite := pairing.NewSuiteBn256()
	random := suite.RandomStream()

	x := suite.G2().Scalar().Pick(random)

	pripoly := share.NewPriPoly(suite.G2(), int(t), x, suite.RandomStream())
	sks := pripoly.Shares(int(n))

	pubpoly := pripoly.Commit(suite.G2().Point().Base())
	vk := pubpoly.Shares(int(n))

	pk := suite.G2().Point().Mul(x, suite.G2().Point().Base())

	return pk, vk, sks
}

func MarshalEncSK(encSk *share.PriShare) ([]byte, error) {
	skStr := make([]string, 2)
	skStr[0] = strconv.Itoa(encSk.I)
	byts, err := encSk.V.MarshalBinary()
	if err != nil {
		return nil, errors.Wrapf(err, "fail to marshal TSSconfig.sk.V")
	}
	skStr[1] = base64.StdEncoding.EncodeToString(byts)
	sk, err := json.Marshal(skStr)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to marshal skStr")
	}
	return sk, nil
}

func UnmarshalEncSK(sk []byte) (*share.PriShare, error) {
	suit := pairing.NewSuiteBn256()
	skStr := make([]string, 2)
	err := json.Unmarshal(sk, &skStr)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to unmarshal sk")
	}

	index, err := strconv.Atoi(skStr[0])
	if err != nil {
		return nil, errors.Wrap(err, "fail to unmarshal sk.i")
	}
	byts, err := base64.StdEncoding.DecodeString(skStr[1])
	if err != nil {
		return nil, errors.Wrap(err, "fail to unmarshal sk.v")
	}
	v := suit.G2().Scalar()
	err = v.UnmarshalBinary(byts)
	if err != nil {
		return nil, errors.Wrap(err, "fail to unmarshal sk.v")
	}
	encSk := &share.PriShare{
		I: index,
		V: v,
	}
	return encSk, nil
}

func MarshalEncPK(encPk kyber.Point) ([]byte, error) {
	pk, err := encPk.MarshalBinary()
	if err != nil {
		return nil, errors.Wrapf(err, "fail to marshal TSSconfig.pk")
	}
	return pk, nil
}

func UnmarshalEncPK(pk []byte) (kyber.Point, error) {
	suit := pairing.NewSuiteBn256()
	encPk := suit.G2().Point()
	err := encPk.UnmarshalBinary(pk)
	if err != nil {
		return nil, errors.Wrap(err, "fail to unmarshal pk")
	}
	return encPk, nil
}

func MarshalEncVK(encVK *share.PubShare) ([]byte, error) {
	vkStr := make([]string, 2)
	vkStr[0] = strconv.Itoa(encVK.I)
	byts, err := encVK.V.MarshalBinary()
	if err != nil {
		return nil, errors.Wrapf(err, "fail to marshal TSSconfig.sk.V")
	}
	vkStr[1] = base64.StdEncoding.EncodeToString(byts)
	vk, err := json.Marshal(vkStr)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to marshal skStr")
	}
	return vk, nil
}

func UnmarshalEncVK(vk []byte) (*share.PubShare, error) {
	suit := pairing.NewSuiteBn256()
	vkStr := make([]string, 2)
	err := json.Unmarshal(vk, &vkStr)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to unmarshal vk")
	}
	index, err := strconv.Atoi(vkStr[0])
	if err != nil {
		return nil, errors.Wrapf(err, "fail to unmarshal vk.i")
	}

	byts, err := base64.StdEncoding.DecodeString(vkStr[1])
	if err != nil {
		return nil, errors.Wrapf(err, "fail to unmarshal vk.v")
	}
	v := suit.G2().Point()
	err = v.UnmarshalBinary(byts)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to unmarshal vk.v")
	}
	encVK := &share.PubShare{
		I: index,
		V: v,
	}
	return encVK, nil
}
