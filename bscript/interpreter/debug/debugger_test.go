package debug_test

import (
	"encoding/hex"
	"testing"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/bscript/interpreter"
	"github.com/libsv/go-bt/v2/bscript/interpreter/debug"
	"github.com/stretchr/testify/assert"
)

func TestDebugger_BeforeExecute(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStack           []string
		expOpcode          string
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
			expStack:           []string{},
			expOpcode:          "OP_4",
		},
		"complex script": {
			lockingScriptHex:   "76a97ca8a687",
			unlockingScriptHex: "00",
			expStack:           []string{},
			expOpcode:          "OP_0",
		},
		"error script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5457",
			expStack:           []string{},
			expOpcode:          "OP_4",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ls, err := bscript.NewFromHexString(test.lockingScriptHex)
			assert.NoError(t, err)

			uls, err := bscript.NewFromHexString(test.unlockingScriptHex)
			assert.NoError(t, err)

			var timesCalled int
			debugger := debug.NewDebugger()
			debugger.AttachBeforeExecute(func(state *interpreter.State) {
				timesCalled++
				stack := make([]string, len(state.DataStack))
				for i, d := range state.DataStack {
					stack[i] = hex.EncodeToString(d)
				}
				assert.Equal(t, test.expStack, stack)
				assert.Equal(t, test.expOpcode, state.Opcode().Name())
			})

			interpreter.NewEngine().Execute(
				interpreter.WithScripts(ls, uls),
				interpreter.WithAfterGenesis(),
				interpreter.WithDebugger(debugger),
			)

			assert.Equal(t, 1, timesCalled)
		})
	}
}

func TestDebugger_BeforeStep(t *testing.T) {
	t.Parallel()

	type stateHistory struct {
		dstack  [][]string
		astack  [][]string
		opcodes []string
	}

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStackHistory    [][]string
		expOpcodes         []string
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
			expStackHistory: [][]string{
				{},
				{"04"},
				{"04", "06"},
				{"04", "06", "02"},
				{"04", "06", "02", "03"},
				{"04", "06", "06"},
				{"04"},
				{"04", "02"},
				{"04", "02", "02"},
				{"04", "04"},
			},
			expOpcodes: []string{
				"OP_4", "OP_6",
				"OP_2", "OP_3", "OP_MUL", "OP_EQUALVERIFY",
				"OP_2", "OP_2", "OP_ADD", "OP_EQUAL",
			},
		},
		"complex script": {
			lockingScriptHex:   "76a97ca8a687",
			unlockingScriptHex: "00",
			expStackHistory: [][]string{
				{},
				{""},
				{"", ""},
				{"", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", ""},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
			},
			expOpcodes: []string{"OP_0", "OP_DUP", "OP_HASH160", "OP_SWAP", "OP_SHA256", "OP_RIPEMD160", "OP_EQUAL"},
		},
		"error script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5457",
			expStackHistory: [][]string{
				{},
				{"04"},
				{"04", "07"},
				{"04", "07", "02"},
				{"04", "07", "02", "03"},
				{"04", "07", "06"},
			},
			expOpcodes: []string{"OP_4", "OP_7", "OP_2", "OP_3", "OP_MUL", "OP_EQUALVERIFY"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ls, err := bscript.NewFromHexString(test.lockingScriptHex)
			assert.NoError(t, err)

			uls, err := bscript.NewFromHexString(test.unlockingScriptHex)
			assert.NoError(t, err)

			history := &stateHistory{
				dstack:  make([][]string, 0),
				astack:  make([][]string, 0),
				opcodes: make([]string, 0),
			}

			debugger := debug.NewDebugger()
			debugger.AttachBeforeStep(func(state *interpreter.State) {
				stack := make([]string, len(state.DataStack))
				for i, d := range state.DataStack {
					stack[i] = hex.EncodeToString(d)
				}
				history.dstack = append(history.dstack, stack)
				history.opcodes = append(history.opcodes, state.Opcode().Name())
			})

			interpreter.NewEngine().Execute(
				interpreter.WithScripts(ls, uls),
				interpreter.WithAfterGenesis(),
				interpreter.WithDebugger(debugger),
			)

			assert.Equal(t, test.expStackHistory, history.dstack)
			assert.Equal(t, test.expOpcodes, history.opcodes)
		})
	}
}

