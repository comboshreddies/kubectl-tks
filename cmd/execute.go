package cmd

import (
	"github.com/spf13/cobra"
)

type CliOptions struct {
	ScriptFile      string
	ExpandSeq       bool
	K8sConfig       string
	K8sContext      string
	K8sNamespace    string
	K8sSelector     string
	KTxDryRun       bool
	KTxSync         bool
	KTxQuiet        bool
	KTxTermPrevSess bool
	KTxTermCurrSess bool
	KTxAttachSess   bool
	KTxPodList      string
	KTxPrompt       string
	KTxPromptSleep  int
	KTxSessionName  string
}

var o = CliOptions{
	ScriptFile:      "",
	ExpandSeq:       false,
	K8sConfig:       "~/.kube/config",
	K8sContext:      "",
	K8sNamespace:    "",
	K8sSelector:     "",
	KTxDryRun:       false,
	KTxSync:         false,
	KTxQuiet:        false,
	KTxTermPrevSess: false,
	KTxTermCurrSess: false,
	KTxAttachSess:   false,
	KTxPodList:      "",
	KTxPrompt:       "",
	KTxPromptSleep:  0,
	KTxSessionName:  "",
}

//var scriptFile     string

var rootCmd = &cobra.Command{Use: "kubect-tks"}

func init() {
}

func Execute() error {
	return rootCmd.Execute()
}
