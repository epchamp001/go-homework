package models

type ActorType string

const (
	ActorClient  ActorType = "client"
	ActorCourier ActorType = "courier"
)

func (a ActorType) valid() bool {
	return a == ActorClient || a == ActorCourier
}
