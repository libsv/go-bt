package interpreter

// EngineOptFunc provides a reference to the engine for configuration
type EngineOptFunc func(*Engine)

// WithSignatureCache sets a SigCache
func WithSignatureCache(s *SigCache) EngineOptFunc {
	return func(e *Engine) {
		e.sigCache = s
	}
}

// WithHashCache sets a TxSigHashes as a hashcache
func WithHashCache(h *TxSigHashes) EngineOptFunc {
	return func(e *Engine) {
		e.hashCache = h
	}
}

// WithFlags sets flags
func WithFlags(s ScriptFlags) EngineOptFunc {
	return func(e *Engine) {
		e.flags = s
	}
}

// WithParser sets a custom OpCodeParser
func WithParser(o OpcodeParser) EngineOptFunc {
	return func(e *Engine) {
		e.scriptParser = o
	}
}

// WithDefaultParser sets the internal OpCodeParser
func WithDefaultParser() EngineOptFunc {
	return func(e *Engine) {
		e.scriptParser = NewOpcodeParser()
	}
}
