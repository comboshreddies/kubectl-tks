package internal

import (
	"fmt"
	"github.com/GianlucaP106/gotmux/gotmux"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type TmuxInData struct {
	SeqName      string
	ScriptLines  []string
	K8sConfig    string
	K8sContext   string
	K8sNamespace string
	PodList      []PodsInfo
	Shorts       map[string]string
	ShortsKeys   []string
	PodCs        []podMap
	Prompt       string
	PromptSleep  int
	SessionName  string
}

func StartTmux(ti TmuxInData, dry, syncExec, delTxSess, quiet bool) {
	//	fmt.Println(ti.SeqName)
	//	fmt.Println(ti.ScriptLines)
	//	fmt.Println(ti.K8sConfig)
	//	fmt.Println(ti.K8sContext)
	//	fmt.Println(ti.K8sNamespace)
	//	fmt.Println(ti.PodList)
	//	fmt.Println(ti.SessionName)
	//	fmt.Println(dry, syncExec, delTxSess)
	//	fmt.Println("------====----")

	if dry == true {
		if !quiet {
			fmt.Printf("#### Starting execution: sync : %t, dry : %t\n", syncExec, dry)
		}
		dryRunPrintOut(ti, syncExec)
		return
	}

	tmux, err := gotmux.DefaultTmux()
	if err != nil {
		fmt.Println("error opening default tmux")
		return
	}

	var tmuxSessionName string
	if ti.SessionName == "" {
		tmuxSessionName = generateTmuxSessionName(ti)
	} else {
		tmuxSessionName = ti.SessionName
	}

	if tmux.HasSession(tmuxSessionName) {
		if !quiet {
			fmt.Printf("# there is already session with this name (%s), ", tmuxSessionName)
		}
		if delTxSess == true {
			cmd := exec.Command(os.Getenv("SHELL"), "-c", fmt.Sprintf("tmux kill-session -t %s\n", tmuxSessionName))
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
				fmt.Printf("unable to terminate session %s, exiting", tmuxSessionName)
				return
			}
			if !quiet {
				fmt.Printf("terminating previous session\n")
			}
		} else {
			fmt.Printf("tmux session exists, exiting\n")
			return
		}
	}

	if !quiet {
		fmt.Printf("#### Creating new session %s\n", tmuxSessionName)
	}
	tmuxSession, err := tmux.NewSession(&gotmux.SessionOptions{
		Name: tmuxSessionName,
	})
	if err != nil {
		fmt.Println("error while creating new session")
		return
	}

	windows, err := tmuxSession.ListWindows()
	windows[0].Rename("base")

	if !quiet {
		fmt.Printf("#### Creating windows per pod\n")
	}
	// open window per pod
	for i := 0; i < len(ti.PodList); i++ {
		newWinOpts := gotmux.NewWindowOptions{}
		newWinOpts.WindowName = ti.PodList[i].PodName
		_, err = tmuxSession.NewWindow(&newWinOpts)
		if err != nil {
			fmt.Println("error creating new window")
			fmt.Println(err)
			return
		}
	}

	// catch prompt, we could track it for each pod/window
	// sleep for a while
	time.Sleep(time.Second * time.Duration(ti.PromptSleep))

	windows, err = tmuxSession.ListWindows()

	if !quiet {
		fmt.Printf("#### Collecting prompts for each window\n")
	}
	prompts := make(map[int]string)
	for i := 0; i < len(windows); i++ {
		prompts[i], err = tmux_get_pane_prompt(windows[i])
		if err != nil {
			fmt.Println("error in fetching pane prompt")
			return
		}
	}
	winName2Idx := make(map[string]int)
	windows, err = tmuxSession.ListWindows()
	for i := 0; i < len(windows); i++ {
		winName2Idx[windows[i].Name] = i
	}

	podIdx2WinIdx := make(map[int]int)
	for i := 0; i < len(ti.PodList); i++ {
		podIdx2WinIdx[i] = winName2Idx[ti.PodList[i].PodName]
	}

	if !quiet {
		fmt.Printf("#### Starting execution: sync : %t, dry : %t\n", syncExec, dry)
	}
	if syncExec == true {
		line := ""
		var operation OpDecoded
		for scrIdx := 0; scrIdx < len(ti.ScriptLines); scrIdx++ {
			operation, line = opDecode(ti.ScriptLines[scrIdx])
			switch operation {
			case OpTerminate:
				break
			case OpAttach:
				break
			case OpDetach:
				break
			case OpFinally:
				break
			case OpExecute:
				for podIdx := 0; podIdx < len(ti.PodList); podIdx++ {
					execLine := RenderLineForExec(ti, podIdx, scrIdx)
					fmt.Printf("#EXECUTE: #%d %s: %s\n", scrIdx, ti.PodList[podIdx].PodName, line)
					err := windowsSendKeys(windows[podIdx2WinIdx[podIdx]], execLine+"\n")
					if err != nil {
						fmt.Println(err)
						return
					}
				}
				doPromptCheck := true
				if len(ti.ScriptLines) > scrIdx+1 {
					nextOperation, _ := opDecode(ti.ScriptLines[scrIdx+1])
					if nextOperation == OpNoPrompt || nextOperation == OpRefreshPrompt {
						doPromptCheck = false
					}
				}
				if doPromptCheck == true {
					allComplete := false
					for allComplete == false {
						allComplete = true
						for podIdx := 0; podIdx < len(ti.PodList); podIdx++ {
							current_prompt, err := tmux_get_pane_prompt(windows[podIdx2WinIdx[podIdx]])
							if err != nil {
								break
							}
							if current_prompt != prompts[podIdx] {
								allComplete = false
							}
							time.Sleep(time.Millisecond * time.Duration(200))
						}
					}
				}
				fmt.Printf("#STEP %d complete on all pods\n", scrIdx)
			case OpInfo:
				fmt.Printf("#INFO: %s\n", line)
			case OpComment:
				fmt.Printf("#Comment: %s\n", line)
				for podIdx := 0; podIdx < len(ti.PodList); podIdx++ {
					fmt.Printf("#COMMENT: %s\n", RenderLineForExec(ti, podIdx, scrIdx))
				}
			case OpNoPrompt:
				fmt.Printf("#NO_PROMPT\n")
				continue
			case OpSleep:
				fmt.Printf("#SLEEP: %s\n", line)
				internalSleep(line)
			case OpRefreshPrompt:
				fmt.Printf("#REFRESH_PROMPT: %s\n", line)
				time.Sleep(time.Second * time.Duration(ti.PromptSleep))
				for podIdx := 0; podIdx < len(ti.PodList); podIdx++ {
					prompts[podIdx], err = tmux_get_pane_prompt(windows[podIdx2WinIdx[podIdx]])
					if err != nil {
						fmt.Println("error in fetching pane prompt")
						break
					}
				}
			case OpUnknown:
				fmt.Printf("# Unknown operation, skipping - %s\n", ti.ScriptLines[scrIdx])
			}
			switch operation {
			case OpTerminate:
				fmt.Println("#TERMINATE")
				windows, err = tmuxSession.ListWindows()
				for i := 0; i < len(windows); i++ {
					windows[i].Kill()
				}
				fmt.Println("windows within session terminated")
			case OpAttach:
				fmt.Println("#ATTACH")
				opts := gotmux.AttachSessionOptions{}
				tmuxSession.AttachSession(&opts)
			case OpDetach:
				fmt.Println("#DETACH")
			case OpFinally:
				fmt.Printf("#FINALY: %s\n", line)
				cmd := exec.Command(os.Getenv("SHELL"), "-c", line)
				err := cmd.Run()
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	} else { // syncExec == false
		var wg sync.WaitGroup
		var podStep []int
		finalOp := make(map[int]OpDecoded)
		finalLine := make(map[int]string)
		for podIdx := 0; podIdx < len(ti.PodList); podIdx++ {
			podStep = append(podStep, 0)
			finalOp[podIdx] = OpUnknown
			finalLine[podIdx] = ""
		}
		shouldWaitForPrompt := make(map[int]map[int]bool)
		executionSent := make(map[int]map[int]bool)
		for podIdx := 0; podIdx < len(ti.PodList); podIdx++ {
			shouldWaitForPrompt[podIdx] = map[int]bool{}
			executionSent[podIdx] = map[int]bool{}
			for scrIdx := 0; scrIdx < len(ti.ScriptLines); scrIdx++ {
				shouldWaitForPrompt[podIdx][scrIdx] = false
				executionSent[podIdx][scrIdx] = false
			}
		}

		allPodsComplete := false
		lastOperation := OpUnknown
		lastLine := ""
		for allPodsComplete == false {
			podIdx := 0
			run := func(podIdx int, wg *sync.WaitGroup) {
				for podStep[podIdx] < len(ti.ScriptLines) {

					scrIdx := podStep[podIdx]
					podName := ti.PodList[podIdx].PodName
					line := ""
					operation := OpUnknown
					operation, line = opDecode(ti.ScriptLines[podStep[podIdx]])
					switch operation {
					case OpTerminate:
						lastOperation = operation
						podStep[podIdx] = len(ti.ScriptLines)
						finalOp[podIdx] = OpTerminate
						break
					case OpAttach:
						lastOperation = operation
						podStep[podIdx] = len(ti.ScriptLines)
						finalOp[podIdx] = OpAttach
						break
					case OpFinally:
						lastOperation = operation
						lastLine = line
						podStep[podIdx] = len(ti.ScriptLines)
						finalOp[podIdx] = OpFinally
						finalLine[podIdx] = line
						break
					case OpExecute:
						if executionSent[podIdx][scrIdx] != true {
							execLine := RenderLineForExec(ti, podIdx, scrIdx)
							fmt.Printf("#EXECUTE #%d %s: %s\n", scrIdx, podName, execLine)
							err := windowsSendKeys(windows[podIdx2WinIdx[podIdx]], execLine)
							if err != nil {
								fmt.Println(err)
								finalOp[podIdx] = OpUnknown
								podStep[podIdx] = len(ti.ScriptLines)
								break
							}
							executionSent[podIdx][scrIdx] = true
						}
						shouldWaitForPrompt[podIdx][scrIdx] = true
						// look ahead
						if len(ti.ScriptLines) > scrIdx+1 {
							nextOperation, _ := opDecode(ti.ScriptLines[scrIdx+1])
							if nextOperation == OpNoPrompt || nextOperation == OpRefreshPrompt {
								shouldWaitForPrompt[podIdx][scrIdx] = false
							}
						}
						if shouldWaitForPrompt[podIdx][scrIdx] == true {
							time.Sleep(time.Millisecond * time.Duration(200))
							current_prompt, err := tmux_get_pane_prompt(windows[podIdx2WinIdx[podIdx]])
							if err != nil {
								break // find way to singal error on execution
							}
							if current_prompt == prompts[podIdx] {
								shouldWaitForPrompt[podIdx][scrIdx] = false
							}
						}
						if shouldWaitForPrompt[podIdx][scrIdx] == false {
							executionSent[podIdx][scrIdx] = false
							podStep[podIdx] += 1
							continue
						}
					case OpInfo:
						fmt.Printf("#INFO #%d, %s: %s\n", scrIdx, podName, line)
						podStep[podIdx] += 1
					case OpComment:
						fmt.Printf("#COMMENT #%d, %s: %s\n", scrIdx, podName, RenderLineForExec(ti, podIdx, scrIdx))
						podStep[podIdx] += 1
					case OpNoPrompt:
						fmt.Printf("#NO_PROMPT #%d, %s:\n", scrIdx, podName)
						podStep[podIdx] += 1
					case OpSleep:
						fmt.Printf("#SLEEP #%d, %s: %s\n", scrIdx, podName, line)
						internalSleep(line)
						podStep[podIdx] += 1
					case OpRefreshPrompt:
						fmt.Printf("#REFRESH_PROMPT #%d, %s:\n", scrIdx, podName)
						time.Sleep(time.Second * time.Duration(ti.PromptSleep))
						prompts[podIdx], err = tmux_get_pane_prompt(windows[podIdx2WinIdx[podIdx]])
						if err != nil {
							fmt.Println("error in fetching pane prompt")
							finalOp[podIdx] = OpUnknown
							podStep[podIdx] = len(ti.ScriptLines)
							break
						}
						podStep[podIdx] += 1
					case OpUnknown:
						fmt.Printf("# Unknown operation #%d: skipping - %s\n", scrIdx, ti.ScriptLines[podStep[podIdx]])
						podStep[podIdx] += 1
					}
				}
				wg.Done()
			}
			for podIdx = 0; podIdx < len(ti.PodList); podIdx++ {
				wg.Add(1)
				go run(podIdx, &wg)
			}
			wg.Wait()
			allPodsComplete = true
		}
		if lastOperation != OpUnknown {
			switch lastOperation {
			case OpTerminate:
				fmt.Println("#TERMINATE")
				windows, err = tmuxSession.ListWindows()
				for i := 0; i < len(windows); i++ {
					windows[i].Kill()
				}
				fmt.Println("# windows within session terminated")
			case OpAttach:
				fmt.Println("#ATTACH")
				opts := gotmux.AttachSessionOptions{}
				tmuxSession.AttachSession(&opts)
			case OpFinally:
				fmt.Printf("#FINAL_EXEC: %s\n", lastLine)
				cmd := exec.Command(os.Getenv("SHELL"), "-c", lastLine)
				err := cmd.Run()
				if err != nil {
					fmt.Println(err)
				}

			}
		}
	}
	fmt.Println("#COMPLETED")
	return
}

func tmux_get_pane_prompt(window *gotmux.Window) (prompt string, err error) {
	pane, err := window.GetPaneByIndex(0)
	if err != nil {
		fmt.Println("error get window - get prompt")
		return "", err
	}

	cap, err := pane.Capture()
	if err != nil {
		fmt.Println("error pane capture")
		return "", err
	}

	lines := strings.Split(cap, "\n")
	prevprompt := ""
	for i := 0; i < len(lines); i++ {
		if prevprompt != lines[i] && lines[i] != "" {
			prevprompt = lines[i]
		}
	}
	return prevprompt, nil
}

func dryRenderLine(ti TmuxInData, podListIndex, scriptLineIndex int, syncExec bool) {
	podName := ti.PodList[podListIndex].PodName
	original := ti.ScriptLines[scriptLineIndex]
	var line string
	op := OpExecute

	if (len(original) > 3 && original[:3] == "{{_") || (len(original) > 5 && original[:2] == "{{") {
		op, line = OpLineTagToOpString(original)
		original = line
	}
	line = ExpandShortcuts(original, ti.Shorts, ti.ShortsKeys)
	line = ExpandUnderscore(line, ti.K8sConfig, ti.K8sContext, ti.K8sNamespace)
	line = ExpandK8s(line, ti.K8sConfig, ti.K8sContext, ti.K8sNamespace, podName)
	line = ExpandPodMapper(line, podName, ti.PodCs)

	div := "|"
	if syncExec == true {
		if podListIndex == 0 {
			div = "/"
		}
		if podListIndex == len(ti.PodList)-1 {
			div = "\\"
		}
		if 1 == len(ti.PodList) {
			div = "*"
		}
	} else {
		if scriptLineIndex == 0 {
			div = "/"
		}
		if scriptLineIndex == len(ti.ScriptLines)-1 {
			div = "\\"
		}
		if 1 == len(ti.ScriptLines) {
			div = "*"
		}
	}

	fmt.Printf("%s %s%d %s: %s\n", OpPrint[op], div, scriptLineIndex, podName, line)

}

func RenderLineForExec(ti TmuxInData, podListIndex, scriptLineIndex int) string {
	podName := ti.PodList[podListIndex].PodName
	original := ti.ScriptLines[scriptLineIndex]
	var line string
	if (len(original) > 3 && original[:3] == "{{_") || (len(original) > 5 && original[:5] == "{{OP_") {
		// should not match, but if we do, put # shell comment
		op, line := OpLineTagToOpString(original)
		if op == OpComment { // opComment is only non OpExec that is Rendered for exec
			original = fmt.Sprintf("#%s", line)
		} else {
			original = line
		}
	}
	line = ExpandShortcuts(original, ti.Shorts, ti.ShortsKeys)
	line = ExpandUnderscore(line, ti.K8sConfig, ti.K8sContext, ti.K8sNamespace)
	line = ExpandK8s(line, ti.K8sConfig, ti.K8sContext, ti.K8sNamespace, podName)
	line = ExpandPodMapper(line, podName, ti.PodCs)
	return line
}

func dryRunPrintOut(ti TmuxInData, syncExec bool) {
	if syncExec == true {
		for i := 0; i < len(ti.ScriptLines); i++ {
			for j := 0; j < len(ti.PodList); j++ {
				dryRenderLine(ti, j, i, syncExec)
			}
		}
	} else {
		for i := 0; i < len(ti.PodList); i++ {
			for j := 0; j < len(ti.ScriptLines); j++ {
				dryRenderLine(ti, i, j, syncExec)
			}
		}
	}
}

func windowsSendKeys(window *gotmux.Window, line string) error {
	pane, err := window.GetPaneByIndex(0)
	if err != nil {
		fmt.Println("error get window - send keys")
		return err
	}

	err = pane.SendKeys(fmt.Sprintf("%s\n", line))
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func generateTmuxSessionName(ti TmuxInData) string {
	return fmt.Sprintf("%s-%s-%s", ti.SeqName, ti.K8sContext, ti.K8sNamespace)
}

func internalSleep(line string) {
	sleepSeconds := 1
	n, err := fmt.Sscanf(line, " %d", &sleepSeconds)
	if n != 1 || err != nil {
		time.Sleep(time.Second * time.Duration(1))
	} else {
		time.Sleep(time.Second * time.Duration(sleepSeconds))
	}
}
