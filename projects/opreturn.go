package main

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

func OpReturnTxBuilder(opReturnData []byte, txId string, addressTo string, valueOut int64, indexId uint32, privateKey string) *wire.MsgTx {

	// this function is similar to EZTxBuilder in that it builds a transaction.
	// however, it has 2 outputs.  One is just sending money back to the same address,
	// and the other is an "op_return" output, which cannot be spent, but can
	// include arbitrary data that will be saved but ignored by the bitcoin network

	// create a new, empty transaction, set version to 2. Same as EZ
	tx := wire.NewMsgTx(2)

	// put the input txid here (your own)
	hashStr := txId

	outpointTxid, err := chainhash.NewHashFromStr(hashStr)
	if err != nil {
		panic(err)
	}
	// spend output index 0, which is the only output for that tx
	outPoint := wire.NewOutPoint(outpointTxid, indexId)

	// create the TxIn, with empty sigscript field
	input := wire.NewTxIn(outPoint, nil, nil)

	// done with inputs for now.  Build outputs.  There will be 2 outputs,
	// the OP_RETURN output and the normal pubkey hash output.
	// The OP_RETURN output will be unspendable, so we should put 0 coins there.
	// build the op_return output script
	// this is the OP_RETURN opcode, followed by a data push opcode, then the data.
	opReturnScript, err :=
		txscript.NewScriptBuilder().AddOp(txscript.OP_RETURN).AddData(opReturnData).Script()
	if err != nil {
		panic(err)
	}

	// build the op_return output.
	opReturnOutput := wire.NewTxOut(0, opReturnScript) // keep it as 0 value

	// next, build the pubkey hash output.  This the same as before in the EZ function.
	// put the address you're sending to here.  It's the same as the address you're
	// spending from!
	sendToAddressString := addressTo
	sendToAddress, err := btcutil.DecodeAddress(sendToAddressString, testnet3Parameters)
	if err != nil {
		panic(err)
	}

	sendToScript, err := txscript.PayToAddrScript(sendToAddress)
	if err != nil {
		panic(err)
	}

	// put a bit less than your input amount, so that there is a fee for the miners
	// this will ensure miners put your transaction in a block.
	p2pkhOutput := wire.NewTxOut(valueOut, sendToScript)

	// put the tx together, 1 input, 2 outputs.
	tx.AddTxIn(input)
	tx.AddTxOut(opReturnOutput)
	tx.AddTxOut(p2pkhOutput)

	// finally we need to sign.  Same as EZ func.
	// we already know the address you're sending from

	prevAddressString := "mpQQryVrYmGNPxVqNeE5RgoYAv2v66Psao" // put the address you're sending "from" here
	prevAddress, err := btcutil.DecodeAddress(prevAddressString, testnet3Parameters)
	if err != nil {
		panic(err)
	}

	spendFromScript, err := txscript.PayToAddrScript(prevAddress)

	phraseHash := chainhash.DoubleHashB([]byte(privateKey))
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), phraseHash)

	pubSig, err := txscript.SignatureScript(
		tx, 0, spendFromScript, txscript.SigHashAll, priv, true)
	if err != nil {
		panic(err)
	}

	tx.TxIn[0].SignatureScript = pubSig

	return tx
}
