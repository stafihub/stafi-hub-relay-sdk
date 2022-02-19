package chain

type ConfigOption struct {
	BlockstorePath string `json:"blockstorePath"`
	StartBlock     int    `json:"startBlock"`
	Account        string `json:"account"`
	ChainID        string `json:"chainId"`
	Denom          string `json:"denom"`
	GasPrice       string `json:"gasPrice"`
	CaredSymbol    string `json:"caredSymbol"`
}
