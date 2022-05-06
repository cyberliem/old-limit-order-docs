package sign_test

import (
	"fmt"
	"log"
	"math/big"
	"testing"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

//signWithVValue function sign a data with ethereum signed message prefix
//in addition to that, it also append v value from ECDSA signature to the returned signature
func signWithVValue(data string, pk string) ([]byte, error) {
	dataWithPrefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	hash := crypto.Keccak256Hash([]byte(dataWithPrefix))
	log.Printf("Hash is %x", hash)
	key, err := crypto.HexToECDSA(pk)
	if err != nil {
		return nil, err
	}
	signed, err := crypto.Sign(hash.Bytes(), key)
	if err != nil {
		return nil, err
	}
	_, _, v, err := signatureValues(signed)
	if err != nil {
		return nil, err
	}
	withVValue := append(signed[:len(signed)-1], v.Bytes()...)
	return withVValue, nil
}

func signatureValues(sig []byte) (r, s, v *big.Int, err error) {
	if len(sig) != 65 {
		panic(fmt.Sprintf("wrong size for signature: got %d, want 65", len(sig)))
	}
	r = new(big.Int).SetBytes(sig[:32])
	s = new(big.Int).SetBytes(sig[32:64])
	v = new(big.Int).SetBytes([]byte{sig[64] + 27})
	return r, s, v, nil
}

func to32LengthByteArr(bts []byte) []byte {
	var result = make([]byte, 32)
	copy(result[len(result)-len(bts):], bts)
	return result
}

func TestSignKyberOrder(t *testing.T) {
	// test data
	const (
		expectedHash = "0x749abf05e4127e63d42e5ee1e0d008e00dedb658590acdd301b8e47be9e0e655"
		expectedSign = "0x6ee278f2037da07a9c879ba604edfb5824e7b75cc74fb4b0076c98d8dfcf219357f4d7d610c36d20cc7280c06de0b1c2997f9840dfc81c9af5f271af00107d3e1b"
		pkS          = "275bc23940a2061ecf0fa34341c0ca2b5d7b5e961032965610fbfda72b0572b7"
		nonceStr     = "0x7fd3e50013e911e7c479a10b8525728f00000000000000000000016afd268cd7"
	)

	var (
		user        = ethereum.HexToAddress("0xe122cd8d3d09271d1e999f766b19ada8d06b8ee9")
		sourceToken = ethereum.HexToAddress("0xbCA556c912754Bc8E7D4Aad20Ad69a1B1444F42d")
		srcAmount   = big.NewInt(0).SetUint64(50000000000000000)
		destToken   = ethereum.HexToAddress("0x4E470dc7321E84CA96FcAEDD0C8aBCebbAEB68C6")
		destAddress = ethereum.HexToAddress("0xe122cd8d3d09271d1e999f766b19ada8d06b8ee9")
		minRate     = big.NewInt(0)
		fee         = big.NewInt(10000)
		// reference tx: https://ropsten.etherscan.io/tx/0x35d7e8ed1ac25c8d1562d764d3f8aad131374cf84313620cf42d75db1346d284
	)

	nonce, err := hexutil.DecodeBig(nonceStr)
	assert.NoError(t, err, "nonce must be a number")

	hashMsg := crypto.Keccak256(
		user.Bytes(),
		nonce.Bytes(),
		sourceToken.Bytes(),
		to32LengthByteArr(srcAmount.Bytes()),
		destToken.Bytes(),
		destAddress.Bytes(),
		to32LengthByteArr(minRate.Bytes()),
		to32LengthByteArr(fee.Bytes()))
	log.Printf("hashMsg: %s", hexutil.Encode(hashMsg))

	assert.Equal(t, expectedHash, hexutil.Encode(hashMsg), "hash msg is not the same")
	sign, err := signWithVValue(string(hashMsg), pkS)
	assert.NoError(t, err, "must return signature")

	assert.Equal(t, expectedSign, hexutil.Encode(sign), "signature must be the same")
}

func TestSignSimpleData(t *testing.T) {
	var (
		web3TestSignMessage = "0xb91467e570a6466aa9e9876cbcd013baba02900b8979d43fe208a4a4f339f5fd6007e74cd82e037b800186422fc2da167c747ef045e5d18a5f5d4300f8e1a0291c"
		web3SimpleData      = "Some data"
		web3Pk              = "4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	)

	//We must make sure that signing by our program will produce the same result as the
	//web3's doc https://web3js.readthedocs.io/en/1.0/web3-eth-accounts.html#sign
	signedSimpleData, err := signWithVValue(web3SimpleData, web3Pk)
	assert.NoError(t, err)
	assert.Equal(t, web3TestSignMessage, hexutil.Encode(signedSimpleData))
}
