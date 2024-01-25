package grpc

type Socket interface {
}

func check(e any) {
	if e != nil {
		panic(e)
	}
}
