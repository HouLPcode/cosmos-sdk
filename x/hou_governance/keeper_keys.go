package hou_governance

import (
	"encoding/binary"
	"github.com/cosmos/cosmos-sdk/types"
)

// 用来生成存储的索引
// nolint
var (
	KeyNextProposalID        = []byte("newProposalID")
	KeyActiveProposalQueue   = []byte("activeProposalQueue")
	KeyInactiveProposalQueue = []byte("inactiveProposalQueue")
)

// GenerateProposalKey creates a key of the form "proposals"|{proposalID}
func GenerateProposalKey(proposalID int64) []byte {
	var key []byte
	proposalIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(proposalIDBytes, uint64(proposalID))

	key = []byte("proposals")
	key = append(key, proposalIDBytes...)
	return key
}

// GenerateProposalVotesKey creates a key of the form "proposals"|{proposalID}|"votes"
func GenerateProposalVotesKey(proposalID int64) []byte {
	key := GenerateProposalKey(proposalID)
	key = append(key, []byte("votes")...)
	return key
}

// GenerateProposalVoteKey creates a key of the form "proposals"|{proposalID}|"votes"|{voterAddress}
func GenerateProposalVoteKey(proposalID int64, voterAddr types.AccAddress) []byte {
	key := GenerateProposalVotesKey(proposalID)
	key = append(key, voterAddr.Bytes()...)
	return key
}