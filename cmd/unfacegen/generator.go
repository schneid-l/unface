// Package main implements unfacegen — a code generator that emits
// unface.Unfacer methods for user types. Two generation modes are
// supported:
//
//   - dispatch (default): the emitted method is a type switch that
//     dispatches src to whichever Un*er methods the type already
//     implements (Unstringer → Unstring, Unbooler → Unbool, ...).
//   - walker: the emitted method inspects the struct's fields and tags
//     and inlines a zero-reflection StructPlugin-equivalent walker.
//
// See the package-level main.go for CLI usage.
package main

import (
	"bytes"
	"fmt"
	"go/types"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Generator inspects a Go package and emits Unfacer methods for requested
// named types.
type Generator struct {
	PackageName string
	Types       []TypeInfo
	Mode        string // "dispatch" or "walker"
	Strict      bool   // emit unface.Strict.Unface fallback instead of unface.Unface
	Tags        string // optional //go:build constraint tags
	MatchMode   string // "exact"|"fold"|"insensitive" — walker-only; only "exact" is honored
}

// TypeInfo describes one target type and the Un*er interfaces it
// satisfies; for walker mode it also carries the parsed field list.
type TypeInfo struct {
	Name    string
	Recv    string // e.g. "*URL" — the receiver we'll generate the method on
	Methods map[string]bool

	// Walker-mode metadata.
	IsStruct bool
	Fields   []FieldInfo
}

// FieldInfo is one struct field in walker mode.
type FieldInfo struct {
	GoName   string   // Go field name
	Key      string   // primary lookup key
	Aliases  []string // alias=... additional keys
	Required bool
	Skip     bool
}

// Request describes generator input.
type Request struct {
	Dir       string
	Types     []string
	Mode      string // "dispatch" (default) or "walker"
	Strict    bool
	Tags      string
	MatchMode string
}

// Generate loads the package at req.Dir, inspects each requested type,
// and returns the generated file content.
func Generate(req Request) ([]byte, error) {
	mode := req.Mode
	if mode == "" {
		mode = "dispatch"
	}
	if mode != "dispatch" && mode != "walker" {
		return nil, fmt.Errorf("unknown mode %q (want dispatch|walker)", mode)
	}
	matchMode := req.MatchMode
	if matchMode == "" {
		matchMode = "exact"
	}
	switch matchMode {
	case "exact", "fold", "insensitive":
	default:
		return nil, fmt.Errorf("unknown matchmode %q (want exact|fold|insensitive)", matchMode)
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedSyntax | packages.NeedDeps | packages.NeedImports,
		Dir: req.Dir,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("load package: %w", err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no package found at %s", req.Dir)
	}
	pkg := pkgs[0]
	if len(pkg.Errors) > 0 {
		for _, e := range pkg.Errors {
			if e.Kind == packages.ParseError {
				return nil, fmt.Errorf("parse: %s", e.Msg)
			}
		}
	}

	g := &Generator{
		PackageName: pkg.Name,
		Mode:        mode,
		Strict:      req.Strict,
		Tags:        req.Tags,
		MatchMode:   matchMode,
	}
	for _, name := range req.Types {
		obj := pkg.Types.Scope().Lookup(name)
		if obj == nil {
			return nil, fmt.Errorf("type %q not found in package %q", name, pkg.Name)
		}
		tn, ok := obj.(*types.TypeName)
		if !ok {
			return nil, fmt.Errorf("%q is not a type", name)
		}
		ti, err := inspect(tn, mode)
		if err != nil {
			return nil, err
		}
		g.Types = append(g.Types, ti)
	}

	return g.render()
}

// inspect computes the method set of *T and (for walker mode) the struct
// field list + parsed tags.
func inspect(tn *types.TypeName, mode string) (TypeInfo, error) {
	named, ok := tn.Type().(*types.Named)
	if !ok {
		return TypeInfo{}, fmt.Errorf("%s is not a named type", tn.Name())
	}

	ti := TypeInfo{
		Name:    tn.Name(),
		Recv:    "*" + tn.Name(),
		Methods: map[string]bool{},
	}

	mset := types.NewMethodSet(types.NewPointer(named))
	for i := 0; i < mset.Len(); i++ {
		m := mset.At(i)
		fn, ok := m.Obj().(*types.Func)
		if !ok {
			continue
		}
		switch fn.Name() {
		case "Unstring", "Unbool", "Unnumber", "Unbytes", "Unrune",
			"Unmap", "Unlist", "Unnil", "Untime", "Unduration", "Unface":
			ti.Methods[fn.Name()] = true
		}
	}

	if mode == "walker" {
		st, ok := named.Underlying().(*types.Struct)
		if !ok {
			return ti, fmt.Errorf("walker mode: %s is not a struct type", tn.Name())
		}
		ti.IsStruct = true
		for i := 0; i < st.NumFields(); i++ {
			f := st.Field(i)
			if f.Name() == "_" {
				// skip marker fields; struct-wide options are not yet
				// consumed by the walker codegen.
				continue
			}
			if !f.Exported() {
				continue
			}
			fi, err := parseFieldTag(f.Name(), st.Tag(i))
			if err != nil {
				return ti, fmt.Errorf("%s.%s: %w", tn.Name(), f.Name(), err)
			}
			ti.Fields = append(ti.Fields, fi)
		}
	}

	return ti, nil
}

// parseFieldTag parses the `unface:"..."` tag value for one struct field.
// Supported modifiers (v1 of the walker codegen): name (positional),
// `-` (skip), `required`, `alias=X`. Other modifiers are rejected so
// users see a clear error instead of silent behavior drift.
func parseFieldTag(goName, rawTag string) (FieldInfo, error) {
	fi := FieldInfo{GoName: goName, Key: goName}
	raw := reflect.StructTag(rawTag).Get("unface")
	if raw == "" {
		return fi, nil
	}
	parts := strings.Split(raw, ",")
	if parts[0] == "-" {
		fi.Skip = true
		return fi, nil
	}
	if parts[0] != "" {
		fi.Key = parts[0]
	}
	for _, p := range parts[1:] {
		kv := strings.SplitN(p, "=", 2)
		key := kv[0]
		val := ""
		if len(kv) == 2 {
			val = kv[1]
		}
		switch key {
		case "required":
			fi.Required = true
		case "alias":
			if val != "" {
				fi.Aliases = append(fi.Aliases, val)
			}
		case "remainder", "inline", "strict", "match", "omitempty":
			return fi, fmt.Errorf("walker codegen: modifier %q is not supported in this version "+
				"(supported: name, -, required, alias=); use the runtime StructPlugin instead", key)
		default:
			return fi, fmt.Errorf("unknown field-tag modifier %q", key)
		}
	}
	return fi, nil
}

// render produces the (unformatted) Go source for the generated file.
func (g *Generator) render() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Fprintln(&buf, "// Code generated by unfacegen; DO NOT EDIT.")
	if g.Mode == "walker" {
		fmt.Fprintln(&buf, "//")
		fmt.Fprintln(&buf, "// Walker-mode limitations (v1): only the positional name, `-` (skip),")
		fmt.Fprintln(&buf, "// `required`, and `alias=` tag modifiers are honored. `remainder`,")
		fmt.Fprintln(&buf, "// `inline`, `strict`, and `match=` are NOT supported — use the runtime")
		fmt.Fprintln(&buf, "// StructPlugin for those. Match mode is always exact; `-matchmode=fold`")
		fmt.Fprintln(&buf, "// and `insensitive` are accepted for CLI symmetry but have no effect on")
		fmt.Fprintln(&buf, "// the generated walker. Switch to the runtime walker if you need them.")
	}
	if g.Tags != "" {
		fmt.Fprintln(&buf)
		fmt.Fprintf(&buf, "//go:build %s\n", g.Tags)
	}
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "package %s\n\n", g.PackageName)

	imports := g.requiredImports()
	if len(imports) > 0 {
		fmt.Fprintln(&buf, "import (")
		// stdlib first block.
		var std, third []string
		for _, imp := range imports {
			if strings.Contains(imp, ".") {
				third = append(third, imp)
			} else {
				std = append(std, imp)
			}
		}
		sort.Strings(std)
		sort.Strings(third)
		for _, imp := range std {
			fmt.Fprintf(&buf, "\t%q\n", imp)
		}
		if len(std) > 0 && len(third) > 0 {
			fmt.Fprintln(&buf)
		}
		for _, imp := range third {
			fmt.Fprintf(&buf, "\t%q\n", imp)
		}
		fmt.Fprintln(&buf, ")")
		fmt.Fprintln(&buf)
	}

	for _, ti := range g.Types {
		if ti.Methods["Unface"] {
			fmt.Fprintf(&buf, "// %s already implements Unfacer; skipping.\n\n", ti.Name)
			continue
		}
		switch g.Mode {
		case "walker":
			g.renderWalker(&buf, ti)
		default:
			renderType(&buf, ti)
		}
	}
	return buf.Bytes(), nil
}

