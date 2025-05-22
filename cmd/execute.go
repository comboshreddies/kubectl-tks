package cmd

import (
	"github.com/spf13/cobra"
)

type CliOptions struct {
	ScriptFile     string
	ExpandSeq      bool
	K8sConfig      string
	K8sContext     string
	K8sNamespace   string
	K8sSelector    string
	KTxDryRun      bool
	KTxSync        bool
	KTxTermSess    bool
	KTxPodList     string
	KTxPrompt      string
	KTxPromptSleep int
	KTxSessionName string
}

var o = CliOptions{
	ScriptFile:     "~/.tks/sequences.json",
	ExpandSeq:      false,
	K8sConfig:      "~/.kube/config",
	K8sContext:     "",
	K8sNamespace:   "",
	K8sSelector:    "",
	KTxDryRun:      false,
	KTxSync:        false,
	KTxPodList:     "",
	KTxPrompt:      "",
	KTxPromptSleep: 0,
	KTxSessionName: "",
}

//var scriptFile     string

var rootCmd = &cobra.Command{Use: "kubect-tks"}

func init() {
	// rootCmd.Flags().StringVarP(&o.ScriptFile, "scriptFile", "f", "", "scriptFile (default is $HOME/.tmux-k8s-scripts.yaml)")
}

func Execute() error {
	return rootCmd.Execute()
}
