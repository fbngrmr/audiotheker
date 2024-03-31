package main

import (
	"github.com/fbngrmr/audiotheker/cmd"
)

func main() {
	cmd.SetupCli()

	cmd.Execute()
}
