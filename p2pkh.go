package main

import (
	"bytes"
	"encoding/hex"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
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

func genMsgTx(txid string, utxoid uint32, receiver string, value int64, privkey string, client *rpcclient.Client) *wire.MsgTx {
	msgTx := wire.NewMsgTx(wire.TxVersion)

	// Add tx input
	// prev tx
	prevtxhash, err := chainhash.NewHashFromStr(txid)
	checkError(err)
	op := wire.NewOutPoint(prevtxhash, utxoid)
	txin := wire.NewTxIn(op, nil, nil)
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

	// generate signature
	prevtx, err := client.GetRawTransaction(prevtxhash)
	checkError(err)
	wif, _ := btcutil.DecodeWIF(privkey)
	privKey := wif.PrivKey
	sigScript, err := txscript.SignatureScript(msgTx, 0, prevtx.MsgTx().TxOut[1].PkScript, txscript.SigHashAll, privKey, wif.CompressPubKey)
	checkError(err)
	msgTx.TxIn[0].SignatureScript = sigScript

	return msgTx
}

func validate(msgTx *wire.MsgTx) bool {
	vm, err := txscript.NewEngine(nil, msgTx, 0, 0, nil, nil, -1)
	checkError(err)
	if err := vm.Execute(); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func show(msgTx *wire.MsgTx) {
	buf := new(bytes.Buffer)
	checkError(msgTx.Serialize(buf))
	log.Println(hex.EncodeToString(buf.Bytes()))
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

	p2pkh := func() {
		prevtxid := "ee2a68ba404e09ab888a3dabb6143d2ae464c534e3c3855dc2f2b8867bdb452d"
		utxoid := uint32(1)
		// P2PKH address
		receiver := "mrm6soHe9svDVh7YzjtSY26PbGXSBp8eDA"
		privkey := os.Getenv("privkey")
		const value int64 = 4500000

		// Generate new tx with sign
		msgTx := genMsgTx(prevtxid, utxoid, receiver, value, privkey, client)

		// show tx
		// bitcoin-cli decoderawtransaction <result of hex.EncodeToString(buf.Bytes()))>
		show(msgTx)

		if !validate(msgTx) {
			log.Println("invalid transaction")
			return
		}

		txh, err := client.SendRawTransaction(msgTx, true)
		checkError(err)
		log.Println(txh)
	}

	p2pkh()
}
