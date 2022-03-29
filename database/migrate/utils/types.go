package types

type TransactionRow struct {
	Hash        string `db:"hash"`
	Height      int64  `db:"height"`
	Success     string `db:"success"`
	Messages    string `db:"messages"`
	Memo        string `db:"memo"`
	Signatures  string `db:"signatures"`
	SignerInfos string `db:"signer_infos"`
	Fee         string `db:"fee"`
	GasWanted   string `db:"gas_wanted"`
	GasUsed     string `db:"gas_used"`
	RawLog      string `db:"raw_log"`
	Logs        string `db:"logs"`
}
