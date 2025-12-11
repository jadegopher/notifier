package cmd

import "notifier"

func main() {
	n := notifier.Default("")

	n.Start()

}
