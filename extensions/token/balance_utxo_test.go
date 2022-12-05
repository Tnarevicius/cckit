package token_test

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/hyperledger-labs/cckit/extensions/token"
	"github.com/hyperledger-labs/cckit/identity"
	"github.com/hyperledger-labs/cckit/identity/testdata"
	"github.com/hyperledger-labs/cckit/router"
	testcc "github.com/hyperledger-labs/cckit/testing"
)

func TestBalance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Balance test suite")
}

var (
	ownerIdentity = testdata.Certificates[0].MustIdentity(testdata.DefaultMSP)
	user1Identity = testdata.Certificates[1].MustIdentity(testdata.DefaultMSP)
	user2Identity = testdata.Certificates[2].MustIdentity(testdata.DefaultMSP)

	ownerAddress = base64.StdEncoding.EncodeToString(identity.MarshalPublicKey(ownerIdentity.Cert.PublicKey))
	user1Address = base64.StdEncoding.EncodeToString(identity.MarshalPublicKey(user1Identity.Cert.PublicKey))
	user2Address = base64.StdEncoding.EncodeToString(identity.MarshalPublicKey(user2Identity.Cert.PublicKey))

	Symbol = `AA`
	Group  = `001`

	// TotalSupply = uint64(10000000)
	TotalSupply   = big.NewInt(1000)
	TotalSupplyX2 = big.NewInt(2000)
	Int50         = big.NewInt(50)
	Int100        = big.NewInt(100)
	Int150        = big.NewInt(150)
	Int200        = big.NewInt(200)
	Int300        = big.NewInt(300)
	Int0          = big.NewInt(0)
)

type Wallet struct {
	cc      *testcc.TxHandler
	ctx     router.Context   // wallet storage here
	store   *token.UTXOStore //  balance access
	address string
	symbol  string
	lockId  *token.LockId
}

func NewWallet(cc *testcc.TxHandler, ctx router.Context, store *token.UTXOStore, address, symbol string) *Wallet {
	return &Wallet{
		cc:      cc,
		ctx:     ctx,
		store:   store,
		address: address,
		symbol:  symbol,
	}
}

