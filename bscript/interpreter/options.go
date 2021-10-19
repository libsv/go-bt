package interpreter

import (
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/bscript/interpreter/scriptflag"
)

// ExecutionOptionFunc for setting execution options.
type ExecutionOptionFunc func(p *execOpts)

// WithTx configure the execution to run again a tx.
func WithTx(tx *bt.Tx, inputIdx int, prevOutput *bt.Output) ExecutionOptionFunc {
	return func(p *execOpts) {
		p.tx = tx
		p.previousTxOut = prevOutput
		p.inputIdx = inputIdx
	}
}

// WithScripts configure the execution to run again a set of *bscript.Script.
func WithScripts(lockingScript *bscript.Script, unlockingScript *bscript.Script) ExecutionOptionFunc {
	return func(p *execOpts) {
		p.lockingScript = lockingScript
		p.unlockingScript = unlockingScript
	}
}

// WithAfterGenesis configure the execution to operate in an after-genesis context.
func WithAfterGenesis() ExecutionOptionFunc {
	return func(p *execOpts) {
		p.flags.AddFlag(scriptflag.UTXOAfterGenesis)
	}
}

// WithForkID configure the execution to allow a tx with a fork id.
func WithForkID() ExecutionOptionFunc {
	return func(p *execOpts) {
		p.flags.AddFlag(scriptflag.EnableSighashForkID)
	}
}

// WithP2SH configure the execution to allow a P2SH output.
func WithP2SH() ExecutionOptionFunc {
	return func(p *execOpts) {
		p.flags.AddFlag(scriptflag.Bip16)
	}
}

// WithFlags configure the execution with the provided flags.
func WithFlags(flags scriptflag.Flag) ExecutionOptionFunc {
	return func(p *execOpts) {
		p.flags.AddFlag(flags)
	}
}

// WithDebugger enable execution debugging with the provided configured debugger.
// It is important to note that when this setting is applied, it enables thread
// state cloning, at every configured debug step.
func WithDebugger(debugger Debugger) ExecutionOptionFunc {
	return func(p *execOpts) {
		p.debugger = debugger
	}
}

// WithState inject the provided state into the execution thread. This assumes
// that the state is correct for the scripts provided.
//
// NOTE: This is highly experimental and is unstable when used with unintended states,
// and likely still when used in a happy path scenario. Therefore, it is recommended
// to only be used for debugging purposes.
//
// The safest recommended *interpreter.State records for a given script can be
// are those which can be captured during `debugger.BeforeStep` and `debugger.AfterStep`.
func WithState(state *State) ExecutionOptionFunc {
	return func(p *execOpts) {
		p.state = state
	}
}
