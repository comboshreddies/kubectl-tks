package cmd

import (
	"fmt"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"kubectl-tks/internal"
)

var cmdList = &cobra.Command{
	Use:   "list [scripts|shortcuts|podConverter|control|kctl]",
	Short: "list objects: scripts, shortcuts, podConverter, control, kctl",
	Long:  "list available objects from a sequence file.",
	Args:  cobra.MinimumNArgs(1),
	Run:   processList,
}

func init() {
	cmdList.Flags().StringVarP(&o.ScriptFile, "scriptFile", "f", "", "scriptFile (default is $HOME/.tmux-k8s-scripts.yaml)")
	rootCmd.AddCommand(cmdList)
}

func processList(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("list needs argument: scripts, shortcuts, podConverter, control")
	}

	if args[0] == "kctl" {
		fmt.Printf("Kubectl params:\n")
		for i := 0; i < len(internal.SupportedKctl); i++ {
			fmt.Printf(" %s or short %s\n", internal.KctlLong[internal.SupportedKctl[i]], internal.KctlShort[internal.SupportedKctl[i]])
		}
		return
	}

	if args[0] == "control" {
		fmt.Printf("Controls:\n")
		for i := 0; i < len(internal.SupportedOps); i++ {
			fmt.Printf(" %s - %s - %s\n", internal.OpInstruction[internal.SupportedOps[i]], internal.OpShort[internal.SupportedOps[i]], internal.OpName[internal.SupportedOps[i]])
		}
		return
	}

	seq := internal.SequenceConfig{}

	envpath := os.Getenv("TKSSEQUENCE")
	fmt.Println(envpath)
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
		fmt.Printf("# No script file %s found, try using -f <path_to_sequence.file>\n", "sequences.json")
		return
	} else {
		seq, err = internal.OpenAndReadSequencefile(o.ScriptFile, false)
		if err != nil {
			fmt.Printf("Can't read conf file\n")
			fmt.Println(err)
			return
		}
	}

	var keys []string
	if args[0] == "scripts" {
		for i := 0; i < len(seq.Scripts); i++ {
			help := ""
			if len(seq.Scripts[i].Items) > 1 && strings.HasPrefix(seq.Scripts[i].Items[0], "{{OP_INFO}}") {
				help = fmt.Sprintf("  %s  : %s", seq.Scripts[i].Name, seq.Scripts[i].Items[0][len("{{OP_INFO}}"):])
			}
			if len(seq.Scripts[i].Items) > 1 && strings.HasPrefix(seq.Scripts[i].Items[0], "{{_I}}") {
				help = fmt.Sprintf("  %s  : %s", seq.Scripts[i].Name, seq.Scripts[i].Items[0][len("{{_I}}"):])
			}
			if help == "" {
				help = fmt.Sprintf("  %s  : ", seq.Scripts[i].Name)
			}
			keys = append(keys, help)
		}
		sort.Strings(keys)
		//      fmt.Println(keys)
		for _, i := range keys {
			fmt.Println(i)
		}
	}
	if args[0] == "shortcuts" {
		for i := 0; i < len(seq.ShortsKeys); i++ {
			key := seq.ShortsKeys[i]
			value := seq.Shorts[key]
			fmt.Printf("%s : %s\n", key, value)
		}
	}
	if args[0] == "podConverter" {
		for i := 0; i < len(seq.PodCs); i++ {
			fmt.Printf("%s :\n", seq.PodCs[i].Name)
			for j := 0; j < len(seq.PodCs[i].Keys); j++ {
				k := seq.PodCs[i].Keys[j]
				fmt.Printf("  '%s' <- '%s'\n", k, seq.PodCs[i].Rules[k])
			}
		}
	}
}
