package cmd

import (
	"fmt"

	"github.com/inovacc/omni/internal/cli/kubehacks"
	"github.com/spf13/cobra"
)

// Kubernetes hack commands with short aliases

// kga - kubectl get all
var kgaCmd = &cobra.Command{
	Use:   "kga",
	Short: "Get all resources in namespace",
	Long: `Get all common resources in a namespace.
Equivalent to: kubectl get pods,svc,deploy,... -o wide

Examples:
  omni kga
  omni kga -n kube-system
  omni kga -A`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, _ := cmd.Flags().GetString("namespace")
		allNs, _ := cmd.Flags().GetBool("all-namespaces")
		return kubehacks.GetAll(ns, allNs)
	},
}

// klf - kubectl logs follow
var klfCmd = &cobra.Command{
	Use:   "klf <pod>",
	Short: "Follow pod logs with timestamps",
	Long: `Follow logs for a pod with timestamps.
Equivalent to: kubectl logs -f --timestamps <pod>

Examples:
  omni klf mypod
  omni klf mypod -n default -c mycontainer
  omni klf mypod --tail 100`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, _ := cmd.Flags().GetString("namespace")
		container, _ := cmd.Flags().GetString("container")
		tail, _ := cmd.Flags().GetInt("tail")
		return kubehacks.LogsFollow(args[0], ns, container, tail)
	},
}

// keb - kubectl exec bash
var kebCmd = &cobra.Command{
	Use:   "keb <pod>",
	Short: "Exec into pod with bash",
	Long: `Exec into a pod with bash (falls back to sh).
Equivalent to: kubectl exec -it <pod> -- /bin/bash

Examples:
  omni keb mypod
  omni keb mypod -n default
  omni keb mypod -c mycontainer`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, _ := cmd.Flags().GetString("namespace")
		container, _ := cmd.Flags().GetString("container")
		return kubehacks.ExecBash(args[0], ns, container)
	},
}

// kpf - kubectl port-forward
var kpfCmd = &cobra.Command{
	Use:   "kpf <pod|svc/name> <local:remote>",
	Short: "Quick port forward",
	Long: `Quick port forward to a pod or service.
Equivalent to: kubectl port-forward <target> <local>:<remote>

Examples:
  omni kpf mypod 8080:80
  omni kpf svc/myservice 3000:80 -n default`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, _ := cmd.Flags().GetString("namespace")
		var local, remote int
		_, _ = fmt.Sscanf(args[1], "%d:%d", &local, &remote)
		if local == 0 || remote == 0 {
			return fmt.Errorf("invalid port format, use local:remote (e.g., 8080:80)")
		}
		return kubehacks.PortForward(args[0], ns, local, remote)
	},
}

// kdp - kubectl delete pods
var kdpCmd = &cobra.Command{
	Use:   "kdp <selector>",
	Short: "Delete pods by selector",
	Long: `Delete pods by label selector.
Equivalent to: kubectl delete pods -l <selector>

Examples:
  omni kdp app=nginx
  omni kdp app=nginx -n default --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, _ := cmd.Flags().GetString("namespace")
		force, _ := cmd.Flags().GetBool("force")
		return kubehacks.DeletePods(args[0], ns, force)
	},
}

// krr - kubectl rollout restart
var krrCmd = &cobra.Command{
	Use:   "krr <deployment>",
	Short: "Rollout restart deployment",
	Long: `Restart a deployment using rollout restart.
Equivalent to: kubectl rollout restart deployment/<name>

Examples:
  omni krr mydeployment
  omni krr mydeployment -n default`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, _ := cmd.Flags().GetString("namespace")
		return kubehacks.RolloutRestart(args[0], ns)
	},
}

// kge - kubectl get events
var kgeCmd = &cobra.Command{
	Use:   "kge",
	Short: "Get events sorted by time",
	Long: `Get events sorted by last timestamp.
Equivalent to: kubectl get events --sort-by='.lastTimestamp'

Examples:
  omni kge
  omni kge -n kube-system
  omni kge -A`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, _ := cmd.Flags().GetString("namespace")
		allNs, _ := cmd.Flags().GetBool("all-namespaces")
		return kubehacks.GetEvents(ns, allNs)
	},
}

// ktp - kubectl top pods
var ktpCmd = &cobra.Command{
	Use:   "ktp",
	Short: "Top pods by resource usage",
	Long: `Show top pods by resource usage.
Equivalent to: kubectl top pods

Examples:
  omni ktp
  omni ktp -n default
  omni ktp --sort-by cpu`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, _ := cmd.Flags().GetString("namespace")
		allNs, _ := cmd.Flags().GetBool("all-namespaces")
		sortBy, _ := cmd.Flags().GetString("sort-by")
		return kubehacks.TopPods(ns, allNs, sortBy)
	},
}

// ktn - kubectl top nodes
var ktnCmd = &cobra.Command{
	Use:   "ktn",
	Short: "Top nodes by resource usage",
	Long: `Show top nodes by resource usage.
Equivalent to: kubectl top nodes

Examples:
  omni ktn
  omni ktn --sort-by cpu`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sortBy, _ := cmd.Flags().GetString("sort-by")
		return kubehacks.TopNodes(sortBy)
	},
}

// kcs - kubectl context switch
var kcsCmd = &cobra.Command{
	Use:   "kcs [context]",
	Short: "Switch kubectl context",
	Long: `Switch to a different kubectl context.
Without arguments, lists available contexts.

Examples:
  omni kcs              # list contexts
  omni kcs production   # switch to production`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return kubehacks.ContextList()
		}
		return kubehacks.ContextSwitch(args[0])
	},
}

