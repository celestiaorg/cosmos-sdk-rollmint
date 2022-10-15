package baseapp

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/mem"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/store/v2alpha1/smt"
	abci "github.com/tendermint/tendermint/abci/types"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
)

// Represents a single-round fraudProof
type FraudProof struct {
	// The block height to load state of
	blockHeight int64

	// TODO: Add Proof that appHash is inside merklized ISRs in block header at block height

	appHash []byte
	// A map from module name to state witness
	stateWitness map[string]StateWitness

	// Fraudulent state transition has to be one of these
	// Only one have of these three can be non-nil
	fraudulentBeginBlock *abci.RequestBeginBlock
	fraudulentDeliverTx  *abci.RequestDeliverTx
	fraudulentEndBlock   *abci.RequestEndBlock

	// TODO: Add Proof that fraudulent state transition is inside merkelizied transactions in block header
}

// State witness with a list of all witness data
type StateWitness struct {
	// store level proof
	Proof    tmcrypto.ProofOp
	RootHash []byte
	// List of witness data
	WitnessData []WitnessData
}

// Witness data containing a key/value pair and a SMT proof for said key/value pair
type WitnessData struct {
	Key   []byte
	Value []byte
	Proof tmcrypto.ProofOp
}

func (fraudProof *FraudProof) getModules() []string {
	keys := make([]string, 0, len(fraudProof.stateWitness))
	for k := range fraudProof.stateWitness {
		keys = append(keys, k)
	}
	return keys
}

func (fraudProof *FraudProof) getDeepIAVLTrees() (map[string]*iavl.Store, error) {
	storeKeyToIAVLTree := make(map[string]*iavl.Store)
	for storeKey, stateWitness := range fraudProof.stateWitness {
		rootHash := stateWitness.RootHash
		// TODO(manav): Replace with IAVL Deep Subtrees once implementation is done
		substoreDeepSMT := smtlib.NewDeepSparseMerkleSubTree(smtlib.NewSimpleMap(), smtlib.NewSimpleMap(), sha256.New(), rootHash)
		for _, witnessData := range stateWitness.WitnessData {
			proofOp, key, val := witnessData.Proof, witnessData.Key, witnessData.Value
			proof, err := types.CommitmentOpDecoder(proofOp)
			if err != nil {
				return nil, err
			}
			// TODO(manav): Replace with an IAVL proof instead of SMT here
			smtProof := proof.(*smt.ProofOp).GetProof()
			substoreDeepSMT.AddBranch(smtProof, key, val)
		}
		// TODO(manav): Replace with actual iavl stores
		storeKeyToIAVLTree[storeKey] = &iavl.Store{}
	}
	return storeKeyToIAVLTree, nil
}

func (fraudProof *FraudProof) extractStore() map[string]types.KVStore {
	store := make(map[string]types.KVStore)
	for storeKey, stateWitness := range fraudProof.stateWitness {
		subStore := mem.NewStore()
		for _, witnessData := range stateWitness.WitnessData {
			key, val := witnessData.Key, witnessData.Value
			subStore.Set(key, val)
		}
		store[storeKey] = subStore
	}
	return store
}

// Returns true only if only one of the three pointers is nil
func (fraudProof *FraudProof) checkFraudulentStateTransition() bool {
	if fraudProof.fraudulentBeginBlock != nil {
		return fraudProof.fraudulentDeliverTx == nil && fraudProof.fraudulentEndBlock == nil
	}
	if fraudProof.fraudulentDeliverTx != nil {
		return fraudProof.fraudulentEndBlock == nil
	}
	return fraudProof.fraudulentEndBlock != nil
}

