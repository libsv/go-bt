package debug

import "github.com/libsv/go-bt/v2/bscript/interpreter"

type (
	// ThreadStateFunc debug handler for a threads state.
	ThreadStateFunc func(state *interpreter.State)

	// StackFunc debug handler for stack operations.
	StackFunc func(state *interpreter.State, data []byte)

	// ExecutionErrorFunc debug handler for execution failure.
	ExecutionErrorFunc func(state *interpreter.State, err error)
)

type debugOpts struct {
	rewind bool
}

// DefaultDebugger exposes attachment points via the way of functions, which
// are to be appended to via a series of function calls.
type DefaultDebugger interface {
	AttachBeforeExecute(ThreadStateFunc)
	AttachAfterExecute(ThreadStateFunc)
	AttachBeforeStep(ThreadStateFunc)
	AttachAfterStep(ThreadStateFunc)
	AttachBeforeExecuteOpcode(ThreadStateFunc)
	AttachAfterExecuteOpcode(ThreadStateFunc)
	AttachBeforeScriptChange(ThreadStateFunc)
	AttachAfterScriptChange(ThreadStateFunc)
	AttachAfterSuccess(ThreadStateFunc)
	AttachAfterError(ExecutionErrorFunc)

	AttachBeforeStackPush(StackFunc)
	AttachAfterStackPush(StackFunc)
	AttachBeforeStackPop(ThreadStateFunc)
	AttachAfterStackPop(StackFunc)

	interpreter.Debugger
}

type debugger struct {
	sh interpreter.StateHandler

	beforeExecuteFns []ThreadStateFunc
	afterExecuteFns  []ThreadStateFunc

	beforeStepFns []ThreadStateFunc
	afterStepFns  []ThreadStateFunc

	beforeExecuteOpcodeFns []ThreadStateFunc
	afterExecuteOpcodeFns  []ThreadStateFunc

	beforeScriptChangeFns []ThreadStateFunc
	afterScriptChangeFns  []ThreadStateFunc

	afterSuccessFns []ThreadStateFunc
	afterErrorFns   []ExecutionErrorFunc

	beforeStackPushFns []StackFunc
	afterStackPushFns  []StackFunc

	beforeStackPopFns []ThreadStateFunc
	afterStackPopFns  []StackFunc
}

// NewDebugger returns an empty debugger which is to be configured with the `Attach`
// functions.
//
// Example usage:
//  debugger := debug.NewDebugger()
//  debugger.AttachBeforeExecuteOpcode(func (state *interpreter.State) {
//      fmt.Println(state.DataStack)
//  })
//  debugger.AttachAfterStackPush(func (state *interpreter.State, data []byte) {
//      fmt.Println(hex.EncodeToString(data))
//  })
//  engine.Execute(interpreter.WithDebugger(debugger))
func NewDebugger(oo ...DebuggerOptionFunc) DefaultDebugger {
	opts := &debugOpts{}
	for _, o := range oo {
		o(opts)
	}

	return &debugger{
		beforeExecuteFns: make([]ThreadStateFunc, 0),
		afterExecuteFns:  make([]ThreadStateFunc, 0),

		beforeStepFns: make([]ThreadStateFunc, 0),
		afterStepFns:  make([]ThreadStateFunc, 0),

		beforeExecuteOpcodeFns: make([]ThreadStateFunc, 0),
		afterExecuteOpcodeFns:  make([]ThreadStateFunc, 0),

		afterSuccessFns: make([]ThreadStateFunc, 0),
		afterErrorFns:   make([]ExecutionErrorFunc, 0),

		beforeStackPushFns: make([]StackFunc, 0),
		afterStackPushFns:  make([]StackFunc, 0),

		beforeStackPopFns: make([]ThreadStateFunc, 0),
		afterStackPopFns:  make([]StackFunc, 0),
	}
}

// AttachBeforeExecute attach the provided function to be executed before
// interpreter execution.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *debugger) AttachBeforeExecute(fn ThreadStateFunc) {
	d.beforeExecuteFns = append(d.beforeExecuteFns, fn)
}

// AttachAfterExecute attach the provided function to be executed after
// all scripts have completed execution, but before the final stack value
// is checked.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *debugger) AttachAfterExecute(fn ThreadStateFunc) {
	d.afterExecuteFns = append(d.afterExecuteFns, fn)
}

// AttachBeforeStep attach the provided function to be executed before a thread
// begins a step.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *debugger) AttachBeforeStep(fn ThreadStateFunc) {
	d.beforeStepFns = append(d.beforeStepFns, fn)
}

// AttachAfterStep attach the provided function to be executed after a thread
// finishes a step.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *debugger) AttachAfterStep(fn ThreadStateFunc) {
	d.afterStepFns = append(d.afterStepFns, fn)
}

// AttachBeforeExecuteOpcode attach the provided function to be executed before
// an opcodes execution.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *debugger) AttachBeforeExecuteOpcode(fn ThreadStateFunc) {
	d.beforeExecuteOpcodeFns = append(d.beforeExecuteOpcodeFns, fn)
}

