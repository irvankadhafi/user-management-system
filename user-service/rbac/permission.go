package rbac

import "sync"

var mutex = &sync.RWMutex{}

// ResourceAction pair of resource and action
type ResourceAction struct {
	Resource Resource
	Action   Action
}

// Permission to store RBAC permission
type Permission struct {
	RRA map[Role][]ResourceAction
}

// NewPermission constructor
func NewPermission() *Permission {
	return &Permission{}
}

// Add add permission for Role to the resource action
func (p *Permission) Add(role Role, rsc Resource, act Action) {
	if p.RRA == nil {
		p.RRA = make(map[Role][]ResourceAction)
	}

	mutex.RLock()
	rra := append(p.RRA[role], ResourceAction{
		Resource: rsc,
		Action:   act,
	})
	mutex.RUnlock()

	mutex.Lock()
	p.RRA[role] = rra
	mutex.Unlock()
}

// RolePermission hold authorization for a role
type RolePermission struct {
	resourceActionPair map[string]bool
	ResourceAction     []ResourceAction
}

// GetResourceAction the role ResourceAction
func (r *RolePermission) GetResourceAction() []ResourceAction {
	if r == nil {
		return []ResourceAction{}
	}
	return r.ResourceAction
}

// NewRolePermission get the permission for the given role
func NewRolePermission(role Role, permission *Permission) *RolePermission {
	resourceActions := permission.RRA[role]
	return &RolePermission{
		ResourceAction: resourceActions,
	}
}

// HasAccess check if the Role has an access to do the Action to the Resource
func (r *RolePermission) HasAccess(resource Resource, action Action) bool {
	if r == nil {
		return false
	}

	if r.resourceActionPair != nil {
		mutex.RLock()
		defer mutex.RUnlock()
		return r.resourceActionPair[newResourceActionKey(resource, action)]
	}

	r.resourceActionPair = make(map[string]bool)
	mutex.Lock()
	for _, ra := range r.ResourceAction {
		r.resourceActionPair[newResourceActionKey(ra.Resource, ra.Action)] = true
	}
	mutex.Unlock()

	mutex.RLock()
	defer mutex.RUnlock()
	return r.resourceActionPair[newResourceActionKey(resource, action)]
}

func newResourceActionKey(rsc Resource, act Action) string {
	return string(rsc) + "_" + string(act)
}
