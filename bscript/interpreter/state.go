package interpreter

import "github.com/libsv/go-bt/v2/bscript/interpreter/scriptflag"

// State a snapshot of a threads state during execution.
type State struct {
	DataStack            [][]byte
	AltStack             [][]byte
	ElseStack            [][]byte
	CondStack            []int
	SavedFirstStack      [][]byte
	Scripts              []ParsedScript
	ScriptIdx            int
	OpcodeIdx            int
	LastCodeSeparatorIdx int
	NumOps               int
	Flags                scriptflag.Flag
	IsFinished           bool
	Genesis              struct {
		AfterGenesis bool
		EarlyReturn  bool
	}
}

// Opcode the current interpreter.ParsedOpcode from the
// threads program counter.
func (s *State) Opcode() ParsedOpcode {
	return s.Scripts[s.ScriptIdx][s.OpcodeIdx]
}

// RemainingScript the remaining script to be executed.
func (s *State) RemainingScript() ParsedScript {
	return s.Scripts[s.ScriptIdx][s.OpcodeIdx:]
}

// StateHandler interfaces getting and applying state.
type StateHandler interface {
	State() *State
	SetState(state *State)
}

type nopStateHandler struct{}

func (n *nopStateHandler) State() *State {
	return &State{}
}
func (n *nopStateHandler) SetState(state *State) {}

func (t *thread) State() *State {
	scriptIdx := t.scriptIdx
	offsetIdx := t.scriptOff
	if scriptIdx >= len(t.scripts) {
		scriptIdx = len(t.scripts) - 1
		offsetIdx = len(t.scripts[scriptIdx]) - 1
	}

	if offsetIdx >= len(t.scripts[scriptIdx]) {
		offsetIdx = len(t.scripts[scriptIdx]) - 1
	}
	ts := State{
		DataStack:            make([][]byte, int(t.dstack.Depth())),
		AltStack:             make([][]byte, int(t.astack.Depth())),
		ElseStack:            make([][]byte, int(t.elseStack.Depth())),
		CondStack:            make([]int, len(t.condStack)),
		SavedFirstStack:      make([][]byte, len(t.savedFirstStack)),
		Scripts:              make([]ParsedScript, len(t.scripts)),
		ScriptIdx:            scriptIdx,
		OpcodeIdx:            offsetIdx,
		LastCodeSeparatorIdx: t.lastCodeSep,
		NumOps:               t.numOps,
		Flags:                t.flags,
		IsFinished:           t.scriptIdx > scriptIdx,
		Genesis: struct {
			AfterGenesis bool
			EarlyReturn  bool
		}{
			AfterGenesis: t.afterGenesis,
			EarlyReturn:  t.earlyReturnAfterGenesis,
		},
	}

	for i, dd := range t.dstack.stk {
		ts.DataStack[i] = make([]byte, len(dd))
		copy(ts.DataStack[i], dd)
	}

	for i, aa := range t.astack.stk {
		ts.AltStack[i] = make([]byte, len(aa))
		copy(ts.AltStack[i], aa)
	}

	if stk, ok := t.elseStack.(*stack); ok {
		for i, ee := range stk.stk {
			ts.ElseStack[i] = make([]byte, len(ee))
			copy(ts.ElseStack[i], ee)
		}
	}

	for i, ss := range t.savedFirstStack {
		ts.SavedFirstStack[i] = make([]byte, len(ss))
		copy(ts.SavedFirstStack[i], ss)
	}

	copy(ts.CondStack, t.condStack)

	for i, script := range t.scripts {
		ts.Scripts[i] = make(ParsedScript, len(script))
		copy(ts.Scripts[i], script)
	}

	return &ts
}

func (t *thread) SetState(state *State) {
	setStack(&t.dstack, state.DataStack)
	setStack(&t.astack, state.AltStack)
	t.elseStack = &nopBoolStack{}
	if state.Genesis.AfterGenesis {
		es := &stack{debug: &nopDebugger{}, sh: &nopStateHandler{}}
		setStack(es, state.ElseStack)
		t.elseStack = es
	}
	t.condStack = make([]int, len(state.CondStack))
	copy(t.condStack, state.CondStack)
	t.savedFirstStack = state.SavedFirstStack

	t.scripts = state.Scripts
	t.scriptIdx = state.ScriptIdx
	t.scriptOff = state.OpcodeIdx
	t.lastCodeSep = state.LastCodeSeparatorIdx
	t.numOps = state.NumOps
	t.flags = state.Flags
	t.afterGenesis = state.Genesis.AfterGenesis
	t.earlyReturnAfterGenesis = state.Genesis.EarlyReturn
}
