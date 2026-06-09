package actions

import (
	"errors"
	"fmt"
)

var (
	ErrorActionRuleIDRequired           = errors.New("action rule ID required")
	ErrorActionRulePriorityLessThanZero = errors.New("action rule priority must be >= 0")
	ErrorActionRuleActionRequired       = errors.New("action rule requires at least one action")
	ErrorActionRuleOnMatchValue         = errors.New("action rule `on_match` invalid value")

	ErrorMatchOnlyOneConditionAllowed     = errors.New("match: only one of all/any/not condition allowed")
	ErrorMatchMixLogicalOperatorsWithLeaf = errors.New("cannot mix logical operators with leaf conditions")

	ErrorActionOperationIsInvalid  = errors.New("action tries use invalid operation")
	ErrorActionOperationDecodeAgrs = errors.New("action operations cannot decode arguments")
)

func (r *ActionRule) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("action rule has an error: %w", ErrorActionRuleIDRequired)
	}

	if r.Priority < 0 {
		return fmt.Errorf("action rule '%s' has an error: %w", r.ID, ErrorActionRulePriorityLessThanZero)
	}

	if len(r.Action) == 0 {
		return fmt.Errorf("action rule '%s' has an error: %w", r.ID, ErrorActionRuleActionRequired)
	}

	err := r.Match.validateMatch()
	if err != nil {
		return fmt.Errorf("action rule '%s' has an error: %w", r.ID, err)
	}

	if r.OnMatch == "" {
		r.OnMatch = OnMatchContinue
	}

	if r.OnMatch != OnMatchContinue && r.OnMatch != OnMatchStop {
		return fmt.Errorf("action rule '%s' has an error: %w", r.ID, ErrorActionRuleOnMatchValue)
	}

	errs := make([]error, 0, len(r.Action))
	for _, a := range r.Action {
		err = a.validateAction()
		if err != nil {
			errs = append(errs, fmt.Errorf("action rule '%s' has an error: %w", r.ID, err))
		}
	}

	return errors.Join(errs...)
}

func (m *Match) validateMatch() error {
	if m.All != nil && (m.Any != nil || m.Not != nil) {
		return ErrorMatchOnlyOneConditionAllowed
	}

	hasLogic := len(m.All) > 0 || len(m.Any) > 0 || m.Not != nil

	hasLeaf :=
		m.Method != "" ||
			m.Domain != "" ||
			m.Path != "" ||
			m.IP != "" ||
			len(m.Headers) > 0

	if hasLogic && hasLeaf {
		return ErrorMatchMixLogicalOperatorsWithLeaf
	}

	return nil
}

func (a *Action) validateAction() error {
	if !a.Operation.IsValid() {
		return ErrorActionOperationIsInvalid
	}

	switch a.Operation {
	case OpDeny, OpAllow:
		if a.Args != nil {
			return fmt.Errorf("there are no arguments for a operation '%s'", a.Operation)
		}

	case OpSetHeader:
		v, err := decode[SetHeaderArgs](a.Args)
		if err != nil {
			return fmt.Errorf("operation '%s': %w", a.Operation, ErrorActionOperationDecodeAgrs)
		}
		if v.Name == "" {
			return fmt.Errorf("operation '%s' requires header name", a.Operation)
		}

	case OpRemoveHeader:
		v, err := decode[RemoveHeaderArgs](a.Args)
		if err != nil {
			return fmt.Errorf("operation '%s': %w", a.Operation, ErrorActionOperationDecodeAgrs)
		}
		if len(v.Names) == 0 {
			return fmt.Errorf("operation '%s' requires at least one header name", a.Operation)
		}

	case OpRedactBody:
		v, err := decode[RedactBodyArgs](a.Args)
		if err != nil {
			return fmt.Errorf("operation '%s': %w", a.Operation, ErrorActionOperationDecodeAgrs)
		}
		if len(v.Fields) == 0 {
			return fmt.Errorf("operation '%s' requires at least one field name", a.Operation)
		}

	case OpLog:
		v, err := decode[LogArgs](a.Args)
		if err != nil {
			return fmt.Errorf("operation '%s': %w", a.Operation, ErrorActionOperationDecodeAgrs)
		}
		if v.Level == "" {
			return fmt.Errorf("operation '%s' requires log level", a.Operation)
		}
		if v.Message == "" {
			return fmt.Errorf("operation '%s' requires a message", a.Operation)
		}
	}

	return nil
}

func (op Operation) IsValid() bool {
	switch op {
	case OpDeny,
		OpAllow,
		OpSetHeader,
		OpRemoveHeader,
		OpRedactBody,
		OpLog:
		return true
	default:
		return false
	}
}
