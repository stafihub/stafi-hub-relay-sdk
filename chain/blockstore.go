package chain

import (
	"github.com/stafihub/rtoken-relay-core/common/utils"
)

func NewBlockstore(path string, relayer string) (*utils.Blockstore, error) {
	return utils.NewBlockstore(path, 0, relayer)
}

func StartBlock(bs *utils.Blockstore, blk uint64) (uint64, error) {
	return checkBlockstore(bs, blk)
}

// checkBlockstore queries the blockstore for the latest known block. If the latest block is
// greater than startBlock, then the latest block is returned, otherwise startBlock is.
func checkBlockstore(bs *utils.Blockstore, startBlock uint64) (uint64, error) {
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
