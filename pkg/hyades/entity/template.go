package entity

import (
	"hyades/dbg"
	"hyades/mem"
	"reflect"
)

// A ComponentTemplate is an object that creates components for new entities.
type ComponentTemplate interface {
	// Derive returns a new ComponentTemplate that is the self template
	// modified by the contents of the child template. Generally the new
	// component is a copy of the parent (self), with new values specified
	// overriding old ones.
	Derive(child ComponentTemplate) ComponentTemplate

	// MakeComponent creates the component specified by this template and adds
	// it to the given entity.
	MakeComponent(manager *Manager, guid Id)
}

// DefaultTemplate is a convenience type for specifying contentless templates.
type DefaultTemplate struct {
	component reflect.Type
	family    ComponentFamily
	values    map[string]interface{}
}

// NewDefaultTemplate makes a default template using a value of the component
// type (may be nil, as long as it's cast to the correct type), and a map of
// default values (may be nil).
func NewDefaultTemplate(componentRef interface{}, family ComponentFamily, values map[string]interface{}) (result *DefaultTemplate) {
	result = new(DefaultTemplate)
	result.component = reflect.Typeof(componentRef)
	result.family = family
	result.values = make(map[string]interface{})
	if values != nil {
		for k, v := range values {
			result.values[k] = v
		}
	}

	return result
}

func (self *DefaultTemplate) Derive(child ComponentTemplate) ComponentTemplate {
	childDefault := child.(*DefaultTemplate)
	dbg.Assert(childDefault.component == self.component,
		"Trying to derive DefaultComponent for a different type.")
	dbg.Assert(childDefault.family == self.family,
		"Trying to derive DefaultComponent for a different component family.")

	result := new(DefaultTemplate)
	result.component = self.component
	result.family = self.family
	result.values = make(map[string]interface{})
	for k, v := range self.values {
		result.values[k] = v
	}
	for k, v := range childDefault.values {
		result.values[k] = v
	}

	return result
}

func (self *DefaultTemplate) MakeComponent(manager *Manager, guid Id) {
	componentVal := mem.BlankCopyOfType(self.component)

	err := mem.AssignFields(componentVal, self.values)
	dbg.AssertNoError(err)

	manager.Handler(self.family).Add(guid, componentVal.Interface())
}


// An assemblage is a collection of component templates. They specify entity
// templates templates.
type Assemblage map[ComponentFamily]ComponentTemplate

// MakeEntity applies all component templates in the assemblage to a new
// entity in an unspecified order and returns the id of the resulting new
// entity.
func (self Assemblage) MakeEntity(manager *Manager) (result Id) {
	result = manager.NewEntity()
	for _, v := range self {
		v.MakeComponent(manager, result)
	}
	return
}

// Derive creates a new Assemblage which is a union of self and child Assemblages where components occurring in both self (as p) and in child (as c) are converted into p.Derive(c)
func (self Assemblage) Derive(child Assemblage) (result Assemblage) {
	result = copyAssemblage(self)
	for k, v := range child {
		if oldV, ok := result[k]; ok {
			v = oldV.Derive(v)
		}
		result[k] = v
	}
	return
}

func copyAssemblage(assemblage Assemblage) (result Assemblage) {
	result = make(Assemblage)
	for k, v := range assemblage {
		result[k] = v
	}
	return
}