func TestDebugger_AfterStep(t *testing.T) {
	t.Parallel()

	type stateHistory struct {
		dstack  [][]string
		astack  [][]string
		opcodes []string
	}

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStackHistory    [][]string
		expOpcodes         []string
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
			expStackHistory: [][]string{
				{"04"},
				{"04", "06"},
				{"04", "06", "02"},
				{"04", "06", "02", "03"},
				{"04", "06", "06"},
				{"04"},
				{"04", "02"},
				{"04", "02", "02"},
				{"04", "04"},
				{"01"},
			},
			expOpcodes: []string{
				"OP_6",
				"OP_2", "OP_3", "OP_MUL", "OP_EQUALVERIFY",
				"OP_2", "OP_2", "OP_ADD", "OP_EQUAL", "OP_EQUAL",
			},
		},
		"complex script": {
			lockingScriptHex:   "76a97ca8a687",
			unlockingScriptHex: "00",
			expStackHistory: [][]string{
				{""},
				{"", ""},
				{"", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", ""},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{"01"},
			},
			expOpcodes: []string{"OP_DUP", "OP_HASH160", "OP_SWAP", "OP_SHA256", "OP_RIPEMD160", "OP_EQUAL", "OP_EQUAL"},
		},
		"error script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5457",
			expStackHistory: [][]string{
				{"04"},
				{"04", "07"},
				{"04", "07", "02"},
				{"04", "07", "02", "03"},
				{"04", "07", "06"},
			},
			expOpcodes: []string{"OP_7", "OP_2", "OP_3", "OP_MUL", "OP_EQUALVERIFY"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ls, err := bscript.NewFromHexString(test.lockingScriptHex)
			assert.NoError(t, err)

			uls, err := bscript.NewFromHexString(test.unlockingScriptHex)
			assert.NoError(t, err)

			history := &stateHistory{
				dstack:  make([][]string, 0),
				astack:  make([][]string, 0),
				opcodes: make([]string, 0),
			}

			debugger := debug.NewDebugger()
			debugger.AttachAfterStep(func(state *interpreter.State) {
				stack := make([]string, len(state.DataStack))
				for i, d := range state.DataStack {
					stack[i] = hex.EncodeToString(d)
				}
				history.dstack = append(history.dstack, stack)
				history.opcodes = append(history.opcodes, state.Opcode().Name())
			})

			interpreter.NewEngine().Execute(
				interpreter.WithScripts(ls, uls),
				interpreter.WithAfterGenesis(),
				interpreter.WithDebugger(debugger),
			)

			assert.Equal(t, test.expStackHistory, history.dstack)
			assert.Equal(t, test.expOpcodes, history.opcodes)
		})
	}
}

