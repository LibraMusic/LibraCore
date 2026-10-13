package sources

import "fmt"

var (
	registry    []Factory
	definitions = map[string]Source{}
	bindings    []Binding
)

type Factory interface {
	ID() string
	Satisfies(id string) bool
	SupportsMultiple() bool
	New(alias string, options map[string]any, info SharedInfo) (Source, error)
}

// Register registers a source factory.
// This is called during init() by each source package.
func Register(factory Factory) {
	registry = append(registry, factory)
}

func Define(id, alias string, options map[string]any, info SharedInfo) (Source, error) {
	if alias != "" {
		if _, exists := definitions[alias]; exists {
			return nil, fmt.Errorf("source alias %q is already defined", alias)
		}
	}

	var factory Factory
	for _, f := range registry {
		if f.Satisfies(id) {
			factory = f
			break
		}
	}
	if factory == nil {
		return nil, ErrUnknownSource
	}

	if !factory.SupportsMultiple() {
		for _, source := range definitions {
			if factory.Satisfies(source.ID()) {
				return nil, fmt.Errorf("source %q does not support multiple instances", id)
			}
		}
	}

	if alias == "" {
		alias = factory.ID()
		if factory.SupportsMultiple() {
			count := 1
			for _, source := range definitions {
				if factory.Satisfies(source.ID()) {
					count++
				}
			}
			alias = fmt.Sprintf("%s_%d", alias, count)
		}
	}

	source, err := factory.New(alias, options, info)
	if err != nil {
		return nil, err
	}

	definitions[alias] = source
	return source, nil
}

func BindAll(infos []BindingInfo) error {
	for _, info := range infos {
		if err := Bind(info); err != nil {
			return err
		}
	}
	return nil
}

func Bind(info BindingInfo) error {
	source, exists := definitions[info.Alias]
	if !exists {
		return fmt.Errorf("source alias %q is not defined", info.Alias)
	}

	bindings = append(bindings, Binding{Source: source, Info: info})
	return nil
}
