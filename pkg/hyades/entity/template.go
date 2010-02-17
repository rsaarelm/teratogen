package entity

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