// kns - kubectl namespace switch
var knsCmd = &cobra.Command{
	Use:   "kns [namespace]",
	Short: "Switch default namespace",
	Long: `Switch the default namespace for the current context.
Without arguments, lists all namespaces.

Examples:
  omni kns           # list namespaces
  omni kns default   # switch to default namespace`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return kubehacks.NamespaceList()
		}
		return kubehacks.NamespaceSwitch(args[0])
	},
}

// kwp - kubectl watch pods
var kwpCmd = &cobra.Command{
	Use:   "kwp",
	Short: "Watch pods continuously",
	Long: `Watch pods continuously.
Equivalent to: kubectl get pods -w

Examples:
  omni kwp
  omni kwp -n default
  omni kwp -l app=nginx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, _ := cmd.Flags().GetString("namespace")
		allNs, _ := cmd.Flags().GetBool("all-namespaces")
		selector, _ := cmd.Flags().GetString("selector")
		return kubehacks.WatchPods(ns, allNs, selector)
	},
}

// kscale - kubectl scale
var kscaleCmd = &cobra.Command{
	Use:   "kscale <deployment> <replicas>",
	Short: "Scale deployment",
	Long: `Scale a deployment to the specified number of replicas.

Examples:
  omni kscale mydeployment 3
  omni kscale mydeployment 0 -n default`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, _ := cmd.Flags().GetString("namespace")
		var replicas int
		_, _ = fmt.Sscanf(args[1], "%d", &replicas)
		return kubehacks.Scale(args[0], replicas, ns)
	},
}

// kdebug - kubectl debug
var kdebugCmd = &cobra.Command{
	Use:   "kdebug <pod>",
	Short: "Debug pod with ephemeral container",
	Long: `Run an ephemeral debug container in a pod.
Equivalent to: kubectl debug -it <pod> --image=<image>

Examples:
  omni kdebug mypod
  omni kdebug mypod --image=nicolaka/netshoot`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, _ := cmd.Flags().GetString("namespace")
		image, _ := cmd.Flags().GetString("image")
		return kubehacks.Debug(args[0], ns, image)
	},
}

// kdrain - kubectl drain
var kdrainCmd = &cobra.Command{
	Use:   "kdrain <node>",
	Short: "Drain node for maintenance",
	Long: `Drain a node for maintenance.
Equivalent to: kubectl drain <node>

Examples:
  omni kdrain mynode
  omni kdrain mynode --ignore-daemonsets`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ignoreDaemonsets, _ := cmd.Flags().GetBool("ignore-daemonsets")
		deleteEmptydir, _ := cmd.Flags().GetBool("delete-emptydir")
		return kubehacks.Drain(args[0], ignoreDaemonsets, deleteEmptydir)
	},
}

// krun - kubectl run
var krunCmd = &cobra.Command{
	Use:   "krun <name> --image=<image> [-- command]",
	Short: "Run a one-off pod",
	Long: `Run a one-off pod that auto-deletes after completion.
Equivalent to: kubectl run <name> --image=<image> --rm -it --restart=Never

Examples:
  omni krun test --image=busybox -- sh
  omni krun curl --image=curlimages/curl -- curl google.com`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, _ := cmd.Flags().GetString("namespace")
		image, _ := cmd.Flags().GetString("image")
		if image == "" {
			return fmt.Errorf("--image is required")
		}
		var cmdArgs []string
		if len(args) > 1 {
			cmdArgs = args[1:]
		}
		return kubehacks.Run(args[0], image, ns, cmdArgs)
	},
}

// kconfig - kubectl config info
var kconfigCmd = &cobra.Command{
	Use:   "kconfig",
	Short: "Show kubeconfig info",
	Long: `Show current kubeconfig context and cluster info.

Examples:
  omni kconfig`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return kubehacks.GetConfig()
	},
}

func init() {
	// kga
	kgaCmd.Flags().StringP("namespace", "n", "", "Namespace")
	kgaCmd.Flags().BoolP("all-namespaces", "A", false, "All namespaces")
	rootCmd.AddCommand(kgaCmd)

	// klf
	klfCmd.Flags().StringP("namespace", "n", "", "Namespace")
	klfCmd.Flags().StringP("container", "c", "", "Container name")
	klfCmd.Flags().Int("tail", 0, "Lines to show from end of logs")
	rootCmd.AddCommand(klfCmd)

	// keb
	kebCmd.Flags().StringP("namespace", "n", "", "Namespace")
	kebCmd.Flags().StringP("container", "c", "", "Container name")
	rootCmd.AddCommand(kebCmd)

	// kpf
	kpfCmd.Flags().StringP("namespace", "n", "", "Namespace")
	rootCmd.AddCommand(kpfCmd)

	// kdp
	kdpCmd.Flags().StringP("namespace", "n", "", "Namespace")
	kdpCmd.Flags().Bool("force", false, "Force delete")
	rootCmd.AddCommand(kdpCmd)

	// krr
	krrCmd.Flags().StringP("namespace", "n", "", "Namespace")
	rootCmd.AddCommand(krrCmd)

	// kge
	kgeCmd.Flags().StringP("namespace", "n", "", "Namespace")
	kgeCmd.Flags().BoolP("all-namespaces", "A", false, "All namespaces")
	rootCmd.AddCommand(kgeCmd)

	// ktp
	ktpCmd.Flags().StringP("namespace", "n", "", "Namespace")
	ktpCmd.Flags().BoolP("all-namespaces", "A", false, "All namespaces")
	ktpCmd.Flags().String("sort-by", "", "Sort by (cpu or memory)")
	rootCmd.AddCommand(ktpCmd)

	// ktn
	ktnCmd.Flags().String("sort-by", "", "Sort by (cpu or memory)")
	rootCmd.AddCommand(ktnCmd)

	// kcs
	rootCmd.AddCommand(kcsCmd)

	// kns
	rootCmd.AddCommand(knsCmd)

	// kwp
	kwpCmd.Flags().StringP("namespace", "n", "", "Namespace")
	kwpCmd.Flags().BoolP("all-namespaces", "A", false, "All namespaces")
	kwpCmd.Flags().StringP("selector", "l", "", "Label selector")
	rootCmd.AddCommand(kwpCmd)

	// kscale
	kscaleCmd.Flags().StringP("namespace", "n", "", "Namespace")
	rootCmd.AddCommand(kscaleCmd)

	// kdebug
	kdebugCmd.Flags().StringP("namespace", "n", "", "Namespace")
	kdebugCmd.Flags().String("image", "", "Debug container image (default: busybox)")
	rootCmd.AddCommand(kdebugCmd)

	// kdrain
	kdrainCmd.Flags().Bool("ignore-daemonsets", false, "Ignore daemonsets")
	kdrainCmd.Flags().Bool("delete-emptydir", false, "Delete emptydir data")
	rootCmd.AddCommand(kdrainCmd)

	// krun
	krunCmd.Flags().StringP("namespace", "n", "", "Namespace")
	krunCmd.Flags().String("image", "", "Container image (required)")
	rootCmd.AddCommand(krunCmd)

	// kconfig
	rootCmd.AddCommand(kconfigCmd)
}
