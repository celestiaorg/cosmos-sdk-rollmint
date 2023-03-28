package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"

	abciclient "github.com/cometbft/cometbft/abci/client"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	cmttypes "github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"golang.org/x/sync/errgroup"

	"github.com/cosmos/cosmos-sdk/server/api"
	servergrpc "github.com/cosmos/cosmos-sdk/server/grpc"
	servercmtlog "github.com/cosmos/cosmos-sdk/server/log"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	rollconf "github.com/rollkit/rollkit/config"
	rollconv "github.com/rollkit/rollkit/conv"
	rollnode "github.com/rollkit/rollkit/node"
	rollrpc "github.com/rollkit/rollkit/rpc"
)

func startInProcess(cfg Config, val *Validator) error {
	logger := val.Ctx.Logger
	cmtCfg := val.Ctx.Config
	cmtCfg.Instrumentation.Prometheus = false

	if err := val.AppConfig.ValidateBasic(); err != nil {
		return err
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(cmtCfg.NodeKeyFile())
	if err != nil {
		return err
	}

	pval := pvm.LoadOrGenFilePV(cmtCfg.PrivValidatorKeyFile(), cmtCfg.PrivValidatorStateFile())
	// keys in Rollkit format
	p2pKey, err := rollconv.GetNodeKey(nodeKey)
	if err != nil {
		return err
	}
	signingKey, err := rollconv.GetNodeKey(&p2p.NodeKey{PrivKey: pval.Key.PrivKey})
	if err != nil {
		return err
	}

	app := cfg.AppConstructor(*val)
	genDocProvider := func() (*cmttypes.GenesisDoc, error) {
		appGenesis, err := genutiltypes.AppGenesisFromFile(cmtCfg.GenesisFile())
		if err != nil {
			return nil, err
		}

		return appGenesis.ToGenesisDoc()
	}

	genDoc, err := genDocProvider()
	if err != nil {
		return err
	}

	nodeConfig := rollconf.NodeConfig{}
	err = nodeConfig.GetViperConfig(val.Ctx.Viper)
	nodeConfig.Aggregator = true
	nodeConfig.DALayer = "mock"
	if err != nil {
		return err
	}
	rollconv.GetNodeConfig(&nodeConfig, cmtCfg)
	err = rollconv.TranslateAddresses(&nodeConfig)
	if err != nil {
		return err
	}
	val.tmNode, err = rollnode.NewNode(
		context.Background(),
		nodeConfig,
		p2pKey,
		signingKey,
		abciclient.NewLocalClient(nil, app),
		genDoc,
		servercmtlog.CometZeroLogWrapper{Logger: val.Ctx.Logger},
	)
	if err != nil {
		return err
	}

	if err := val.tmNode.Start(); err != nil {
		return err
	}

	if val.RPCAddress != "" {
		server := rollrpc.NewServer(val.tmNode, cmtCfg.RPC, servercmtlog.CometZeroLogWrapper{Logger: val.Ctx.Logger})
		err = server.Start()
		if err != nil {
			return err
		}
		val.RPCClient = server.Client()
	}

	// We'll need a RPC client if the validator exposes a gRPC or REST endpoint.
	if val.APIAddress != "" || val.AppConfig.GRPC.Enable {
		val.ClientCtx = val.ClientCtx.
			WithClient(val.RPCClient)

		app.RegisterTxService(val.ClientCtx)
		app.RegisterTendermintService(val.ClientCtx)
		app.RegisterNodeService(val.ClientCtx)
	}

	ctx := context.Background()
	ctx, val.cancelFn = context.WithCancel(ctx)
	val.errGroup, ctx = errgroup.WithContext(ctx)

	grpcCfg := val.AppConfig.GRPC

	if grpcCfg.Enable {
		grpcSrv, err := servergrpc.NewGRPCServer(val.ClientCtx, app, grpcCfg)
		if err != nil {
			return err
		}

		// Start the gRPC server in a goroutine. Note, the provided ctx will ensure
		// that the server is gracefully shut down.
		val.errGroup.Go(func() error {
			return servergrpc.StartGRPCServer(ctx, logger.With("module", "grpc-server"), grpcCfg, grpcSrv)
		})

		val.grpc = grpcSrv
	}

	if val.APIAddress != "" {
		apiSrv := api.New(val.ClientCtx, logger.With("module", "api-server"), val.grpc)
		app.RegisterAPIRoutes(apiSrv, val.AppConfig.API)

		val.errGroup.Go(func() error {
			return apiSrv.Start(ctx, *val.AppConfig)
		})

		val.api = apiSrv
	}

	return nil
}

func collectGenFiles(cfg Config, vals []*Validator, outputDir string) error {
	genTime := cmttime.Now()

	for i := 0; i < cfg.NumValidators; i++ {
		cmtCfg := vals[i].Ctx.Config

		nodeDir := filepath.Join(outputDir, vals[i].Moniker, "simd")
		gentxsDir := filepath.Join(outputDir, "gentxs")

		cmtCfg.Moniker = vals[i].Moniker
		cmtCfg.SetRoot(nodeDir)

		initCfg := genutiltypes.NewInitConfig(cfg.ChainID, gentxsDir, vals[i].NodeID, vals[i].PubKey)

		genFile := cmtCfg.GenesisFile()
		appGenesis, err := genutiltypes.AppGenesisFromFile(genFile)
		if err != nil {
			return err
		}

		appState, err := genutil.GenAppStateFromConfig(cfg.Codec, cfg.TxConfig,
			cmtCfg, initCfg, appGenesis, banktypes.GenesisBalancesIterator{}, genutiltypes.DefaultMessageValidator)
		if err != nil {
			return err
		}

		// overwrite each validator's genesis file to have a canonical genesis time
		if err := genutil.ExportGenesisFileWithTime(genFile, cfg.ChainID, nil, appState, genTime); err != nil {
			return err
		}
	}

	return nil
}

func initGenFiles(cfg Config, genAccounts []authtypes.GenesisAccount, genBalances []banktypes.Balance, genFiles []string) error {
	// set the accounts in the genesis state
	var authGenState authtypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[authtypes.ModuleName], &authGenState)

	accounts, err := authtypes.PackAccounts(genAccounts)
	if err != nil {
		return err
	}

	authGenState.Accounts = append(authGenState.Accounts, accounts...)
	cfg.GenesisState[authtypes.ModuleName] = cfg.Codec.MustMarshalJSON(&authGenState)

	// set the balances in the genesis state
	var bankGenState banktypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[banktypes.ModuleName], &bankGenState)

	bankGenState.Balances = append(bankGenState.Balances, genBalances...)
	cfg.GenesisState[banktypes.ModuleName] = cfg.Codec.MustMarshalJSON(&bankGenState)

	appGenStateJSON, err := json.MarshalIndent(cfg.GenesisState, "", "  ")
	if err != nil {
		return err
	}

	appGenesis := genutiltypes.AppGenesis{
		ChainID:  cfg.ChainID,
		AppState: appGenStateJSON,
		Consensus: &genutiltypes.ConsensusGenesis{
			Validators: nil,
		},
	}

	// generate empty genesis files for each validator and save
	for i := 0; i < cfg.NumValidators; i++ {
		if err := appGenesis.SaveAs(genFiles[i]); err != nil {
			return err
		}
	}

	return nil
}

func writeFile(name string, dir string, contents []byte) error {
	file := filepath.Join(dir, name)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("could not create directory %q: %w", dir, err)
	}

	if err := os.WriteFile(file, contents, 0o644); err != nil { //nolint: gosec
		return err
	}

	return nil
}

// Get a free address for a test CometBFT server
// protocol is either tcp, http, etc
func FreeTCPAddr() (addr, port string, closeFn func() error, err error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", "", nil, err
	}

	closeFn = func() error {
		return l.Close()
	}

	portI := l.Addr().(*net.TCPAddr).Port
	port = fmt.Sprintf("%d", portI)
	addr = fmt.Sprintf("tcp://0.0.0.0:%s", port)
	return
}
