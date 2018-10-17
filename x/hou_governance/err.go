package hou_governance

import (
	"github.com/cosmos/cosmos-sdk/types"
	"strconv"
)

type CodeType = types.CodeType

const (
	DefaultCodespace types.CodespaceType = 7

	// Simple Gov errors reserve 700 ~ 799.
	CodeInvalidOption         CodeType = 701
	CodeInvalidProposalID     CodeType = 702
	CodeVotingPeriodClosed    CodeType = 703
	CodeEmptyProposalQueue    CodeType = 704
	CodeInvalidTitle          CodeType = 705
	CodeInvalidDescription    CodeType = 706
	CodeProposalNotFound      CodeType = 707
	CodeVoteNotFound          CodeType = 708
	CodeProposalQueueNotFound CodeType = 709
	CodeInvalidDeposit        CodeType = 710
)

//----------------------------------------
// Error constructors

// ErrInvalidOption throws an error on invalid option
func ErrInvalidOption(msg string) types.Error {
	return newError(DefaultCodespace, CodeInvalidOption, msg)
}

// ErrProposalNotFound throws an error when the searched proposal is not found
func ErrProposalNotFound(proposalID int64) types.Error {
	return newError(DefaultCodespace, CodeProposalNotFound, "Proposal with id "+
		strconv.Itoa(int(proposalID))+" not found")
}
// ErrMinimumDeposit throws an error when deposit is less than the default minimum
func ErrMinimumDeposit() types.Error {
	return newError(DefaultCodespace, CodeInvalidDeposit, "Deposit is lower than the minimum")
}
// ErrVoteNotFound throws an error when the searched vote is not found
func ErrVoteNotFound(msg string) types.Error {
	return newError(DefaultCodespace, CodeVoteNotFound, msg)
}
// ErrInvalidProposalID throws an error on invalid proposaID
func ErrInvalidProposalID(msg string) types.Error {
	return newError(DefaultCodespace, CodeInvalidProposalID, msg)
}

// ErrInvalidTitle throws an error on invalid title
func ErrInvalidTitle(msg string) types.Error {
	return newError(DefaultCodespace, CodeInvalidTitle, msg)
}

// ErrInvalidDescription throws an error on invalid description
func ErrInvalidDescription(msg string) types.Error {
	return newError(DefaultCodespace, CodeInvalidDescription, msg)
}


//----------------------------------------
func codeToDefaultMsg(code CodeType) string {
	switch code {
	case CodeInvalidOption:
		return "Invalid option"
	case CodeInvalidProposalID:
		return "Invalid proposalID"
	case CodeVotingPeriodClosed:
		return "Voting Period Closed"
	case CodeEmptyProposalQueue:
		return "ProposalQueue is empty"
	case CodeInvalidTitle:
		return "Invalid proposal title"
	case CodeInvalidDescription:
		return "Invalid proposal description"
	case CodeProposalNotFound:
		return "Proposal not found"
	case CodeVoteNotFound:
		return "Vote not found"
	case CodeProposalQueueNotFound:
		return "Proposal Queue not found"
	case CodeInvalidDeposit:
		return "Invalid deposit"
	default:
		return types.CodeToDefaultMsg(code)
	}
}

func msgOrDefaultMsg(msg string, code CodeType) string {
	if msg != "" {
		return msg
	}
	return codeToDefaultMsg(code)
}

func newError(codespace types.CodespaceType, code CodeType, msg string) types.Error {
	msg = msgOrDefaultMsg(msg, code)
	return types.NewError(codespace, code, msg)
}