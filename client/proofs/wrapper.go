package proofs

import (
	"fmt"

	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/tmlibs/events"

	"github.com/tendermint/tendermint/certifiers"
	certclient "github.com/tendermint/tendermint/certifiers/client"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
)

var _ rpcclient.Client = Wrapper{}

type Wrapper struct {
	rpcclient.Client
	cert *certifiers.Inquiring
}

// SecureClient uses a given certifier to wrap an connection to an untrusted
// host and return a cryptographically secure rpc client.
func SecureClient(c rpcclient.Client, cert *certifiers.Inquiring) Wrapper {
	wrap := Wrapper{c, cert}
	// if we wrap http client, then we can swap out the event switch to filter
	if hc, ok := c.(*rpcclient.HTTP); ok {
		evt := hc.WSEvents.EventSwitch
		hc.WSEvents.EventSwitch = WrappedSwitch{evt, wrap}
	}
	return wrap
}

func (w Wrapper) ABCIQueryWithOptions(path string, data data.Bytes, opts rpcclient.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	res, _, err := client.GetWithProofOptions(path, data, opts, w.Client, w.cert)
	return res, err
}

func (w Wrapper) ABCIQuery(path string, data data.Bytes) (*ctypes.ResultABCIQuery, error) {
	return w.ABCIQueryWithOptions(path, data, rpcclient.DefaultABCIQueryOptions)
}

func (w Wrapper) Tx(hash []byte, prove bool) (*ctypes.ResultTx, error) {
	res, err := w.Client.Tx(hash, prove)
	if !prove || err != nil {
		return res, err
	}
	check, err := client.GetCertifiedCommit(res.Height, w.Client, w.cert)
	if err != nil {
		return res, err
	}
	err = res.Proof.Validate(check.Header.DataHash)
	return res, err
}

func (w Wrapper) BlockchainInfo(minHeight, maxHeight int) (*ctypes.ResultBlockchainInfo, error) {
	r, err := w.Client.BlockchainInfo(minHeight, maxHeight)
	if err != nil {
		return nil, err
	}

	// go and verify every blockmeta in the result....
	for _, meta := range r.BlockMetas {
		// get a checkpoint to verify from
		c, err := w.Commit(&meta.Header.Height)
		if err != nil {
			return nil, err
		}
		check := certclient.CommitFromResult(c)
		// TODO: 3
		err = ValidateBlockMeta(meta, check)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (w Wrapper) Block(height *int) (*ctypes.ResultBlock, error) {
	r, err := w.Client.Block(height)
	if err != nil {
		return nil, err
	}
	// get a checkpoint to verify from
	c, err := w.Commit(height)
	if err != nil {
		return nil, err
	}
	check := certclient.CommitFromResult(c)

	// now verify
	err = ValidateBlockMeta(r.BlockMeta, check)
	if err != nil {
		return nil, err
	}
	err = ValidateBlock(r.Block, check)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Commit downloads the Commit and certifies it with the certifiers.
//
// This is the foundation for all other verification in this module
func (w Wrapper) Commit(height *int) (*ctypes.ResultCommit, error) {
	rpcclient.WaitForHeight(w.Client, *height, nil)
	r, err := w.Client.Commit(height)
	// if we got it, then certify it
	if err == nil {
		check := certclient.CommitFromResult(r)
		err = w.cert.Certify(check)
	}
	return r, err
}

type WrappedSwitch struct {
	types.EventSwitch
	client rpcclient.Client
}

func (s WrappedSwitch) FireEvent(event string, data events.EventData) {
	tm, ok := data.(types.TMEventData)
	if !ok {
		fmt.Printf("bad type %#v\n", data)
		return
	}

	// check to validate it if possible, and drop if not valid
	switch t := tm.Unwrap().(type) {
	case types.EventDataNewBlockHeader:
		err := verifyHeader(s.client, t.Header)
		if err != nil {
			fmt.Printf("Invalid header: %#v\n", err)
			return
		}
	case types.EventDataNewBlock:
		err := verifyBlock(s.client, t.Block)
		if err != nil {
			fmt.Printf("Invalid block: %#v\n", err)
			return
		}
	}

	// looks good, we fire it
	s.EventSwitch.FireEvent(event, data)
}

func verifyHeader(c rpcclient.Client, head *types.Header) error {
	// get a checkpoint to verify from
	commit, err := c.Commit(&head.Height)
	if err != nil {
		return err
	}
	check := certclient.CommitFromResult(commit)
	return ValidateHeader(head, check)
}

func verifyBlock(c rpcclient.Client, block *types.Block) error {
	// get a checkpoint to verify from
	commit, err := c.Commit(&block.Height)
	if err != nil {
		return err
	}
	check := certclient.CommitFromResult(commit)
	return ValidateBlock(block, check)
}
