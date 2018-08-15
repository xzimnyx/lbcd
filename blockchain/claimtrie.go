package blockchain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"

	"github.com/lbryio/claimtrie"
	"github.com/lbryio/claimtrie/cfg"
	"github.com/lbryio/claimtrie/change"
	"github.com/lbryio/claimtrie/claim"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	dbCS *leveldb.DB
)

func init() {
	path := cfg.DefaultConfig(cfg.ClaimScriptDB)
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		panic(err)
	}
	dbCS = db
}

func (b *BlockChain) CheckClaimScripts(block *btcutil.Block, node *blockNode, view *UtxoViewpoint) error {
	ht := block.Height()
	// if ht > 100000 {
	// 	if dbCS != nil {
	// 		if err := dbCS.Close(); err != nil {
	// 			fmt.Printf("failed to close dbCS: %s, err\n", err)
	// 		}
	// 	}
	// 	os.Exit(1)
	// }
	var chgs []*change.Change
	for _, tx := range block.Transactions() {
		h := handler{ht, tx, view, map[string]bool{}, nil}
		if err := h.handleTxIns(b.claimTrie); err != nil {
			return err
		}
		if err := h.handleTxOuts(b.claimTrie); err != nil {
			return err
		}
		chgs = append(chgs, h.chgs...)
	}

	b.claimTrie.Commit(claim.Height(ht))
	hash := b.claimTrie.MerkleHash()
	if len(chgs) > 0 {
		key := strconv.Itoa(int(ht))
		blk := change.Block{
			Hash:    node.claimTrie,
			Changes: chgs,
		}
		buf := bytes.NewBuffer(nil)
		if err := gob.NewEncoder(buf).Encode(blk); err != nil {
			return errors.Wrapf(err, "gob.Encode()", err)
		}
		if err := dbCS.Put([]byte(key), buf.Bytes(), nil); err != nil {
			return errors.Wrapf(err, "dbCS.Put(%s)", key)
		}
	}
	if node.claimTrie != *hash {
		return fmt.Errorf("height: %d, ct.MerkleHash: %s != node.ClaimTrie: %s", ht, *hash, node.claimTrie)
	}
	return nil
}

type handler struct {
	ht    int32
	tx    *btcutil.Tx
	view  *UtxoViewpoint
	spent map[string]bool
	chgs  []*change.Change
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
		} else if err != nil {
			return err
		}
		chg := &change.Change{
			Height: claim.Height(h.ht),
			Name:   string(cs.Name()),
			OP:     claim.OutPoint{OutPoint: op},
			Amt:    claim.Amount(e.Amount()),
			Value:  cs.Value(),
		}

		switch cs.Opcode() {
		case txscript.OP_CLAIMNAME:
			chg.Cmd = change.SpendClaim
			chg.ID = claim.NewID(chg.OP)
			h.spent[chg.ID.String()] = true
			err = ct.SpendClaim(chg.Name, chg.OP)
		case txscript.OP_UPDATECLAIM:
			chg.Cmd = change.SpendClaim
			copy(chg.ID[:], cs.ClaimID())
			h.spent[chg.ID.String()] = true
			err = ct.SpendClaim(chg.Name, chg.OP)
		case txscript.OP_SUPPORTCLAIM:
			chg.Cmd = change.SpendSupport
			copy(chg.ID[:], cs.ClaimID())
			err = ct.SpendSupport(chg.Name, chg.OP)
		}
		if err != nil {
			return errors.Wrapf(err, "handleTxIns: %s", chg)
		}
		h.chgs = append(h.chgs, chg)
		// fmt.Printf("%s\n", chg)
	}
	return nil
}

func (h *handler) handleTxOuts(ct *claimtrie.ClaimTrie) error {
	for i, txOut := range h.tx.MsgTx().TxOut {
		op := wire.NewOutPoint(h.tx.Hash(), uint32(i))
		cs, err := txscript.DecodeClaimScript(txOut.PkScript)
		if err == txscript.ErrNotClaimScript {
			continue
		} else if err != nil {
			return err
		}
		chg := &change.Change{
			Height: claim.Height(h.ht),
			Name:   string(cs.Name()),
			OP:     claim.OutPoint{OutPoint: *op},
			Amt:    claim.Amount(txOut.Value),
			Value:  cs.Value(),
		}

		switch cs.Opcode() {
		case txscript.OP_CLAIMNAME:
			chg.Cmd = change.AddClaim
			chg.ID = claim.NewID(chg.OP)
			err = ct.AddClaim(chg.Name, chg.OP, chg.Amt, chg.Value)
		case txscript.OP_SUPPORTCLAIM:
			chg.Cmd = change.AddSupport
			copy(chg.ID[:], cs.ClaimID())
			err = ct.AddSupport(chg.Name, chg.OP, chg.Amt, chg.ID)
		case txscript.OP_UPDATECLAIM:
			chg.Cmd = change.UpdateClaim
			copy(chg.ID[:], cs.ClaimID())
			if !h.spent[chg.ID.String()] {
				fmt.Printf("%d can't find id: %s\n", h.ht, chg.ID)
				continue
			}
			delete(h.spent, chg.ID.String())
			err = ct.UpdateClaim(chg.Name, chg.OP, chg.Amt, chg.ID, chg.Value)
		}
		if err != nil {
			return errors.Wrapf(err, "handleTxOuts: %s", chg)
		}
		h.chgs = append(h.chgs, chg)
		// fmt.Printf("%s\n", chg)
	}
	return nil
}
