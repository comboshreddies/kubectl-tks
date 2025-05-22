package cmd

import (
	"fmt"
	"k8s.io/client-go/util/homedir"
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
	fmt.Println("list: .... ")
	if len(args) != 1 {
		fmt.Println("list needs argument: scripts, shortcuts, podConverter, control")
	}

	if o.ScriptFile == "" {
		if home := homedir.HomeDir(); home != "" {
			o.ScriptFile = filepath.Join(home, ".tks/sequences.json")
		} else {
			o.ScriptFile = "sequences.json"
		}
	}

	seq, err := internal.OpenAndReadSequencefile(o.ScriptFile)
	if err != nil {
		fmt.Println(err)
	}
	var keys []string
	if args[0] == "scripts" {
		for i := range seq.Scripts {
			var help string
			if len(seq.Scripts[i].Items) > 1 && strings.HasPrefix(seq.Scripts[i].Items[0], "{{OP_INFO}}") {
				help = fmt.Sprintf("  %s  - %s", seq.Scripts[i].Name, seq.Scripts[i].Items[0][len("{{PO_INFO}}"):])
			} else {
				help = fmt.Sprintf("  %s  - ", seq.Scripts[i].Name)
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
		for i := 0; i < len(seq.Shorts); i++ {
			fmt.Printf("%s - %s\n", seq.Shorts[i].Name, seq.Shorts[i].Value)
		}
	}
	if args[0] == "podConverter" {
		for i := 0; i < len(seq.PodCs); i++ {
			fmt.Printf("%s -\n", seq.PodCs[i].Name)
		}
	}
	if args[0] == "control" {
		for i := 0; i < len(seq.Predefs); i++ {
			if seq.Predefs[i].Name == "control" {
				fmt.Printf("%s :\n", seq.Predefs[i].Name)
				for j := 0; j < len(seq.Predefs[i].Tags); j++ {
					fmt.Printf(" - %s", seq.Predefs[i].Tags[j])
					switch seq.Predefs[i].Tags[j] {
					case "OP_INFO":
						fmt.Printf(" : Info for that script, kind of introduction help line\n")
					case "OP_COMMENT":
						fmt.Printf(" : All content of this script instruction line will be just echoed to stdout\n")
					case "OP_NO_RETURN":
						fmt.Printf(" : Do not wait for return, do not read expected prompt line, jump to next script line\n")
					case "OP_FINAL_EXEC":
						fmt.Printf(" : After scripts are executed on all pods, and tmux terminated, execute this script line \n")
					case "OP_ATTACH":
						fmt.Printf(" : After script execution attach to tmux\n")
					case "OP_TERMINATE":
						fmt.Printf(" : After scripts execution terminate tmux and it's session\n")
					case "OP_SLEEP":
						fmt.Printf(" : Sleep for desired time on control process, before attaching, terminating or being interrupted\n")
					case "OP_REFRESH_PROMPT":
						fmt.Printf(" : Load new prompt line after previuos script line execution (remote session or changing prompt line)\n")
					case "OP_SYNC":
						fmt.Printf(" : Synchronize execution on all tmux terminals before proceeding with next script line\n")
					default:
						fmt.Println()
					}
				}
			}
		}
	}
	if args[0] == "kctl" {
		for i := 0; i < len(seq.Predefs); i++ {
			if seq.Predefs[i].Name == "kctl" {
				fmt.Printf("%s :\n", seq.Predefs[i].Name)
				for j := 0; j < len(seq.Predefs[i].Tags); j++ {
					fmt.Printf(" - %s\n", seq.Predefs[i].Tags[j])
				}
			}
		}
	}

}
