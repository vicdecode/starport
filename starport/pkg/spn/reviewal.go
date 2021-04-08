package spn

import (
	"context"
	"errors"

	"github.com/cosmos/cosmos-sdk/types"
	launchtypes "github.com/tendermint/spn/x/launch/types"
)

// reviewal keeps a proposal's reviewal.
type reviewal struct {
	id         int
	isApproved bool
}

// Reviewal configures reviewal to create a review for a proposal.
type Reviewal func(*reviewal)

// ApproveProposal returns approval for a proposal with id.
func ApproveProposal(id int) Reviewal {
	return func(r *reviewal) {
		r.id = id
		r.isApproved = true
	}
}

// RejectProposal returns rejection for a proposals with id.
func RejectProposal(id int) Reviewal {
	return func(r *reviewal) {
		r.id = id
	}
}

// SubmitReviewals submits reviewals for proposals in batch for chainID by using SPN accountName.
func (c *Client) SubmitReviewals(ctx context.Context, accountName, chainID string, reviewals ...Reviewal) (gas uint64, broadcast func() error, err error) {
	if len(reviewals) == 0 {
		return 0, nil, errors.New("at least one reviewal required")
	}

	clientCtx, err := c.buildClientCtx(accountName)
	if err != nil {
		return 0, nil, err
	}

	var msgs []types.Msg

	for _, r := range reviewals {
		var rev reviewal
		r(&rev)

		if rev.isApproved {
			msgs = append(msgs, launchtypes.NewMsgApprove(chainID, int32(rev.id), clientCtx.GetFromAddress().String()))
		} else {
			msgs = append(msgs, launchtypes.NewMsgReject(chainID, int32(rev.id), clientCtx.GetFromAddress().String()))
		}
	}

	return c.broadcastProvision(ctx, clientCtx, msgs...)
}