func TestDebugger_BeforeExecuteOpcode(t *testing.T) {
	t.Parallel()

	type stateHistory struct {
		dstack  [][]string
		astack  [][]string
		opcodes []string
	}

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStackHistory    [][]string
		expOpcodes         []string
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
			expStackHistory: [][]string{
				{},
				{"04"},
				{"04", "06"},
				{"04", "06", "02"},
				{"04", "06", "02", "03"},
				{"04", "06", "06"},
				{"04"},
				{"04", "02"},
				{"04", "02", "02"},
				{"04", "04"},
			},
			expOpcodes: []string{
				"OP_4", "OP_6",
				"OP_2", "OP_3", "OP_MUL", "OP_EQUALVERIFY",
				"OP_2", "OP_2", "OP_ADD", "OP_EQUAL",
			},
		},
		"complex script": {
			lockingScriptHex:   "76a97ca8a687",
			unlockingScriptHex: "00",
			expStackHistory: [][]string{
				{},
				{""},
				{"", ""},
				{"", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", ""},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
			},
			expOpcodes: []string{"OP_0", "OP_DUP", "OP_HASH160", "OP_SWAP", "OP_SHA256", "OP_RIPEMD160", "OP_EQUAL"},
		},
		"error script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5457",
			expStackHistory: [][]string{
				{},
				{"04"},
				{"04", "07"},
				{"04", "07", "02"},
				{"04", "07", "02", "03"},
				{"04", "07", "06"},
			},
			expOpcodes: []string{"OP_4", "OP_7", "OP_2", "OP_3", "OP_MUL", "OP_EQUALVERIFY"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ls, err := bscript.NewFromHexString(test.lockingScriptHex)
			assert.NoError(t, err)

			uls, err := bscript.NewFromHexString(test.unlockingScriptHex)
			assert.NoError(t, err)

			history := &stateHistory{
				dstack:  make([][]string, 0),
				astack:  make([][]string, 0),
				opcodes: make([]string, 0),
			}

			debugger := debug.NewDebugger()
			debugger.AttachBeforeExecuteOpcode(func(state *interpreter.State) {
				stack := make([]string, len(state.DataStack))
				for i, d := range state.DataStack {
					stack[i] = hex.EncodeToString(d)
				}
				history.dstack = append(history.dstack, stack)
				history.opcodes = append(history.opcodes, state.Opcode().Name())
			})

			interpreter.NewEngine().Execute(
				interpreter.WithScripts(ls, uls),
				interpreter.WithAfterGenesis(),
				interpreter.WithDebugger(debugger),
			)

			assert.Equal(t, test.expStackHistory, history.dstack)
			assert.Equal(t, test.expOpcodes, history.opcodes)
		})
	}
}

func TestDebugger_AfterExecuteOpcode(t *testing.T) {
	t.Parallel()

	type stateHistory struct {
		dstack  [][]string
		astack  [][]string
		opcodes []string
	}

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStackHistory    [][]string
		expOpcodes         []string
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
			expStackHistory: [][]string{
				{"04"},
				{"04", "06"},
				{"04", "06", "02"},
				{"04", "06", "02", "03"},
				{"04", "06", "06"},
				{"04"},
				{"04", "02"},
				{"04", "02", "02"},
				{"04", "04"},
				{"01"},
			},
			expOpcodes: []string{
				"OP_4", "OP_6",
				"OP_2", "OP_3", "OP_MUL", "OP_EQUALVERIFY",
				"OP_2", "OP_2", "OP_ADD", "OP_EQUAL",
			},
		},
		"complex script": {
			lockingScriptHex:   "76a97ca8a687",
			unlockingScriptHex: "00",
			expStackHistory: [][]string{
				{""},
				{"", ""},
				{"", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", ""},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{"01"},
			},
			expOpcodes: []string{"OP_0", "OP_DUP", "OP_HASH160", "OP_SWAP", "OP_SHA256", "OP_RIPEMD160", "OP_EQUAL"},
		},
		"error script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5457",
			expStackHistory: [][]string{
				{"04"},
				{"04", "07"},
				{"04", "07", "02"},
				{"04", "07", "02", "03"},
				{"04", "07", "06"},
			},
			expOpcodes: []string{"OP_4", "OP_7", "OP_2", "OP_3", "OP_MUL"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ls, err := bscript.NewFromHexString(test.lockingScriptHex)
			assert.NoError(t, err)

			uls, err := bscript.NewFromHexString(test.unlockingScriptHex)
			assert.NoError(t, err)

			history := &stateHistory{
				dstack:  make([][]string, 0),
				astack:  make([][]string, 0),
				opcodes: make([]string, 0),
			}

			debugger := debug.NewDebugger()
			debugger.AttachAfterExecuteOpcode(func(state *interpreter.State) {
				stack := make([]string, len(state.DataStack))
				for i, d := range state.DataStack {
					stack[i] = hex.EncodeToString(d)
				}
				history.dstack = append(history.dstack, stack)
				history.opcodes = append(history.opcodes, state.Opcode().Name())
			})

			interpreter.NewEngine().Execute(
				interpreter.WithScripts(ls, uls),
				interpreter.WithAfterGenesis(),
				interpreter.WithDebugger(debugger),
			)

			assert.Equal(t, test.expStackHistory, history.dstack)
			assert.Equal(t, test.expOpcodes, history.opcodes)
		})
	}
}

