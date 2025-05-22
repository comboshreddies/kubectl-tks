package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

type scriptItem struct {
	Name  string
	Items []string
}

// map ?
type shortcuts struct {
	Name  string
	Value string
}

type predefined struct {
	Name string
	Tags []string
}

type podConverter struct {
	Name  string
	Rules []p2cRule
}

// map ?
type p2cRule struct {
	Name   string
	Regexp string
}

type sequenceConfig struct {
	Predefs []predefined
	PodCs   []podConverter
	Shorts  []shortcuts
	Scripts []scriptItem
}

func OpenAndReadSequencefile(fileName string) (conf sequenceConfig, err error) {
	var seq sequenceConfig

	jsonFile, err := os.Open(fileName)
	if err != nil {
		return sequenceConfig{}, errors.New("unable to open sequence json file " + fileName)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(byteValue), &result)
	if err != nil {
		newerr := errors.New("unable to load json from sequence json file")
		return sequenceConfig{}, newerr
	}

	for key, val := range result {
		if key == "internal" {
			for key1, val1 := range val.(map[string]interface{}) {
				var t predefined
				t.Name = key1
				for _, val2 := range val1.([]interface{}) {
					t.Tags = append(t.Tags, val2.(string))
				}
				seq.Predefs = append(seq.Predefs, t)
			}
		}
		if key == "podConverter" {
			for key1, val1 := range val.(map[string]interface{}) {
				var x podConverter
				x.Name = key1
				for key2, val2 := range val1.(map[string]interface{}) {
					var t p2cRule
					t.Name = key2
					t.Regexp = val2.(string)
					x.Rules = append(x.Rules, t)
				}
				seq.PodCs = append(seq.PodCs, x)
			}
		}
		if key == "shortcuts" {
			for key1, val1 := range val.(map[string]interface{}) {
				var t shortcuts
				t.Name = key1
				t.Value = val1.(string)
				seq.Shorts = append(seq.Shorts, t)
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
		}
	}
	return seq, nil
}

func OpLineTagToOpString(line string) (print, operation string) {
	ret_print := "UNKNOWN_OPERATION"
	ret_operation := "OP_UNKNOWN"
	if line[:5] != "{{OP_" {
		return ret_print, ret_operation
	}
	check_operation := strings.Split(line[2:], "}}")[0]
	switch check_operation {
	case "OP_INFO", "OP_COMMENT", "OP_NO_RETURN", "OP_FINAL_EXEC", "OP_ATTACH", "OP_TERMINATE", "OP_SLEEP", "OP_REFRESH_PROMPT", "OP_SYNC":
		ret_print = check_operation[3:]
		ret_operation = check_operation
	}
	//fmt.Println(ret_operation)
	return ret_print, ret_operation
}

func OpLineTagToString(line string) string {
	to_print, operation := OpLineTagToOpString(line)
	//fmt.Printf("#%s: ->%s<- %d\n", to_print, line, len(line))
	return fmt.Sprintf("#%s:%s", to_print, line[len(operation)+4:])
}

func ExpandShortcuts(line string, shorts []shortcuts) string {
	newLine := line
	for l := 0; l < 100; l++ {
		changes := false
		for k := range shorts {
			after := strings.Replace(newLine, "{{"+shorts[k].Name+"}}", shorts[k].Value, 10)
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

func ExpandK8s(line, K8sContext, K8sNamespace, K8sPod string) string {
	newLine := line
	for l := 0; l < 100; l++ {
		changes := false
		after := strings.Replace(newLine, "{{k8s_context}}", K8sContext, 10)
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

func ExpandP2cRule(rules []p2cRule, pod string) string {
	for i := 0; i < len(rules); i++ {
		var err error = nil
		matched := false
		matched, err = regexp.Match(rules[i].Regexp, []byte(pod))
		if err != nil {
			fmt.Println("Non fatal Error, failed in compiling p2c rules")
			fmt.Println(err)
			continue
		}
		if matched {
			return rules[i].Name
		}
	}
	return "default"
}

func ExpandPodConverter(line, K8s_pod string, p2c []podConverter) string {
	newLine := line
	for l := 0; l < 100; l++ {
		changes := false
		for k := range p2c {
			after := strings.Replace(newLine, "{{"+p2c[k].Name+"}}", ExpandP2cRule(p2c[k].Rules, K8s_pod), 10)
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
		return -1, errors.New("no matching script in sequence file")
	}
	return seqOffset, nil
}

type OpDecoded int

const (
	OpTerminate OpDecoded = iota
	OpAttach
	OpFinally
	OpExecute
	OpComment
	OpInfo
	OpNoPrompt
	OpSleep
	OpRefreshPrompt
	OpUnknown
)

var OpName = map[OpDecoded]string{
	OpTerminate:     "Terminate tmux, script end",
	OpAttach:        "Attach tmux, script end",
	OpFinally:       "Finally execute, script end",
	OpExecute:       "Execute line",
	OpInfo:          "Print info",
	OpComment:       "Print comment, render",
	OpNoPrompt:      "Do not wait for prompt for last command",
	OpSleep:         "Sleep for n seconds",
	OpRefreshPrompt: "Load new prompt",
	OpUnknown:       "Operation Uknown",
}

func opDecode(inputLine string) (op OpDecoded, line string) {
	if inputLine[:5] == "{{OP_" {
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
		if len(line) >= 12 && line[:12] == "#FINAL_EXEC:" {
			return OpFinally, line[12:]
		}
		if len(line) >= 11 && line[:11] == "#NO_RETURN:" {
			return OpNoPrompt, ""
		}
		if len(line) >= 16 && line[:16] == "#REFRESH_PROMPT:" {
			return OpRefreshPrompt, ""
		}
		return OpUnknown, ""
	}
	return OpExecute, inputLine
}
