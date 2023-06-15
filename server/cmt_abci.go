package server

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

type cometABCIWrapper struct {
	app servertypes.ABCI
}

func NewCometABCIWrapper(app servertypes.ABCI) abci.Application {
	return cometABCIWrapper{app: app}
}

func (w cometABCIWrapper) Info(_ context.Context, req *abci.RequestInfo) (*abci.ResponseInfo, error) {
	return w.app.Info(req)
}

func (w cometABCIWrapper) Query(ctx context.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error) {
	return w.app.Query(ctx, req)
}

func (w cometABCIWrapper) CheckTx(_ context.Context, req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	return w.app.CheckTx(req)
}

func (w cometABCIWrapper) InitChain(_ context.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	return w.app.InitChain(req)
}

func (w cometABCIWrapper) PrepareProposal(_ context.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
	return w.app.PrepareProposal(req)
}

func (w cometABCIWrapper) ProcessProposal(_ context.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	return w.app.ProcessProposal(req)
}

func (w cometABCIWrapper) FinalizeBlock(_ context.Context, req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	return w.app.FinalizeBlock(req)
}

func (w cometABCIWrapper) ExtendVote(ctx context.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
	return w.app.ExtendVote(ctx, req)
}

func (w cometABCIWrapper) VerifyVoteExtension(_ context.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
	return w.app.VerifyVoteExtension(req)
}

func (w cometABCIWrapper) Commit(_ context.Context, _ *abci.RequestCommit) (*abci.ResponseCommit, error) {
	return w.app.Commit()
}

func (w cometABCIWrapper) ListSnapshots(_ context.Context, req *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	return w.app.ListSnapshots(req)
}

func (w cometABCIWrapper) OfferSnapshot(_ context.Context, req *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	return w.app.OfferSnapshot(req)
}

func (w cometABCIWrapper) LoadSnapshotChunk(_ context.Context, req *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	return w.app.LoadSnapshotChunk(req)
}

func (w cometABCIWrapper) ApplySnapshotChunk(_ context.Context, req *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	return w.app.ApplySnapshotChunk(req)
}

func (w cometABCIWrapper) BeginBlock(_ context.Context, req *abci.RequestBeginBlock) (*abci.ResponseBeginBlock, error) {
	return nil, fmt.Errorf("BeginBlock no longer supported")
}

func (w cometABCIWrapper) DeliverTx(_ context.Context, req *abci.RequestDeliverTx) (*abci.ResponseDeliverTx, error) {
	return nil, fmt.Errorf("DeliverTx no longer supported")
}

func (w cometABCIWrapper) EndBlock(_ context.Context, req *abci.RequestEndBlock) (*abci.ResponseEndBlock, error) {
	return nil, fmt.Errorf("EndBlock no longer supported")
}

func (w cometABCIWrapper) GetAppHash(_ context.Context, req *abci.RequestGetAppHash) (*abci.ResponseGetAppHash, error) {
	//TODO implement me
	panic("implement me")
}

func (w cometABCIWrapper) GenerateFraudProof(_ context.Context, req *abci.RequestGenerateFraudProof) (*abci.ResponseGenerateFraudProof, error) {
	//TODO implement me
	panic("implement me")
}

func (w cometABCIWrapper) VerifyFraudProof(_ context.Context, req *abci.RequestVerifyFraudProof) (*abci.ResponseVerifyFraudProof, error) {
	//TODO implement me
	panic("implement me")
}
