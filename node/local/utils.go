package local

import (
	"fmt"
	"time"

	cfg "github.com/cometbft/cometbft/config"
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/viper"

	"github.com/forbole/juno/v5/types"
)

func ParseConfig() (*cfg.Config, error) {
	conf := cfg.DefaultConfig()
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}
	conf.SetRoot(conf.RootDir)

	err = conf.ValidateBasic()
	if err != nil {
		return nil, fmt.Errorf("error in config file: %v", err)
	}

	return conf, nil
}

// Deprecated: this interface is used only internally for scenario we are
// deprecating (StdTxConfig support)
type intoAny interface {
	AsAny() *codectypes.Any
}

func makeTxResult(txConfig client.TxConfig, resTx *tmctypes.ResultTx, resBlock *tmctypes.ResultBlock) (*sdk.TxResponse, error) {
	txb, err := txConfig.TxDecoder()(resTx.Tx)
	if err != nil {
		return nil, err
	}
	p, ok := txb.(intoAny)
	if !ok {
		return nil, fmt.Errorf("expecting a type implementing intoAny, got: %T", txb)
	}
	any := p.AsAny()
	return sdk.NewResponseResultTx(resTx, any, resBlock.Block.Time.Format(time.RFC3339)), nil
}

// -------------------------------------------------------------------------------------------------------------------

// NewTxResponseFromSdkTxResponse allows to build a new TxResponse instance from the given sdk.TxResponse
func NewTxResponseFromSdkTxResponse(txResponse *sdk.TxResponse, tx *types.Tx) *types.TxResponse {
	return &types.TxResponse{
		TxResponse: txResponse,
		Tx:         tx,
		Height:     uint64(txResponse.Height),
		GasWanted:  uint64(txResponse.GasWanted),
		GasUsed:    uint64(txResponse.GasUsed),
	}
}

// NewTxFromSdkTx allows to build a new Tx instance from the given tx.Tx
func NewTxFromSdkTx(cdc codec.Codec, tx *tx.Tx) *types.Tx {
	return &types.Tx{
		Tx:       tx,
		Body:     NewTxBodyFromSdkTxBody(cdc, tx.Body),
		AuthInfo: NewAuthInfoFromSdkAuthInfo(cdc, tx.AuthInfo),
	}
}

// NewTxBodyFromSdkTxBody allows to build a new TxBody instance from the given tx.TxBody
func NewTxBodyFromSdkTxBody(cdc codec.Codec, body *tx.TxBody) *types.TxBody {
	messages := make([]types.Message, len(body.Messages))
	for i, msg := range body.Messages {
		err := cdc.UnpackAny(msg, &messages[i])
		if err != nil {
			panic(fmt.Errorf("error while unpacking message: %s", err))
		}

		messages[i] = types.NewStandardMessage(i, proto.MessageName(msg), cdc.MustMarshalJSON(msg))
	}
	return &types.TxBody{
		TxBody:        body,
		TimeoutHeight: uint64(body.TimeoutHeight),
		Messages:      messages,
	}
}

// NewAuthInfoFromSdkAuthInfo allows to build a new AuthInfo instance from the given tx.AuthInfo
func NewAuthInfoFromSdkAuthInfo(cdc codec.Codec, authInfo *tx.AuthInfo) *types.AuthInfo {
	signerInfos := make([]*types.SignerInfo, len(authInfo.SignerInfos))
	for i, si := range authInfo.SignerInfos {
		signerInfos[i] = NewSignerInfoFromSdkSignerInfo(cdc, si)
	}

	return &types.AuthInfo{
		AuthInfo:    authInfo,
		SignerInfos: signerInfos,
		Fee:         NewFeeFromSdkFee(authInfo.Fee),
	}
}

// NewSignerInfoFromSdkSignerInfo allows to build a new SignerInfo instance from the given tx.SignerInfo
func NewSignerInfoFromSdkSignerInfo(cdc codec.Codec, signerInfo *tx.SignerInfo) *types.SignerInfo {
	return &types.SignerInfo{
		SignerInfo: signerInfo,
		PublicKey:  cdc.MustMarshalJSON(signerInfo.PublicKey),
		Sequence:   uint64(signerInfo.Sequence),
	}
}

// NewFeeFromSdkFee allows to build a new Fee instance from the given tx.Fee
func NewFeeFromSdkFee(fee *tx.Fee) *types.Fee {
	return &types.Fee{
		Fee:      fee,
		GasLimit: uint64(fee.GasLimit),
	}
}
