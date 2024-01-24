package core

import (
	"fmt"
	"github.com/barreleye-labs/barreleye/common"
	types2 "github.com/barreleye-labs/barreleye/core/types"
	"testing"

	"github.com/barreleye-labs/barreleye/crypto"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
)

func TestSendNativeTransferTamper(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	signer := crypto.GeneratePrivateKey()

	block := types2.randomBlock(t, uint32(1), getPrevBlockHash(t, bc, uint32(1)))
	assert.Nil(t, block.Sign(signer))

	privKeyBob := crypto.GeneratePrivateKey()
	privKeyAlice := crypto.GeneratePrivateKey()
	amount := uint64(100)

	accountBob := bc.accountState.CreateAccount(privKeyBob.PublicKey().Address())
	accountBob.Balance = amount

	tx := types2.NewTransaction([]byte{})
	tx.From = privKeyBob.PublicKey()
	tx.To = privKeyAlice.PublicKey()
	tx.Value = amount
	tx.Sign(privKeyBob)
	tx.hash = common.Hash{}

	hackerPrivKey := crypto.GeneratePrivateKey()
	tx.To = hackerPrivKey.PublicKey()

	block.AddTransaction(tx)
	assert.NotNil(t, bc.AddBlock(block)) // this should fail

	_, err := bc.accountState.GetAccount(hackerPrivKey.PublicKey().Address())
	assert.Equal(t, err, ErrAccountNotFound)
}

func TestSendNativeTransferInsuffientBalance(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	signer := crypto.GeneratePrivateKey()

	block := types2.randomBlock(t, uint32(1), getPrevBlockHash(t, bc, uint32(1)))
	assert.Nil(t, block.Sign(signer))

	privKeyBob := crypto.GeneratePrivateKey()
	privKeyAlice := crypto.GeneratePrivateKey()
	amount := uint64(100)

	accountBob := bc.accountState.CreateAccount(privKeyBob.PublicKey().Address())
	accountBob.Balance = uint64(99)

	tx := types2.NewTransaction([]byte{})
	tx.From = privKeyBob.PublicKey()
	tx.To = privKeyAlice.PublicKey()
	tx.Value = amount
	tx.Sign(privKeyBob)
	tx.hash = common.Hash{}

	fmt.Printf("alice => %s\n", privKeyAlice.PublicKey().Address())
	fmt.Printf("bob => %s\n", privKeyBob.PublicKey().Address())

	block.AddTransaction(tx)
	assert.Nil(t, bc.AddBlock(block))

	_, err := bc.accountState.GetAccount(privKeyAlice.PublicKey().Address())
	assert.NotNil(t, err)

	hash := tx.Hash(types2.TxHasher{})
	_, err = bc.GetTxByHash(hash)
	assert.NotNil(t, err)
}

func TestSendNativeTransferSuccess(t *testing.T) {
	bc := newBlockchainWithGenesis(t)

	signer := crypto.GeneratePrivateKey()

	block := types2.randomBlock(t, uint32(1), getPrevBlockHash(t, bc, uint32(1)))
	assert.Nil(t, block.Sign(signer))

	privKeyBob := crypto.GeneratePrivateKey()
	privKeyAlice := crypto.GeneratePrivateKey()
	amount := uint64(100)

	accountBob := bc.accountState.CreateAccount(privKeyBob.PublicKey().Address())
	accountBob.Balance = amount

	tx := types2.NewTransaction([]byte{})
	tx.From = privKeyBob.PublicKey()
	tx.To = privKeyAlice.PublicKey()
	tx.Value = amount
	tx.Sign(privKeyBob)
	block.AddTransaction(tx)

	assert.Nil(t, bc.AddBlock(block))

	accountAlice, err := bc.accountState.GetAccount(privKeyAlice.PublicKey().Address())
	assert.Nil(t, err)
	assert.Equal(t, amount, accountAlice.Balance)
}

func TestAddBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)

	lenBlocks := 1000
	for i := 0; i < lenBlocks; i++ {
		block := types2.randomBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		assert.Nil(t, bc.AddBlock(block))
	}

	assert.Equal(t, bc.Height(), uint32(lenBlocks))
	assert.Equal(t, len(bc.headers), lenBlocks+1)
	assert.NotNil(t, bc.AddBlock(types2.randomBlock(t, 89, common.Hash{})))
}

func TestNewBlockchain(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	assert.NotNil(t, bc.validator)
	assert.Equal(t, bc.Height(), uint32(0))
}

func TestHasBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	assert.True(t, bc.HasBlock(0))
	assert.False(t, bc.HasBlock(1))
	assert.False(t, bc.HasBlock(100))
}

func TestGetBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	lenBlocks := 100

	for i := 0; i < lenBlocks; i++ {
		block := types2.randomBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		assert.Nil(t, bc.AddBlock(block))

		fetchedBlock, err := bc.GetBlock(block.Height)
		assert.Nil(t, err)
		assert.Equal(t, fetchedBlock, block)
	}
}

func TestGetHeader(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	lenBlocks := 1000

	for i := 0; i < lenBlocks; i++ {
		block := types2.randomBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		assert.Nil(t, bc.AddBlock(block))
		header, err := bc.GetHeader(block.Height)
		assert.Nil(t, err)
		assert.Equal(t, header, block.Header)
	}
}

func TestAddBlockToHigh(t *testing.T) {
	bc := newBlockchainWithGenesis(t)

	assert.Nil(t, bc.AddBlock(types2.randomBlock(t, 1, getPrevBlockHash(t, bc, uint32(1)))))
	assert.NotNil(t, bc.AddBlock(types2.randomBlock(t, 3, common.Hash{})))
}

func newBlockchainWithGenesis(t *testing.T) *Blockchain {
	bc, err := NewBlockchain(log.NewNopLogger(), types2.randomBlock(t, 0, common.Hash{}))
	assert.Nil(t, err)

	return bc
}

func getPrevBlockHash(t *testing.T, bc *Blockchain, height uint32) common.Hash {
	prevHeader, err := bc.GetHeader(height - 1)
	assert.Nil(t, err)
	return types2.BlockHasher{}.Hash(prevHeader)
}
