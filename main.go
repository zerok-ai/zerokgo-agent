package main

import (
	"os"
)

func main() {
	f, _ := os.Create("out.txt")
	defer f.Close()
	f.WriteString("You are free to start trying.")
	TryItOut(f)
}

func TryItOut(f *os.File) {
	f.WriteString("Be Patient, I'm trying hard.......!!!!!!!!")
}
