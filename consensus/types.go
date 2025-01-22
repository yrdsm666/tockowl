package consensus

import (
	"encoding/json"

	"github.com/yrdsm666/tockowl/crypto"
	"github.com/yrdsm666/tockowl/types"
	"github.com/yrdsm666/tockowl/utils"
)

type TimeVoteInner struct {
	Hash types.TxHash `json:"hash"`
	Time int64        `json:"time"`
	ID   int64        `json:"id"`
}

// func (tvi *TimeVoteInner) GenHash(ks *config.KeySet) []byte {
// 	btvi, err := json.Marshal(tvi)
// 	utils.PanicOnError(err)
// 	h, err := crypto.CreateDocumentHash(btvi, ks.PublicKey)
// 	utils.PanicOnError(err)
// 	return h
// }

// func (tvi *TimeVoteInner) Sign(ks *config.KeySet) *TimeVote {
func (tvi *TimeVoteInner) Sign(ks *crypto.KeySet) *TimeVote {
	marshal, _ := json.Marshal(tvi)
	sig := crypto.ThresholdSign(ks.SecretKey, marshal)

	// h := tvi.GenHash(ks)
	// sig, err := crypto.TSign(h, ks.PrivateKey, ks.PublicKey)
	// utils.PanicOnError(err)
	return &TimeVote{
		TVI: tvi,
		Sig: sig,
	}
}

type TimeVote struct {
	TVI *TimeVoteInner `json:"tvi"`
	// Sig *tcrsa.SigShare `json:"sig"`
	Sig []byte `json:"sig"`
}

// func (tv *TimeVote) Verify(ks *config.KeySet) bool {
func (tv *TimeVote) Verify(ks *crypto.KeySet) bool {
	// marshal, _ := json.Marshal(tv)
	// err := crypto.VerifySignature(ks.VerifyKeys[senderId], marshal, tv.Sig)
	// if err != nil {
	// 	return false
	// }
	return true

}

type Lock struct {
	Block []byte `json:"block"`
	Proof []byte `json:"proof"`
	Cid   int64  `json:"cid"`
}

type TimeWeightRecord struct {
	voted    map[int64]bool
	weight   map[int64]float64
	starting bool
}

func NewTimeWeightRecord() *TimeWeightRecord {
	return &TimeWeightRecord{
		voted:    map[int64]bool{},
		weight:   map[int64]float64{},
		starting: false,
	}
}

type Proof struct {
	Block []byte `json:"block"`
	Proof []byte `json:"proof"`
	Cid   int64  `json:"cid"`
}

func GenProof(lowerBlock types.Block, lowerProof []byte, cid int64) []byte {
	bblock, err := json.Marshal(lowerBlock)
	utils.PanicOnError(err)
	proof := &Proof{
		Block: bblock,
		Proof: lowerProof,
		Cid:   cid,
	}
	bproof, err := json.Marshal(proof)
	utils.PanicOnError(err)
	return bproof
}
