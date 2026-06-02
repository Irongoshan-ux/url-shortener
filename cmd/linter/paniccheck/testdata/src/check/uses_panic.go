package check

func UsesPanic() {
	panic("oops") // want "avoid panic"
}
