package interpreter

type EngineOptFunc func(*Engine)

func WithSignatureCache(s *SigCache) EngineOptFunc {
	return func(e *Engine) {
		e.sigCache = s
	}
}

func WithHashCache(h *TxSigHashes) EngineOptFunc {
	return func(e *Engine) {
		e.hashCache = h
	}
}

func WithFlags(s ScriptFlags) EngineOptFunc {
	return func(e *Engine) {
		e.flags = s
	}
}

func WithParser(o OpCodeParser) EngineOptFunc {
	return func(e *Engine) {
		e.scriptParser = o
	}
}

func WithDefaultParser() EngineOptFunc {
	return func(e *Engine) {
		e.scriptParser = NewOpCodeParser()
	}
}
