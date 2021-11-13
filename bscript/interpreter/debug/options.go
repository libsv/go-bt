package debug

// DebuggerOptionFunc for setting debugger options.
type DebuggerOptionFunc func(o *debugOpts)

// WithRewind configure the debugger to enable rewind functionality. When
// enabled, the debugger will save each stack frame from BeforeStep to memory.
func WithRewind() DebuggerOptionFunc {
	return func(o *debugOpts) {
		o.rewind = true
	}
}
