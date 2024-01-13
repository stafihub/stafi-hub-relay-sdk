package client_test

import (
	"bytes"
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	// "github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	xBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/sirupsen/logrus"
	"github.com/stafihub/rtoken-relay-core/common/log"
	hubClient "github.com/stafihub/stafi-hub-relay-sdk/client"
	stafiHubXRValidatorTypes "github.com/stafihub/stafihub/x/rvalidator/types"
	"github.com/stretchr/testify/assert"
)

var client *hubClient.Client

func initClient() {
	// key, err := keyring.New(types.KeyringServiceName(), keyring.BackendFile, "/Users/tpkeeper/.stafihub", strings.NewReader("tpkeeper\n"))
	// if err != nil {
	// 	panic(err)
	// }

	var err error
	// client, err = hubClient.NewClient(nil, "", "0.005ufis", []string{"https://test-rpc1.stafihub.io:443"}, log.NewLog("client"))
	client, err = hubClient.NewClient(nil, "", "0.005ufis", []string{"https://public-rpc1.stafihub.io:443"}, log.NewLog("client"))
	// client, err = hubClient.NewClient(nil, "", "0.005ufis", []string{"https://private-rpc1.stafihub.io:443"}, log.NewLog("client"))
	// client, err = hubClient.NewClient(nil, "", "0.005ufis", []string{"https://iris-rpc1.stafihub.io:443"}, log.NewLog("client"))
	// client, err := hubClient.NewClient(key, "relay1", "0.005ufis", []string{"http://localhost:26657"})
	// client, err = hubClient.NewClient(nil, "", "", []string{"http://localhost:26657"})
	if err != nil {
		panic(err)
	}
}

func TestClient_QueryTxByHash(t *testing.T) {
	initClient()

	// res, err := client.QueryTxByHash("7AB804A2E1E28870F534FA4BAA823AF101B54E2DA95293D5BC5382DDD3579211")
	// assert.NoError(t, err)
	// for _, e := range res.Events {
	// 	t.Log("e", e.String())
	// }

	txs, err := client.GetBlockTxs(901223)
	assert.NoError(t, err)
	t.Log("----------", txs)
	for _, tx := range txs {
		if tx.TxHash == "7AB804A2E1E28870F534FA4BAA823AF101B54E2DA95293D5BC5382DDD3579211" {

			for _, log := range tx.Logs {
				for i, event := range log.Events {
					// eventIndex := log.MsgIndex*100 + uint32(i)
					// if event.Type!="transfer"{
					// 	continue
					// }
					t.Log("log.msgIndex", log.MsgIndex, "log.Log", log.Log, "eventIndex", i, "eventType", event.Type, "evetlen:", len(event.Attributes), "event", event.String())

					eventIndex := log.MsgIndex*1000000 + uint32(i*10000)

					if len(event.Attributes)%3 != 0 {
						t.Log("attribute len error")
					}

					groupLen := len(event.Attributes) / 3
					for group := 0; group < groupLen; group++ {
						cursor := group * 3
						recipient := event.Attributes[0+cursor].Value
						from := event.Attributes[1+cursor].Value
						amountStr := event.Attributes[2+cursor].Value

						coins, err := types.ParseCoinsNormalized(amountStr)
						if err != nil {
							t.Log("parsecoin err")
						}

						for coinIndex, coin := range coins {
							willUseEventIndex := eventIndex + uint32(group*100) + uint32(coinIndex)

							t.Log(cursor, coin.Denom, coinIndex, willUseEventIndex, recipient, from, amountStr)
						}
					}
				}
			}

		}
	}
}

func TestChangeEndPoint(t *testing.T) {
	initClient()
	for i := 0; i < 200; i++ {
		height, err := client.GetCurrentBlockHeight()
		if err != nil {
			t.Error(err)
		} else {

			t.Log(height)
		}

	}
}

