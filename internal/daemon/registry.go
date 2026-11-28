package daemon

import (
	"sort"
)

type node struct {
	children map[string]*node
	op       *Op
}

// Registry holds all known ops.
type Registry struct {
	ops  map[string]*Op
	root *node

	sources  []Source
	statuses []func() []string
}

func NewRegistry() *Registry {
	return &Registry{ops: map[string]*Op{}, root: &node{children: map[string]*node{}}}
}

// Default is the registry the application's packages register their ops in.
var Default = NewRegistry()

// Register adds an op to the default registry. Intended to be called from init().
func Register(op Op) {
	Default.Register(op)
}

func (r *Registry) Register(op Op) {
	if _, exists := r.ops[op.Name]; exists {
		panic("daemon: duplicate op " + op.Name)
	}
	o := op
	o.stats = &opStats{}
	r.ops[o.Name] = &o

	n := r.root
	for _, word := range o.path() {
		child, ok := n.children[word]
		if !ok {
			child = &node{children: map[string]*node{}}
			n.children[word] = child
		}
		n = child
	}
	n.op = &o
}

// RegisterSource adds an event source that refreshes watched state of the default registry, and
// optionally (status may be nil) the lines it adds to `daemon status`, e.g. its health.
// Intended to be called from init(), next to the ops the source refreshes.
func RegisterSource(source Source, status func() []string) {
	Default.RegisterSource(source, status)
}

func (r *Registry) RegisterSource(source Source, status func() []string) {
	r.sources = append(r.sources, source)
	if status != nil {
		r.statuses = append(r.statuses, status)
	}
}

// Lookup finds an op by name.
func (r *Registry) Lookup(name string) (*Op, bool) {
	op, ok := r.ops[name]
	return op, ok
}

// Ops returns all registered ops sorted by name.
func (r *Registry) Ops() []*Op {
	result := make([]*Op, 0, len(r.ops))
	for _, op := range r.ops {
		result = append(result, op)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}
