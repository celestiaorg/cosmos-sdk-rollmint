#!/bin/bash
cd ..
apt update
apt install jq -y
curl https://get.ignite.com/cli@v0.26.1! | bash
ignite scaffold chain gm --address-prefix gm
cd gm
go mod edit -replace github.com/cosmos/cosmos-sdk=../cosmos-sdk
go mod edit -replace github.com/tendermint/tendermint=github.com/rollkit/cometbft@v0.0.0-20230524013049-75272ebaee38
go mod tidy
go mod download

VALIDATOR_NAME=validator1
CHAIN_ID=gm
KEY_NAME=gm-key
KEY_2_NAME=gm-key-2
CHAINFLAG="--chain-id ${CHAIN_ID}"
TOKEN_AMOUNT="10000000000000000000000000stake"
STAKING_AMOUNT="1000000000stake"

# create a random Namespace ID for your rollup to post blocks to
NAMESPACE_ID=$(openssl rand -hex 8)
echo $NAMESPACE_ID

# build the gm chain with Rollkit
ignite chain build
# reset any existing genesis/chain data
#/home/runner/go/bin/gmd tendermint unsafe-reset-all
gmd tendermint unsafe-reset-all

# initialize the validator with the chain ID you set
#/home/runner/go/bin/gmd init $VALIDATOR_NAME --chain-id $CHAIN_ID
gmd init $VALIDATOR_NAME --chain-id $CHAIN_ID

# add keys for key 1 and key 2 to keyring-backend test
echo y | /home/runner/go/bin/gmd keys add $KEY_NAME --keyring-backend test
echo y | /home/runner/go/bin/gmd keys add $KEY_2_NAME --keyring-backend test

# add these as genesis accounts
gmd add-genesis-account $KEY_NAME $TOKEN_AMOUNT --keyring-backend test
gmd add-genesis-account $KEY_2_NAME $TOKEN_AMOUNT --keyring-backend test

# set the staking amounts in the genesis transaction
gmd gentx $KEY_NAME $STAKING_AMOUNT --chain-id $CHAIN_ID --keyring-backend test

# collect genesis transactions
gmd collect-gentxs

# query the DA Layer start height, in this case we are querying
# our local devnet at port 26657, the RPC. The RPC endpoint is
# to allow users to interact with Celestia's nodes by querying
# the node's state and broadcasting transactions on the Celestia
# network. The default port is 26657.
DA_BLOCK_HEIGHT=$(curl http://celestia:26657/block | jq -r '.result.block.header.height')
echo $DA_BLOCK_HEIGHT

# start the chain
echo "Starting rollup in foreground!"
/home/runner/go/bin/gmd start --rollkit.aggregator true --rollkit.da_layer celestia --rollkit.da_config='{"base_url":"http://celestia:26659","timeout":60000000000,"fee":6000,"gas_limit":6000000}' --rollkit.namespace_id $NAMESPACE_ID --rollkit.da_start_height $DA_BLOCK_HEIGHT