func TestGetTxs(t *testing.T) {
	initClient()

	rs, err := client.QueryLatestLsmProposalId()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rs)
	return
	// txs, err := client.GetBlockTxs(610)
	txs, err := client.QueryTxByHash("3843CB53C4F49ECB4F4D749B6EC33B1B1475D5BEF33F4AFBD48930E1311BB9C3")
	if err != nil {
		t.Fatal(err)
	}

	// t.Log(len(txs))
	// for _, tx := range txs {
	// 	t.Log("===============")
	// 	t.Logf("%+v", tx)
	// 	for _, log := range tx.Logs {
	// 		for _, event := range log.Events {
	// 			t.Logf("%+v", event)
	// 		}
	// 	}

	// }
	tx, err := client.GetTxConfig().TxDecoder()(txs.Tx.GetValue())
	if err != nil {
		panic(err)
	}

	txJson, err := client.GetTxConfig().TxJSONEncoder()(tx)
	if err != nil {
		panic(err)
	}
	t.Log(string(txJson))
}

func TestGetBlockResults(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)
	initClient()
	// txs, err := client.GetBlockTxs(610)

	t.Log(time.Now().Unix())
	// txs, err := client.GetBlockResults(7707207)
	txs, err := client.GetBlockResults(7712205)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(time.Now().Unix())

	for _, tx := range txs.TxsResults {
		t.Log("===============")
		t.Logf("%+v", tx)
		for _, log := range tx.Log {
			t.Logf("%+v", log)
		}

	}
}

func TestGetPubKey(t *testing.T) {
	initClient()
	test, _ := types.AccAddressFromBech32("cosmos1u22lut8qgqg8znxam72pwgqp8c09rnvme00kea")
	account, _ := client.QueryAccount(test)
	t.Log(hex.EncodeToString(account.GetPubKey().Bytes()))

}

func TestClient_Sign(t *testing.T) {
	initClient()
	bts, err := hex.DecodeString("0E4F8F8FF7A3B67121711DA17FBE5AE8CB25DB272DDBF7DC0E02122947266604")
	assert.NoError(t, err)
	sigs, pubkey, err := client.Sign("recipient", bts)
	assert.NoError(t, err)
	t.Log(hex.EncodeToString(sigs))
	//4c6902bda88424923c62f95b3e3ead40769edab4ec794108d1c18994fac90d490087815823bd1a8af3d6a0271538cef4622b4b500a6253d2bd4c80d38e95aa6d
	t.Log(hex.EncodeToString(pubkey.Bytes()))
	//02e7710b4f7147c10ad90da06b69d2d6b8ff46786ef55a3f1e889c33de2bf0b416
}

func TestAddress(t *testing.T) {
	addrKey1, _ := types.AccAddressFromBech32("cosmos1a8mg9rj4nklhmwkf5vva8dvtgx4ucd9yjasret")
	addrKey2, _ := types.AccAddressFromBech32("cosmos1ztquzhpkve7szl99jkugq4l8jtpnhln76aetam")
	addrKey3, _ := types.AccAddressFromBech32("cosmos12zz2hm02sxe9f4pwt7y5q9wjhcu98vnuwmjz4x")
	addrKey4, _ := types.AccAddressFromBech32("cosmos12yprrdprzat35zhqxe2fcnn3u26gwlt6xcq0pj")
	addrKey5, _ := types.AccAddressFromBech32("cosmos1em384d8ek3y8nlugapz7p5k5skg58j66je3las")
	t.Log(hex.EncodeToString(addrKey1.Bytes()))
	t.Log(hex.EncodeToString(addrKey2.Bytes()))
	t.Log(hex.EncodeToString(addrKey3.Bytes()))
	t.Log(hex.EncodeToString(addrKey4.Bytes()))
	t.Log(hex.EncodeToString(addrKey5.Bytes()))
	//client_test.go:347: e9f6828e559dbf7dbac9a319d3b58b41abcc34a4
	//client_test.go:348: 12c1c15c36667d017ca595b88057e792c33bfe7e
	//client_test.go:349: 5084abedea81b254d42e5f894015d2be3853b27c
}

func TestClient_QueryDelegations(t *testing.T) {
	initClient()
	addr, err := types.AccAddressFromBech32("cosmos12yprrdprzat35zhqxe2fcnn3u26gwlt6xcq0pj")
	assert.NoError(t, err)
	res, err := client.QueryDelegations(addr, 2458080)
	assert.NoError(t, err)
	t.Log(res.String())
	for i, d := range res.GetDelegationResponses() {
		t.Log(i, d.Balance.Amount.IsZero())
	}
}

