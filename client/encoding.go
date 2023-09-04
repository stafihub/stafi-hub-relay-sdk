package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	feegrantModule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	interChain "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts"
	ibcTransfer "github.com/cosmos/ibc-go/v5/modules/apps/transfer"
	ibcCore "github.com/cosmos/ibc-go/v5/modules/core"
	stafiHubXBridge "github.com/stafihub/stafihub/x/bridge"
	stafiHubXClaim "github.com/stafihub/stafihub/x/claim"
	stafiHubXLedger "github.com/stafihub/stafihub/x/ledger"
	stafiHubXMining "github.com/stafihub/stafihub/x/mining"
	stafiHubXRBank "github.com/stafihub/stafihub/x/rbank"
	stafiHubXRDex "github.com/stafihub/stafihub/x/rdex"
	stafiHubXRelayer "github.com/stafihub/stafihub/x/relayers"
	stafiHubXRmintReward "github.com/stafihub/stafihub/x/rmintreward"
	stafiHubXRStaking "github.com/stafihub/stafihub/x/rstaking"
	stafiHubXRvalidator "github.com/stafihub/stafihub/x/rvalidator"
	stafiHubXRvote "github.com/stafihub/stafihub/x/rvote"
	stafiHubXSudo "github.com/stafihub/stafihub/x/sudo"
)

// EncodingConfig specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Marshaler         codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() EncodingConfig {
	encodingConfig := makeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	moduleBasics := module.NewBasicManager( //codec need
		auth.AppModuleBasic{},
		authz.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		crisis.AppModuleBasic{},
		distribution.AppModuleBasic{},
		evidence.AppModuleBasic{},
		feegrantModule.AppModuleBasic{},
		genutil.AppModuleBasic{},
		gov.AppModuleBasic{},
		mint.AppModuleBasic{},
		params.AppModuleBasic{},
		slashing.AppModuleBasic{},
		staking.AppModuleBasic{},
		upgrade.AppModuleBasic{},

		ibcTransfer.AppModuleBasic{},
		ibcCore.AppModuleBasic{},
		interChain.AppModuleBasic{},

		stafiHubXLedger.AppModuleBasic{},
		stafiHubXSudo.AppModuleBasic{},
		stafiHubXRvote.AppModuleBasic{},
		stafiHubXRelayer.AppModuleBasic{},
		stafiHubXBridge.AppModuleBasic{},
		stafiHubXRvalidator.AppModuleBasic{},
		stafiHubXRmintReward.AppModuleBasic{},
		stafiHubXRBank.AppModuleBasic{},
		stafiHubXRDex.AppModuleBasic{},
		stafiHubXRStaking.AppModuleBasic{},
		stafiHubXMining.AppModuleBasic{},
		stafiHubXClaim.AppModuleBasic{},
	)
	moduleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	moduleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	return encodingConfig
}

// MakeEncodingConfig creates an EncodingConfig for an amino based test configuration.
func makeEncodingConfig() EncodingConfig {
	amino := codec.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(marshaler, tx.DefaultSignModes)
	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          txCfg,
		Amino:             amino,
	}
}

func marshalSignatureJSON(txConfig client.TxConfig, txBldr client.TxBuilder, signatureOnly bool) ([]byte, error) {
	parsedTx := txBldr.GetTx()
	if signatureOnly {
		sigs, err := parsedTx.GetSignaturesV2()
		if err != nil {
			return nil, err
		}
		return txConfig.MarshalSignatureJSON(sigs)
	}

	return txConfig.TxJSONEncoder()(parsedTx)
}