func (fraudProof *FraudProof) verifyFraudProof() (bool, error) {
	if !fraudProof.checkFraudulentStateTransition() {
		return false, fmt.Errorf("fraudProof has more than one type of fradulent state transitions marked nil")
	}
	for storeKey, stateWitness := range fraudProof.stateWitness {
		proofOp := stateWitness.Proof
		proof, err := types.CommitmentOpDecoder(proofOp)
		if err != nil {
			return false, err
		}
		if !bytes.Equal(proof.GetKey(), []byte(storeKey)) {
			return false, fmt.Errorf("got storeKey: %s, expected: %s", string(proof.GetKey()), storeKey)
		}
		appHash, err := proof.Run([][]byte{stateWitness.RootHash})
		if err != nil {
			return false, err
		}
		if !bytes.Equal(appHash[0], fraudProof.appHash) {
			return false, fmt.Errorf("got appHash: %s, expected: %s", string(fraudProof.appHash), string(fraudProof.appHash))
		}
		// Fraudproof verification on a substore level
		for _, witness := range stateWitness.WitnessData {
			proofOp, key, value := witness.Proof, witness.Key, witness.Value
			if err != nil {
				return false, err
			}
			if !bytes.Equal(key, proofOp.GetKey()) {
				return false, fmt.Errorf("got key: %s, expected: %s for storeKey: %s", string(key), string(proof.GetKey()), storeKey)
			}
			proof, err := types.CommitmentOpDecoder(proofOp)
			if err != nil {
				return false, err
			}
			rootHash, err := proof.Run([][]byte{value})
			if err != nil {
				return false, err
			}
			if !bytes.Equal(rootHash[0], stateWitness.RootHash) {
				return false, fmt.Errorf("got rootHash: %s, expected: %s for storeKey: %s", string(rootHash[0]), string(stateWitness.RootHash), storeKey)
			}
		}
	}
	return true, nil
}

func (fraudProof *FraudProof) toABCI() abci.FraudProof {
	abciStateWitness := make(map[string]*abci.StateWitness)
	for storeKey, stateWitness := range fraudProof.stateWitness {
		abciWitnessData := make([]*abci.WitnessData, 0, len(stateWitness.WitnessData))
		for _, witnessData := range stateWitness.WitnessData {
			abciWitness := abci.WitnessData{
				Key:   witnessData.Key,
				Value: witnessData.Value,
				Proof: &witnessData.Proof,
			}
			abciWitnessData = append(abciWitnessData, &abciWitness)
		}
		proof := stateWitness.Proof
		abciStateWitness[storeKey] = &abci.StateWitness{
			Proof:       &proof,
			RootHash:    stateWitness.RootHash,
			WitnessData: abciWitnessData,
		}
	}
	return abci.FraudProof{
		BlockHeight:          fraudProof.blockHeight,
		AppHash:              fraudProof.appHash,
		StateWitness:         abciStateWitness,
		FraudulentBeginBlock: fraudProof.fraudulentBeginBlock,
		FraudulentDeliverTx:  fraudProof.fraudulentDeliverTx,
		FraudulentEndBlock:   fraudProof.fraudulentEndBlock,
	}
}

func (fraudProof *FraudProof) fromABCI(abciFraudProof abci.FraudProof) {
	stateWitness := make(map[string]StateWitness)
	for storeKey, abciStateWitness := range abciFraudProof.StateWitness {
		witnessData := make([]WitnessData, 0, len(abciStateWitness.WitnessData))
		for _, abciWitnessData := range abciStateWitness.WitnessData {
			witness := WitnessData{
				Key:   abciWitnessData.Key,
				Value: abciWitnessData.Value,
				Proof: *abciWitnessData.Proof,
			}
			witnessData = append(witnessData, witness)
		}
		stateWitness[storeKey] = StateWitness{
			Proof:       *abciStateWitness.Proof,
			RootHash:    abciStateWitness.RootHash,
			WitnessData: witnessData,
		}
	}
	fraudProof.blockHeight = abciFraudProof.BlockHeight
	fraudProof.appHash = abciFraudProof.AppHash
	fraudProof.stateWitness = stateWitness
	fraudProof.fraudulentBeginBlock = abciFraudProof.FraudulentBeginBlock
	fraudProof.fraudulentDeliverTx = abciFraudProof.FraudulentDeliverTx
	fraudProof.fraudulentEndBlock = abciFraudProof.FraudulentEndBlock
}