func (w *Wallet) ExpectBalance(amount string) {
	b, err := w.store.Get(w.ctx, &token.BalanceId{
		Address: w.address,
		Symbol:  w.symbol,
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(b.Amount).To(Equal(amount))
}

func (w *Wallet) ExpectMint(amount string) {
	w.cc.Tx(func() {
		err := w.store.Mint(w.ctx, &token.BalanceOperation{
			Address: w.address,
			Symbol:  w.symbol,
			Amount:  amount,
		})
		Expect(err).NotTo(HaveOccurred())
	})
}

func (w *Wallet) ExpectBurn(amount string) {
	w.cc.Tx(func() {
		err := w.store.Burn(w.ctx, &token.BalanceOperation{
			Address: w.address,
			Symbol:  w.symbol,
			Amount:  amount,
		})
		Expect(err).NotTo(HaveOccurred())
	})
}

func (w *Wallet) ExpectBurnLock() {
	w.cc.Tx(func() {
		err := w.store.BurnLock(w.ctx, w.lockId)
		Expect(err).NotTo(HaveOccurred())
	})
}

func (w *Wallet) ExpectBurnAllLock() {
	w.cc.Tx(func() {
		err := w.store.BurnAllLock(w.ctx, &token.BalanceOperation{Symbol: w.symbol})
		Expect(err).NotTo(HaveOccurred())
	})
}

func (w *Wallet) ExpectLockAll() {
	w.cc.Tx(func() {
		err := w.store.LockAll(w.ctx, &token.BalanceOperation{Symbol: w.symbol})
		Expect(err).NotTo(HaveOccurred())
	})
}

func (w *Wallet) ExpectTransfer(recipient string, amount string) {
	w.cc.Tx(func() {
		err := w.store.Transfer(w.ctx, &token.TransferOperation{
			Sender:    w.address,
			Recipient: recipient,
			Symbol:    w.symbol,
			Amount:    amount,
		})
		Expect(err).NotTo(HaveOccurred())
	})
}

func (w *Wallet) ExpectNotTransfer(recipient string, amount string) {
	w.cc.Tx(func() {
		err := w.store.Transfer(w.ctx, &token.TransferOperation{
			Sender:    w.address,
			Recipient: recipient,
			Symbol:    w.symbol,
			Amount:    amount,
		})
		Expect(err).To(HaveOccurred())
	})
}

func (w *Wallet) ExpectLock(amount string) {
	w.cc.Tx(func() {
		lockId, err := w.store.Lock(w.ctx, &token.BalanceOperation{
			Address: w.address,
			Symbol:  w.symbol,
			Amount:  amount,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(lockId.Address).To(Equal(w.address))
		Expect(lockId.Symbol).To(Equal(w.symbol))
		Expect(lockId.TxId).NotTo(BeZero())
		w.lockId = lockId
	})
}

func (w *Wallet) ExpectUnlock() {
	w.cc.Tx(func() {
		err := w.store.Unlock(w.ctx, w.lockId)
		Expect(err).NotTo(HaveOccurred())
	})
}

func (w *Wallet) ExpectLockedBalance(amount string) {
	b, err := w.store.GetLocked(w.ctx, &token.BalanceId{
		Symbol:  w.symbol,
		Address: w.address,
	})

	Expect(err).NotTo(HaveOccurred())
	Expect(b.Amount).To(Equal(amount))
}

type transfer struct {
	recipient string
	amount    string
}

func (w *Wallet) ExpectTransferBatch(transfers []*transfer) {

	var transferOperations []*token.TransferOperation

	for _, t := range transfers {
		transferOperations = append(transferOperations, &token.TransferOperation{
			Sender:    w.address,
			Recipient: t.recipient,
			Symbol:    w.symbol,
			Amount:    t.amount,
		})
	}
	w.cc.Tx(func() {
		err := w.store.TransferBatch(w.ctx, transferOperations)
		Expect(err).NotTo(HaveOccurred())
	})
}

func (w *Wallet) ExpectOutputsNum(num int) {
	outputs, err := w.store.ListOutputs(w.ctx, &token.BalanceId{
		Address: w.address,
		Symbol:  w.symbol,
	})

	Expect(err).NotTo(HaveOccurred())
	Expect(len(outputs)).To(Equal(num))
}

var _ = Describe(`UTXO store`, func() {

	cc, ctx := testcc.NewTxHandler(`UTXO`)
	utxo := token.NewUTXOStore()
	ownerWallet := NewWallet(cc, ctx, utxo, ownerAddress, Symbol)
	user1Wallet := NewWallet(cc, ctx, utxo, user1Address, Symbol)
	user2Wallet := NewWallet(cc, ctx, utxo, user2Address, Symbol)

	fmt.Println("ownerAddress", ownerAddress)
	fmt.Println("user1Address", user1Address)
	fmt.Println("user2Address", user2Address)

	It(`allow to get empty balance`, func() {
		ownerWallet.ExpectBalance(Int0.String())
	})

	It(`allow to mint balance`, func() {
		ownerWallet.ExpectMint(TotalSupply.String())
		ownerWallet.ExpectBalance(TotalSupply.String())
		ownerWallet.ExpectOutputsNum(1)
	})

	It(`allow to mint balance once more time`, func() {
		ownerWallet.ExpectMint(TotalSupply.String())
		ownerWallet.ExpectBalance(TotalSupplyX2.String())
		ownerWallet.ExpectOutputsNum(2)
	})

	It(`allow to partially transfer balance`, func() {
		ownerWallet.ExpectTransfer(user1Address, Int100.String())
		ownerWallet.ExpectBalance(new(big.Int).Sub(TotalSupplyX2, Int100).String())
		ownerWallet.ExpectOutputsNum(2)

		user1Wallet.ExpectBalance(big.NewInt(100).String())
		user1Wallet.ExpectOutputsNum(1)
	})

	It(`allow to return all amount back`, func() {
		user1Wallet.ExpectTransfer(ownerAddress, Int100.String())
		ownerWallet.ExpectBalance(TotalSupplyX2.String())
		ownerWallet.ExpectOutputsNum(3)

		user1Wallet.ExpectBalance(Int0.String())
		user1Wallet.ExpectOutputsNum(0)
	})

	It(`allow to burn`, func() {
		ownerWallet.ExpectBurn(TotalSupply.String())
		ownerWallet.ExpectBalance(TotalSupply.String())
		//ownerWallet.ExpectOutputsNum(2)
	})

	It(`allow to transfer batch`, func() {
		ownerWallet.ExpectTransferBatch([]*transfer{
			{recipient: user1Address, amount: Int100.String()},
			{recipient: user2Address, amount: Int200.String()},
		})
		ownerWallet.ExpectBalance(new(big.Int).Sub(TotalSupply, Int300).String()) // must be equal TotalSupply - 100 - 200
		user1Wallet.ExpectBalance(Int100.String())
		user2Wallet.ExpectBalance(Int200.String())
		//ownerWallet.ExpectOutputsNum(2)
	})

	It(`allow to lock`, func() {
		user1Wallet.ExpectLock(Int50.String())
		user1Wallet.ExpectLockedBalance(Int50.String())
	})

	It(`disallow to transfer locked balance`, func() {
		user1Wallet.ExpectNotTransfer(ownerAddress, Int100.String())
	})

	It(`allow to unlock`, func() {
		user1Wallet.ExpectUnlock()
		user1Wallet.ExpectBalance(Int100.String())
		user1Wallet.ExpectLockedBalance(Int0.String())
	})

	It(`allow to burn locked`, func() {
		user2Wallet.ExpectLock(Int50.String())
		user2Wallet.ExpectBurnLock()
		user2Wallet.ExpectLockedBalance(Int0.String())
		user2Wallet.ExpectBalance(Int150.String())
	})

	It(`allow to burn all locked`, func() {
		user1Wallet.ExpectLock(Int50.String())
		user2Wallet.ExpectLock(Int50.String())
		user2Wallet.ExpectBurnAllLock()
		user1Wallet.ExpectLockedBalance(Int0.String())
		user2Wallet.ExpectLockedBalance(Int0.String())
		user1Wallet.ExpectBalance(Int50.String())
		user2Wallet.ExpectBalance(Int100.String())
	})

	It(`allow to lock all`, func() {
		ownerWallet.ExpectLockAll()
		ownerWallet.ExpectBalance(new(big.Int).Sub(TotalSupply, Int300).String()) // must be equal TotalSupply - 100 - 200
		user1Wallet.ExpectLockedBalance(Int50.String())
		user2Wallet.ExpectLockedBalance(Int100.String())
	})

})