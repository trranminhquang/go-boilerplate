package api

type Option interface {
	apply(*API)
}
