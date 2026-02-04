package task

import (
	"fmt"
	"slices"
)

// DependencyResolver handles task dependency resolution
type DependencyResolver struct {
	tf *Taskfile
}

// NewDependencyResolver creates a new dependency resolver
func NewDependencyResolver(tf *Taskfile) *DependencyResolver {
	return &DependencyResolver{tf: tf}
}

// ResolveDeps returns tasks in dependency order (topological sort)
// Uses Kahn's algorithm for topological sorting
func (r *DependencyResolver) ResolveDeps(taskName string) ([]string, error) {
	// Build dependency graph
	graph := make(map[string][]string) // task -> dependencies
	inDegree := make(map[string]int)   // task -> number of incoming edges

	// Collect all tasks to process
	toProcess := []string{taskName}
	processed := make(map[string]bool)

	for len(toProcess) > 0 {
		current := toProcess[0]
		toProcess = toProcess[1:]

		if processed[current] {
			continue
		}

		processed[current] = true

		task := r.tf.GetTask(current)
		if task == nil {
			return nil, fmt.Errorf("task %q not found", current)
		}

		// Initialize in-degree if not present
		if _, ok := inDegree[current]; !ok {
			inDegree[current] = 0
		}

		// Get dependencies
		deps := make([]string, 0, len(task.Deps))
		for _, dep := range task.Deps {
			deps = append(deps, dep.Task)
			toProcess = append(toProcess, dep.Task)

			// Initialize in-degree if not present
			if _, ok := inDegree[dep.Task]; !ok {
				inDegree[dep.Task] = 0
			}
		}

		graph[current] = deps
	}

	// Note: in-degrees are calculated below after building reverse graph

	// Actually, let's build reverse graph: who depends on whom
	// reverseDeps[A] = [B, C] means B and C depend on A
	reverseDeps := make(map[string][]string)

	for task, deps := range graph {
		for _, dep := range deps {
			reverseDeps[dep] = append(reverseDeps[dep], task)
		}
	}

	// Recalculate in-degrees: for each task, count how many tasks it depends on
	for task := range inDegree {
		inDegree[task] = len(graph[task])
	}

	// Kahn's algorithm
	var (
		queue  []string
		result []string
	)

	// Start with tasks that have no dependencies

	for task, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, task)
		}
	}

	for len(queue) > 0 {
		// Pop from queue
		current := queue[0]
		queue = queue[1:]

		result = append(result, current)

		// For each task that depends on current
		for _, dependent := range reverseDeps[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Check for cycles
	if len(result) != len(inDegree) {
		return nil, fmt.Errorf("cyclic dependency detected")
	}

	return result, nil
}

// GetDirectDeps returns direct dependencies of a task
func (r *DependencyResolver) GetDirectDeps(taskName string) ([]Dependency, error) {
	task := r.tf.GetTask(taskName)
	if task == nil {
		return nil, fmt.Errorf("task %q not found", taskName)
	}

	return task.Deps, nil
}

// ValidateDeps checks if all dependencies exist
func (r *DependencyResolver) ValidateDeps(taskName string) error {
	visited := make(map[string]bool)
	return r.validateDepsRecursive(taskName, visited, nil)
}

func (r *DependencyResolver) validateDepsRecursive(taskName string, visited map[string]bool, path []string) error {
	// Check for cycle
	if slices.Contains(path, taskName) {
		return fmt.Errorf("cyclic dependency: %v -> %s", path, taskName)
	}

	if visited[taskName] {
		return nil
	}

	visited[taskName] = true

	task := r.tf.GetTask(taskName)
	if task == nil {
		return fmt.Errorf("task %q not found", taskName)
	}

	newPath := make([]string, len(path)+1)
	copy(newPath, path)
	newPath[len(path)] = taskName

	for _, dep := range task.Deps {
		if err := r.validateDepsRecursive(dep.Task, visited, newPath); err != nil {
			return err
		}
	}

	return nil
}
