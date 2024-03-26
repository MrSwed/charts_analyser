package constant

type Role int

const (
	RoleVessel   Role = 0x1
	RoleOperator Role = 0x1 << iota
	RoleAdmin

	PasswordMinLen = 8
	PasswordMaxLen = 30
)

var (
	PasswordCheckRules = [4]string{"[a-z]", "[A-Z]", "[0-9]", "[^\\d\\w]"}
)

func (r Role) CheckIsRole(e Role) bool {
	return r&e != 0
}
