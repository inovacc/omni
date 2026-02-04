package task

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Taskfile represents the parsed Taskfile.yml
type Taskfile struct {
	Version  string            `yaml:"version"`
	Vars     map[string]any    `yaml:"vars"`
	Env      map[string]string `yaml:"env"`
	Tasks    map[string]*Task  `yaml:"tasks"`
	Includes map[string]string `yaml:"includes"`

	// Internal fields
	dir string // Directory containing this taskfile
}

// Task represents a single task definition
type Task struct {
	Desc         string         `yaml:"desc"`
	Summary      string         `yaml:"summary"`
	Cmds         []Command      `yaml:"cmds"`
	Deps         []Dependency   `yaml:"deps"`
	Vars         map[string]any `yaml:"vars"`
	Status       []string       `yaml:"status"` // Commands to check if task is up-to-date
	Sources      []string       `yaml:"sources"`
	Generates    []string       `yaml:"generates"`
	Dir          string         `yaml:"dir"`
	Silent       bool           `yaml:"silent"`
	Internal     bool           `yaml:"internal"` // Hide from list
	Precondition *Precondition  `yaml:"precondition"`
	Aliases      []string       `yaml:"aliases"`

	// Internal fields
	name string
}

// Command represents a command to execute
type Command struct {
	Cmd      string `yaml:"cmd"`
	Task     string `yaml:"task"`    // Reference to another task
	Silent   bool   `yaml:"silent"`
	IgnoreError bool `yaml:"ignore_error"`
	Defer    bool   `yaml:"defer"`
}

// UnmarshalYAML implements custom unmarshaling for Command
func (c *Command) UnmarshalYAML(node *yaml.Node) error {
	// Handle string shorthand: "omni ls"
	if node.Kind == yaml.ScalarNode {
		c.Cmd = node.Value
		return nil
	}

	// Handle map form
	type rawCommand Command
	return node.Decode((*rawCommand)(c))
}

// Dependency represents a task dependency
type Dependency struct {
	Task string         `yaml:"task"`
	Vars map[string]any `yaml:"vars"`
}

// UnmarshalYAML implements custom unmarshaling for Dependency
func (d *Dependency) UnmarshalYAML(node *yaml.Node) error {
	// Handle string shorthand: "build"
	if node.Kind == yaml.ScalarNode {
		d.Task = node.Value
		return nil
	}

	// Handle map form
	type rawDep Dependency
	return node.Decode((*rawDep)(d))
}

// Precondition represents a precondition check
type Precondition struct {
	Sh  string `yaml:"sh"`
	Msg string `yaml:"msg"`
}

// ParseTaskfile parses a Taskfile.yml file
func ParseTaskfile(path string) (*Taskfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading taskfile: %w", err)
	}

	var tf Taskfile
	if err := yaml.Unmarshal(data, &tf); err != nil {
		return nil, fmt.Errorf("parsing taskfile: %w", err)
	}

	// Set internal fields
	tf.dir = filepath.Dir(path)

	// Set task names
	for name, task := range tf.Tasks {
		if task != nil {
			task.name = name
		}
	}

	// Process includes
	if len(tf.Includes) > 0 {
		if err := tf.processIncludes(); err != nil {
			return nil, err
		}
	}

	return &tf, nil
}

// processIncludes loads and merges included taskfiles
func (tf *Taskfile) processIncludes() error {
	for namespace, includePath := range tf.Includes {
		// Resolve relative path
		if !filepath.IsAbs(includePath) {
			includePath = filepath.Join(tf.dir, includePath)
		}

		// Check if it's a directory (look for Taskfile.yml)
		info, err := os.Stat(includePath)
		if err != nil {
			return fmt.Errorf("include %s not found: %w", includePath, err)
		}

		if info.IsDir() {
			// Look for taskfile in directory
			found := false
			for _, name := range DefaultTaskfiles {
				path := filepath.Join(includePath, name)
				if _, err := os.Stat(path); err == nil {
					includePath = path
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("no taskfile found in included directory: %s", includePath)
			}
		}

		// Parse included taskfile
		included, err := ParseTaskfile(includePath)
		if err != nil {
			return fmt.Errorf("parsing included taskfile %s: %w", includePath, err)
		}

		// Merge tasks with namespace prefix
		for name, task := range included.Tasks {
			nsName := namespace + ":" + name
			task.name = nsName
			tf.Tasks[nsName] = task
		}

		// Merge vars (included vars don't override parent)
		for k, v := range included.Vars {
			if _, exists := tf.Vars[k]; !exists {
				tf.Vars[k] = v
			}
		}
	}

	return nil
}

// GetTask returns a task by name, checking aliases
func (tf *Taskfile) GetTask(name string) *Task {
	// Direct lookup
	if task, ok := tf.Tasks[name]; ok {
		return task
	}

	// Check aliases
	for _, task := range tf.Tasks {
		for _, alias := range task.Aliases {
			if alias == name {
				return task
			}
		}
	}

	return nil
}

// ListTaskNames returns all non-internal task names
func (tf *Taskfile) ListTaskNames() []string {
	var names []string
	for name, task := range tf.Tasks {
		if task != nil && !task.Internal {
			names = append(names, name)
		}
	}
	return names
}
