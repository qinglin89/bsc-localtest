package types

//var (
//	big5   = big.NewInt(5)
//	big8   = big.NewInt(8)
//	big15  = big.NewInt(15)
//	big20  = big.NewInt(20)
//	big25  = big.NewInt(25)
//	big32  = big.NewInt(32)
//	big75  = big.NewInt(75)
//	big80  = big.NewInt(80)
//	big100 = big.NewInt(100)
//	a1     string
//	a2     string
//	a3     string
//)

type Block struct {
	Hash            string
	Number          string
	Miner           string
	Difficulty      string
	TotalDifficulty string `json:"totalDifficulty"`
	GasUsed         string `json:"gasUsed"`
	Transactions    []string
	Timestamp       string
	Uncles          []string
}

type BlockDetails struct {
	Block
	Transactions []*Transaction
}

type Transaction struct {
	Gas      string
	GasPrice string `json:"gasPrice"`
}

type RpcResponse struct {
	ID      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Error   *ErrorField
	Result  interface{}
}

type ResBlockNumber struct {
	RpcResponse
	Result string
}

type ResBlock struct {
	RpcResponse
	Result *Block
}

type ResBlockDetails struct {
	RpcResponse
	Result *BlockDetails
}

type ErrorField struct {
	Code    int
	Message string
	Data    string
}