func TestDebugger_BeforeScriptChange(t *testing.T) {
	t.Parallel()

	type stateHistory struct {
		dstack  [][]string
		astack  [][]string
		opcodes []string
	}

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStackHistory    [][]string
		expOpcodes         []string
		exptimesCalled     int
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
			expStackHistory: [][]string{
				{"04", "06"},
				{"01"},
			},
			expOpcodes:     []string{"OP_6", "OP_EQUAL"},
			exptimesCalled: 2,
		},
		"complex script": {
			lockingScriptHex:   "76a97ca8a687",
			unlockingScriptHex: "00",
			expStackHistory: [][]string{
				{""},
				{"01"},
			},
			expOpcodes:     []string{"OP_0", "OP_EQUAL"},
			exptimesCalled: 2,
		},
		"error script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5457",
			expStackHistory: [][]string{
				{"04", "07"},
			},
			expOpcodes:     []string{"OP_7"},
			exptimesCalled: 1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ls, err := bscript.NewFromHexString(test.lockingScriptHex)
			assert.NoError(t, err)

			uls, err := bscript.NewFromHexString(test.unlockingScriptHex)
			assert.NoError(t, err)

			history := &stateHistory{
				dstack:  make([][]string, 0),
				astack:  make([][]string, 0),
				opcodes: make([]string, 0),
			}

			debugger := debug.NewDebugger()

			var timesCalled int
			debugger.AttachBeforeScriptChange(func(state *interpreter.State) {
				timesCalled++
				stack := make([]string, len(state.DataStack))
				for i, d := range state.DataStack {
					stack[i] = hex.EncodeToString(d)
				}
				history.dstack = append(history.dstack, stack)
				history.opcodes = append(history.opcodes, state.Opcode().Name())
			})

			interpreter.NewEngine().Execute(
				interpreter.WithScripts(ls, uls),
				interpreter.WithAfterGenesis(),
				interpreter.WithDebugger(debugger),
			)

			assert.Equal(t, test.expStackHistory, history.dstack)
			assert.Equal(t, test.expOpcodes, history.opcodes)
			assert.Equal(t, test.exptimesCalled, timesCalled)
		})
	}
}

func TestDebugger_AfterScriptChange(t *testing.T) {
	t.Parallel()

	type stateHistory struct {
		dstack  [][]string
		astack  [][]string
		opcodes []string
	}

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStackHistory    [][]string
		expOpcodes         []string
		exptimesCalled     int
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
			expStackHistory: [][]string{
				{"04", "06"},
				{"01"},
			},
			expOpcodes:     []string{"OP_2", "OP_EQUAL"},
			exptimesCalled: 2,
		},
		"complex script": {
			lockingScriptHex:   "76a97ca8a687",
			unlockingScriptHex: "00",
			expStackHistory: [][]string{
				{""},
				{"01"},
			},
			expOpcodes:     []string{"OP_DUP", "OP_EQUAL"},
			exptimesCalled: 2,
		},
		"error script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5457",
			expStackHistory: [][]string{
				{"04", "07"},
			},
			expOpcodes:     []string{"OP_2"},
			exptimesCalled: 1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ls, err := bscript.NewFromHexString(test.lockingScriptHex)
			assert.NoError(t, err)

			uls, err := bscript.NewFromHexString(test.unlockingScriptHex)
			assert.NoError(t, err)

			history := &stateHistory{
				dstack:  make([][]string, 0),
				astack:  make([][]string, 0),
				opcodes: make([]string, 0),
			}

			debugger := debug.NewDebugger()

			var timesCalled int
			debugger.AttachAfterScriptChange(func(state *interpreter.State) {
				timesCalled++
				stack := make([]string, len(state.DataStack))
				for i, d := range state.DataStack {
					stack[i] = hex.EncodeToString(d)
				}
				history.dstack = append(history.dstack, stack)
				history.opcodes = append(history.opcodes, state.Opcode().Name())
			})

			interpreter.NewEngine().Execute(
				interpreter.WithScripts(ls, uls),
				interpreter.WithAfterGenesis(),
				interpreter.WithDebugger(debugger),
			)

			assert.Equal(t, test.expStackHistory, history.dstack)
			assert.Equal(t, test.expOpcodes, history.opcodes)
			assert.Equal(t, test.exptimesCalled, timesCalled)
		})
	}
}

