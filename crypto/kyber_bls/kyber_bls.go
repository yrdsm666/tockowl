package kyber_bls

import (
	"encoding/json"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bls"
	"go.dedis.ch/kyber/v3/sign/tbls"
)

type PriShare struct {
	I int    // Index of the private share
	V []byte // Value of the private share
}

type PubPoly struct {
	B       []byte   // Base point, nil for standard base
	Commits [][]byte // Commitments to coefficients of the secret sharing polynomial
}

// SigKeyGen return pk and sks, n is the number of parties, t is the threshold of combining signature
func SigKeyGen(n int, t int) ([]*share.PriShare, *share.PubPoly) {
	suit := pairing.NewSuiteBn256()
	random := suit.RandomStream()

	x := suit.G1().Scalar().Pick(random)

	// pripoly
	pripoly := share.NewPriPoly(suit.G2(), t, x, suit.RandomStream())
	// n points in poly
	npoints := pripoly.Shares(n)
	//pub poly
	pubpoly := pripoly.Commit(suit.G2().Point().Base())
	return npoints, pubpoly
}

func Sign(sk *share.PriShare, msg []byte) []byte {
	sigShare, err := tbls.Sign(pairing.NewSuiteBn256(), sk, msg)
	if err != nil {
		panic(err)
	}
	return sigShare
}

// func Recover(pk *share.PubPoly, msg []byte, sigs [][]byte, n int, t int) ([]byte, error) {
// 	return tbls.Recover(pairing.NewSuiteBn256(), pk, msg, sigs, t, n)
// }

func Recover(sigs [][]byte, n int, t int) ([]byte, error) {
	suite := pairing.NewSuiteBn256()
	pubShares := make([]*share.PubShare, 0)
	for _, sig := range sigs {
		s := tbls.SigShare(sig)
		i, err := s.Index()
		if err != nil {
			return nil, err
		}

		point := suite.G1().Point()
		if err := point.UnmarshalBinary(s.Value()); err != nil {
			return nil, err
		}
		pubShares = append(pubShares, &share.PubShare{I: i, V: point})
		if len(pubShares) >= t {
			break
		}
	}
	commit, err := share.RecoverCommit(suite.G1(), pubShares, t, n)
	if err != nil {
		return nil, err
	}
	sig, err := commit.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func VerifyShare(pk *share.PubPoly, msg []byte, sig []byte) error {
	s := tbls.SigShare(sig)
	i, err := s.Index()
	if err != nil {
		return err
	}

	if err := bls.Verify(pairing.NewSuiteBn256(), pk.Eval(i).V, msg, s.Value()); err != nil {
		return err
	}
	return nil
}

func VerifySignature(pk *share.PubPoly, msg []byte, sig []byte) error {
	return bls.Verify(pairing.NewSuiteBn256(), pk.Commit(), msg, sig)
}

func MarshalPriShare(ps *share.PriShare) []byte {
	scalerBytes, err := ps.V.MarshalBinary()
	if err != nil {
		panic(err)
	}

	priShare := PriShare{
		I: ps.I,
		V: scalerBytes,
	}

	priShareBytes, err := json.Marshal(priShare)
	if err != nil {
		panic(err)
	}
	return priShareBytes
}

func UnmarshalPriShare(psBytes []byte) *share.PriShare {
	priShare := new(PriShare)
	err := json.Unmarshal(psBytes, priShare)
	if err != nil {
		panic(err)
	}

	suit := pairing.NewSuiteBn256()
	ps := &share.PriShare{
		I: priShare.I,
		V: suit.G1().Scalar(),
	}

	// ps.V.SetBytes(priShare.V)
	ps.V.UnmarshalBinary(priShare.V)
	if err != nil {
		panic(err)
	}

	return ps
}

func MarshalPubPoly(ps *share.PubPoly) []byte {
	base, commits := ps.Info()
	commitsBytes := make([][]byte, len(commits))
	for i := 0; i < len(commits); i++ {
		commitsBytes[i] = MarshalPoint(commits[i])
	}

	pubPoly := PubPoly{
		B:       MarshalPoint(base),
		Commits: commitsBytes,
	}

	pubPolyBytes, err := json.Marshal(pubPoly)
	if err != nil {
		panic(err)
	}
	return pubPolyBytes
}

func UnmarshalPubPoly(pubPolyBytes []byte) *share.PubPoly {
	pubPoly := new(PubPoly)
	err := json.Unmarshal(pubPolyBytes, pubPoly)
	if err != nil {
		panic(err)
	}

	commits := make([]kyber.Point, len(pubPoly.Commits))
	for i := 0; i < len(commits); i++ {
		commits[i] = UnmarshalPoint(pubPoly.Commits[i])
	}
	b := UnmarshalPoint(pubPoly.B)

	suit := pairing.NewSuiteBn256()
	g2 := suit.G2()
	ps := share.NewPubPoly(g2, b, commits)

	return ps
}

func MarshalPoint(p kyber.Point) []byte {
	pBytes, err := p.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return pBytes
}

func UnmarshalPoint(pBytes []byte) kyber.Point {
	suit := pairing.NewSuiteBn256()
	p := suit.G2().Point()
	p.UnmarshalBinary(pBytes)
	return p
}

func RecoverPubPolyFromPriShares(pss []*share.PriShare, n int, k int) *share.PubPoly {
	suit := pairing.NewSuiteBn256()
	g2 := suit.G2()
	priPoly, _ := share.RecoverPriPoly(g2, pss, k, n)
	pubPoly := priPoly.Commit(suit.G2().Point().Base())
	return pubPoly
}
