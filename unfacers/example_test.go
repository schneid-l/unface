package unfacers_test

import (
	"fmt"

	"github.com/schneid-l/unface/engine"
	"github.com/schneid-l/unface/unfacers"
)

// ExampleStandardPlugin shows that loading StandardPlugin into a Facer
// turns on the full built-in coercion kit: strings parse into ints, map
// keys fold into CamelCase struct fields, and nested maps drive nested
// structs.
func ExampleStandardPlugin() {
	type addr struct {
		Host string `unface:"host"`
		Port int    `unface:"port"`
	}
	type server struct {
		Name   string `unface:"name"`
		Listen addr   `unface:"listen"`
	}
	f := engine.New(engine.With(unfacers.StandardPlugin))
	raw := map[string]any{
		"name": "demo",
		"listen": map[string]any{
			"host": "localhost",
			"port": "8080", // string coerced into int
		},
	}
	var s server
	if err := f.Unface(raw, &s); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%s %s:%d\n", s.Name, s.Listen.Host, s.Listen.Port)
	// Output: demo localhost:8080
}
