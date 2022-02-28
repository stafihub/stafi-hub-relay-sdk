package chain

type ConfigOption struct {
	BlockstorePath string `json:"blockstorePath"`
	StartBlock     int    `json:"startBlock"`
	Account        string `json:"account"`
	GasPrice       string `json:"gasPrice"`
	CaredSymbol    string `json:"caredSymbol"`
}