func TestDebugger_AfterExecution(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStack           []string
		expOpcode          string
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
			expStack:           []string{"01"},
			expOpcode:          "OP_EQUAL",
		},
		"complex script": {
			lockingScriptHex:   "76a97ca8a687",
			unlockingScriptHex: "00",
			expStack:           []string{"01"},
			expOpcode:          "OP_EQUAL",
		},
		"error script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5457",
			expStack:           []string{"04"},
			expOpcode:          "OP_EQUALVERIFY",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ls, err := bscript.NewFromHexString(test.lockingScriptHex)
			assert.NoError(t, err)

			uls, err := bscript.NewFromHexString(test.unlockingScriptHex)
			assert.NoError(t, err)

			stack := make([]string, 0)
			var opcode string

			debugger := debug.NewDebugger()
			debugger.AttachAfterExecute(func(state *interpreter.State) {
				for _, d := range state.DataStack {
					stack = append(stack, hex.EncodeToString(d))
				}
				opcode = state.Opcode().Name()
			})

			interpreter.NewEngine().Execute(
				interpreter.WithScripts(ls, uls),
				interpreter.WithAfterGenesis(),
				interpreter.WithDebugger(debugger),
			)

			assert.Equal(t, test.expStack, stack)
			assert.Equal(t, test.expOpcode, opcode)
		})
	}
}

func TestDebugger_AfterError(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStack           []string
		expOpcode          string
		expCalled          bool
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
		},
		"complex script": {
			lockingScriptHex:   "76a97ca8a687",
			unlockingScriptHex: "00",
		},
		"error script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5457",
			expStack:           []string{"04"},
			expOpcode:          "OP_EQUALVERIFY",
			expCalled:          true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ls, err := bscript.NewFromHexString(test.lockingScriptHex)
			assert.NoError(t, err)

			uls, err := bscript.NewFromHexString(test.unlockingScriptHex)
			assert.NoError(t, err)

			stack := make([]string, 0)
			var opcode string
			var called bool

			debugger := debug.NewDebugger()
			debugger.AttachAfterError(func(state *interpreter.State, err error) {
				called = true
				for _, d := range state.DataStack {
					stack = append(stack, hex.EncodeToString(d))
				}
				opcode = state.Opcode().Name()
			})

			interpreter.NewEngine().Execute(
				interpreter.WithScripts(ls, uls),
				interpreter.WithAfterGenesis(),
				interpreter.WithDebugger(debugger),
			)

			assert.Equal(t, test.expCalled, called)
			if called {
				assert.Equal(t, test.expStack, stack)
				assert.Equal(t, test.expOpcode, opcode)
			}
		})
	}
}

func TestDebugger_AfterSuccess(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStack           []string
		expOpcode          string
		expCalled          bool
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
			expStack:           []string{},
			expOpcode:          "OP_EQUAL",
			expCalled:          true,
		},
		"complex script": {
			lockingScriptHex:   "76a97ca8a687",
			unlockingScriptHex: "00",
			expStack:           []string{},
			expOpcode:          "OP_EQUAL",
			expCalled:          true,
		},
		"error script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5457",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ls, err := bscript.NewFromHexString(test.lockingScriptHex)
			assert.NoError(t, err)

			uls, err := bscript.NewFromHexString(test.unlockingScriptHex)
			assert.NoError(t, err)

			stack := make([]string, 0)
			var opcode string
			var called bool

			debugger := debug.NewDebugger()
			debugger.AttachAfterSuccess(func(state *interpreter.State) {
				called = true
				for _, d := range state.DataStack {
					stack = append(stack, hex.EncodeToString(d))
				}
				opcode = state.Opcode().Name()
			})

			interpreter.NewEngine().Execute(
				interpreter.WithScripts(ls, uls),
				interpreter.WithAfterGenesis(),
				interpreter.WithDebugger(debugger),
			)

			assert.Equal(t, test.expCalled, called)
			if called {
				assert.Equal(t, test.expStack, stack)
				assert.Equal(t, test.expOpcode, opcode)
			}
		})
	}
}

