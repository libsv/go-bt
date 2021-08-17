package interpreter

// EngineOptFunc provides a reference to the engine for configuration
type EngineOptFunc func(*Engine)

// WithSignatureCache sets a SigCache
func WithSignatureCache(s SigCache) EngineOptFunc {
	return func(e *Engine) {
		e.sigCache = s
	}
}

// WithNopSignatureCache sets a nop SigCache, effectively not using one
func WithNopSignatureCache() EngineOptFunc {
	return func(e *Engine) {
		e.sigCache = &nopSigCache{}
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
