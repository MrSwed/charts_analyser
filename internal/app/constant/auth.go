package constant

type Role int

const (
	RoleVessel   Role = 0x1
	RoleOperator Role = 0x1 << iota
	RoleAdmin
)

func (r Role) CheckIsRole(e Role) bool {
	return r&e != 0
}
