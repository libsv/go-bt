// Copyright (c) 2013-2018 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interpreter

// Engine is the virtual machine that executes scripts.
type Engine interface {
	Execute(opts ...ExecutionOptionFunc) error
}

type engine struct{}

// NewEngine returns a new script engine for the provided locking script
// (of a previous transaction out), transaction, and input index.  The
// flags modify the behaviour of the script engine according to the
// description provided by each flag.
func NewEngine() Engine {
	return &engine{}
}

// Execute will execute all scripts in the script engine and return either nil
// for successful validation or an error if one occurred.
//
// Execute with tx example:
//  if err := engine.Execute(
//      interpreter.WithTx(tx, inputIdx, previousOutput),
//      interpreter.WithAfterGenesis(),
//      interpreter.WithForkID(),
//  ); err != nil {
//      // handle err
//  }
//
// Execute with scripts example:
//  if err := engine.Execute(
//      interpreter.WithScripts(lockingScript, unlockingScript),
//      interpreter.WithAfterGenesis(),
//      interpreter.WithForkID(),
//  }); err != nil {
//      // handle err
//  }
//
func (e *engine) Execute(oo ...ExecutionOptionFunc) error {
	opts := &execOpts{}
	for _, o := range oo {
		o(opts)
	}

	t, err := createThread(opts)
	if err != nil {
		return err
	}

	if err := t.execute(); err != nil {
		t.afterError(err)
		return err
	}

	return nil
}
