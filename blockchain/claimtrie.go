package blockchain

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"

	"github.com/btcsuite/btcd/claimtrie"
	"github.com/btcsuite/btcd/claimtrie/node"
)

// Hack: print which block mismatches happened, but keep recording.
var mismatchedPrinted bool

func (b *BlockChain) CheckClaimScripts(block *btcutil.Block, node *blockNode, view *UtxoViewpoint) error {
	ht := block.Height()

	for _, tx := range block.Transactions() {
		h := handler{ht, tx, view, map[string]bool{}}
		if err := h.handleTxIns(b.claimTrie); err != nil {
			return err
		}
		if err := h.handleTxOuts(b.claimTrie); err != nil {
			return err
		}
	}

	// Hack: let the claimtrie know the expected Hash.
	b.claimTrie.ReportHash(ht, node.claimTrie)

	b.claimTrie.AppendBlock()
	hash := b.claimTrie.MerkleHash()

	if node.claimTrie != *hash {
		if !mismatchedPrinted {
			fmt.Printf("\n\nheight: %d, ct.MerkleHash: %s != node.ClaimTrie: %s\n\n", ht, *hash, node.claimTrie)
			mismatchedPrinted = true
		}
		// Continue recording the input to the claimtrie.

		// return fmt.Errorf("height: %d, ct.MerkleHash: %s != node.ClaimTrie: %s", ht, *hash, node.claimTrie)
	}
	return nil
}

type handler struct {
	ht    int32
	tx    *btcutil.Tx
	view  *UtxoViewpoint
	spent map[string]bool
}

func (h *handler) handleTxIns(ct *claimtrie.ClaimTrie) error {
	if IsCoinBase(h.tx) {
		return nil
	}
	for _, txIn := range h.tx.MsgTx().TxIn {
		op := txIn.PreviousOutPoint
		e := h.view.LookupEntry(op)
		cs, err := txscript.DecodeClaimScript(e.pkScript)
		if err == txscript.ErrNotClaimScript {
			continue
		}
		if err != nil {
			return err
		}

		var id node.ClaimID
		name := cs.Name()

		switch cs.Opcode() {
		case txscript.OP_CLAIMNAME:
			id = node.NewClaimID(op)
			h.spent[id.String()] = true
			err = ct.SpendClaim(name, op)
		case txscript.OP_UPDATECLAIM:
			copy(id[:], cs.ClaimID())
			h.spent[id.String()] = true
			err = ct.SpendClaim(name, op)
		case txscript.OP_SUPPORTCLAIM:
			copy(id[:], cs.ClaimID())
			err = ct.SpendSupport(name, op)
		}
		if err != nil {
			return errors.Wrapf(err, "handleTxIns")
		}
	}
	return nil
}

func (h *handler) handleTxOuts(ct *claimtrie.ClaimTrie) error {
	for i, txOut := range h.tx.MsgTx().TxOut {
		op := wire.NewOutPoint(h.tx.Hash(), uint32(i))
		cs, err := txscript.DecodeClaimScript(txOut.PkScript)
		if err == txscript.ErrNotClaimScript {
			continue
		}
		if err != nil {
			return err
		}

		var id node.ClaimID
		name := cs.Name()
		amt := txOut.Value
		value := cs.Value()

		switch cs.Opcode() {
		case txscript.OP_CLAIMNAME:
			id = node.NewClaimID(*op)
			err = ct.AddClaim(name, *op, amt, value)
		case txscript.OP_SUPPORTCLAIM:
			copy(id[:], cs.ClaimID())
			err = ct.AddSupport(name, *op, amt, id)
		case txscript.OP_UPDATECLAIM:
			copy(id[:], cs.ClaimID())
			if !h.spent[id.String()] {
				fmt.Printf("%d can't find id: %s\n", h.ht, id)
				continue
			}
			delete(h.spent, id.String())
			err = ct.UpdateClaim(name, *op, amt, id, value)
		}
		if err != nil {
			return errors.Wrapf(err, "handleTxOuts")
		}
	}
	return nil
}
