package interpreter

import "math"

type config interface {
	AfterGenesis() bool
	MaxOps() int
	MaxStackSize() int
	MaxScriptSize() int
	MaxScriptElementSize() int
	MaxScriptNumberLength() int
	MaxPubKeysPerMultiSig() int
}

// Limits applied to transactions before genesis
const (
	MaxOpsBeforeGenesis                = 500
	MaxStackSizeBeforeGenesis          = 1000
	MaxScriptSizeBeforeGenesis         = 10000
	MaxScriptElementSizeBeforeGenesis  = 520
	MaxScriptNumberLengthBeforeGenesis = 4
	MaxPubKeysPerMultiSigBeforeGenesis = 20
)

type beforeGenesisConfig struct{}
type afterGenesisConfig struct{}

func (a *afterGenesisConfig) AfterGenesis() bool {
	return true
}

func (b *beforeGenesisConfig) AfterGenesis() bool {
	return false
}

func (a *afterGenesisConfig) MaxStackSize() int {
	return math.MaxInt32
}

func (b *beforeGenesisConfig) MaxStackSize() int {
	return MaxStackSizeBeforeGenesis
}

func (a *afterGenesisConfig) MaxScriptSize() int {
	return math.MaxInt32
}

func (b *beforeGenesisConfig) MaxScriptSize() int {
	return MaxScriptSizeBeforeGenesis
}

func (a *afterGenesisConfig) MaxScriptElementSize() int {
	return math.MaxInt32
}

func (b *beforeGenesisConfig) MaxScriptElementSize() int {
	return MaxScriptElementSizeBeforeGenesis
}

func (a *afterGenesisConfig) MaxScriptNumberLength() int {
	return 750 * 1000 // 750 * 1Kb
}

func (b *beforeGenesisConfig) MaxScriptNumberLength() int {
	return MaxScriptNumberLengthBeforeGenesis
}

func (a *afterGenesisConfig) MaxOps() int {
	return math.MaxInt32
}

func (b *beforeGenesisConfig) MaxOps() int {
	return MaxOpsBeforeGenesis
}

func (a *afterGenesisConfig) MaxPubKeysPerMultiSig() int {
	return math.MaxInt32
}

func (b *beforeGenesisConfig) MaxPubKeysPerMultiSig() int {
	return MaxPubKeysPerMultiSigBeforeGenesis
}
