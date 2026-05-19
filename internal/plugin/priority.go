package plugin

const DefaultPriority = 100

type PriorityPlugin interface {
	Priority() int
}

func getPriority(p Plugin) int {
	if pr, ok := p.(PriorityPlugin); ok {
		return pr.Priority()
	}

	return DefaultPriority
}
