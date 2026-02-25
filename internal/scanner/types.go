package scanner

// Ecosystem represents a supported package ecosystem.
type Ecosystem string

const (
	EcosystemGo    Ecosystem = "go"
	EcosystemNpm   Ecosystem = "npm"
	EcosystemPip   Ecosystem = "pip"
	EcosystemCargo Ecosystem = "cargo"
	EcosystemBun   Ecosystem = "bun"
)

// Dependency represents a single dependency in any ecosystem.
type Dependency struct {
	Name       string
	Current    string
	Latest     string
	Ecosystem  Ecosystem
	Outdated   bool
	Vulnerable bool
	VulnInfo   string
	Indirect   bool
}
