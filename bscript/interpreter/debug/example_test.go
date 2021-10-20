package debug_test

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/bscript/interpreter"
	"github.com/libsv/go-bt/v2/bscript/interpreter/debug"
)

func ExampleDebugger_AfterStep() {
	lockingScript, err := bscript.NewFromASM("777f726c64 OP_SWAP OP_CAT OP_SHA256 3376118fc0230e6054e782fb31ae52ebcfd551342d8d026c209997e0127b6f74 OP_EQUAL")
	if err != nil {
		fmt.Println(err)
		return
	}

	unlockingScript, err := bscript.NewFromASM(hex.EncodeToString([]byte("hello")))
	if err != nil {
		fmt.Println(err)
		return
	}

	debugger := debug.NewDebugger()
	debugger.AttachAfterStep(func(state *interpreter.State) {
		frames := make([]string, len(state.DataStack))
		for i, frame := range state.DataStack {
			frames[i] = hex.EncodeToString(frame)
		}
		fmt.Println(strings.Join(frames, "|"))
	})

	if err := interpreter.NewEngine().Execute(
		interpreter.WithScripts(lockingScript, unlockingScript),
		interpreter.WithAfterGenesis(),
		interpreter.WithDebugger(debugger),
	); err != nil {
		fmt.Println(err)
	}

	// Output:
	// 68656c6c6f
	// 68656c6c6f|777f726c64
	// 777f726c64|68656c6c6f
	// 777f726c6468656c6c6f
	// 8a0e597fd66749ca1a2f098f4ef706422c63a96dceef4abfd74517b10cd12f63
	// 8a0e597fd66749ca1a2f098f4ef706422c63a96dceef4abfd74517b10cd12f63|8376118fc0230e6054e782fb31ae52ebcfd551342d8d026c209997e0127b6f74
	//
}
