package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
)

const (
	// MaxExpansionIterations is the maximum number of iterations for template expansion
	MaxExpansionIterations = 10
	// MaxReplacementsPerIteration is the maximum number of replacements per iteration
	MaxReplacementsPerIteration = 10
)

type scriptItem struct {
	Name  string
	Items []string
}

type podMap struct {
	Name  string
	Keys  []string
	Rules map[string]string
}

type SequenceConfig struct {
	PodCs      []podMap
	Shorts     map[string]string
	ShortsKeys []string
	Scripts    []scriptItem
}

func OpenAndReadSequencefile(fileName string, quiet bool) (conf SequenceConfig, err error) {
	var seq SequenceConfig

	jsonFile, err := os.Open(fileName)
	if err != nil {
		return SequenceConfig{}, fmt.Errorf("unable to open sequence json file %s: %w", fileName, err)
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return SequenceConfig{}, fmt.Errorf("unable to read sequence json file %s: %w", fileName, err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		return SequenceConfig{}, fmt.Errorf("unable to load json from sequence json file %s: %w", fileName, err)
	}

	for key, val := range result {
		if key == "podMap" {
			for key1, val1 := range val.(map[string]interface{}) {
				var x podMap
				x.Name = key1
				x.Rules = make(map[string]string)
				for i := 0; i < len(val1.([]interface{})); i++ {
					item := val1.([]interface{})[i]
					for key2, val2 := range item.(map[string]interface{}) {
						x.Keys = append(x.Keys, key2)
						x.Rules[key2] = val2.(string)
					}
				}
				seq.PodCs = append(seq.PodCs, x)
			}
			if !quiet {
				fmt.Printf("#PodMap loaded from %s\n", fileName)
			}
		}
		if key == "shortcuts" {
			seq.Shorts = make(map[string]string)
			for key1, val1 := range val.(map[string]interface{}) {
				seq.ShortsKeys = append(seq.ShortsKeys, key1)
				seq.Shorts[key1] = val1.(string)
			}
			slices.Sort(seq.ShortsKeys)
			if !quiet {
				fmt.Printf("#Shortcuts loaded from %s\n", fileName)
			}
		}
		if key == "scripts" {
			var t scriptItem
			for key1, val1 := range val.(map[string]interface{}) {
				t.Name = key1
				var cmdList []string
				for _, val2 := range val1.([]interface{}) {
					cmdList = append(cmdList, val2.(string))
				}
				t.Items = cmdList
				seq.Scripts = append(seq.Scripts, t)
			}
			if !quiet {
				fmt.Printf("#Sripts loaded from %s\n", fileName)
			}
		}
	}
	return seq, nil
}

func OpLineTagToOpString(line string) (op OpDecoded, opLine string) {
	retOperation := OpExecute
	retLine := line

	// checking shortened operations
	if len(line) >= 6 && line[:3] == "{{_" {
		splitLine := strings.Split(line[2:], "}}")
		if len(splitLine) > 1 {
			check_operation := splitLine[0]
			for i := 0; i < len(SupportedOps); i++ {
				if OpShort[SupportedOps[i]] == check_operation {
					retOperation = SupportedOps[i]
					retLine = strings.Join(splitLine[1:], "}}")
					break
				}
			}
		}
	}
	// checking full length operations
	if len(line) > 7 && line[:5] == "{{OP_" {
		splitLine := strings.Split(line[2:], "}}")
		if len(splitLine) > 1 {
			check_operation := splitLine[0]
			for i := 0; i < len(SupportedOps); i++ {
				if OpInstruction[SupportedOps[i]] == check_operation {
					retOperation = SupportedOps[i]
					retLine = strings.Join(splitLine[1:], "}}")
					break
				}
			}
		}
	}

	return retOperation, retLine
}

func ExpandShortcuts(line string, shorts map[string]string, keys []string) string {
	newLine := line
	for l := 0; l < MaxExpansionIterations; l++ {
		changes := false
		for i := 0; i < len(keys); i++ {
			key := keys[i]
			value := shorts[key]
			after := strings.Replace(newLine, "{{"+key+"}}", value, MaxReplacementsPerIteration)
			if after != newLine {
				changes = true
				newLine = after
			}
		}
		if changes == false {
			break
		}
	}
	return newLine
}

func ExpandUnderscore(line, K8sConfig, K8sContext, K8sNamespace string) string {
	newLine := "kubectl"
	if K8sConfig != "" {
		newLine += " --kubeconfig=" + K8sConfig
	}
	if K8sContext != "" {
		newLine += " --context=" + K8sContext
	}
	if K8sNamespace != "" {
		newLine += " -n " + K8sNamespace
	}

	if len(line) > 1 && line[:2] == "_ " {
		return newLine + line[1:]
	}
	if len(line) == 1 && line[:1] == "_" {
		return newLine
	}
	return line
}

func ExpandK8s(line, K8sConfig, K8sContext, K8sNamespace, K8sPod string) string {
	newLine := line
	for l := 0; l < MaxExpansionIterations; l++ {
		changes := false
		after := strings.Replace(newLine, "{{k8s_config}}", K8sConfig, MaxReplacementsPerIteration)
		after = strings.Replace(after, "{{k8s_context}}", K8sContext, MaxReplacementsPerIteration)
		after = strings.Replace(after, "{{k8s_namespace}}", K8sNamespace, MaxReplacementsPerIteration)
		after = strings.Replace(after, "{{k8s_pod}}", K8sPod, MaxReplacementsPerIteration)
		after = strings.Replace(after, "{{cnf}}", K8sConfig, MaxReplacementsPerIteration)
		after = strings.Replace(after, "{{ctx}}", K8sContext, MaxReplacementsPerIteration)
		after = strings.Replace(after, "{{nsp}}", K8sNamespace, MaxReplacementsPerIteration)
		after = strings.Replace(after, "{{pod}}", K8sPod, MaxReplacementsPerIteration)

		if after != newLine {
			changes = true
			newLine = after
		}
		if changes == false {
			break
		}
	}
	return newLine
}

func ExpandP2cRule(rules map[string]string, keys []string, pod string) string {
	for i := 0; i < len(keys); i++ {
		var err error
		matched := false
		matched, err = regexp.Match(rules[keys[i]], []byte(pod))
		if err != nil {
			fmt.Printf("non-fatal error, failed in compiling p2c rules for key %s: %v\n", keys[i], err)
			continue
		}
		if matched {
			return keys[i]
		}
	}
	return "default"
}

func ExpandPodMapper(line, K8s_pod string, p2c []podMap) string {
	newLine := line
	for l := 0; l < MaxExpansionIterations; l++ {
		changes := false
		for k := range p2c {
			after := strings.Replace(newLine, "{{"+p2c[k].Name+"}}", ExpandP2cRule(p2c[k].Rules, p2c[k].Keys, K8s_pod), MaxReplacementsPerIteration)
			if after != newLine {
				changes = true
				newLine = after
			}
		}
		if changes == false {
			return newLine
		}
	}
	return newLine
}

func IsThereAScript(name string, scripts []scriptItem) (offset int, err error) {
	var match bool = false
	var seqOffset = -1
	for i := range scripts {
		if scripts[i].Name == name {
			match = true
			seqOffset = i
		}
	}

	if !match {
		return -1, fmt.Errorf("no matching script %s in conf file", name)
	}
	return seqOffset, nil
}

type KctlDecode int

const (
	KctlConfig KctlDecode = iota
	KctlContext
	KctlNamespace
	KctlPod
)

var SupportedKctl = []KctlDecode{KctlConfig, KctlContext, KctlNamespace, KctlPod}

var KctlLong = map[KctlDecode]string{
	KctlConfig:    "k8s_config",
	KctlContext:   "k8s_context",
	KctlNamespace: "k8s_namespace",
	KctlPod:       "k8s_pod",
}

var KctlShort = map[KctlDecode]string{
	KctlConfig:    "cnf",
	KctlContext:   "ctx",
	KctlNamespace: "nsp",
	KctlPod:       "pod",
}

type OpDecoded int

const (
	OpTerminate OpDecoded = iota
	OpAttach
	OpDetach
	OpFinally
	OpExecute
	OpComment
	OpInfo
	OpNoPrompt
	OpSleep
	OpRefreshPrompt
	OpUnknown
)

var SupportedOps = []OpDecoded{OpTerminate, OpAttach, OpDetach, OpFinally, OpExecute, OpInfo, OpComment, OpNoPrompt, OpSleep, OpRefreshPrompt}

var OpInstruction = map[OpDecoded]string{
	OpTerminate:     "OP_TERMINATE",
	OpAttach:        "OP_ATTACH",
	OpDetach:        "OP_DETACH",
	OpFinally:       "OP_FINALLY",
	OpExecute:       "OP_EXECUTE",
	OpInfo:          "OP_INFO",
	OpComment:       "OP_COMMENT",
	OpNoPrompt:      "OP_NO_PROMPT_WAIT",
	OpSleep:         "OP_SLEEP",
	OpRefreshPrompt: "OP_REFRESH_PROMPT",
	OpUnknown:       "Operation_Uknown",
}

var OpPrint = map[OpDecoded]string{
	OpTerminate:     "#TERMINATE",
	OpAttach:        "#ATTACH",
	OpDetach:        "#DETACH",
	OpFinally:       "#FINALLY",
	OpExecute:       "#EXECUTE",
	OpInfo:          "#INFO",
	OpComment:       "#COMMENT",
	OpNoPrompt:      "#NO_PROMPT_WAIT",
	OpSleep:         "#SLEEP",
	OpRefreshPrompt: "#REFRESH_PROMPT",
	OpUnknown:       "#OPeration_Uknown",
}

var OpShort = map[OpDecoded]string{
	OpTerminate:     "_T",
	OpAttach:        "_A",
	OpDetach:        "_D",
	OpFinally:       "_F",
	OpExecute:       "_E",
	OpInfo:          "_I",
	OpComment:       "_C",
	OpNoPrompt:      "_N",
	OpSleep:         "_S",
	OpRefreshPrompt: "_R",
	OpUnknown:       "#OPeration_Uknown",
}

var OpName = map[OpDecoded]string{
	OpTerminate:     "Terminate tmux, script end",
	OpAttach:        "Attach tmux, script end",
	OpDetach:        "Detach tmux, script end, default behavior",
	OpFinally:       "Finally execute, script end",
	OpExecute:       "Execute line, no need to specify, default behaviour",
	OpInfo:          "Print info",
	OpComment:       "Print comment, render",
	OpNoPrompt:      "Do not wait for prompt for last command",
	OpSleep:         "Sleep for n seconds",
	OpRefreshPrompt: "Load new prompt",
	OpUnknown:       "Operation Uknown",
}

func opDecode(inputLine string) (op OpDecoded, line string) {
	if len(inputLine) > 5 && inputLine[:2] == "{{" {
		op, line = OpLineTagToOpString(inputLine)
		if op == OpUnknown {
			return OpExecute, line
		} else {
			return op, line
		}
	}
	return OpExecute, inputLine
}
