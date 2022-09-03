package adapter

type Box interface {
	Service
	Router() Router
}