func TestDebugger_BeforeStackPush(t *testing.T) {
	t.Parallel()

	type stateHistory struct {
		dstack  [][]string
		astack  [][]string
		opcodes []string
		entries []string
	}

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStackHistory    [][]string
		expOpcodes         []string
		expPushData        []string
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
			expStackHistory: [][]string{
				{},
				{"04"},
				{"04", "06"},
				{"04", "06", "02"},
				{"04", "06"},
				{"04"},
				{"04"},
				{"04", "02"},
				{"04"},
				{},
			},
			expPushData: []string{"04", "06", "02", "03", "06", "01", "02", "02", "04", "01"},
			expOpcodes: []string{
				"OP_4", "OP_6",
				"OP_2", "OP_3", "OP_MUL", "OP_EQUALVERIFY",
				"OP_2", "OP_2", "OP_ADD", "OP_EQUAL",
			},
		},
		"complex script": {
			lockingScriptHex:   "76a97ca8a687",
			unlockingScriptHex: "00",
			expStackHistory: [][]string{
				{},
				{""},
				{""},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{},
			},
			expPushData: []string{"", "", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "01"},
			expOpcodes:  []string{"OP_0", "OP_DUP", "OP_HASH160", "OP_SWAP", "OP_SHA256", "OP_RIPEMD160", "OP_EQUAL"},
		},
		"error script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5457",
			expStackHistory: [][]string{
				{},
				{"04"},
				{"04", "07"},
				{"04", "07", "02"},
				{"04", "07"},
				{"04"},
			},
			expPushData: []string{"04", "07", "02", "03", "06", ""},
			expOpcodes:  []string{"OP_4", "OP_7", "OP_2", "OP_3", "OP_MUL", "OP_EQUALVERIFY"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ls, err := bscript.NewFromHexString(test.lockingScriptHex)
			assert.NoError(t, err)

			uls, err := bscript.NewFromHexString(test.unlockingScriptHex)
			assert.NoError(t, err)

			history := &stateHistory{
				dstack:  make([][]string, 0),
				astack:  make([][]string, 0),
				opcodes: make([]string, 0),
				entries: make([]string, 0),
			}

			debugger := debug.NewDebugger()
			debugger.AttachBeforeStackPush(func(state *interpreter.State, data []byte) {
				stack := make([]string, len(state.DataStack))
				for i, d := range state.DataStack {
					stack[i] = hex.EncodeToString(d)
				}
				history.dstack = append(history.dstack, stack)
				history.opcodes = append(history.opcodes, state.Opcode().Name())
				history.entries = append(history.entries, hex.EncodeToString(data))
			})

			interpreter.NewEngine().Execute(
				interpreter.WithScripts(ls, uls),
				interpreter.WithAfterGenesis(),
				interpreter.WithDebugger(debugger),
			)

			assert.Equal(t, test.expStackHistory, history.dstack)
			assert.Equal(t, test.expOpcodes, history.opcodes)
			assert.Equal(t, test.expPushData, history.entries)
		})
	}
}

