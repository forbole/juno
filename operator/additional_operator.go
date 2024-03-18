package operator

import "github.com/forbole/juno/v5/interfaces"

var _ interfaces.AdditionalOperator = &Operator{}

type Operator struct {
	modules []interfaces.AdditionalOperationsModule
}

func NewAdditionalOperator() *Operator {
	return &Operator{}
}

func (o *Operator) Register(module interfaces.AdditionalOperationsModule) {
	o.modules = append(o.modules, module)
}

func (o *Operator) Start() error {
	for _, module := range o.modules {
		if err := module.RunAdditionalOperations(); err != nil {
			return err
		}
	}
	return nil
}
