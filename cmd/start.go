package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
	"kubectl-tks/internal"
	"path/filepath"
)

func init() {
	cmdStart.Flags().StringVarP(&o.ScriptFile, "scriptFile", "f", "", "scriptFile (default is $HOME/.tks/sequence.json)")
	cmdStart.Flags().StringVarP(&o.K8sConfig, "kubeconfig", "", "", "kubernetes config file")
	cmdStart.Flags().StringVarP(&o.K8sContext, "context", "", "", "kubernetes context")
	cmdStart.Flags().StringVarP(&o.K8sNamespace, "namespace", "n", "", "kubernetes namespace")
	cmdStart.Flags().StringVarP(&o.K8sSelector, "selector", "l", "", "kubernetes label query selector")
	cmdStart.Flags().BoolVarP(&o.KTxDryRun, "dry", "d", false, "start a dry run, no script execution")
	cmdStart.Flags().BoolVarP(&o.KTxSync, "sync", "s", false, "run in sync step mode")
	cmdStart.Flags().BoolVarP(&o.KTxQuiet, "quiet", "q", false, "run in quiet mode")
	cmdStart.Flags().BoolVarP(&o.KTxTermPrevSess, "termPrev", "T", false, "kill session if exists, before starting")
	cmdStart.Flags().BoolVarP(&o.KTxTermCurrSess, "term", "X", false, "terminate session, override script")
	cmdStart.Flags().BoolVarP(&o.KTxAttachSess, "attach", "A", false, "attach tmux session, override script")
	cmdStart.Flags().StringVarP(&o.KTxSessionName, "sessionName", "S", "", "tmux session name")
	cmdStart.Flags().StringVarP(&o.KTxPodList, "pods", "p", "", "set list of pods, comma separated")
	cmdStart.Flags().StringVarP(&o.KTxPrompt, "Prompt", "P", "", "tmux define prompt")
	cmdStart.Flags().IntVarP(&o.KTxPromptSleep, "sleepTime", "t", 2, "sleep seconds before catching prompt")
	rootCmd.AddCommand(cmdStart)
}

var cmdStart = &cobra.Command{
	Use:   "start script|quotedCommand",
	Short: "execute of the script from sequence file, if no script it will execute argument",
	Long:  `start execution of a selected script from sequence file, if no script it will consider argument is quoted set of commands that are ';' separated`,
	Args:  cobra.MinimumNArgs(1),
	Run:   processStart,
}

func processStart(cmd *cobra.Command, args []string) {

	if o.KTxAttachSess == true && o.KTxTermCurrSess == true {
		fmt.Println("start can't have both -A and -X, choose one, exiting")
		return
	}

	noConfFile := false
	seq := internal.SequenceConfig{}

	envpath := os.Getenv("TKSSEQUENCE")
	if envpath != "" && o.ScriptFile == "" {
		_, err := os.Stat(envpath)
		if err == nil {
			o.ScriptFile = envpath
		}
	}

	if o.ScriptFile == "" {
		if home := homedir.HomeDir(); home != "" {
			o.ScriptFile = filepath.Join(home, ".tks/sequences.json")
		} else {
			o.ScriptFile = "sequences.json"
		}
	}

	_, err := os.Stat(o.ScriptFile)
	if err != nil { // check for krew store path
		if home := homedir.HomeDir(); home != "" {
			o.ScriptFile = filepath.Join(home, ".krew/store/tks", internal.TksVersion, "sequences.json")
		} else { // windows part, not sure at the moment
			o.ScriptFile = "sequences.json"
		}
	}
	_, err = os.Stat(o.ScriptFile)
	if err != nil { // check for krew store path
		if !o.KTxQuiet {
			fmt.Printf("# No script file %s found, try using -f <path_to_sequence.file>\n", "sequences.json")
		}
		noConfFile = true
		seq.Shorts = nil
	} else {
		seq, err = internal.OpenAndReadSequencefile(o.ScriptFile, o.KTxQuiet)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("# Unable to read conf file %s, assuming oneLiner\n", o.ScriptFile)
			noConfFile = true
			seq = internal.SequenceConfig{}
			seq.Shorts = nil
		}
	}

	noScript := false
	seqOffset := -1
	if noConfFile == false {
		seqOffset, err = internal.IsThereAScript(args[0], seq.Scripts)
		if err != nil {
			if !o.KTxQuiet {
				fmt.Println(err)
				fmt.Printf("# assuming oneLiner '%s'\n", args[0])
			}
			noScript = true
		}
	}

	var scriptLines []string
	var scriptName string
	if noConfFile == true || noScript == true {
		splitArg0 := strings.Split(args[0], ";")
		for i := 0; i < len(splitArg0); i++ {
			scriptLines = append(scriptLines, splitArg0[i][0:len(splitArg0[i])])
		}
		scriptName = "OneLiner"
	} else {
		scriptLines = seq.Scripts[seqOffset].Items
		scriptName = args[0]
	}

	podList, err := internal.Kubernetes_pod_list(o.K8sConfig, o.K8sContext, o.K8sNamespace, o.K8sSelector)
	if err != nil {
		fmt.Println(err)
		return
	}

	var filteredPodList []internal.PodsInfo
	cliPods := strings.Split(o.KTxPodList, ",")
	if o.KTxPodList != "" {
		for i := 0; i < len(podList); i++ {
			for j := 0; j < len(cliPods); j++ {
				if podList[i].PodName == cliPods[j] {
					filteredPodList = append(filteredPodList, podList[i])
					break
				}
			}
		}
	} else {
		filteredPodList = podList
	}

	tmuxIn := internal.TmuxInData{}
	tmuxIn.SeqName = scriptName
	tmuxIn.ScriptLines = scriptLines
	tmuxIn.K8sConfig = o.K8sConfig
	tmuxIn.K8sContext = o.K8sContext
	tmuxIn.K8sNamespace = o.K8sNamespace
	tmuxIn.PodList = filteredPodList
	tmuxIn.Shorts = seq.Shorts
	tmuxIn.ShortsKeys = seq.ShortsKeys
	tmuxIn.PodCs = seq.PodCs
	tmuxIn.Prompt = o.KTxPrompt
	tmuxIn.PromptSleep = o.KTxPromptSleep
	tmuxIn.SessionName = o.KTxSessionName

	internal.StartTmux(tmuxIn, o.KTxDryRun, o.KTxSync, o.KTxTermPrevSess, o.KTxTermCurrSess, o.KTxAttachSess, o.KTxQuiet)
}
