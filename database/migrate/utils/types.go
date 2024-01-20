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

type MessageRow struct {
	TransactionHash           string `db:"transaction_hash"`
	Index                     int64  `db:"index"`
	Type                      string `db:"type"`
	Value                     string `db:"value"`
	InvolvedAccountsAddresses string `db:"involved_accounts_addresses"`
	Height                    int64  `db:"height"`
	PartitionID               int64  `db:"partition_id"`
}
