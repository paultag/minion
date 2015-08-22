package main

func main() {
	switch os.Args[1] {
	case "minion":
		BeAMinion()
	case "coordinator":
		BeACoordinator()
	default:
		log.Fatalf("Don't know what to do :(\n")
	}
}
