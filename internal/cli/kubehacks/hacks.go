// Package kubehacks provides Kubernetes shortcut commands for common operations.
package kubehacks

import (
	"fmt"
	"os"

	"github.com/inovacc/omni/internal/cli/kubectl"
)

// GetAll gets all common resources in a namespace.
// Equivalent to: kubectl get all -n <namespace>
func GetAll(namespace string, allNamespaces bool) error {
	args := []string{"get", "pods,svc,deploy,rs,sts,ds,jobs,cronjobs,cm,secret,ing,pvc"}
	if allNamespaces {
		args = append(args, "-A")
	} else if namespace != "" {
		args = append(args, "-n", namespace)
	}

	args = append(args, "-o", "wide")

	return kubectl.Run(args)
}

// LogsFollow follows logs for a pod with timestamps.
// Equivalent to: kubectl logs -f --timestamps <pod>
func LogsFollow(pod, namespace, container string, tail int) error {
	args := []string{"logs", "-f", "--timestamps"}
	if tail > 0 {
		args = append(args, fmt.Sprintf("--tail=%d", tail))
	}

	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if container != "" {
		args = append(args, "-c", container)
	}

	args = append(args, pod)

	return kubectl.Run(args)
}

// ExecBash execs into a pod with bash or sh.
// Equivalent to: kubectl exec -it <pod> -- /bin/bash
func ExecBash(pod, namespace, container string) error {
	args := []string{"exec", "-it"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if container != "" {
		args = append(args, "-c", container)
	}

	args = append(args, pod, "--", "/bin/bash")

	// Try bash first, fall back to sh
	err := kubectl.Run(args)
	if err != nil {
		args[len(args)-1] = "/bin/sh"
		return kubectl.Run(args)
	}

	return nil
}

// PortForward quick port forward.
// Equivalent to: kubectl port-forward <pod> <localPort>:<remotePort>
func PortForward(target, namespace string, localPort, remotePort int) error {
	args := []string{"port-forward"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	args = append(args, target, fmt.Sprintf("%d:%d", localPort, remotePort))

	return kubectl.Run(args)
}

// DeletePods deletes pods by selector.
// Equivalent to: kubectl delete pods -l <selector>
func DeletePods(selector, namespace string, force bool) error {
	args := []string{"delete", "pods", "-l", selector}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if force {
		args = append(args, "--grace-period=0", "--force")
	}

	return kubectl.Run(args)
}

// RolloutRestart restarts a deployment.
// Equivalent to: kubectl rollout restart deployment/<name>
func RolloutRestart(name, namespace string) error {
	args := []string{"rollout", "restart", "deployment/" + name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	return kubectl.Run(args)
}

// GetEvents gets events sorted by time.
// Equivalent to: kubectl get events --sort-by='.lastTimestamp'
func GetEvents(namespace string, allNamespaces bool) error {
	args := []string{"get", "events", "--sort-by=.lastTimestamp"}
	if allNamespaces {
		args = append(args, "-A")
	} else if namespace != "" {
		args = append(args, "-n", namespace)
	}

	return kubectl.Run(args)
}

// TopPods shows top pods by resource usage.
// Equivalent to: kubectl top pods
func TopPods(namespace string, allNamespaces bool, sortBy string) error {
	args := []string{"top", "pods"}
	if allNamespaces {
		args = append(args, "-A")
	} else if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if sortBy != "" {
		args = append(args, "--sort-by="+sortBy)
	}

	return kubectl.Run(args)
}

// TopNodes shows top nodes by resource usage.
func TopNodes(sortBy string) error {
	args := []string{"top", "nodes"}
	if sortBy != "" {
		args = append(args, "--sort-by="+sortBy)
	}

	return kubectl.Run(args)
}

// ContextSwitch switches to a different context.
// Equivalent to: kubectl config use-context <context>
func ContextSwitch(context string) error {
	return kubectl.Run([]string{"config", "use-context", context})
}

// ContextList lists all available contexts.
func ContextList() error {
	return kubectl.Run([]string{"config", "get-contexts"})
}

// NamespaceSwitch switches the default namespace for current context.
// Equivalent to: kubectl config set-context --current --namespace=<ns>
func NamespaceSwitch(namespace string) error {
	return kubectl.Run([]string{"config", "set-context", "--current", "--namespace=" + namespace})
}

// NamespaceList lists all namespaces.
func NamespaceList() error {
	return kubectl.Run([]string{"get", "namespaces"})
}

// Describe describes a resource.
func Describe(resource, name, namespace string) error {
	args := []string{"describe", resource, name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	return kubectl.Run(args)
}

// Apply applies a manifest file.
func Apply(file string, namespace string, dryRun bool) error {
	args := []string{"apply", "-f", file}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if dryRun {
		args = append(args, "--dry-run=client")
	}

	return kubectl.Run(args)
}

// Delete deletes a resource.
func Delete(resource, name, namespace string, force bool) error {
	args := []string{"delete", resource, name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if force {
		args = append(args, "--grace-period=0", "--force")
	}

	return kubectl.Run(args)
}

// Scale scales a deployment.
func Scale(name string, replicas int, namespace string) error {
	args := []string{"scale", "deployment/" + name, fmt.Sprintf("--replicas=%d", replicas)}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	return kubectl.Run(args)
}

// WatchPods watches pods continuously.
func WatchPods(namespace string, allNamespaces bool, selector string) error {
	args := []string{"get", "pods", "-w"}
	if allNamespaces {
		args = append(args, "-A")
	} else if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if selector != "" {
		args = append(args, "-l", selector)
	}

	return kubectl.Run(args)
}

// GetSecret decodes and displays a secret.
func GetSecret(name, namespace, key string) error {
	if key != "" {
		args := []string{"get", "secret", name, "-o", fmt.Sprintf("jsonpath={.data.%s}", key)}
		if namespace != "" {
			args = append(args, "-n", namespace)
		}
		// Need to decode base64
		return kubectl.Run(args)
	}

	args := []string{"get", "secret", name, "-o", "yaml"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	return kubectl.Run(args)
}

// Drain drains a node for maintenance.
func Drain(node string, ignoreDaemonsets, deleteEmptydir bool) error {
	args := []string{"drain", node}
	if ignoreDaemonsets {
		args = append(args, "--ignore-daemonsets")
	}

	if deleteEmptydir {
		args = append(args, "--delete-emptydir-data")
	}

	return kubectl.Run(args)
}

// Cordon cordons a node.
func Cordon(node string) error {
	return kubectl.Run([]string{"cordon", node})
}

// Uncordon uncordons a node.
func Uncordon(node string) error {
	return kubectl.Run([]string{"uncordon", node})
}

// Debug runs an ephemeral debug container.
func Debug(pod, namespace, image string) error {
	args := []string{"debug", "-it", pod}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if image != "" {
		args = append(args, "--image="+image)
	} else {
		args = append(args, "--image=busybox")
	}

	return kubectl.Run(args)
}

// PrintEnv prints environment variables for a container.
func PrintEnv(pod, namespace, container string) error {
	args := []string{"exec"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if container != "" {
		args = append(args, "-c", container)
	}

	args = append(args, pod, "--", "env")

	return kubectl.Run(args)
}

// CopyFrom copies files from a pod.
func CopyFrom(pod, namespace, srcPath, destPath string) error {
	args := []string{"cp"}
	if namespace != "" {
		args = append(args, namespace+"/"+pod+":"+srcPath, destPath)
	} else {
		args = append(args, pod+":"+srcPath, destPath)
	}

	return kubectl.Run(args)
}

// CopyTo copies files to a pod.
func CopyTo(pod, namespace, srcPath, destPath string) error {
	args := []string{"cp"}
	if namespace != "" {
		args = append(args, srcPath, namespace+"/"+pod+":"+destPath)
	} else {
		args = append(args, srcPath, pod+":"+destPath)
	}

	return kubectl.Run(args)
}

// Run runs a one-off pod.
func Run(name, image, namespace string, args []string) error {
	cmdArgs := []string{"run", name, "--image=" + image, "--rm", "-it", "--restart=Never"}
	if namespace != "" {
		cmdArgs = append(cmdArgs, "-n", namespace)
	}

	if len(args) > 0 {
		cmdArgs = append(cmdArgs, "--")
		cmdArgs = append(cmdArgs, args...)
	}

	return kubectl.Run(cmdArgs)
}

// GetConfig prints the current kubeconfig info.
func GetConfig() error {
	fmt.Fprintln(os.Stdout, "Current Context:")

	if err := kubectl.Run([]string{"config", "current-context"}); err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, "\nCluster Info:")

	return kubectl.Run([]string{"cluster-info"})
}
