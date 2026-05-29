package plugin

const DefaultPriority = 100

type PriorityPlugin interface {
	Priority() int
}

func getPriority(p Plugin) int {
	var priority int
	if pr, ok := p.(PriorityPlugin); ok {
		priority = pr.Priority()
	}

	if priority > 0 {
		return priority
	}

	return DefaultPriority
}