func TestDebugger_AfterStackPush(t *testing.T) {
	t.Parallel()

	type stateHistory struct {
		dstack  [][]string
		astack  [][]string
		opcodes []string
		entries []string
	}

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStackHistory    [][]string
		expOpcodes         []string
		expPushData        []string
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
			expStackHistory: [][]string{
				{"04"},
				{"04", "06"},
				{"04", "06", "02"},
				{"04", "06", "02", "03"},
				{"04", "06", "06"},
				{"04", "01"},
				{"04", "02"},
				{"04", "02", "02"},
				{"04", "04"},
				{"01"},
			},
			expPushData: []string{"04", "06", "02", "03", "06", "01", "02", "02", "04", "01"},
			expOpcodes: []string{
				"OP_4", "OP_6",
				"OP_2", "OP_3", "OP_MUL", "OP_EQUALVERIFY",
				"OP_2", "OP_2", "OP_ADD", "OP_EQUAL",
			},
		},
		"complex script": {
			lockingScriptHex:   "76a97ca8a687",
			unlockingScriptHex: "00",
			expStackHistory: [][]string{
				{""},
				{"", ""},
				{"", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", ""},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{"01"},
			},
			expPushData: []string{"", "", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "01"},
			expOpcodes:  []string{"OP_0", "OP_DUP", "OP_HASH160", "OP_SWAP", "OP_SHA256", "OP_RIPEMD160", "OP_EQUAL"},
		},
		"error script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5457",
			expStackHistory: [][]string{
				{"04"},
				{"04", "07"},
				{"04", "07", "02"},
				{"04", "07", "02", "03"},
				{"04", "07", "06"},
				{"04", ""},
			},
			expPushData: []string{"04", "07", "02", "03", "06", ""},
			expOpcodes:  []string{"OP_4", "OP_7", "OP_2", "OP_3", "OP_MUL", "OP_EQUALVERIFY"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ls, err := bscript.NewFromHexString(test.lockingScriptHex)
			assert.NoError(t, err)

			uls, err := bscript.NewFromHexString(test.unlockingScriptHex)
			assert.NoError(t, err)

			history := &stateHistory{
				dstack:  make([][]string, 0),
				astack:  make([][]string, 0),
				opcodes: make([]string, 0),
				entries: make([]string, 0),
			}

			debugger := debug.NewDebugger()
			debugger.AttachAfterStackPush(func(state *interpreter.State, data []byte) {
				stack := make([]string, len(state.DataStack))
				for i, d := range state.DataStack {
					stack[i] = hex.EncodeToString(d)
				}
				history.dstack = append(history.dstack, stack)
				history.opcodes = append(history.opcodes, state.Opcode().Name())
				history.entries = append(history.entries, hex.EncodeToString(data))
			})

			interpreter.NewEngine().Execute(
				interpreter.WithScripts(ls, uls),
				interpreter.WithAfterGenesis(),
				interpreter.WithDebugger(debugger),
			)

			assert.Equal(t, test.expStackHistory, history.dstack)
			assert.Equal(t, test.expOpcodes, history.opcodes)
			assert.Equal(t, test.expPushData, history.entries)
		})
	}
}

func TestDebugger_BeforeStackPop(t *testing.T) {
	t.Parallel()

	type stateHistory struct {
		dstack  [][]string
		astack  [][]string
		opcodes []string
	}

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStackHistory    [][]string
		expOpcodes         []string
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
			expStackHistory: [][]string{
				{"04", "06", "02", "03"},
				{"04", "06", "02"},
				{"04", "06", "06"},
				{"04", "06"},
				{"04", "01"},
				{"04", "02", "02"},
				{"04", "02"},
				{"04", "04"},
				{"04"},
				{"01"},
			},
			expOpcodes: []string{
				"OP_MUL", "OP_MUL", "OP_EQUALVERIFY", "OP_EQUALVERIFY", "OP_EQUALVERIFY",
				"OP_ADD", "OP_ADD", "OP_EQUAL", "OP_EQUAL", "OP_EQUAL",
			},
		},
		"complex script": {
			lockingScriptHex:   "76a97ca8a687",
			unlockingScriptHex: "00",
			expStackHistory: [][]string{
				{"", ""},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", ""},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{"01"},
			},
			expOpcodes: []string{"OP_HASH160", "OP_SHA256", "OP_RIPEMD160", "OP_EQUAL", "OP_EQUAL", "OP_EQUAL"},
		},
		"error script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5457",
			expStackHistory: [][]string{
				{"04", "07", "02", "03"},
				{"04", "07", "02"},
				{"04", "07", "06"},
				{"04", "07"},
				{"04", ""},
			},
			expOpcodes: []string{"OP_MUL", "OP_MUL", "OP_EQUALVERIFY", "OP_EQUALVERIFY", "OP_EQUALVERIFY"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ls, err := bscript.NewFromHexString(test.lockingScriptHex)
			assert.NoError(t, err)

			uls, err := bscript.NewFromHexString(test.unlockingScriptHex)
			assert.NoError(t, err)

			history := &stateHistory{
				dstack:  make([][]string, 0),
				astack:  make([][]string, 0),
				opcodes: make([]string, 0),
			}

			debugger := debug.NewDebugger()
			debugger.AttachBeforeStackPop(func(state *interpreter.State) {
				stack := make([]string, len(state.DataStack))
				for i, d := range state.DataStack {
					stack[i] = hex.EncodeToString(d)
				}
				history.dstack = append(history.dstack, stack)
				history.opcodes = append(history.opcodes, state.Opcode().Name())
			})

			interpreter.NewEngine().Execute(
				interpreter.WithScripts(ls, uls),
				interpreter.WithAfterGenesis(),
				interpreter.WithDebugger(debugger),
			)

			assert.Equal(t, test.expStackHistory, history.dstack)
			assert.Equal(t, test.expOpcodes, history.opcodes)
		})
	}
}

