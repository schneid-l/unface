// Package walker is the unfacegen struct-walker golden input. Config
// declares fields with various supported tag modifiers (positional name,
// required, alias=, -) so the generated walker exercises each branch.
package walker

// Config is the fixture struct for walker-mode golden tests.
type Config struct {
	Host    string `unface:"host,required"`
	Port    int    `unface:"port"`
	Tags    []string
	Secret  string `unface:"-"`
	Backend string `unface:"backend,alias=server,alias=upstream"`
}