// AttachAfterExecuteOpcode attach the provided function to be executed after
// an opcodes execution.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *debugger) AttachAfterExecuteOpcode(fn ThreadStateFunc) {
	d.afterExecuteOpcodeFns = append(d.afterExecuteOpcodeFns, fn)
}

// AttachBeforeScriptChange attach the provided function to be executed after
// a scripts execution has finished, just before the thread shifts to the
// next.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *debugger) AttachBeforeScriptChange(fn ThreadStateFunc) {
	d.beforeScriptChangeFns = append(d.beforeScriptChangeFns, fn)
}

// AttachAfterScriptChange attach the provided function to be executed after
// a scripts execution has finished, just after the thread shifts to the
// next.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *debugger) AttachAfterScriptChange(fn ThreadStateFunc) {
	d.afterScriptChangeFns = append(d.afterScriptChangeFns, fn)
}

// AttachAfterSuccess attach the provided function to be executed on
// successful execution.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *debugger) AttachAfterSuccess(fn ThreadStateFunc) {
	d.afterSuccessFns = append(d.afterSuccessFns, fn)
}

// AttachAfterError attach the provided function to be executed on execution
// error.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *debugger) AttachAfterError(fn ExecutionErrorFunc) {
	d.afterErrorFns = append(d.afterErrorFns, fn)
}

// AttachBeforeStackPush attach the provided function to be executed just before
// a push to a stack.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *debugger) AttachBeforeStackPush(fn StackFunc) {
	d.beforeStackPushFns = append(d.beforeStackPushFns, fn)
}

// AttachAfterStackPush attach the provided function to be executed immediately
// after a push to a stack.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *debugger) AttachAfterStackPush(fn StackFunc) {
	d.afterStackPushFns = append(d.afterStackPushFns, fn)
}

// AttachBeforeStackPop attach the provided function to be executed just before
// a pop to a stack.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *debugger) AttachBeforeStackPop(fn ThreadStateFunc) {
	d.beforeStackPopFns = append(d.beforeStackPopFns, fn)
}

// AttachAfterStackPop attach the provided function to be executed immediately
// after a pop from a stack.
// If this is called multiple times, provided funcs are executed on a
// FIFO basis.
func (d *debugger) AttachAfterStackPop(fn StackFunc) {
	d.afterStackPopFns = append(d.afterStackPopFns, fn)
}

// BeforeExecute execute all before execute attachments.
func (d *debugger) BeforeExecute(state *interpreter.State) {
	for _, fn := range d.beforeExecuteFns {
		fn(state)
	}
}

// AfterExecute execute all after execute attachments.
func (d *debugger) AfterExecute(state *interpreter.State) {
	for _, fn := range d.afterExecuteFns {
		fn(state)
	}
}

// BeforeStep execute all before step attachments.
func (d *debugger) BeforeStep(state *interpreter.State) {
	for _, fn := range d.beforeStepFns {
		fn(state)
	}
}

// AfterStep execute all after step attachments.
func (d *debugger) AfterStep(state *interpreter.State) {
	for _, fn := range d.afterStepFns {
		fn(state)
	}
}

// BeforeExecuteOpcode execute all before execute opcode attachments.
func (d *debugger) BeforeExecuteOpcode(state *interpreter.State) {
	for _, fn := range d.beforeExecuteOpcodeFns {
		fn(state)
	}
}

// AfterExecuteOpcode execute all after execute opcode attachments.
func (d *debugger) AfterExecuteOpcode(state *interpreter.State) {
	for _, fn := range d.afterExecuteOpcodeFns {
		fn(state)
	}
}

// BeforeScriptChange execute all before script change attachments.
func (d *debugger) BeforeScriptChange(state *interpreter.State) {
	for _, fn := range d.beforeScriptChangeFns {
		fn(state)
	}
}

// AfterScriptChange execute all after script change attachments.
func (d *debugger) AfterScriptChange(state *interpreter.State) {
	for _, fn := range d.afterScriptChangeFns {
		fn(state)
	}
}

// AfterSuccess execute all after success attachments.
func (d *debugger) AfterSuccess(state *interpreter.State) {
	for _, fn := range d.afterSuccessFns {
		fn(state)
	}
}

// AfterError execute all after error attachments.
func (d *debugger) AfterError(state *interpreter.State, err error) {
	for _, fn := range d.afterErrorFns {
		fn(state, err)
	}
}

// BeforeStackPush execute all before stack push attachments.
func (d *debugger) BeforeStackPush(state *interpreter.State, data []byte) {
	for _, fn := range d.beforeStackPushFns {
		fn(state, data)
	}
}

// AfterStackPush execute all after stack push attachments.
func (d *debugger) AfterStackPush(state *interpreter.State, data []byte) {
	for _, fn := range d.afterStackPushFns {
		fn(state, data)
	}
}

// BeforeStackPop execute all before stack pop attachments.
func (d *debugger) BeforeStackPop(state *interpreter.State) {
	for _, fn := range d.beforeStackPopFns {
		fn(state)
	}
}

// AfterStackPop execute all after stack pop attachments.
func (d *debugger) AfterStackPop(state *interpreter.State, data []byte) {
	for _, fn := range d.afterStackPopFns {
		fn(state, data)
	}
}
