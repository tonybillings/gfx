package _test

func PanicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
