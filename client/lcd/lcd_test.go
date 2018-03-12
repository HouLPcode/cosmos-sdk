package lcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	keys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/app"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	cryptoKeys "github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/tendermint/p2p"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
)

func TestKeys(t *testing.T) {
	_, db, err := initKeybase(t)
	require.Nil(t, err, "Couldn't init Keybase")

	cdc := app.MakeCodec()
	r := initRouter(cdc)

	// empty keys
	res := request(t, r, "GET", "/keys", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	body := res.Body.String()
	assert.Equal(t, body, "[]", "Expected an empty array")

	// get seed
	res = request(t, r, "GET", "/keys/seed", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	seed := res.Body.String()
	reg, err := regexp.Compile(`([a-z]+ ){12}`)
	require.Nil(t, err)
	match := reg.MatchString(seed)
	assert.True(t, match, "Returned seed has wrong foramt", seed)

	// add key
	var jsonStr = []byte(`{"name":"test_fail", "password":"1234567890"}`)
	res = request(t, r, "POST", "/keys", jsonStr)

	assert.Equal(t, http.StatusBadRequest, res.Code, "Account creation should require a seed")

	jsonStr = []byte(fmt.Sprintf(`{"name":"test", "password":"1234567890", "seed": "%s"}`, seed))
	res = request(t, r, "POST", "/keys", jsonStr)

	assert.Equal(t, http.StatusOK, res.Code, res.Body.String())
	addr := res.Body.String()
	assert.Len(t, addr, 40, "Returned address has wrong format", addr)

	// existing keys
	res = request(t, r, "GET", "/keys", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var m [1]keys.KeyOutput
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&m)
	require.NoError(t, err)

	assert.Equal(t, m[0].Name, "test", "Did not serve keys name correctly")
	assert.Equal(t, m[0].Address, addr, "Did not serve keys Address correctly")

	// select key
	res = request(t, r, "GET", "/keys/test", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var m2 keys.KeyOutput
	decoder = json.NewDecoder(res.Body)
	err = decoder.Decode(&m2)

	assert.Equal(t, m2.Name, "test", "Did not serve keys name correctly")
	assert.Equal(t, m2.Address, addr, "Did not serve keys Address correctly")

	// update key
	jsonStr = []byte(`{"old_password":"1234567890", "new_password":"12345678901"}`)
	res = request(t, r, "PUT", "/keys/test", jsonStr)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	// here it should say unauthorized as we changed the password before
	res = request(t, r, "PUT", "/keys/test", jsonStr)
	require.Equal(t, http.StatusUnauthorized, res.Code, res.Body.String())

	// delete key
	jsonStr = []byte(`{"password":"12345678901"}`)
	res = request(t, r, "DELETE", "/keys/test", jsonStr)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	db.Close()
}

func TestVersion(t *testing.T) {
	prepareClient(t)
	cdc := app.MakeCodec()
	r := initRouter(cdc)

	// node info
	res := request(t, r, "GET", "/version", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	reg, err := regexp.Compile(`\d+\.\d+\.\d+(-dev)?`)
	require.Nil(t, err)
	match := reg.MatchString(res.Body.String())
	assert.True(t, match, res.Body.String())
}

func TestNodeStatus(t *testing.T) {
	//ch := server.StartServer(t)
	//defer close(ch)
	//prepareClient(t)

	//cdc := app.MakeCodec() // TODO refactor so that cdc is included from a common test init function
	//r := initRouter(cdc)

	err := tests.InitServer()
	require.Nil(t, err)
	err := tests.StartServer()
	require.Nil(t, err)

	// node info
	res := request(t, r, "GET", "/node_info", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var m p2p.NodeInfo
	decoder := json.NewDecoder(res.Body)
	err := decoder.Decode(&m)
	require.Nil(t, err, "Couldn't parse node info")

	assert.NotEqual(t, p2p.NodeInfo{}, m, "res: %v", res)

	// syncing
	res = request(t, r, "GET", "/syncing", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	assert.Equal(t, "true", res.Body.String())
}

func TestBlock(t *testing.T) {
	ch := server.StartServer(t)
	defer close(ch)
	prepareClient(t)

	cdc := app.MakeCodec()
	r := initRouter(cdc)

	// res := request(t, r, "GET", "/blocks/latest", nil)
	// require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	// var m ctypes.ResultBlock
	// decoder := json.NewDecoder(res.Body)
	// err := decoder.Decode(&m)
	// require.Nil(t, err, "Couldn't parse block")

	// assert.NotEqual(t, ctypes.ResultBlock{}, m)

	// --

	res := request(t, r, "GET", "/blocks/1", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var m ctypes.ResultBlock
	decoder := json.NewDecoder(res.Body)
	err := decoder.Decode(&m)
	require.Nil(t, err, "Couldn't parse block")

	assert.NotEqual(t, ctypes.ResultBlock{}, m)

	// --

	res = request(t, r, "GET", "/blocks/2", nil)
	require.Equal(t, http.StatusNotFound, res.Code, res.Body.String())
}

func TestValidators(t *testing.T) {
	ch := server.StartServer(t)
	defer close(ch)

	prepareClient(t)
	cdc := app.MakeCodec()
	r := initRouter(cdc)

	// res := request(t, r, "GET", "/validatorsets/latest", nil)
	// require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	// var m ctypes.ResultValidators
	// decoder := json.NewDecoder(res.Body)
	// err := decoder.Decode(&m)
	// require.Nil(t, err, "Couldn't parse validatorset")

	// assert.NotEqual(t, ctypes.ResultValidators{}, m)

	// --

	res := request(t, r, "GET", "/validatorsets/1", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var m ctypes.ResultValidators
	decoder := json.NewDecoder(res.Body)
	err := decoder.Decode(&m)
	require.Nil(t, err, "Couldn't parse validatorset")

	assert.NotEqual(t, ctypes.ResultValidators{}, m)

	// --

	res = request(t, r, "GET", "/validatorsets/2", nil)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestCoinSend(t *testing.T) {
	ch := server.StartServer(t)
	defer close(ch)

	prepareClient(t)
	cdc := app.MakeCodec()
	r := initRouter(cdc)

	// TODO make that account has coins
	kb := client.MockKeyBase()
	info, seed, err := kb.Create("account_with_coins", "1234567890", cryptoKeys.CryptoAlgo("ed25519"))
	require.NoError(t, err)
	addr := string(info.Address())

	// query empty
	res := request(t, r, "GET", "/accounts/1234567890123456789012345678901234567890", nil)
	require.Equal(t, http.StatusNoContent, res.Code, res.Body.String())

	// query
	res = request(t, r, "GET", "/accounts/"+addr, nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	assert.Equal(t, `{
		"coins": [
			{
				"denom": "mycoin",
				"amount": 9007199254740992
			}
		]
	}`, res.Body.String())

	// create account to send in keybase
	var jsonStr = []byte(fmt.Sprintf(`{"name":"test", "password":"1234567890", "seed": "%s"}`, seed))
	res = request(t, r, "POST", "/keys", jsonStr)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	// create receive address
	receiveInfo, _, err := kb.Create("receive_address", "1234567890", cryptoKeys.CryptoAlgo("ed25519"))
	require.NoError(t, err)
	receiveAddr := string(receiveInfo.Address())

	// send
	jsonStr = []byte(`{
		"name":"test", 
		"password":"1234567890", 
		"amount":[{
			"denom": "mycoin",
			"amount": 1
		}]
	}`)
	res = request(t, r, "POST", "/accounts/"+receiveAddr+"/send", jsonStr)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	// check if received
	res = request(t, r, "GET", "/accounts/"+receiveAddr, nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	assert.Equal(t, `{
		"coins": [
			{
				"denom": "mycoin",
				"amount": 1
			}
		]
	}`, res.Body.String())
}

//__________________________________________________________
// helpers

func defaultLogger() log.Logger {
	return log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
}

func prepareClient(t *testing.T) {
	db := dbm.NewMemDB()
	app := baseapp.NewBaseApp(t.Name(), defaultLogger(), db)
	viper.Set(client.FlagNode, "localhost:46657")
	header := abci.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	app.Commit()
}

func request(t *testing.T, r http.Handler, method string, path string, payload []byte) *httptest.ResponseRecorder {

	req, err := http.NewRequest(method, path, bytes.NewBuffer(payload))
	require.Nil(t, err)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	return res
}

func initKeybase(t *testing.T) (cryptoKeys.Keybase, *dbm.GoLevelDB, error) {
	os.RemoveAll("./testKeybase")
	db, err := dbm.NewGoLevelDB("keys", "./testKeybase")
	require.Nil(t, err)
	kb := client.GetKeyBase(db)
	keys.SetKeyBase(kb)
	return kb, db, nil
}