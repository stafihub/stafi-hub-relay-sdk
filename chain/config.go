package chain

type ConfigOption struct {
	BlockstorePath string `json:"blockstorePath"`
	StartBlock     int    `json:"startBlock"`
	ChainID        string `json:"chainId"`
	Denom          string `json:"denom"`
	GasPrice       string `json:"gasPrice"`
	Account        string `json:"account"`
}
