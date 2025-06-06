package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
	"kubectl-tks/internal"
	"path/filepath"
)

var cmdInfo = &cobra.Command{
	Use:   "info script",
	Short: "info script",
	Long:  `show details of selected script from sequence file`,
	Args:  cobra.MinimumNArgs(1),
	Run:   processInfo,
}

func init() {
	cmdInfo.Flags().StringVarP(&o.ScriptFile, "scriptFile", "f", "", "scriptFile (default is $HOME/.tks/sequences.json)")
	cmdInfo.Flags().BoolVarP(&o.ExpandSeq, "expand shortcuts", "x", false, "show script that is expanded with shortcuts")
	rootCmd.AddCommand(cmdInfo)
}

func processInfo(cmd *cobra.Command, args []string) {

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

	seq := internal.SequenceConfig{}

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
			fmt.Println(err)
			return
		}
	}
	var match bool = false

	for i := range seq.Scripts {
		if seq.Scripts[i].Name == args[0] {
			match = true
			fmt.Println("--------")
			for j := 0; j < len(seq.Scripts[i].Items); j++ {
				item := seq.Scripts[i].Items[j]
				fmt.Printf("Script line %d :\n", j)
				if (len(item) > 3 && item[:3] == "{{_") || (len(item) > 5 && item[:4] == "{{OP_") {
					op, line := internal.OpLineTagToOpString(item)
					fmt.Printf("%s:%s\n", internal.OpInstruction[op], line)
				} else {
					if o.ExpandSeq {
						newItem := internal.ExpandShortcuts(item, seq.Shorts, seq.ShortsKeys)
						fmt.Printf("Origial: %s\n", item)
						fmt.Printf("Expanded: %s\n", newItem)
					} else {
						fmt.Println(item)
					}
				}
			}
		}
	}
	if !match {
		fmt.Printf("No matching sequence name %s in sequence file %s, check available sequences with list command\n", args[0], o.ScriptFile)
	}
}
