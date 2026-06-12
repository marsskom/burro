package pluginapi

type Event struct {
	Name string
	From string
	Data any
}

type EventHandler func(event Event)

type EventBus interface {
	Emit(event Event) error
	On(name string, handler EventHandler) func()
	Off(name string, id int)
}
