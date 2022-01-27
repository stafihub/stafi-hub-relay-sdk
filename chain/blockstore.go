package chain

import (
	"github.com/stafiprotocol/chainbridge/utils/blockstore"
	"github.com/stafiprotocol/chainbridge/utils/msg"
)

func NewBlockstore(path string, relayer string) (*blockstore.Blockstore, error) {
	//todo change chainId for different rToken
	return blockstore.NewBlockstore(path, msg.ChainId(000), relayer)
}

func StartBlock(bs *blockstore.Blockstore, blk uint64) (uint64, error) {
	return checkBlockstore(bs, blk)
}

// checkBlockstore queries the blockstore for the latest known block. If the latest block is
// greater than startBlock, then the latest block is returned, otherwise startBlock is.
func checkBlockstore(bs *blockstore.Blockstore, startBlock uint64) (uint64, error) {
	latestBlock, err := bs.TryLoadLatestBlock()
	if err != nil {
		return 0, err
	}

	if latestBlock.Uint64() > startBlock {
		return latestBlock.Uint64(), nil
	} else {
		return startBlock, nil
	}
}
