package spn

import (
	"context"
	"errors"

	"github.com/cosmos/cosmos-sdk/types"
	launchtypes "github.com/tendermint/spn/x/launch/types"
	"github.com/tendermint/starport/starport/pkg/jsondoc"
)

// ProposalType represents the type of the proposal
type ProposalType string

const (
	ProposalTypeAll          ProposalType = ""
	ProposalTypeAddAccount   ProposalType = "add-account"
	ProposalTypeAddValidator ProposalType = "add-validator"
)

// ProposalStatus represents the status of the proposal
type ProposalStatus string

const (
	ProposalStatusAll      ProposalStatus = ""
	ProposalStatusPending  ProposalStatus = "pending"
	ProposalStatusApproved ProposalStatus = "approved"
	ProposalStatusRejected ProposalStatus = "rejected"
)

// Proposal represents a proposal.
type Proposal struct {
	ID        int                   `yaml:",omitempty"`
	Status    ProposalStatus        `yaml:",omitempty"`
	Account   *ProposalAddAccount   `yaml:",omitempty"`
	Validator *ProposalAddValidator `yaml:",omitempty"`
}

// ProposalAddAccount used to propose adding an account.
type ProposalAddAccount struct {
	Address string
	Coins   types.Coins
}

// ProposalAddValidator used to propose adding a validator.
type ProposalAddValidator struct {
	Gentx            jsondoc.Doc
	ValidatorAddress string
	SelfDelegation   types.Coin
	P2PAddress       string
}

var statusFromSPN = map[launchtypes.ProposalStatus]ProposalStatus{
	launchtypes.ProposalStatus_PENDING:  ProposalStatusPending,
	launchtypes.ProposalStatus_APPROVED: ProposalStatusApproved,
	launchtypes.ProposalStatus_REJECTED: ProposalStatusRejected,
}

var statusToSPN = map[ProposalStatus]launchtypes.ProposalStatus{
	ProposalStatusAll:      launchtypes.ProposalStatus_ANY_STATUS,
	ProposalStatusPending:  launchtypes.ProposalStatus_PENDING,
	ProposalStatusApproved: launchtypes.ProposalStatus_APPROVED,
	ProposalStatusRejected: launchtypes.ProposalStatus_REJECTED,
}

var proposalTypeToSPN = map[ProposalType]launchtypes.ProposalType{
	ProposalTypeAll:          launchtypes.ProposalType_ANY_TYPE,
	ProposalTypeAddAccount:   launchtypes.ProposalType_ADD_ACCOUNT,
	ProposalTypeAddValidator: launchtypes.ProposalType_ADD_VALIDATOR,
}

// proposalListOptions holds proposal listing options.
type proposalListOptions struct {
	typ    ProposalType
	status ProposalStatus
}

// ProposalListOption configures proposal listing options.
type ProposalListOption func(*proposalListOptions)

// ProposalListStatus sets proposal status filter for proposal listing.
func ProposalListStatus(status ProposalStatus) ProposalListOption {
	return func(o *proposalListOptions) {
		o.status = status
	}
}

// ProposalListType sets proposal type filter for proposal listing.
func ProposalListType(typ ProposalType) ProposalListOption {
	return func(o *proposalListOptions) {
		o.typ = typ
	}
}

// ProposalList lists proposals on a chain by status.
func (c *Client) ProposalList(ctx context.Context, acccountName, chainID string, options ...ProposalListOption) ([]Proposal, error) {
	o := &proposalListOptions{}
	for _, apply := range options {
		apply(o)
	}

	// Get spn proposal status
	spnStatus, ok := statusToSPN[o.status]
	if !ok {
		return nil, errors.New("unrecognized status")
	}

	// Get spn proposal type
	spnType, ok := proposalTypeToSPN[o.typ]
	if !ok {
		return nil, errors.New("unrecognized type")
	}

	var proposals []Proposal
	var spnProposals []*launchtypes.Proposal

	queryClient := launchtypes.NewQueryClient(c.clientCtx)

	// Send query
	res, err := queryClient.ListProposals(ctx, &launchtypes.QueryListProposalsRequest{
		ChainID: chainID,
		Status:  spnStatus,
		Type:    spnType,
	})
	if err != nil {
		return nil, err
	}
	spnProposals = res.Proposals

	// Format proposals
	for _, gp := range spnProposals {
		proposal, err := c.toProposal(*gp)
		if err != nil {
			return nil, err
		}

		proposals = append(proposals, proposal)
	}

	return proposals, nil
}

