package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"slices"
	"strings"
)

type scriptItem struct {
	Name  string
	Items []string
}

type podConverter struct {
	Name  string
	Keys  []string
	Rules map[string]string
}

type SequenceConfig struct {
	PodCs      []podConverter
	Shorts     map[string]string
	ShortsKeys []string
	Scripts    []scriptItem
}

func OpenAndReadSequencefile(fileName string) (conf SequenceConfig, err error) {
	var seq SequenceConfig

	jsonFile, err := os.Open(fileName)
	if err != nil {
		return SequenceConfig{}, errors.New("# unable to open sequence json file " + fileName)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(byteValue), &result)
	if err != nil {
		newerr := errors.New("# unable to load json from sequence json file")
		return SequenceConfig{}, newerr
	}

	for key, val := range result {
		if key == "podConverter" {
			for key1, val1 := range val.(map[string]interface{}) {
				var x podConverter
				x.Name = key1
				x.Rules = make(map[string]string)
				for key2, val2 := range val1.(map[string]interface{}) {
					x.Keys = append(x.Keys, key2)
					x.Rules[key2] = val2.(string)
				}
				seq.PodCs = append(seq.PodCs, x)
			}
		}
		if key == "shortcuts" {
			seq.Shorts = make(map[string]string)
			for key1, val1 := range val.(map[string]interface{}) {
				fmt.Println(key1, val1)
				seq.ShortsKeys = append(seq.ShortsKeys, key1)
				seq.Shorts[key1] = val1.(string)
			}
			slices.Sort(seq.ShortsKeys)
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
		}
	}
	return seq, nil
}

func OpLineTagToOpString(line string) (print, operation string) {
	ret_print := "UNKNOWN_OPERATION"
	ret_operation := "OP_UNKNOWN"
	if len(line) > 7 && line[:5] != "{{OP_" {
		return ret_print, ret_operation
	}
	check_operation := strings.Split(line[2:], "}}")[0]
	for i := 0; i < len(SupportedOps); i++ {
		if OpInstruction[SupportedOps[i]] == check_operation {
			ret_print = check_operation[3:]
			ret_operation = OpInstruction[SupportedOps[i]]
			break
		}
	}
	// fmt.Println("_ ",ret_print, ret_operation)
	return ret_print, ret_operation
}

func OpLineTagToString(line string) string {
	to_print, operation := OpLineTagToOpString(line)
	return fmt.Sprintf("#%s:%s", to_print, line[len(operation)+4:])
}

func ExpandShortcuts(line string, shorts map[string]string, keys []string) string {
	newLine := line
	for l := 0; l < 100; l++ {
		changes := false
		for i := 0; i < len(keys); i++ {
			key := keys[i]
			value := shorts[key]
			after := strings.Replace(newLine, "{{"+key+"}}", value, 10)
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

func ExpandK8s(line, K8sConfig, K8sContext, K8sNamespace, K8sPod string) string {
	newLine := line
	for l := 0; l < 100; l++ {
		changes := false
		after := strings.Replace(newLine, "{{k8s_config}}", K8sConfig, 10)
		after = strings.Replace(after, "{{k8s_context}}", K8sContext, 10)
		after = strings.Replace(after, "{{k8s_namespace}}", K8sNamespace, 10)
		after = strings.Replace(after, "{{k8s_pod}}", K8sPod, 10)
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
		var err error = nil
		matched := false
		matched, err = regexp.Match(rules[keys[i]], []byte(pod))
		if err != nil {
			fmt.Println("Non fatal Error, failed in compiling p2c rules")
			fmt.Println(err)
			continue
		}
		if matched {
			return keys[i]
		}
	}
	return "default"
}

func ExpandPodConverter(line, K8s_pod string, p2c []podConverter) string {
	newLine := line
	for l := 0; l < 100; l++ {
		changes := false
		for k := range p2c {
			after := strings.Replace(newLine, "{{"+p2c[k].Name+"}}", ExpandP2cRule(p2c[k].Rules, p2c[k].Keys, K8s_pod), 10)
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
		return -1, errors.New(fmt.Sprintf("# No matching script %s in conf file", name))
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

var KctlVariables = map[KctlDecode]string{
	KctlConfig:    "k8s_config",
	KctlContext:   "k8s_context",
	KctlNamespace: "k8s_namespace",
	KctlPod:       "k8s_pod",
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
	if len(inputLine) > 5 && inputLine[:5] == "{{OP_" {
		line = OpLineTagToString(inputLine)
		if len(line) >= 6 && line[:6] == "#INFO:" {
			return OpInfo, line[6:]
		}
		if len(line) >= 9 && line[:9] == "#COMMENT:" {
			return OpComment, line[9:]
		}
		if len(line) >= 7 && line[:7] == "#SLEEP:" {
			return OpSleep, line[7:]
		}
		if len(line) >= 11 && line[:11] == "#TERMINATE:" {
			return OpTerminate, ""
		}
		if len(line) >= 8 && line[:8] == "#ATTACH:" {
			return OpAttach, ""
		}
		if len(line) >= 8 && line[:8] == "#DETACH:" {
			return OpDetach, ""
		}
		if len(line) >= 9 && line[:9] == "#FINALLY:" {
			return OpFinally, line[12:]
		}
		if len(line) >= 16 && line[:16] == "#NO_PROMPT_WAIT:" {
			return OpNoPrompt, ""
		}
		if len(line) >= 16 && line[:16] == "#REFRESH_PROMPT:" {
			return OpRefreshPrompt, ""
		}
		return OpUnknown, ""
	}
	return OpExecute, inputLine
}