func TestClient_QueryDelegationTotalRewards(t *testing.T) {
	initClient()
	addr, err := types.AccAddressFromBech32("cosmos12yprrdprzat35zhqxe2fcnn3u26gwlt6xcq0pj")
	assert.NoError(t, err)
	t.Log(client.GetDenom())
	res, err := client.QueryDelegationTotalRewards(addr, 2458080)
	assert.NoError(t, err)
	for i, _ := range res.Rewards {
		t.Log(i, res.Rewards[i].Reward.AmountOf(client.GetDenom()))
		t.Log(i, res.Rewards[i].Reward.AmountOf(client.GetDenom()).TruncateInt())

	}
	t.Log("total ", res.GetTotal().AmountOf(client.GetDenom()).TruncateInt())
}

func TestMemo(t *testing.T) {
	initClient()
	// res, err := client.QueryTxByHash("c7e3f7baf5a5f1d8cbc112080f32070dddd7cca5fe4272e06f8d42c17b25193f")
	// assert.NoError(t, err)
	// txBts, err := hex.DecodeString("0ada010ac0010a2c2f73746166696875622e73746166696875622e6c65646765722e4d73674c6971756964697479556e626f6e64128f010a2c737461666931356c6e653730796b3235347330706d32646136673539723832636a796d7a6a71767671787a37122a69616131356e706c743477663639366430356c666667706a6b7366397632796c66307335767a7a6d336c 1a 0731 757269726973 22 2a69616131356c6e653730796b3235347330706d32646136673539723832636a796d7a6a717a3973613568121575736520796f757220706f77657220776973656c7912660a500a460a1f2f636f736d6f732e63727970746f2e736563703235366b312e5075624b657912230a2102b55d8e6a0b7a57364cf2437c793ebed3b103e7dec4c56b87258972230252127812040a020801180312120a0c0a047566697312043435303010a0fe0a1a403b89ceb1ede49a88270f215136e62fab924e3cee762a3404ea5693b3096a806f508f47ce5870249ee7cc9c91a05058f7694dd95263ba7b431899722c5d51a44c")
	// txBts, err := hex.DecodeString("0ada010ac0010a2c2f73746166696875622e73746166696875622e6c65646765722e4d73674c6971756964697479556e626f6e64128f010a2c737461666931356c6e653730796b3235347330706d32646136673539723832636a796d7a6a71767671787a37122a69616131356e706c743477663639366430356c666667706a6b7366397632796c66307335767a7a6d336c 1a 0731 757269726973 22 2a69616131356c6e653730796b3235347330706d32646136673539723832636a796d7a6a717a3973613568121575736520796f757220706f77657220776973656c7912660a500a460a1f2f636f736d6f732e63727970746f2e736563703235366b312e5075624b657912230a2102b55d8e6a0b7a57364cf2437c793ebed3b103e7dec4c56b87258972230252127812040a020801180312120a0c0a047566697312043435303010a0fe0a1a403b89ceb1ede49a88270f215136e62fab924e3cee762a3404ea5693b3096a806f508f47ce5870249ee7cc9c91a05058f7694dd95263ba7b431899722c5d51a44c")
	// txBts, err := hex.DecodeString("0ac7010ac4010a2c2f73746166696875622e73746166696875622e6c65646765722e4d73674c6971756964697479556e626f6e641293010a2c737461666931356c6e653730796b3235347330706d32646136673539723832636a796d7a6a71767671787a37122a69616131356e706c743477663639366430356c666667706a6b7366397632796c66307335767a7a6d336c 1a 0b0a06 757269726973 120131 22 2a69616131767a6632386b706d7332747a72666b61613634356b6172707a3471386776613074657579773712650a500a460a1f2f636f736d6f732e63727970746f2e736563703235366b312e5075624b657912230a2102b55d8e6a0b7a57364cf2437c793ebed3b103e7dec4c56b87258972230252127812040a020801180212110a0b0a0475666973120335303010c09a0c1a405ffb2ce2bc9d120956b24c20749240c1e49953274c6d400e04da234c48059aef758c66895b78e2bee28b9c5ac6a619278db40a0708dfa00e94786aaa7f0d5e71")
	txBts, err := hex.DecodeString("0ae6010ae3010a2c2f73746166696875622e73746166696875622e6c65646765722e4d73674c6971756964697479556e626f6e6412b2010a2c737461666931777a39617839786c786a747739616b787966323961666c61753466363370356475766b67397a1241636f736d6f73316773746834367a35307732353670346b71333678717568347139306d666a713074346c6d3973636c6e367a75636736346570797175647a717a6d1a100a06757261746f6d1206313030303030222d636f736d6f7331777a39617839786c786a747739616b787966323961666c6175346636337035643838787a333612660a500a460a1f2f636f736d6f732e63727970746f2e736563703235366b312e5075624b657912230a21028b6d9485ca194f6daf2f7a59efa19c58d3cbc95bbcc6b6fe26ac8f4d21fadeaa12040a02087f180212120a0c0a047566697312043431323710c8890a1a40960cade6b6888087ebb9e88a2ef5570730c38105046220be051b7d4418ba0c16012c59360c104d851f66b574279887217598cd1644b02de028a402cd615db3f9")
	assert.NoError(t, err)
	t.Log(string(txBts))

	tx, err := client.GetTxConfig().TxDecoder()(txBts)
	assert.NoError(t, err)

	memoTx, ok := tx.(types.TxWithMemo)
	assert.Equal(t, true, ok)
	t.Log(memoTx.GetMemo())
}

