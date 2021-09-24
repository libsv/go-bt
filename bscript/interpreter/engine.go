// Copyright (c) 2013-2018 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interpreter

// Engine is the virtual machine that executes scripts.
type Engine interface {
	Execute(ExecutionParams) error
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
func (e *engine) Execute(params ExecutionParams) error {
	th := &thread{
		scriptParser: &DefaultOpcodeParser{},
		cfg:          &beforeGenesisConfig{},
	}

	if err := th.apply(params); err != nil {
		return err
	}

	return th.execute()
}
