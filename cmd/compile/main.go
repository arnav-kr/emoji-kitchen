package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	var sourceFile, outPath string
	if len(os.Args) < 2 {
		fmt.Println("usage: compile <source> <output>")
		os.Exit(1)
	}
	sourceFile = os.Args[1]
	if len(os.Args) > 2 {
		outPath = os.Args[2]
	} else {
		outPath = fmt.Sprintf("./dist/%s.ekbf", strings.Split(sourceFile, ".")[0])
	}

	fmt.Println("source:", sourceFile)
	fmt.Println("output:", outPath)
}