func TestDecodeTx(t *testing.T) {
	initClient()
	// unbond
	// amino
	// ui
	t.Log(Decode("0ae6010ae3010a2c2f73746166696875622e73746166696875622e6c65646765722e4d73674c6971756964697479556e626f6e6412b2010a2c737461666931777a39617839786c786a747739616b787966323961666c61753466363370356475766b67397a1241636f736d6f73316773746834367a35307732353670346b71333678717568347139306d666a713074346c6d3973636c6e367a75636736346570797175647a717a6d1a100a06757261746f6d1206323030303030222d636f736d6f7331777a39617839786c786a747739616b787966323961666c6175346636337035643838787a333612670a500a460a1f2f636f736d6f732e63727970746f2e736563703235366b312e5075624b657912230a21028b6d9485ca194f6daf2f7a59efa19c58d3cbc95bbcc6b6fe26ac8f4d21fadeaa12040a02087f180312130a0d0a04756669731205313235303010a0c21e1a4011f5b96d19525ea6d124fa5e3f8d6b5fcf76f7c6f87960926bb900508dceb84f05d056c96edcc5dcd720a4f7d04c42e5a680ebc74aac8fa1284756b79ab12035"))
	// cli
	t.Log(Decode("0ae6010ae3010a2c2f73746166696875622e73746166696875622e6c65646765722e4d73674c6971756964697479556e626f6e6412b2010a2c737461666931777a39617839786c786a747739616b787966323961666c61753466363370356475766b67397a1241636f736d6f73316773746834367a35307732353670346b71333678717568347139306d666a713074346c6d3973636c6e367a75636736346570797175647a717a6d1a100a06757261746f6d1206323030303030222d636f736d6f7331777a39617839786c786a747739616b787966323961666c6175346636337035643838787a333612670a500a460a1f2f636f736d6f732e63727970746f2e736563703235366b312e5075624b657912230a21028b6d9485ca194f6daf2f7a59efa19c58d3cbc95bbcc6b6fe26ac8f4d21fadeaa12040a02087f180312130a0d0a04756669731205313235303010a0c21e1a40487ae9c0fcb4f2538d87283069541c2f018a90fa964a335211b5fdbbfa3390af7ebbc6e7059635e19d09ed55eb63f6760162143137794125fc1fb32171b428b8"))

	toBeSignedUi, _ := hex.DecodeString("7b226163636f756e745f6e756d626572223a223339222c22636861696e5f6964223a2273746166696875622d746573746e65742d32222c22666565223a7b22616d6f756e74223a5b7b22616d6f756e74223a223132353030222c2264656e6f6d223a2275666973227d5d2c22676173223a22353030303030227d2c226d656d6f223a22222c226d736773223a5b7b2274797065223a226c65646765722f4c6971756964697479556e626f6e64222c2276616c7565223a7b2263726561746f72223a22737461666931777a39617839786c786a747739616b787966323961666c61753466363370356475766b67397a222c22706f6f6c223a22636f736d6f73316773746834367a35307732353670346b71333678717568347139306d666a713074346c6d3973636c6e367a75636736346570797175647a717a6d222c22726563697069656e74223a22636f736d6f7331777a39617839786c786a747739616b787966323961666c6175346636337035643838787a3336222c2276616c7565223a7b22616d6f756e74223a22323030303030222c2264656e6f6d223a22757261746f6d227d7d7d5d2c2273657175656e6365223a2233227d")
	toBeSignedCi, _ := hex.DecodeString("7b226163636f756e745f6e756d626572223a223339222c22636861696e5f6964223a2273746166696875622d746573746e65742d32222c22666565223a7b22616d6f756e74223a5b7b22616d6f756e74223a223132353030222c2264656e6f6d223a2275666973227d5d2c22676173223a22353030303030227d2c226d656d6f223a22222c226d736773223a5b7b2274797065223a226c65646765722f4c6971756964697479556e626f6e64222c2276616c7565223a7b2263726561746f72223a22737461666931777a39617839786c786a747739616b787966323961666c61753466363370356475766b67397a222c22706f6f6c223a22636f736d6f73316773746834367a35307732353670346b71333678717568347139306d666a713074346c6d3973636c6e367a75636736346570797175647a717a6d222c22726563697069656e74223a22636f736d6f7331777a39617839786c786a747739616b787966323961666c6175346636337035643838787a3336222c2276616c7565223a7b22616d6f756e74223a22323030303030222c2264656e6f6d223a22757261746f6d227d7d7d5d2c2273657175656e6365223a2233227d")
	t.Log(string(toBeSignedUi))
	t.Log(string(toBeSignedCi))

	// proto
	t.Log(Decode("0ae4010ae1010a2c2f73746166696875622e73746166696875622e6c65646765722e4d73674c6971756964697479556e626f6e6412b0010a2c737461666931717a743071616a7a7239646633656e35736b3036786c6b32366e333030303363387568646b671241636f736d6f73316773746834367a35307732353670346b71333678717568347139306d666a713074346c6d3973636c6e367a75636736346570797175647a717a6d1a0e0a06757261746f6d120432303030222d636f736d6f73316b346d727a733637746e6437666c366178336c34343077706133616d6174377465757267396112670a500a460a1f2f636f736d6f732e63727970746f2e736563703235366b312e5075624b657912230a2102e175fb10e71787328b1c6cd560eef117180c61a4e3f55d4ef3e7b7ae96bf778712040a020801180212130a0d0a04756669731205313235303010a0c21e1a40648fc60304a9dbd54069a031fd2d23ec3b9005ab9f18dd24c766a0c6f6ae341677b7a20dc557298291f2e67db15517ff36b9170794fe3e161251a7d5df4a47fe"))
	toBesignedCiProto, _ := hex.DecodeString("0ae4010ae1010a2c2f73746166696875622e73746166696875622e6c65646765722e4d73674c6971756964697479556e626f6e6412b0010a2c737461666931717a743071616a7a7239646633656e35736b3036786c6b32366e333030303363387568646b671241636f736d6f73316773746834367a35307732353670346b71333678717568347139306d666a713074346c6d3973636c6e367a75636736346570797175647a717a6d1a0e0a06757261746f6d120432303030222d636f736d6f73316b346d727a733637746e6437666c366178336c34343077706133616d6174377465757267396112670a500a460a1f2f636f736d6f732e63727970746f2e736563703235366b312e5075624b657912230a2102e175fb10e71787328b1c6cd560eef117180c61a4e3f55d4ef3e7b7ae96bf778712040a020801180212130a0d0a04756669731205313235303010a0c21e1a1273746166696875622d746573746e65742d322028")
	t.Log(string(toBesignedCiProto))

	// ibc transfer
	// amini
	// cli
	toBeSignedIbcTransfer, _ := hex.DecodeString("7b226163636f756e745f6e756d626572223a223339222c22636861696e5f6964223a2273746166696875622d746573746e65742d32222c22666565223a7b22616d6f756e74223a5b7b22616d6f756e74223a223132353030222c2264656e6f6d223a2275666973227d5d2c22676173223a22353030303030227d2c226d656d6f223a22222c226d736773223a5b7b2274797065223a22636f736d6f732d73646b2f4d73675472616e73666572222c2276616c7565223a7b227265636569766572223a22636f736d6f7331777a39617839786c786a747739616b787966323961666c6175346636337035643838787a3336222c2273656e646572223a22737461666931777a39617839786c786a747739616b787966323961666c61753466363370356475766b67397a222c22736f757263655f6368616e6e656c223a226368616e6e656c2d30222c22736f757263655f706f7274223a227472616e73666572222c2274696d656f75745f686569676874223a7b227265766973696f6e5f686569676874223a22343730343232222c227265766973696f6e5f6e756d626572223a2231227d2c2274696d656f75745f74696d657374616d70223a2231363639323938373435383936393633303030222c22746f6b656e223a7b22616d6f756e74223a22313030222c2264656e6f6d223a22757261746f6d227d7d7d5d2c2273657175656e6365223a2238227d")
	t.Log(string(toBeSignedIbcTransfer))
	// ui
	toBeSignedIbcTransfUi, _ := hex.DecodeString("7b226163636f756e745f6e756d626572223a223339222c22636861696e5f6964223a2273746166696875622d746573746e65742d32222c22666565223a7b22616d6f756e74223a5b7b22616d6f756e74223a2231222c2264656e6f6d223a2275666973227d5d2c22676173223a22313030343632227d2c226d656d6f223a22222c226d736773223a5b7b2274797065223a22636f736d6f732d73646b2f4d73675472616e73666572222c2276616c7565223a7b227265636569766572223a22636f736d6f7331777a39617839786c786a747739616b787966323961666c6175346636337035643838787a3336222c2273656e646572223a22737461666931777a39617839786c786a747739616b787966323961666c61753466363370356475766b67397a222c22736f757263655f6368616e6e656c223a226368616e6e656c2d30222c22736f757263655f706f7274223a227472616e73666572222c2274696d656f75745f686569676874223a7b227265766973696f6e5f686569676874223a22353639343232222c227265766973696f6e5f6e756d626572223a2231227d2c2274696d656f75745f74696d657374616d70223a2230222c22746f6b656e223a7b22616d6f756e74223a22313030222c2264656e6f6d223a22757261746f6d227d7d7d5d2c2273657175656e6365223a2238227d")
	t.Log(string(toBeSignedIbcTransfUi))

	// claim mint reward
	// cli
	toBeSignedClaimMintReward, _ := hex.DecodeString("7b226163636f756e745f6e756d626572223a223339222c22636861696e5f6964223a2273746166696875622d746573746e65742d32222c22666565223a7b22616d6f756e74223a5b7b22616d6f756e74223a223132353030222c2264656e6f6d223a2275666973227d5d2c22676173223a22353030303030227d2c226d656d6f223a22222c226d736773223a5b7b2274797065223a22726d696e747265776172642f436c61696d4d696e74526577617264222c2276616c7565223a7b2263726561746f72223a22737461666931777a39617839786c786a747739616b787966323961666c61753466363370356475766b67397a222c2264656e6f6d223a22757261746f6d227d7d5d2c2273657175656e6365223a223132227d")
	t.Log(string(toBeSignedClaimMintReward))

}

