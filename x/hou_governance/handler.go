package hou_governance

import (
	"encoding/binary"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"reflect"
)

// Minimum proposal deposit
var minDeposit = types.NewInt(int64(100))

func int64ToBytes(i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return b
}

// 消息处理函数
// NewHandler creates a new handler for all simple_gov type messages.
func NewHandler(k Keeper) types.Handler {
	return func(ctx types.Context, msg types.Msg) types.Result {
		switch msg := msg.(type) {
		case SubmitProposalMsg:
			return handleSubmitProposalMsg(ctx, k, msg)
		case VoteMsg:
			return handleVoteMsg(ctx, k, msg)
		default:
			errMsg := "Unrecognized gov Msg type: " + reflect.TypeOf(msg).Name()
			return types.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// handleVoteMsg handles the logic of a SubmitProposalMsg
func handleSubmitProposalMsg(ctx types.Context, k Keeper, msg SubmitProposalMsg) types.Result {
	err := msg.ValidateBasic()
	if err != nil {
		return err.Result()
	}

	// Subtract coins from the submitter balance and updates it
	_, _, err = k.ck.SubtractCoins(ctx, msg.Submitter, msg.Deposit)
	if err != nil {
		return err.Result()
	}

	//if msg.Deposit.AmountOf("Atom").GT(minDeposit) ||
	//	msg.Deposit.AmountOf("Atom").Equal(minDeposit) {
	//	proposal := k.NewProposal(ctx, msg.Title, msg.Description)
	//	k.SetProposal(ctx, proposal)
	//	return types.Result{
	//		Tags: types.NewTags(
	//			"action", []byte("propose"),
	//			"proposal", int64ToBytes(proposal.ID),
	//			"submitter", msg.Submitter.Bytes(),
	//		),
	//	}
	//}
	return ErrMinimumDeposit().Result()
}

// handleVoteMsg handles the logic of a VoteMsg
func handleVoteMsg(ctx types.Context, k Keeper, msg VoteMsg) types.Result {
	err := msg.ValidateBasic()
	if err != nil {
		return err.Result()
	}

	proposal, err := k.GetProposal(ctx, msg.ProposalID)
	if err != nil {
		return err.Result()
	}

	//if ctx.BlockHeight() > proposal.SubmitBlock+votingPeriod ||
	//	!proposal.IsOpen() {
	//	return ErrVotingPeriodClosed().Result()
	//}

	delegatedTo := k.sm.GetDelegations(ctx, msg.Voter, 10)

	if len(delegatedTo) <= 0 {
		return stake.ErrNoDelegatorForAddress(stake.DefaultCodespace).Result()
	}
	// Check if address already voted
	voterOption, err := k.GetVote(ctx, msg.ProposalID, msg.Voter)
	if voterOption == "" && err != nil {
		// voter has not voted yet
		for _, delegation := range delegatedTo {
			bondShares := delegation.GetBondShares().EvaluateBig().Int64()
			err = proposal.updateTally(msg.Option, bondShares)
			if err != nil {
				return err.Result()
			}
		}
	} else {
		// voter has already voted
		for _, delegation := range delegatedTo {
			bondShares := delegation.GetBondShares().EvaluateBig().Int64()
			// update previous vote with new one
			err = proposal.updateTally(voterOption, -bondShares)
			if err != nil {
				return err.Result()
			}
			err = proposal.updateTally(msg.Option, bondShares)
			if err != nil {
				return err.Result()
			}
		}
	}

	k.SetVote(ctx, msg.ProposalID, msg.Voter, msg.Option)
	k.SetProposal(ctx, proposal)

	return types.Result{
		Tags: types.NewTags(
			"action", []byte("vote"),
			"proposal", int64ToBytes(msg.ProposalID),
			"voter", msg.Voter.Bytes(),
			"option", []byte(msg.Option),
		),
	}

}
