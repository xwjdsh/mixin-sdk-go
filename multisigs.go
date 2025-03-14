package mixin

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

const (
	UTXOStateUnspent = "unspent"
	UTXOStateSigned  = "signed"
	UTXOStateSpent   = "spent"
)

const (
	MultisigActionSign   = "sign"
	MultisigActionUnlock = "unlock"
)

const (
	MultisigStateInitial = "initial"
	MultisigStateSigned  = "signed"
)

type (
	MultisigUTXO struct {
		Type            string          `json:"type"`
		UserID          string          `json:"user_id"`
		UTXOID          string          `json:"utxo_id"`
		AssetID         string          `json:"asset_id"`
		TransactionHash Hash            `json:"transaction_hash"`
		OutputIndex     int             `json:"output_index"`
		Sender          string          `json:"sender,omitempty"`
		Amount          decimal.Decimal `json:"amount"`
		Threshold       uint8           `json:"threshold"`
		Members         []string        `json:"members"`
		Memo            string          `json:"memo"`
		State           string          `json:"state"`
		CreatedAt       time.Time       `json:"created_at"`
		UpdatedAt       time.Time       `json:"updated_at"`
		SignedBy        string          `json:"signed_by"`
		SignedTx        string          `json:"signed_tx"`
	}

	MultisigRequest struct {
		Type            string          `json:"type"`
		RequestID       string          `json:"request_id"`
		UserID          string          `json:"user_id"`
		AssetID         string          `json:"asset_id"`
		Amount          decimal.Decimal `json:"amount"`
		Threshold       uint8           `json:"threshold"`
		Senders         []string        `json:"senders"`
		Receivers       []string        `json:"receivers"`
		Signers         []string        `json:"signers"`
		Memo            string          `json:"memo"`
		Action          string          `json:"action"`
		State           string          `json:"state"`
		TransactionHash Hash            `json:"transaction_hash"`
		RawTransaction  string          `json:"raw_transaction"`
		CreatedAt       time.Time       `json:"created_at"`
		UpdatedAt       time.Time       `json:"updated_at"`
		CodeID          string          `json:"code_id"`
	}
)

func (utxo MultisigUTXO) Asset() Hash {
	return NewHash([]byte(utxo.AssetID))
}

func HashMembers(ids []string) string {
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	var in string
	for _, id := range ids {
		in = in + id
	}
	return NewHash([]byte(in)).String()
}

// ReadMultisigs return a list of multisig utxos
func (c *Client) ReadMultisigs(ctx context.Context, offset time.Time, limit int) ([]*MultisigUTXO, error) {
	params := make(map[string]string)
	if !offset.IsZero() {
		params["offset"] = offset.UTC().Format(time.RFC3339Nano)
	}

	if limit > 0 {
		params["limit"] = fmt.Sprint(limit)
	}

	var utxos []*MultisigUTXO
	if err := c.Get(ctx, "/multisigs", params, &utxos); err != nil {
		return nil, err
	}

	return utxos, nil
}

// ReadMultisigOutputs return a list of multisig outputs, including unspent, signed, spent utxos
func (c *Client) ReadMultisigOutputs(ctx context.Context, members []string, threshold uint8, offset time.Time, limit int) ([]*MultisigUTXO, error) {
	params := make(map[string]string)
	if !offset.IsZero() {
		params["offset"] = offset.UTC().Format(time.RFC3339Nano)
	}

	if limit > 0 {
		params["limit"] = fmt.Sprint(limit)
	}

	if len(members) > 0 {
		if threshold < 1 || int(threshold) > len(members) {
			return nil, errors.New("invalid members")
		}
		params["members"] = HashMembers(members)
		params["threshold"] = fmt.Sprint(threshold)
	}

	var utxos []*MultisigUTXO
	if err := c.Get(ctx, "/multisigs/outputs", params, &utxos); err != nil {
		return nil, err
	}

	return utxos, nil
}

// CreateMultisig create a multisig request
func (c *Client) CreateMultisig(ctx context.Context, action, raw string) (*MultisigRequest, error) {
	params := map[string]string{
		"action": action,
		"raw":    raw,
	}

	var req MultisigRequest
	if err := c.Post(ctx, "/multisigs/requests", params, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

// SignMultisig sign a multisig request
func (c *Client) SignMultisig(ctx context.Context, reqID, pin string) (*MultisigRequest, error) {
	uri := "/multisigs/requests/" + reqID + "/sign"
	params := map[string]string{
		"pin": c.EncryptPin(pin),
	}

	var req MultisigRequest
	if err := c.Post(ctx, uri, params, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

// CancelMultisig cancel a multisig request
func (c *Client) CancelMultisig(ctx context.Context, reqID string) error {
	uri := "/multisigs/requests/" + reqID + "/cancel"
	if err := c.Post(ctx, uri, nil, nil); err != nil {
		return err
	}

	return nil
}

// UnlockMultisig unlock a multisig request
func (c *Client) UnlockMultisig(ctx context.Context, reqID, pin string) error {
	var (
		uri    = "/multisigs/requests/" + reqID + "/unlock"
		params = map[string]string{
			"pin": c.EncryptPin(pin),
		}
	)
	if err := c.Post(ctx, uri, params, nil); err != nil {
		return err
	}

	return nil
}