// requiredImports computes the import list based on the selected mode
// and the per-type method sets.
func (g *Generator) requiredImports() []string {
	set := map[string]bool{}
	if g.Mode == "walker" {
		set["fmt"] = true
		set["github.com/schneid-l/unface"] = true
		set["github.com/schneid-l/unface/plugin"] = true
		return sortedKeys(set)
	}
	// dispatch mode
	set["github.com/schneid-l/unface"] = true
	for _, ti := range g.Types {
		if ti.Methods["Untime"] || ti.Methods["Unduration"] {
			set["time"] = true
		}
	}
	return sortedKeys(set)
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// renderType writes the dispatch-mode Unface method for one TypeInfo.
func renderType(buf *bytes.Buffer, ti TypeInfo) {
	names := make([]string, 0, len(ti.Methods))
	for k := range ti.Methods {
		if k == "Unface" {
			continue
		}
		names = append(names, k)
	}
	sort.Strings(names)

	var switchBody bytes.Buffer
	caseFor := func(kind, method string) {
		fmt.Fprintf(&switchBody, "\tcase %s:\n\t\treturn recv.%s(v)\n", kind, method)
	}
	for _, m := range names {
		switch m {
		case "Unstring":
			caseFor("string", "Unstring")
		case "Unbool":
			caseFor("bool", "Unbool")
		case "Unbytes":
			caseFor("[]byte", "Unbytes")
		case "Unrune":
			caseFor("rune", "Unrune")
		case "Untime":
			caseFor("time.Time", "Untime")
		case "Unduration":
			caseFor("time.Duration", "Unduration")
		case "Unnil":
			fmt.Fprint(&switchBody, "\tcase nil:\n\t\treturn recv.Unnil()\n")
		}
	}

	fmt.Fprintf(buf, "// Unface is generated to dispatch src into the specific Un*er methods\n")
	fmt.Fprintf(buf, "// implemented on %s. Unrecognised src types return ErrNotHandled so the\n", ti.Recv)
	fmt.Fprintf(buf, "// library's plugin pipeline can try defaults.\n")
	fmt.Fprintf(buf, "func (recv %s) Unface(src any) error {\n", ti.Recv)

	if switchBody.Len() > 0 {
		fmt.Fprintf(buf, "\tswitch v := src.(type) {\n")
		buf.Write(switchBody.Bytes())
		fmt.Fprintf(buf, "\t}\n")
	}

	if ti.Methods["Unnumber"] {
		fmt.Fprintf(buf, "\tif n, ok := unface.NumberOf(src); ok {\n")
		fmt.Fprintf(buf, "\t\treturn recv.Unnumber(n)\n")
		fmt.Fprintf(buf, "\t}\n")
	}
	if ti.Methods["Unlist"] {
		fmt.Fprintf(buf, "\tif l, ok := unface.ListOf(src); ok {\n")
		fmt.Fprintf(buf, "\t\treturn recv.Unlist(l)\n")
		fmt.Fprintf(buf, "\t}\n")
	}
	if ti.Methods["Unmap"] {
		fmt.Fprintf(buf, "\tif m, ok := unface.MapOf(src); ok {\n")
		fmt.Fprintf(buf, "\t\treturn recv.Unmap(m)\n")
		fmt.Fprintf(buf, "\t}\n")
	}
	fmt.Fprintf(buf, "\treturn unface.ErrNotHandled\n")
	fmt.Fprintf(buf, "}\n\n")
}

// renderWalker writes the walker-mode Unface method for one TypeInfo.
func (g *Generator) renderWalker(buf *bytes.Buffer, ti TypeInfo) {
	unfaceCall := "unface.Unface"
	if g.Strict {
		unfaceCall = "unface.Strict.Unface"
	}

	// Active (non-skipped) fields, preserving source order.
	active := make([]FieldInfo, 0, len(ti.Fields))
	for _, f := range ti.Fields {
		if f.Skip {
			continue
		}
		active = append(active, f)
	}

	fmt.Fprintf(buf, "// Unface is generated from %s's struct tags. It is behaviorally\n", ti.Name)
	fmt.Fprintf(buf, "// equivalent to the runtime StructPlugin for the supported tag subset\n")
	fmt.Fprintf(buf, "// (positional name, -, required, alias=), but performs no reflection.\n")
	fmt.Fprintf(buf, "func (recv %s) Unface(src any) error {\n", ti.Recv)
	fmt.Fprintf(buf, "\tm, ok := plugin.MapOf(src)\n")
	fmt.Fprintf(buf, "\tif !ok {\n")
	fmt.Fprintf(buf, "\t\treturn fmt.Errorf(\"unface: cannot walk %%T into %s\", src)\n", ti.Recv)
	fmt.Fprintf(buf, "\t}\n")

	// Determine whether we need a `seen` map (only for required fields).
	needSeen := false
	for _, f := range active {
		if f.Required {
			needSeen = true
			break
		}
	}
	if needSeen {
		fmt.Fprintf(buf, "\tseen := make(map[string]bool, %d)\n", len(active))
	}

	for _, f := range active {
		keys := append([]string{f.Key}, f.Aliases...)
		for i, k := range keys {
			if i == 0 {
				fmt.Fprintf(buf, "\tif v, ok := m.Get(%q); ok {\n", k)
			} else {
				fmt.Fprintf(buf, "\t} else if v, ok := m.Get(%q); ok {\n", k)
			}
			if f.Required && i == 0 {
				fmt.Fprintf(buf, "\t\tseen[%q] = true\n", f.Key)
			} else if f.Required {
				fmt.Fprintf(buf, "\t\tseen[%q] = true\n", f.Key)
			}
			fmt.Fprintf(buf, "\t\tif err := %s(v, &recv.%s); err != nil {\n", unfaceCall, f.GoName)
			fmt.Fprintf(buf, "\t\t\treturn fmt.Errorf(\"%s: %%w\", err)\n", f.Key)
			fmt.Fprintf(buf, "\t\t}\n")
		}
		fmt.Fprintf(buf, "\t}\n")
	}

	if needSeen {
		for _, f := range active {
			if !f.Required {
				continue
			}
			fmt.Fprintf(buf, "\tif !seen[%q] {\n", f.Key)
			fmt.Fprintf(buf, "\t\treturn fmt.Errorf(\"%%w: %s\", plugin.ErrRequired)\n", f.Key)
			fmt.Fprintf(buf, "\t}\n")
		}
	}

	fmt.Fprintf(buf, "\treturn nil\n")
	fmt.Fprintf(buf, "}\n\n")
}
