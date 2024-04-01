package main

import (
	"fmt"
	"os/exec"
)

func main() {
	str := "python3 -m http.server 8000"
	out, _ := exec.Command("pgrep", "-f", str).Output()
	fmt.Println(len(out) > 0)
}
