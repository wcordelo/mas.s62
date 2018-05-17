package main

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"math"
	"strconv"
	"encoding/json"
	"strings"
)

var (
	testnet3Parameters = &chaincfg.TestNet3Params
)

//App Entrance
func main() {
	// Test Data is used as a template for fingerprint data
	fmt.Printf("KeyChain\n")
	testData := []float64{12.3434, 15.9090, 10.43434, 0.0345, 0.004, 0.132, 0.454, 34.343}

	// Private Key is the key that we want to encrypt in the fuzzy vault with our fingerprint
	privateKey := "WHAT'S UP?"
	fmt.Println("Private key before Fuzzy Vault is: ", privateKey)

	//  Lock creates the fuzzy vault using private key and fingerprint
	vault := Lock(privateKey, testData)

	// We need to convert the vault to a string that can be put into the opt return part for Bitcoin. We parse the vault
	// into a string
	vaultString := "["
	for index, vaultRow := range vault {
		vaultString += "[" + strconv.FormatFloat(vaultRow[0], 'E', -1, 64) + "," + strconv.FormatFloat(vaultRow[1], 'E', -1, 64) + "]"
		if index < len((vault))-1 {
			vaultString += ","
		}
	}
	vaultString += "]"

	// We convert the string into a byte array that can be placed into the opt return part of Bitcoin.
	opReturnData := []byte(vaultString)

	// Because the data is too large, we first compress the opt return data so that there is less transactions made when
	// placing the data in the opt return part of Bitcoin
	// TODO: Use a Merkle Tree to hash the data so that there are less transactions made to place the fuzzy vault in Bitcoin
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(opReturnData); err != nil {
		panic(err)
	}
	if err := gz.Flush(); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}

	// We encode the compressed information to a string and convert to a compressed byte array
	compressedOpReturnData := base64.StdEncoding.EncodeToString(b.Bytes())
	compressedOpReturnDataInByte := []byte(compressedOpReturnData)

	// We need several transactions to be made because the compressed byte array is still too large to be placed in the
	// opt return part of Bitcoin. Therefore, we separate the compressed byte array to several parts which then are
	// placed in a different transaction on the Bitcoin testnet. Opt Return has a max limit of 520 bytes.
	// If a Merkle tree is used, then we can ignore this part because we would be able to place the entire hashed
	// information without having to separate the components
	var compressedVaultPieces [][]byte
	for pieceIndex := 0; pieceIndex < len(compressedOpReturnDataInByte); pieceIndex += 520 {
		piece := compressedOpReturnDataInByte[pieceIndex:int(math.Min(float64(pieceIndex+520), float64(len(compressedOpReturnDataInByte))))]
		doubleByteOfPiece := [][]byte{piece}
		compressedVaultPieces = append(compressedVaultPieces, doubleByteOfPiece...)
	}

	// We populate the transaction information with our test data. This information is similar to PSET 3 of MAS.S62.
	// Read PSET 3 for more information on how OpReturnTxBuilder works.
	publicAddress, _ := GenerateAddress("KeyChain")
	txFrom := "1f497ac245eb25cd94157c290f62d042e3bdda1e57920b6d1d2c5cfa362c12da"
	index := uint32(32)
	addressTo := publicAddress
	valueOut := int64(10000)

	// A transaction is made for each compressed vault piece so that they can individually be placed in the Bitcoin testnet.
	var transactionStrings []string
	for _, compressedVaultPiece := range compressedVaultPieces {
		optx := OpReturnTxBuilder(compressedVaultPiece, txFrom, addressTo, valueOut, index, privateKey)
		hexOpt := TxToHex(optx)
		transactionStrings = append(transactionStrings, hexOpt)
	}

	// Assuming that the vault pieces have been retrieved from the Opt Return part of the Bitcoin testnet, then we can
	// append the pieces together. In this example, two transactions were made so we append both compressed vault pieces
	// together to made a single vault piece.
	retrievedCompressedVaultPieces := append(compressedVaultPieces[0], compressedVaultPieces[1]...)
	decompressedOpReturnData0, _ := base64.StdEncoding.DecodeString(string(retrievedCompressedVaultPieces))

	// We decompress the information and revert the data back into a string format that can be decoded
	rdata0 := bytes.NewReader(decompressedOpReturnData0)
	r0, _ := gzip.NewReader(rdata0)
	decodedVaultPiece, _ := (ioutil.ReadAll(r0))
	decodedVaultString := string(decodedVaultPiece)

	// We decode the vault string into a double array that is in the format of the original fuzzy vault needed to
	// retrieve the private key.
	var vaultArray [][]float64
	dec := json.NewDecoder(strings.NewReader(string(decodedVaultString)))
	err := dec.Decode(&vaultArray)
	fmt.Println(err, vaultArray)

	// We use the decoded fuzzy vault and the test fingerprint data to unlock the fuzzy vault and obtain the coefficients
	// needed to obtain the private key. Then the coefficients are decoded and the private key is finally obtained.
	coeffs := Unlock(testData, vaultArray)
	decodedPrivateKey := Decode(coeffs)
	fmt.Println("Decoded Private Key from Fuzzy Vault: ", decodedPrivateKey)
	return
}
