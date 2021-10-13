package interpreter

// ThreadState a snapshot of a threads state during execution.
type ThreadState struct {
	DStack        [][]byte
	AStack        [][]byte
	CurrentOpcode ParsedOp
	Scripts       []ParsedScript
}

// DebugThreadStateFunc debug handler for a threads state.
type DebugThreadStateFunc func(state *ThreadState)

// DebugStackFunc debug handler for stack operations.
type DebugStackFunc func(state *ThreadState, data []byte)

// DebugExecutionErrorFunc debug handler for execution failure.
type DebugExecutionErrorFunc func(state *ThreadState, err error)

type threadState interface {
	state() *ThreadState
}

type noopThreadState struct{}

func (n *noopThreadState) state() *ThreadState {
	return &ThreadState{}
}

// Debugger for debugging execution.
type Debugger struct {
	ts threadState

	beforeExecuteOpcodeFns []DebugThreadStateFunc
	afterExecuteOpcodefns  []DebugThreadStateFunc

	afterExecutionFns []DebugThreadStateFunc
	afterSuccessFns   []DebugThreadStateFunc
	afterErrorFns     []DebugExecutionErrorFunc

	beforeStackPushFns []DebugStackFunc
	afterStackPushFns  []DebugStackFunc

	beforeStackPopFns []DebugThreadStateFunc
	afterStackPopFns  []DebugStackFunc
}

// NewDebugger returns an empty debugger, to be configured with `Attach`
// functions.
// Example usage:
//  debugger := interpreter.NewDebugger()
//  debugger.AttachBeforeExecuteOpcode(func (state *interpreter.ThreadState) {
//      fmt.Println(state.DStack)
//  })
//  debugger.AttachAfterStackPush(func (state *interpreter.ThreadState, data []byte) {
//      fmt.Println(hex.EncodeToString(data))
//  })
//  engine.Execute(interpreter.WithDebugger(debugger))
func NewDebugger() *Debugger {
	return &Debugger{
		ts:                     &noopThreadState{},
		beforeExecuteOpcodeFns: make([]DebugThreadStateFunc, 0),
		afterExecuteOpcodefns:  make([]DebugThreadStateFunc, 0),
		afterExecutionFns:      make([]DebugThreadStateFunc, 0),
		afterSuccessFns:        make([]DebugThreadStateFunc, 0),
		afterErrorFns:          make([]DebugExecutionErrorFunc, 0),
		beforeStackPushFns:     make([]DebugStackFunc, 0),
		afterStackPushFns:      make([]DebugStackFunc, 0),
		beforeStackPopFns:      make([]DebugThreadStateFunc, 0),
		afterStackPopFns:       make([]DebugStackFunc, 0),
	}
}

func (d *Debugger) attach(t threadState) {
	d.ts = t
}

// AttachBeforeExecuteOpcode attach the provided function to be executed before
// an opcodes execution.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *Debugger) AttachBeforeExecuteOpcode(fn DebugThreadStateFunc) {
	d.beforeExecuteOpcodeFns = append(d.beforeExecuteOpcodeFns, fn)
}

// AttachAfterExecuteOpcode attach the provided function to be executed after
// an opcodes execution.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *Debugger) AttachAfterExecuteOpcode(fn DebugThreadStateFunc) {
	d.afterExecuteOpcodefns = append(d.afterExecuteOpcodefns, fn)
}

// AttachAfterExecution attach the provided function to be executed after
// all scripts have completed execution, but before the final stack value
// is checked.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *Debugger) AttachAfterExecution(fn DebugThreadStateFunc) {
	d.afterExecutionFns = append(d.afterExecutionFns, fn)
}

// AttachAfterSuccess attach the provided function to be executed on
// successful execution.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *Debugger) AttachAfterSuccess(fn DebugThreadStateFunc) {
	d.afterSuccessFns = append(d.afterSuccessFns, fn)
}

// AttachAfterError attach the provided function to be executed on execution
// error.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *Debugger) AttachAfterError(fn DebugExecutionErrorFunc) {
	d.afterErrorFns = append(d.afterErrorFns, fn)
}

// AttachBeforeStackPush attach the provided function to be executed just before
// a push to a stack.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *Debugger) AttachBeforeStackPush(fn DebugStackFunc) {
	d.beforeStackPushFns = append(d.beforeStackPushFns, fn)
}

// AttachAfterStackPush attach the provided function to be executed immediately
// after a push to a stack.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *Debugger) AttachAfterStackPush(fn DebugStackFunc) {
	d.afterStackPushFns = append(d.afterStackPushFns, fn)
}

// AttachBeforeStackPop attach the provided function to be executed just before
// a pop to a stack.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *Debugger) AttachBeforeStackPop(fn DebugThreadStateFunc) {
	d.beforeStackPopFns = append(d.beforeStackPopFns, fn)
}

// AttachAfterStackPop attach the provided function to be executed immediately
// after a pop from a stack.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *Debugger) AttachAfterStackPop(fn DebugStackFunc) {
	d.afterStackPopFns = append(d.afterStackPopFns, fn)
}

func (d *Debugger) beforeExecuteOpcode() {
	state := d.ts.state()
	for _, fn := range d.beforeExecuteOpcodeFns {
		fn(state)
	}
}

func (d *Debugger) afterExecuteOpcode() {
	state := d.ts.state()
	for _, fn := range d.afterExecuteOpcodefns {
		fn(state)
	}
}

func (d *Debugger) afterExecution() {
	state := d.ts.state()
	for _, fn := range d.afterExecutionFns {
		fn(state)
	}
}

func (d *Debugger) afterSuccess() {
	state := d.ts.state()
	for _, fn := range d.afterSuccessFns {
		fn(state)
	}
}

func (d *Debugger) afterError(err error) {
	state := d.ts.state()
	for _, fn := range d.afterErrorFns {
		fn(state, err)
	}
}

func (d *Debugger) beforeStackPush(data []byte) {
	state := d.ts.state()
	for _, fn := range d.beforeStackPushFns {
		fn(state, data)
	}
}

func (d *Debugger) afterStackPush(data []byte) {
	state := d.ts.state()
	for _, fn := range d.afterStackPushFns {
		fn(state, data)
	}
}

func (d *Debugger) beforeStackPop() {
	state := d.ts.state()
	for _, fn := range d.beforeStackPopFns {
		fn(state)
	}
}

func (d *Debugger) afterStackPop(data []byte) {
	state := d.ts.state()
	for _, fn := range d.afterStackPopFns {
		fn(state, data)
	}
}
