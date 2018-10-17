package hou_governance

// 提供存储功能
// TODO 键值对存储？？？
// Proposals --> 'proposals'|<proposalID>.
// Votes (Yes, No, Abstain) --> 'proposals'|<proposalID>|'votes'|<voterAddress>.

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/tendermint/go-amino"
)

type Story interface {
	GetProposal(ctx types.Context, proposalID int64) (Proposal, types.Error)
	SetProposal(ctx types.Context, proposal Proposal)
	GetVote(ctx types.Context, proposalID int64, voter types.AccAddress) (option string, err types.Error)
	SetVote(ctx types.Context, proposalID int64, voterAddr types.AccAddress, option string)
}


//TODO 什么时候用的？？？
type ProposalQueue interface {
	GetProposalQueue(ctx types.Context) (ProposalQueue, types.Error)
	SetProposalQueue(ctx types.Context, proposalQueue ProposalQueue)
	ProposalQueueHead(ctx types.Context) (Proposal, types.Error)
	ProposalQueuePop(ctx types.Context) (Proposal, types.Error)
	ProposalQueuePush(ctx types.Context, proposaID int64) types.Error
}

// nolint
type Keeper struct {
	storeKey  types.StoreKey      // Key to our module's store
	codespace types.CodespaceType // Reserves space for error codes
	cdc       *wire.Codec         // Codec to encore/decode structs

	ck bank.Keeper  // Needed to handle deposits. This module only requires read/writes to Atom balance
	sm stake.Keeper // Needed to compute voting power. This module only needs read access to the staking store
}

// NewKeeper crates a new keeper with write and read access
func NewKeeper(cdc *amino.Codec, simpleGovKey types.StoreKey, ck bank.Keeper, sm stake.Keeper, codespace types.CodespaceType) Keeper {

	return Keeper{
		storeKey:  simpleGovKey,
		cdc:       cdc,
		ck:        ck,
		sm:        sm,
		codespace: codespace,
	}
}

// GetProposal gets the proposal with the given id from the context.
func (k Keeper) GetProposal(ctx types.Context, proposalID int64) (Proposal, types.Error) {
	store := ctx.KVStore(k.storeKey)

	key := GenerateProposalKey(proposalID)
	bp := store.Get(key)
	if bp == nil {
		return Proposal{}, ErrProposalNotFound(proposalID)
	}
	proposal := &Proposal{}
	k.cdc.MustUnmarshalBinary(bp, proposal)

	return *proposal, nil
}

// SetProposal sets a proposal to the context
func (k Keeper) SetProposal(ctx types.Context, proposal Proposal) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(proposal)
	key := GenerateProposalKey(proposal.ID)
	store.Set(key, bz)
}

// GetVote returns the given option of a proposal stored in the keeper
// Used to check if an address already voted
func (k Keeper) GetVote(ctx types.Context, proposalID int64, voter types.AccAddress) (option string, err types.Error) {

	key := GenerateProposalVoteKey(proposalID, voter)
	store := ctx.KVStore(k.storeKey)
	bv := store.Get(key)
	if bv == nil {
		return "", ErrVoteNotFound("")
	}
	k.cdc.MustUnmarshalBinary(bv, &option)
	return option, nil
}

// SetVote sets the vote option to the proposal stored in the context store
func (k Keeper) SetVote(ctx types.Context, proposalID int64, voterAddr types.AccAddress, option string) {
	key := GenerateProposalVoteKey(proposalID, voterAddr)
	store := ctx.KVStore(k.storeKey)
	bv, err := k.cdc.MarshalBinary(option)
	if err != nil {
		panic(err)
	}
	store.Set(key, bv)
}

//--------------------------------------------------------------------------------------

// KeeperRead is a Keeper only with read access
type KeeperRead struct {
	Keeper
}

// NewKeeperRead crates a new keeper with read access
func NewKeeperRead(cdc *amino.Codec, simpleGovKey types.StoreKey, ck bank.Keeper, sm stake.Keeper, codespace types.CodespaceType) KeeperRead {
	return KeeperRead{Keeper{
		storeKey:  simpleGovKey,
		cdc:       cdc,
		ck:        ck,
		sm:        sm,
		codespace: codespace,
	}}
}

// SetProposal sets a proposal to the context
func (k KeeperRead) SetProposal(ctx types.Context, proposal Proposal) types.Error {
	return types.ErrUnauthorized("").TraceSDK("This keeper does not have write access for the simple governance store")
}

// SetVote sets the vote option to the proposal stored in the context store
func (k KeeperRead) SetVote(ctx types.Context, key []byte, option string) types.Error {
	return types.ErrUnauthorized("").TraceSDK("This keeper does not have write access for the simple governance store")
}
