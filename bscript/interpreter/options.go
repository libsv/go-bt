package interpreter

// EngineOptFunc provides a reference to the engine for configuration
type EngineOptFunc func(*Engine)

// WithFlags sets flags
func WithFlags(s ScriptFlags) EngineOptFunc {
	return func(e *Engine) {
		e.flags = s
	}
}

// WithParser sets a custom OpcodeParser
func WithParser(o OpcodeParser) EngineOptFunc {
	return func(e *Engine) {
		e.scriptParser = o
	}
}

// WithDefaultParser sets the internal OpcodeParser
func WithDefaultParser() EngineOptFunc {
	return func(e *Engine) {
		e.scriptParser = NewOpcodeParser()
	}
}
