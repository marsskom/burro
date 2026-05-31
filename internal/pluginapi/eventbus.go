package pluginapi

type Event struct {
	Name string
	Data any
	From string
}

type EventHandler func(event Event)

type EventBus interface {
	Emit(event Event) error
	On(name string, handler EventHandler)
}
