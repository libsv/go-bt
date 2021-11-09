package main

import (
	"fmt"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/bscript/interpreter"
	"github.com/libsv/go-bt/v2/bscript/interpreter/scriptflag"
)

func main() {
	e := interpreter.NewEngine()

	ls, err := bscript.NewFromHexString("964f87")
	//ls, err := bscript.NewFromASM("OP_2 OP_2 OP_SUB OP_4 OP_ADD OP_EQUAL")
	if err != nil {
		fmt.Println(err)
	}
	uls, _ := bscript.NewFromHexString("01a0011d")
	//uls, err := bscript.NewFromASM("OP_2")
	if err != nil {
		fmt.Println(err)
	}

	asm, _ := uls.ToASM()
	asm2, _ := ls.ToASM()
	fmt.Println(asm, asm2)

	if err := e.Execute(
		interpreter.WithScripts(
			ls, uls,
		),
		interpreter.WithFlags(scriptflag.VerifyStrictEncoding),
	); err != nil {
		fmt.Println(err)
	}
}
