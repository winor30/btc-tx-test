package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
}

func checkErrorMsg(err error, msg string) {
	if err != nil {
		log.Fatalln(msg)
		os.Exit(1)
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln(err)
		return
	}

	host := os.Getenv("host")
	user := os.Getenv("user")
	pass := os.Getenv("pass")

	// Connect to local bitcoin core RPC server using HTTP POST mode.
	connCfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown()

	prevtxid := "4eb8629ffb3bdf1035951d6df78fdb0bf5770a1b6b5744995ad593a52b8c2dc3"
	utxoid := uint32(0)
	// P2PKH address
	receiver := "mrm6soHe9svDVh7YzjtSY26PbGXSBp8eDA"
	privkey := os.Getenv("privkey")
	const value int64 = 4500000

	// Generate new tx
	msgTx := wire.NewMsgTx(wire.TxVersion)

	// Add tx input
	// prev tx
	prevtxhash, err := chainhash.NewHashFromStr(prevtxid)
	checkError(err)
	op := wire.NewOutPoint(prevtxhash, utxoid)
	txin := wire.NewTxIn(op, []byte{}, [][]byte{})
	msgTx.AddTxIn(txin)

	// Add tx output
	// value
	addr, err := btcutil.DecodeAddress(receiver, &chaincfg.TestNet3Params)
	checkError(err)
	// generate locking script consider by address type
	pkscript, err := txscript.PayToAddrScript(addr)
	checkError(err)
	txout := wire.NewTxOut(value, pkscript)
	msgTx.AddTxOut(txout)

	// Generate sign
	prevtx, err := client.GetRawTransaction(prevtxhash)
	checkError(err)
	privKeyBytes := base58.Decode(privkey)
	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)
	lookupKey := func(addres btcutil.Address) (*btcec.PrivateKey, bool, error) {
		return privKey, true, nil
	}
	sigScript, err := txscript.SignTxOutput(&chaincfg.TestNet3Params, prevtx.MsgTx(), 0, pkscript, txscript.SigHashAll, txscript.KeyClosure(lookupKey), nil, nil)
	checkError(err)
	msgTx.TxIn[0].SignatureScript = sigScript

	// show tx
	// bitcoin-cli decoderawtransaction <result of hex.EncodeToString(buf.Bytes()))>
	buf := new(bytes.Buffer)
	checkError(msgTx.Serialize(buf))
	fmt.Println(hex.EncodeToString(buf.Bytes()))
}
