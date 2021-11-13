package interpreter

// Debugger implement to enable debugging.
// If enabled, copies of state are provided to each of the functions on
// call.
//
// Each function is called during its stage of a threads lifecycle.
// A high level overview of this lifecycle is:
//
//   BeforeExecute
//   for step
//      BeforeStep
//      BeforeExecuteOpcode
//      for each stack push
//        BeforeStackPush
//        AfterStackPush
//      end for
//      for each stack pop
//        BeforeStackPop
//        AfterStackPop
//      end for
//      AfterExecuteOpcode
//      if end of script
//        BeforeScriptChange
//        AfterScriptChange
//      end if
//      if bip16 and end of final script
//        BeforeStackPush
//        AfterStackPush
//      end if
//      AfterStep
//   end for
//   AfterExecute
//   if success
//     AfterSuccess
//   end if
//   if error
//     AfterError
//   end if
type Debugger interface {
	BeforeExecute(*State)
	AfterExecute(*State)
	BeforeStep(*State)
	AfterStep(*State)
	BeforeExecuteOpcode(*State)
	AfterExecuteOpcode(*State)
	BeforeScriptChange(*State)
	AfterScriptChange(*State)
	AfterSuccess(*State)
	AfterError(*State, error)

	BeforeStackPush(*State, []byte)
	AfterStackPush(*State, []byte)
	BeforeStackPop(*State)
	AfterStackPop(*State, []byte)
}

type nopDebugger struct{}

func (n *nopDebugger) BeforeExecute(*State) {}

func (n *nopDebugger) AfterExecute(*State) {}

func (n *nopDebugger) BeforeStep(*State) {}

func (n *nopDebugger) AfterStep(*State) {}

func (n *nopDebugger) BeforeExecuteOpcode(*State) {}

func (n *nopDebugger) AfterExecuteOpcode(*State) {}

func (n *nopDebugger) BeforeScriptChange(*State) {}

func (n *nopDebugger) AfterScriptChange(*State) {}

func (n *nopDebugger) BeforeStackPush(*State, []byte) {}

func (n *nopDebugger) AfterStackPush(*State, []byte) {}

func (n *nopDebugger) BeforeStackPop(*State) {}

func (n *nopDebugger) AfterStackPop(*State, []byte) {}

func (n *nopDebugger) AfterSuccess(*State) {}

func (n *nopDebugger) AfterError(*State, error) {}
