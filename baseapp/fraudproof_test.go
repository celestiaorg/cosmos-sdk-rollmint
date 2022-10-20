package baseapp

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/abci/types"
)

// Tests the checkFraudulentStateTransition for all possible cases.
func TestCheckFraudulentStateTransition(t *testing.T) {

	//Case when only BeginBlock is set
	fraudProof := FraudProof{}
	fraudProof.fraudulentBeginBlock = &types.RequestBeginBlock{}
	result := fraudProof.checkFraudulentStateTransition()
	require.True(t, result)

	//Case when only DeliverTx is set
	fraudProof = FraudProof{}
	fraudProof.fraudulentDeliverTx = &types.RequestDeliverTx{}
	result = fraudProof.checkFraudulentStateTransition()
	require.True(t, result)

	//Case when only EndBlock is set
	fraudProof = FraudProof{}
	fraudProof.fraudulentEndBlock = &types.RequestEndBlock{}
	result = fraudProof.checkFraudulentStateTransition()
	require.True(t, result)

	//Case when both BeginBlock and DeliverTx are set
	fraudProof = FraudProof{}
	fraudProof.fraudulentBeginBlock = &types.RequestBeginBlock{}
	fraudProof.fraudulentDeliverTx = &types.RequestDeliverTx{}
	result = fraudProof.checkFraudulentStateTransition()
	require.False(t, result)

	//Case when both BeginBlock and EndBlock are set
	fraudProof = FraudProof{}
	fraudProof.fraudulentBeginBlock = &types.RequestBeginBlock{}
	fraudProof.fraudulentEndBlock = &types.RequestEndBlock{}
	result = fraudProof.checkFraudulentStateTransition()
	require.False(t, result)

	//Case when both DeliverTx and EndBlock are set
	fraudProof = FraudProof{}
	fraudProof.fraudulentDeliverTx = &types.RequestDeliverTx{}
	fraudProof.fraudulentEndBlock = &types.RequestEndBlock{}
	result = fraudProof.checkFraudulentStateTransition()
	require.False(t, result)

	//Case when both DeliverTx and EndBlock are set
	fraudProof = FraudProof{}
	fraudProof.fraudulentBeginBlock = &types.RequestBeginBlock{}
	fraudProof.fraudulentDeliverTx = &types.RequestDeliverTx{}
	fraudProof.fraudulentEndBlock = &types.RequestEndBlock{}
	result = fraudProof.checkFraudulentStateTransition()
	require.False(t, result)

}