func Decode(hexStr string) string {
	txBts, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	tx, err := client.GetTxConfig().TxDecoder()(txBts)
	if err != nil {
		panic(err)
	}

	txJson, err := client.GetTxConfig().TxJSONEncoder()(tx)
	if err != nil {
		panic(err)
	}

	return string(txJson)
}

func TestMultiThread(t *testing.T) {
	initClient()
	wg := sync.WaitGroup{}
	wg.Add(50)

	for i := 0; i < 50; i++ {
		go func(i int) {
			t.Log(i)
			time.Sleep(5 * time.Second)
			height, err := client.GetAccount()
			if err != nil {
				t.Log("fail", i, err)
			} else {
				t.Log("success", i, height.GetSequence())
			}
			time.Sleep(15 * time.Second)
			height, err = client.GetAccount()
			if err != nil {
				t.Log("fail", i, err)
			} else {
				t.Log("success", i, height.GetSequence())
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestSort(t *testing.T) {
	a := []string{"cosmos1kuyde8vpt8c0ty4pxqgxw3makse7md80umthvg"}
	t.Log(a)
	sort.SliceStable(a, func(i, j int) bool {
		return bytes.Compare([]byte(a[i]), []byte(a[j])) < 0
	})
	t.Log(a)
	// rawTx := "7b22626f6479223a7b226d65737361676573223a5b7b224074797065223a222f636f736d6f732e62616e6b2e763162657461312e4d73674d756c746953656e64222c22696e70757473223a5b7b2261646472657373223a22636f736d6f7331776d6b39797334397a78676d78373770717337636a6e70616d6e6e7875737071753272383779222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a22313134373730227d5d7d5d2c226f757470757473223a5b7b2261646472657373223a22636f736d6f733135366b6b326b71747777776670733836673534377377646c7263326377367163746d36633877222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a2231393936227d5d7d2c7b2261646472657373223a22636f736d6f73316b7579646538767074386330747934707871677877336d616b7365376d643830756d74687667222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a223939383030227d5d7d2c7b2261646472657373223a22636f736d6f73316a6b6b68666c753871656471743463796173643674673730676a7778346a6b6872736536727a222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a223132393734227d5d7d5d7d5d2c226d656d6f223a22222c2274696d656f75745f686569676874223a2230222c22657874656e73696f6e5f6f7074696f6e73223a5b5d2c226e6f6e5f637269746963616c5f657874656e73696f6e5f6f7074696f6e73223a5b5d7d2c22617574685f696e666f223a7b227369676e65725f696e666f73223a5b5d2c22666565223a7b22616d6f756e74223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a2237353030227d5d2c226761735f6c696d6974223a2231353030303030222c227061796572223a22222c226772616e746572223a22227d7d2c227369676e617475726573223a5b5d7d"
	rawTx := "7b22626f6479223a7b226d65737361676573223a5b7b224074797065223a222f636f736d6f732e62616e6b2e763162657461312e4d73674d756c746953656e64222c22696e70757473223a5b7b2261646472657373223a22636f736d6f7331776d6b39797334397a78676d78373770717337636a6e70616d6e6e7875737071753272383779222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a2231303539363338303631227d5d7d5d2c226f757470757473223a5b7b2261646472657373223a22636f736d6f733135366b6b326b71747777776670733836673534377377646c7263326377367163746d36633877222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a2231303539363338303631227d5d7d5d7d5d2c226d656d6f223a22222c2274696d656f75745f686569676874223a2230222c22657874656e73696f6e5f6f7074696f6e73223a5b5d2c226e6f6e5f637269746963616c5f657874656e73696f6e5f6f7074696f6e73223a5b5d7d2c22617574685f696e666f223a7b227369676e65725f696e666f73223a5b5d2c22666565223a7b22616d6f756e74223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a2237353030227d5d2c226761735f6c696d6974223a2231353030303030222c227061796572223a22222c226772616e746572223a22227d7d2c227369676e617475726573223a5b5d7d"
	// rawTx:="7b22626f6479223a7b226d65737361676573223a5b7b224074797065223a222f636f736d6f732e62616e6b2e763162657461312e4d73674d756c746953656e64222c22696e70757473223a5b7b2261646472657373223a22636f736d6f7331776d6b39797334397a78676d78373770717337636a6e70616d6e6e7875737071753272383779222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a22313134373730227d5d7d5d2c226f757470757473223a5b7b2261646472657373223a22636f736d6f73316a6b6b68666c753871656471743463796173643674673730676a7778346a6b6872736536727a222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a223132393734227d5d7d2c7b2261646472657373223a22636f736d6f733135366b6b326b71747777776670733836673534377377646c7263326377367163746d36633877222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a2231393936227d5d7d2c7b2261646472657373223a22636f736d6f73316b7579646538767074386330747934707871677877336d616b7365376d643830756d74687667222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a223939383030227d5d7d5d7d5d2c226d656d6f223a22222c2274696d656f75745f686569676874223a2230222c22657874656e73696f6e5f6f7074696f6e73223a5b5d2c226e6f6e5f637269746963616c5f657874656e73696f6e5f6f7074696f6e73223a5b5d7d2c22617574685f696e666f223a7b227369676e65725f696e666f73223a5b5d2c22666565223a7b22616d6f756e74223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a2237353030227d5d2c226761735f6c696d6974223a2231353030303030222c227061796572223a22222c226772616e746572223a22227d7d2c227369676e617475726573223a5b5d7d"
	txBts, err := hex.DecodeString(rawTx)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(txBts))
}

func TestZeroAddress(t *testing.T) {
	addressBts := [20]byte{}
	ac := types.AccAddress(addressBts[:])
	t.Log(ac.String())
}

func TestGetBalance(t *testing.T) {

	initClient()
	addr, _ := types.AccAddressFromBech32("stafi1qzt0qajzr9df3en5sk06xlk26n30003c8uhdkg")
	balance, err := client.QueryBalance(addr, "ufis", 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(balance)
}

func TestCalculGas(t *testing.T) {
	initClient()
	addr, err := types.AccAddressFromBech32("stafi1qzt0qajzr9df3en5sk06xlk26n30003c8uhdkg")
	if err != nil {
		t.Fatal(err)
	}
	msg := xBankTypes.NewMsgSend(client.GetFromAddress(), addr, types.NewCoins(types.NewCoin("ufis", types.NewInt(5))))

	bts, err := client.ConstructAndSignTx(msg)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bts))
}

func TestSendUpdataRvalidator(t *testing.T) {
	initClient()
	content := stafiHubXRValidatorTypes.NewUpdateRValidatorProposal(
		client.GetFromAddress().String(),
		"uratom",
		"",
		"cosmosvaloper17h2x3j7u44qkrq0sk8ul0r2qr440rwgjkfg0gh",
		"cosmosvaloper1cc99d3xcukhedg4wcw53j7a9q68uza707vpfe7",
		&stafiHubXRValidatorTypes.Cycle{
			Denom:   "uratom",
			Version: 0,
			Number:  0,
		})
	txHashStr, _, err := client.SubmitProposal(content)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(txHashStr)
}

func TestQueryLatestVotedCycle(t *testing.T) {
	initClient()
	latest, err := client.QueryLatestVotedCycle("uratom", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(latest)
}

func TestQuerySignInfos(t *testing.T) {
	initClient()

	signingInfos, err := client.QueryAllSigningInfos(1445000)
	if err != nil {
		t.Fatal(err)
	}

	validators, err := client.QueryValidators(1445000)
	if err != nil {
		t.Fatal(err)
	}

	missedBlocks := make(map[string]int)
	for _, signInfo := range signingInfos.Info {
		addr, err := types.ConsAddressFromBech32(signInfo.Address)
		if err != nil {
			t.Fatal(err)
		}
		missedBlocks[strings.ToUpper(hex.EncodeToString(addr.Bytes()))] = int(signInfo.MissedBlocksCounter)
	}

	for _, val := range validators.Validators {

		// get hex addr
		consPubkeyJson, err := client.Ctx().Codec.MarshalJSON(val.ConsensusPubkey)
		if err != nil {
			t.Fatal(err)
		}
		var pk cryptotypes.PubKey
		if err := client.Ctx().Codec.UnmarshalInterfaceJSON(consPubkeyJson, &pk); err != nil {
			t.Fatal(err)
		}
		addrHexStr := pk.Address().String()
		missedNumber := missedBlocks[addrHexStr]
		if val.Status == 3 {
			signed := 43200 - missedNumber
			t.Log("val ", val.Description.Moniker, "uptime ", uint64(signed/432))
		} else {
			t.Log("val ", val.Description.Moniker, "uptime ", 0)
		}

	}
}