func TestDebugger_AfterStackPop(t *testing.T) {
	t.Parallel()

	type stateHistory struct {
		dstack  [][]string
		astack  [][]string
		opcodes []string
		entries []string
	}

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStackHistory    [][]string
		expOpcodes         []string
		expPopData         []string
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
			expStackHistory: [][]string{
				{"04", "06", "02"},
				{"04", "06"},
				{"04", "06"},
				{"04"},
				{"04"},
				{"04", "02"},
				{"04"},
				{"04"},
				{},
				{},
			},
			expOpcodes: []string{
				"OP_MUL", "OP_MUL", "OP_EQUALVERIFY", "OP_EQUALVERIFY", "OP_EQUALVERIFY",
				"OP_ADD", "OP_ADD", "OP_EQUAL", "OP_EQUAL", "OP_EQUAL",
			},
			expPopData: []string{"03", "02", "06", "06", "01", "02", "02", "04", "04", "01"},
		},
		"complex script": {
			lockingScriptHex:   "76a97ca8a687",
			unlockingScriptHex: "00",
			expStackHistory: [][]string{
				{""},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{},
				{},
			},
			expOpcodes: []string{"OP_HASH160", "OP_SHA256", "OP_RIPEMD160", "OP_EQUAL", "OP_EQUAL", "OP_EQUAL"},
			expPopData: []string{"", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "01"},
		},
		"error script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5457",
			expStackHistory: [][]string{
				{"04", "07", "02"},
				{"04", "07"},
				{"04", "07"},
				{"04"},
				{"04"},
			},
			expOpcodes: []string{"OP_MUL", "OP_MUL", "OP_EQUALVERIFY", "OP_EQUALVERIFY", "OP_EQUALVERIFY"},
			expPopData: []string{"03", "02", "06", "07", ""},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ls, err := bscript.NewFromHexString(test.lockingScriptHex)
			assert.NoError(t, err)

			uls, err := bscript.NewFromHexString(test.unlockingScriptHex)
			assert.NoError(t, err)

			history := &stateHistory{
				dstack:  make([][]string, 0),
				astack:  make([][]string, 0),
				opcodes: make([]string, 0),
				entries: make([]string, 0),
			}

			debugger := debug.NewDebugger()
			debugger.AttachAfterStackPop(func(state *interpreter.State, data []byte) {
				stack := make([]string, len(state.DataStack))
				for i, d := range state.DataStack {
					stack[i] = hex.EncodeToString(d)
				}
				history.dstack = append(history.dstack, stack)
				history.opcodes = append(history.opcodes, state.Opcode().Name())
				history.entries = append(history.entries, hex.EncodeToString(data))
			})

			interpreter.NewEngine().Execute(
				interpreter.WithScripts(ls, uls),
				interpreter.WithAfterGenesis(),
				interpreter.WithDebugger(debugger),
			)

			assert.Equal(t, test.expStackHistory, history.dstack)
			assert.Equal(t, test.expOpcodes, history.opcodes)
			assert.Equal(t, test.expPopData, history.entries)
		})
	}
}