func (c *Client) toProposal(proposal launchtypes.Proposal) (Proposal, error) {
	p := Proposal{
		ID:     int(proposal.ProposalInformation.ProposalID),
		Status: statusFromSPN[proposal.ProposalState.GetStatus()],
	}
	switch payload := proposal.Payload.(type) {
	case *launchtypes.Proposal_AddAccountPayload:
		p.Account = &ProposalAddAccount{
			Address: payload.AddAccountPayload.Address,
			Coins:   payload.AddAccountPayload.Coins,
		}

	case *launchtypes.Proposal_AddValidatorPayload:
		p.Validator = &ProposalAddValidator{
			P2PAddress:       payload.AddValidatorPayload.Peer,
			Gentx:            payload.AddValidatorPayload.GenTx,
			ValidatorAddress: payload.AddValidatorPayload.ValidatorAddress,
			SelfDelegation:   *payload.AddValidatorPayload.SelfDelegation,
		}
	}

	return p, nil
}

func (c *Client) ProposalGet(ctx context.Context, accountName, chainID string, id int) (Proposal, error) {
	queryClient := launchtypes.NewQueryClient(c.clientCtx)

	// Query the proposal
	param := &launchtypes.QueryShowProposalRequest{
		ChainID:    chainID,
		ProposalID: int32(id),
	}
	res, err := queryClient.ShowProposal(ctx, param)
	if err != nil {
		return Proposal{}, err
	}

	return c.toProposal(*res.Proposal)
}

// ProposalOption configures Proposal to set a spesific type of proposal.
type ProposalOption func(*Proposal)

// AddAccountProposal creates an add account proposal option.
func AddAccountProposal(address string, coins types.Coins) ProposalOption {
	return func(p *Proposal) {
		p.Account = &ProposalAddAccount{address, coins}
	}
}

// AddValidatorProposal creates an add validator proposal option.
func AddValidatorProposal(gentx jsondoc.Doc, validatorAddress string, selfDelegation types.Coin, p2pAddress string) ProposalOption {
	return func(p *Proposal) {
		p.Validator = &ProposalAddValidator{gentx, validatorAddress, selfDelegation, p2pAddress}
	}
}

// Propose proposes given proposals in batch for chainID by using SPN accountName.
func (c *Client) Propose(ctx context.Context, accountName, chainID string, proposals ...ProposalOption) error {
	if len(proposals) == 0 {
		return errors.New("at least one proposal required")
	}

	clientCtx, err := c.buildClientCtx(accountName)
	if err != nil {
		return err
	}

	var msgs []types.Msg

	for _, p := range proposals {
		var proposal Proposal
		p(&proposal)

		switch {
		case proposal.Account != nil:
			// Create the proposal payload
			payload := launchtypes.NewProposalAddAccountPayload(
				proposal.Account.Address,
				proposal.Account.Coins,
			)

			msgs = append(msgs, launchtypes.NewMsgProposalAddAccount(
				chainID,
				clientCtx.GetFromAddress().String(),
				payload,
			))

		case proposal.Validator != nil:
			// Create the proposal payload
			payload := launchtypes.NewProposalAddValidatorPayload(
				proposal.Validator.Gentx,
				proposal.Validator.ValidatorAddress,
				proposal.Validator.SelfDelegation,
				proposal.Validator.P2PAddress,
			)

			msgs = append(msgs, launchtypes.NewMsgProposalAddValidator(
				chainID,
				clientCtx.GetFromAddress().String(),
				payload,
			))
		}
	}

	return c.broadcast(ctx, clientCtx, msgs...)
}
