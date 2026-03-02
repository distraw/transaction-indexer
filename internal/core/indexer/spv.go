package indexer

import (
	"math/big"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/pkg/errors"
)

func bitsToTarget(bits uint32) (*big.Int, error) {
	if bits&0x00800000 != 0 {
		return nil, errors.New("invalid bits (negative target)")
	}

	exponent := bits >> 24
	coefficient := bits & 0x007fffff

	if exponent <= 3 {
		return nil, errors.New("invalid bits (exponent is less than 3)")
	}
	if coefficient == 0 {
		return nil, errors.New("invalid bits (coefficient is zero)")
	}

	target := new(big.Int).SetUint64(uint64(coefficient))
	shift := 8 * (int(exponent) - 3)
	target.Lsh(target, uint(shift))

	return target, nil
}

func reverseHash(hash chainhash.Hash) chainhash.Hash {
	for i := 0; i < len(hash)/2; i++ {
		hash[i], hash[len(hash)-1-i] =
			hash[len(hash)-1-i], hash[i]
	}

	return hash
}

func isDifficultyMet(bits uint32, blockHash chainhash.Hash) (bool, error) {
	target, err := bitsToTarget(bits)
	if err != nil {
		return false, errors.Wrap(err, "failed to convert block bits to target")
	}

	bigEndianHash := reverseHash(blockHash)

	hashInt := new(big.Int).SetBytes(bigEndianHash[:])

	return hashInt.Cmp(target) <= 0, nil
}

func (i *indexer) validatePreviousHash(header *wire.BlockHeader) error {
	headerHash := header.BlockHash()

	if headerHash.IsEqual(i.netParams.GenesisHash) && i.initialBlockHash.IsEqual(&chainhash.Hash{}) {
		return nil
	}

	// Starting point, no local blocks before initial height
	if headerHash.IsEqual(i.initialBlockHash) {
		return nil
	}

	exists, err := i.storage.Blocks().Exists(header.PrevBlock.String())
	if err != nil {
		return errors.Wrap(err, "failed to check block existence with given hash in db")
	}
	if !exists {
		return errors.New("previous hash is" + header.PrevBlock.String() + "not tracked in local best chain")
	}

	return nil
}

// validateBlock checks if all SPV requirements are met for the block.
//
// Returns nil if block is valid or error otherwise
func (i *indexer) validateBlockHeader(header *wire.BlockHeader) error {
	err := i.validatePreviousHash(header)
	if err != nil {
		i.log.Infof("prevHash=%s, initialHash=%s", header.PrevBlock, i.initialBlockHash)
		return errors.Wrap(err, "failed to validate previous hash")
	}

	met, err := isDifficultyMet(header.Bits, header.BlockHash())
	if err != nil {
		return errors.Wrap(err, "failed to check if difficulty is met in the block")
	}
	if !met {
		return errors.New("block difficulty is not met")
	}

	return nil
}
