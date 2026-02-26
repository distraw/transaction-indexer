package indexer

import (
	"math/big"
	"slices"
	"strconv"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/pkg/errors"
)

func hashToBigEndian(rawHash string) (*big.Int, error) {
	hash, err := chainhash.NewHashFromStr(rawHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate hash from raw string")
	}

	if len(hash) != 32 {
		return nil, errors.New("hash hex is of incorrect size, must be 32 bytes long")
	}

	// Little-endian to big-endian
	slices.Reverse(hash[:])

	return new(big.Int).SetBytes(hash[:]), nil
}

func hexBitsToUint32(bitsHex string) (uint32, error) {
	bits64, err := strconv.ParseUint(bitsHex, 16, 32)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse bitsHex to uint64")
	}

	return uint32(bits64), nil
}

func bitsToTarget(bitsHex string) (*big.Int, error) {
	bits, err := hexBitsToUint32(bitsHex)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert raw bitsHex from string to uint32")
	}

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

func isDifficultyMet(bits string, blockHash string) (bool, error) {
	target, err := bitsToTarget(bits)
	if err != nil {
		return false, errors.Wrap(err, "failed to convert block bits to target")
	}

	bigEndianBlockHash, err := hashToBigEndian(blockHash)
	if err != nil {
		return false, errors.Wrap(err, "failed to parse block hash from string to big-endian big int")
	}

	return bigEndianBlockHash.Cmp(target) <= 0, nil
}

func (i *indexer) validatePreviousHash(header *btcjson.GetBlockHeaderVerboseResult) error {
	// Starting point, no local blocks before initial height
	if header.Hash == i.initialBlockHash.String() {
		return nil
	}

	exists, err := i.storage.Blocks().Exists(header.PreviousHash)
	if err != nil {
		return errors.Wrap(err, "failed to check block existence with given hash in db")
	}
	if !exists {
		return errors.New("previous hash is not tracked in local best chain")
	}

	return nil
}

// validateBlock checks if all SPV requirements are met for the block.
//
// Returns nil if block is valid or error otherwise
func (i *indexer) validateBlockHeader(header *btcjson.GetBlockHeaderVerboseResult) error {
	err := i.validatePreviousHash(header)
	if err != nil {
		i.log.Infof("prevHash=%s, initialHash=%s", header.PreviousHash, i.initialBlockHash.String())
		return errors.Wrap(err, "failed to validate previous hash")
	}

	met, err := isDifficultyMet(header.Bits, header.Hash)
	if err != nil {
		return errors.Wrap(err, "failed to check if difficulty is met in the block")
	}
	if !met {
		return errors.New("block difficulty is not met")
	}

	return nil
}
