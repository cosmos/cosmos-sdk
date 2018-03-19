package proxy

import (
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/tendermint/lite"
	certclient "github.com/tendermint/tendermint/lite/client"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ rpcclient.Client = Wrapper{}

// Wrapper wraps a rpcclient with a Certifier and double-checks any input that is
// provable before passing it along. Allows you to make any rpcclient fully secure.
type Wrapper struct {
	rpcclient.Client
	cert *lite.InquiringCertifier
}

// SecureClient uses a given certifier to wrap an connection to an untrusted
// host and return a cryptographically secure rpc client.
//
// If it is wrapping an HTTP rpcclient, it will also wrap the websocket interface
func SecureClient(c rpcclient.Client, cert *lite.InquiringCertifier) Wrapper {
	wrap := Wrapper{c, cert}
	// TODO: no longer possible as no more such interface exposed....
	// if we wrap http client, then we can swap out the event switch to filter
	// if hc, ok := c.(*rpcclient.HTTP); ok {
	// 	evt := hc.WSEvents.EventSwitch
	// 	hc.WSEvents.EventSwitch = WrappedSwitch{evt, wrap}
	// }
	return wrap
}

// ABCIQueryWithOptions exposes all options for the ABCI query and verifies the returned proof
func (w Wrapper) ABCIQueryWithOptions(path string, data cmn.HexBytes,
	opts rpcclient.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {

	res, _, err := GetWithProofOptions(path, data, opts, w.Client, w.cert)
	return res, err
}

// ABCIQuery uses default options for the ABCI query and verifies the returned proof
func (w Wrapper) ABCIQuery(path string, data cmn.HexBytes) (*ctypes.ResultABCIQuery, error) {
	return w.ABCIQueryWithOptions(path, data, rpcclient.DefaultABCIQueryOptions)
}

// Tx queries for a given tx and verifies the proof if it was requested
func (w Wrapper) Tx(hash []byte, prove bool) (*ctypes.ResultTx, error) {
	res, err := w.Client.Tx(hash, prove)
	if !prove || err != nil {
		return res, err
	}
	h := int64(res.Height)
	check, err := GetCertifiedCommit(h, w.Client, w.cert)
	if err != nil {
		return res, err
	}
	err = res.Proof.Validate(check.Header.DataHash)
	return res, err
}

// BlockchainInfo requests a list of headers and verifies them all...
// Rather expensive.
//
// TODO: optimize this if used for anything needing performance
func (w Wrapper) BlockchainInfo(minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error) {
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
		err = ValidateBlockMeta(meta, check)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

// Block returns an entire block and verifies all signatures
func (w Wrapper) Block(height *int64) (*ctypes.ResultBlock, error) {
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

// Commit downloads the Commit and certifies it with the lite.
//
// This is the foundation for all other verification in this module
func (w Wrapper) Commit(height *int64) (*ctypes.ResultCommit, error) {
	rpcclient.WaitForHeight(w.Client, *height, nil)
	r, err := w.Client.Commit(height)
	// if we got it, then certify it
	if err == nil {
		check := certclient.CommitFromResult(r)
		err = w.cert.Certify(check)
	}
	return r, err
}

// // WrappedSwitch creates a websocket connection that auto-verifies any info
// // coming through before passing it along.
// //
// // Since the verification takes 1-2 rpc calls, this is obviously only for
// // relatively low-throughput situations that can tolerate a bit extra latency
// type WrappedSwitch struct {
// 	types.EventSwitch
// 	client rpcclient.Client
// }

// // FireEvent verifies any block or header returned from the eventswitch
// func (s WrappedSwitch) FireEvent(event string, data events.EventData) {
// 	tm, ok := data.(types.TMEventData)
// 	if !ok {
// 		fmt.Printf("bad type %#v\n", data)
// 		return
// 	}

// 	// check to validate it if possible, and drop if not valid
// 	switch t := tm.Unwrap().(type) {
// 	case types.EventDataNewBlockHeader:
// 		err := verifyHeader(s.client, t.Header)
// 		if err != nil {
// 			fmt.Printf("Invalid header: %#v\n", err)
// 			return
// 		}
// 	case types.EventDataNewBlock:
// 		err := verifyBlock(s.client, t.Block)
// 		if err != nil {
// 			fmt.Printf("Invalid block: %#v\n", err)
// 			return
// 		}
// 		// TODO: can we verify tx as well? anything else
// 	}

// 	// looks good, we fire it
// 	s.EventSwitch.FireEvent(event, data)
// }

// func verifyHeader(c rpcclient.Client, head *types.Header) error {
// 	// get a checkpoint to verify from
// 	commit, err := c.Commit(&head.Height)
// 	if err != nil {
// 		return err
// 	}
// 	check := certclient.CommitFromResult(commit)
// 	return ValidateHeader(head, check)
// }
//
// func verifyBlock(c rpcclient.Client, block *types.Block) error {
// 	// get a checkpoint to verify from
// 	commit, err := c.Commit(&block.Height)
// 	if err != nil {
// 		return err
// 	}
// 	check := certclient.CommitFromResult(commit)
// 	return ValidateBlock(block, check)
// }
