package rbac

// Role map role to the constant
type Role string

// user role constants
const (
	RoleMember Role = "MEMBER"
	RoleAdmin  Role = "ADMIN"
)

// Resource is a resource
type Resource string

// Resource constants
const (
	ResourceUser   Resource = "user"   // this resource is used for admin-specific usecases
	ResourceMember Resource = "member" // this resource is used for 'space member' role specific usecases
)

// Action is an action
type Action string

// Action constants
const (
	ActionCreateAny  Action = "create_any"
	ActionViewAny    Action = "view_any"
	ActionEditAny    Action = "edit_any"
	ActionDeleteAny  Action = "delete_any"
	ActionChangeRole Action = "change_role"
)

// TraversePermission traverse the built in permission
func TraversePermission(cb func(role Role, rsc Resource, act Action)) {
	for rra, roles := range _permissions {
		for _, role := range roles {
			cb(role, rra.Resource, rra.Action)
		}
	}
}

type resourceAction struct {
	Resource Resource
	Action   Action
}

// describe the permission for the Role here
var _permissions = map[resourceAction][]Role{
	{ResourceUser, ActionViewAny}:   {RoleAdmin},
	{ResourceUser, ActionEditAny}:   {RoleAdmin},
	{ResourceUser, ActionDeleteAny}: {RoleAdmin},

	{ResourceMember, ActionViewAny}:    {RoleAdmin, RoleMember},
	{ResourceMember, ActionEditAny}:    {RoleAdmin},
	{ResourceMember, ActionChangeRole}: {RoleAdmin},
	{ResourceMember, ActionDeleteAny}:  {RoleAdmin},
}
