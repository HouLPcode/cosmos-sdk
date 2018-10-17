package hou_governance
// 定义了一个提案Proposal和两个消息SubmitProposalMsg和VoteMsg

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/types"
	"strings"
)

// Proposal defines the basic properties of a staking proposal
type Proposal struct {
	ID          int64          `json:"id"`           // ID of the proposal
	Title       string         `json:"title"`        // Title of the proposal
	Description string         `json:"description"`  // Description of the proposal
	Submitter   types.AccAddress `json:"submitter"`    // Account address of the proposer
	SubmitBlock int64          `json:"submit_block"` // Block height from which the proposal is open for votations
	State       string         `json:"state"`        // One of Open, Accepted, Rejected
	Deposit     types.Coins      `json:"deposit"`      // Coins deposited in escrow

	//通过updateTally函数统计投票数量
	YesVotes     int64 `json:"yes_votes"`     // Total Yes votes
	NoVotes      int64 `json:"no_votes"`      // Total No votes
	AbstainVotes int64 `json:"abstain_votes"` // Total Abstain votes
}

// updateTally updates the counter for each of the available options
// 更新投票数据
func (p *Proposal) updateTally(option string, amount int64) types.Error {
	switch option {
	case "Yes":
		p.YesVotes += amount
		return nil
	case "No":
		p.NoVotes += amount
		return nil
	case "Abstain":
		p.AbstainVotes += amount
		return nil
	default:
		return ErrInvalidOption("Invalid option: " + option)
	}
}

//消息需要实现types中的Msg接口

//SubmitProposalMsg defines a message to create a proposal
type SubmitProposalMsg struct {
	Title       string         // Title of the proposal
	Description string         // Description of the proposal
	Deposit     types.Coins      // Deposit paid by submitter. Must be > MinDeposit to enter voting period
	Submitter   types.AccAddress // Address of the submitter
}

// NewSubmitProposalMsg submits a message with a new proposal
func NewSubmitProposalMsg(title string, description string, deposit types.Coins, submitter types.AccAddress) SubmitProposalMsg {
	return SubmitProposalMsg{
		Title:       title,
		Description: description,
		Deposit:     deposit,
		Submitter:   submitter,
	}
}

// Return the message type.
// Must be alphanumeric or empty.
func (msg SubmitProposalMsg)Type() string{
	return "hou governance"
}

// ValidateBasic does a simple validation check that
// doesn't require access to any other information.
func (msg SubmitProposalMsg)ValidateBasic() types.Error{
	if len(msg.Submitter) == 0 {
		return types.ErrInvalidAddress("Invalid address: " + msg.Submitter.String())
	}
	if len(strings.TrimSpace(msg.Title)) <= 0 {
		return ErrInvalidTitle("Cannot submit a proposal with empty title")
	}

	if len(strings.TrimSpace(msg.Description)) <= 0 {
		return ErrInvalidDescription("Cannot submit a proposal with empty description")
	}

	if !msg.Deposit.IsValid() {
		return types.ErrInvalidCoins("Deposit is not valid")
	}

	if !msg.Deposit.IsPositive() {
		return types.ErrInvalidCoins("Deposit cannot be negative")
	}

	return nil
}

// Get the canonical byte representation of the Msg.
func (msg SubmitProposalMsg)GetSignBytes() []byte{
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// Signers returns the addrs of signers that must sign.
// CONTRACT: All signatures must be present to be valid.
// CONTRACT: Returns addrs in some deterministic order.
func (msg SubmitProposalMsg)GetSigners() []types.AccAddress{
	return []types.AccAddress{msg.Submitter}
}

// VoteMsg defines the msg of a staker containing the vote option to an
// specific proposal
type VoteMsg struct {
	ProposalID int64          // ID of the proposal
	Option     string         // Option chosen by voter
	Voter      types.AccAddress // Address of the voter
}

// NewVoteMsg creates a VoteMsg instance
func NewVoteMsg(proposalID int64, option string, voter types.AccAddress) VoteMsg {
	// by default a nil option is an abstention
	if option == "" {
		option = "Abstain"
	}
	return VoteMsg{
		ProposalID: proposalID,
		Option:     option,
		Voter:      voter,
	}
}

// Type Implements Msg
func (msg VoteMsg) Type() string {
	return "simpleGov"
}

// GetSigners Implements Msg
func (msg VoteMsg) GetSigners() []types.AccAddress {
	return []types.AccAddress{msg.Voter}
}


// GetSignBytes Implements Msg
func (msg VoteMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

func isValidOption(option string) bool {
	options := []string{"Yes", "No", "Abstain"}
	for _, value := range options {
		if value == option {
			return true
		}
	}
	return false
}

// ValidateBasic Implements Msg
func (msg VoteMsg) ValidateBasic() types.Error {
	if len(msg.Voter) == 0 {
		return types.ErrInvalidAddress("Invalid address: " + msg.Voter.String())
	}
	if msg.ProposalID <= 0 {
		return ErrInvalidProposalID("ProposalID cannot be negative")
	}
	if !isValidOption(msg.Option) {
		return ErrInvalidOption("Invalid voting option: " + msg.Option)
	}
	if len(strings.TrimSpace(msg.Option)) <= 0 {
		return ErrInvalidOption("Option can't be blank")
	}

	return nil
}