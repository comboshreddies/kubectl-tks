
##   Tmux Kubectl Scripts - tks
# tks kubectl plugin - or stand-alone tool
## A plugin for executing scripts on pods within:

Tks plugin runs multiple executions scripts (sequences) in multiple tmux windows.
Each window runs script on one pod. You can decide if you want to attach and inspect
executions per each pod, or you want to just execute and exit tmux.

With this tool you can:
1) run any command for specific pod selections, like: 
```console
"kubectl -n {{k8s_namespace}} exec {{k8s_pod}} -c {{p2c}}  -- /bin/sh `env > {{k8s_pod}}.env`"
```
tks will automatically change pod name namespace and for each execution step

2) you can select by namespace, label, then filter specific pods, like
```console
kubectl tks --context minikube -n test-run start env -l app=nginx
kubectl tks --context minikube -n test-run start env -p busybox1,busybox2  
kubectl tks --context minikube -n test-run start env -l app=busybox -p busybox1,busybox2  
```

3) You can decide would you like to terminate, attach tmux or do some command at the and of execution

4) You can run scripts in:
* synchronous mode - so each script line step is done on all pods, and once all are complete (prompt is returned) tks will go to next step, add -s switch on start command
* asynchronous mode (default) - each pod runs it's own set of sequences (no -s switch on start command)

 
With this tool you can attach tmux and take a look, or you can gather results of executions.
If you have some actions that you frequently do, like gathering some info from pods, you
can add sequence that fits your needs, and next time you can run it fast. Tmux solves
problem of having multiple outputs of executions on same terminal, as each tmux screen/window
is dedicated to one pod execution, so you have terminal for each pod, do ctrl+b+n and
go to next pod window.


## Installation
1) go build
2) copy kubectl-tks to your path

Additionally, but can be done later
3) mkdir  ~/.tks
4) copy sequences.json to ~/.tks/

# purpose
One can easily make a a shell script to execute some command on each pod/container in parallel, then gather results.
Executing on more than few pods might fail on some pods. Having output of many pod executions on one screen is not practical. You might need to take a look at each execution on your own, one pod at the time. You might want to keep terminal session for each pod that you've connected to. Using tmux could help.


# intro

To practice first steps create a deployment

```console
$ kubectl create ns test-run
$ kubectl -n test-run apply -f k8s_yamls/sample_deploy1.yaml
$ kubectl -n test-run get pod --show-labels
NAME                             READY   STATUS    RESTARTS   AGE   LABELS
nginx-sample1-6475dd48b7-bgtsx   1/1     Running   0          16h   app=nginx,pod-template-hash=6475dd48b7,ver=v1
nginx-sample1-6475dd48b7-br6jc   1/1     Running   0          16h   app=nginx,pod-template-hash=6475dd48b7,ver=v1
nginx-sample1-6475dd48b7-n6mtg   1/1     Running   0          16h   app=nginx,pod-template-hash=6475dd48b7,ver=v1
$ kubectl -n test-run get pod -l app=nginx
nginx-sample1-6475dd48b7-bgtsx   1/1     Running   0          16h   app=nginx,pod-template-hash=6475dd48b7,ver=v1
nginx-sample1-6475dd48b7-br6jc   1/1     Running   0          16h   app=nginx,pod-template-hash=6475dd48b7,ver=v1
nginx-sample1-6475dd48b7-n6mtg   1/1     Running   0          16h   app=nginx,pod-template-hash=6475dd48b7,ver=v1
```

## One-liner example
let's run simple one-liner
```console
$ kubectl tks -n test-run start -l app=nginx  "kubectl -n test-run exec {{k8s_pod}} -c nginx -- env"
# Unable to read conf file /Users/none/.tks/sequences.json, assuming oneLiner
# unable to open sequence json file /Users/none/.tks/sequences.json
#### Creating new session OneLiner--test-run
#### Creating windows per pod
#### Collecting prompts for each window
#### Starting execution: sync : false, dry false
#EXECUTE #0 nginx-sample1-6475dd48b7-br6jc: kubectl -n test-run exec nginx-sample1-6475dd48b7-br6jc -c nginx -- env
#EXECUTE #0 nginx-sample1-6475dd48b7-n6mtg: kubectl -n test-run exec nginx-sample1-6475dd48b7-n6mtg -c nginx -- env
#EXECUTE #0 nginx-sample1-6475dd48b7-bgtsx: kubectl -n test-run exec nginx-sample1-6475dd48b7-bgtsx -c nginx -- env
#COMPLETED
```

you can check what has been executed by attaching to tmux
```console
$ tmux a
```

## one-liner leaves tmux session open
if you run same command second time (and if you have not terminated previous tmux session) you will get
```console
$ kubectl tks -n test-run start -l app=nginx  "kubectl -n test-run exec {{k8s_pod}} -c nginx -- env"
# Unable to read conf file /Users/none/.tks/sequences.json, assuming oneLiner
# unable to open sequence json file /Users/none/.tks/sequences.json
# there is already session with this name (OneLiner--test-run), exiting
```
you can use
```console
tmux ls
```
to check what sessions are running.


## one-liner with previous session removal
If you want to start and remove previous session, remember to use -T option.
You can always remove previous session with tmux kill-session -t <session_name> .
Flag -T  will terminate previous session of a same name. Session Names are
generated by concatenating sequence-name (for one-liners it's OneLiner), kubernetes context (if specified) and
kubernetes namespace. Default behaviour of tks plugin is to leave tmux session in detached state.

```console
$ kubectl tks -n test-run start -l app=nginx  "kubectl -n test-run exec -t {{k8s_pod}} -c nginx -- env" -T
# Unable to read conf file /Users/none/.tks/sequences.json, assuming oneLiner
# unable to open sequence json file /Users/none/.tks/sequences.json
# there is already session with this name (OneLiner--test-run), terminating old one
#### Creating new session OneLiner--test-run
#### Creating windows per pod
#### Collecting prompts for each window
#### Starting execution: sync : false, dry false
#EXECUTE #0 nginx-sample1-6475dd48b7-n6mtg: kubectl -n test-run exec -t nginx-sample1-6475dd48b7-n6mtg -c nginx -- env
#EXECUTE #0 nginx-sample1-6475dd48b7-br6jc: kubectl -n test-run exec -t nginx-sample1-6475dd48b7-br6jc -c nginx -- env
#EXECUTE #0 nginx-sample1-6475dd48b7-bgtsx: kubectl -n test-run exec -t nginx-sample1-6475dd48b7-bgtsx -c nginx -- env
#COMPLETED
```

## one-liner with more kubernetes related templated fiels

Now you can see we used k8s_pod as temlpate field for each pod was there in test-run namespace.
We can also use k8s_namespace template field to specify namespace (so you don't have to repeat test-run)
There are also k8s_config and k8s_context template fields that would be filled if specified within tks command line
parameters - if they are not specified they will be empty string.

```console
$ kubectl tks -n test-run start -l app=nginx  "kubectl -n {{k8s_namespace}} exec -t {{k8s_pod}} -c nginx -- env" -T
# Unable to read conf file /Users/none/.tks/sequences.json, assuming oneLiner
# unable to open sequence json file /Users/none/.tks/sequences.json
# there is already session with this name (OneLiner--test-run), terminating old one
#### Creating new session OneLiner--test-run
#### Creating windows per pod
#### Collecting prompts for each window
#### Starting execution: sync : false, dry false
#EXECUTE #0 nginx-sample1-6475dd48b7-br6jc: kubectl -n test-run exec -t nginx-sample1-6475dd48b7-br6jc -c nginx -- env
#EXECUTE #0 nginx-sample1-6475dd48b7-bgtsx: kubectl -n test-run exec -t nginx-sample1-6475dd48b7-bgtsx -c nginx -- env
#EXECUTE #0 nginx-sample1-6475dd48b7-n6mtg: kubectl -n test-run exec -t nginx-sample1-6475dd48b7-n6mtg -c nginx -- env
#COMPLETED
```


## one-liner with dry run mode
You can always run command in dry-run mode with -d flag, only output will be printend, but templated fields will
encoded. Table shows podname, pod number, script line number, then command that would have been executed

```console
$ kubectl tks -n test-run start -l app=nginx  "kubectl -n {{k8s_namespace}} exec {{k8s_pod}} -c nginx -- env" -d
# Unable to read conf file /Users/none/.tks/sequences.json, assuming oneLiner
# unable to open sequence json file /Users/none/.tks/sequences.json
nginx-sample1-6475dd48b7-bgtsx 0 0 kubectl -n test-run exec -t nginx-sample1-6475dd48b7-bgtsx -c nginx -- env
nginx-sample1-6475dd48b7-br6jc 1 0 kubectl -n test-run exec -t nginx-sample1-6475dd48b7-br6jc -c nginx -- env
nginx-sample1-6475dd48b7-n6mtg 2 0 kubectl -n test-run exec -t nginx-sample1-6475dd48b7-n6mtg -c nginx -- env
```


## one-liner with more then one executions
You can specify more than one command in One-Liner execution. Commands are separated by ';' sign.
```console
$ kubectl tks -n test-run start -l app=nginx  "kubectl -n {{k8s_namespace}} exec {{k8s_pod}} -c nginx -- env;echo {{k8s_pod}}" -T
```
You can do tmux attach to inspect execution, by switching between tmux terminal windows.

## one-liner with attaching to tmux session
If you do not want to manually attach to execution every time you can add OP_ command for attaching 

```console
$ kubectl tks -n test-run start -l app=nginx  "kubectl -n {{k8s_namespace}} exec {{k8s_pod}} -c nginx -- env;{{OP_ATTACH}}" -T
```
Now at the end of execution tks will attach tmux session it created. 
There are more OP_ commands available, you can find more in question sectoin Q: What other OP_ commands are available ?


## using sequence file for storing complex scripts
If you are tired of creating One-Liners you have to remember, you could copy sequence.json to ~/tks directory
(as mentioned in installation step 3 and 4 above)

## sequence.json contents
sequence.json file cotains 3 sections:
- scripts
- shortcuts
- podConverter

### sequence.json - scripts section 
Scripts section define script name and list of actions.
For example:
```console
"scripts" : {
    "env-nginx-simple": [
        "{{OP_INFO}} execute env on each pod and put to pod.env file",
        "kubectl -n {{k8s_namespace} exec {{k8s_pod}} -c nginx  -- env > {{k8s_pod}}.env",
        "cat {{k8s_pod}}.env"
    ],
```
Scripts are working same way as OneLiners but you can call them by name (env-nginx-simple) instead of typing
long OneLiners. 

so you can run
```console
$ kubectl tks -n test-run start env-nginx-simple -l app=nginx -T
#### Creating new session env-nginx-simple--test-run
#### Creating windows per pod
#### Collecting prompts for each window
#### Starting execution: sync : false, dry false
#INFO #0, nginx-sample1-6475dd48b7-n6mtg:  execute env on each pod and put to pod.env file
#INFO #0, nginx-sample1-6475dd48b7-bgtsx:  execute env on each pod and put to pod.env file
#INFO #0, nginx-sample1-6475dd48b7-br6jc:  execute env on each pod and put to pod.env file
#EXECUTE #1 nginx-sample1-6475dd48b7-br6jc: kubectl -n test-run exec nginx-sample1-6475dd48b7-br6jc -c nginx  -- env > nginx-sample1-6475dd48b7-br6jc.env
#EXECUTE #1 nginx-sample1-6475dd48b7-bgtsx: kubectl -n test-run exec nginx-sample1-6475dd48b7-bgtsx -c nginx  -- env > nginx-sample1-6475dd48b7-bgtsx.env
#EXECUTE #1 nginx-sample1-6475dd48b7-n6mtg: kubectl -n test-run exec nginx-sample1-6475dd48b7-n6mtg -c nginx  -- env > nginx-sample1-6475dd48b7-n6mtg.env
#EXECUTE #2 nginx-sample1-6475dd48b7-br6jc: cat nginx-sample1-6475dd48b7-br6jc.env
#EXECUTE #2 nginx-sample1-6475dd48b7-n6mtg: cat nginx-sample1-6475dd48b7-n6mtg.env
#EXECUTE #2 nginx-sample1-6475dd48b7-bgtsx: cat nginx-sample1-6475dd48b7-bgtsx.env
#COMPLETED
``` 
You can checkk files in local directory, there should be nginx-sample1- .env files for each pod
Also you can attach tmux, as it's left running.

### sequence.json - scripts - scripts with OP_ commands 
In sequence.json scripts there is :
```console
    "env-nginx-simple-t": [
        "{{OP_INFO}} execute env on each pod and put to pod.env file, terminate tmux",
        "kubectl -n {{k8s_namespace}} exec {{k8s_pod}} -c nginx  -- env > {{k8s_pod}}.env",
        "cat {{k8s_pod}}.env",
        "{{OP_TERMINATE}}"
    ],
```
If you run this one, OP_TERMINATE will instruct tks to terminate tmux session, you will get only env files.

### sequence.json - list available scripts
You can check a list of available scripts within sequence.json like this
```console
kubectl tks list scripts
...
...

```
you will get list of scripts with their info , scripts that are available in sequence.json file

If you want to use some other sequence.json you can always use -f other_sequence_file.json

Do 
```console
kubectl tks list
```
for more details on what can be listed.

Currently we know for following OP_ comands:
OP_INFO line is used as information printed when tks list scripts is called
OP_TERMINATE terminates tmux session
OP_ATTACH attached to tmux session
Check also answer below for question Q: What other OP_ commands are available ? 

### sequence.json - shortcuts
If you get tired of writing full kubectl line every time you can use shortcuts.

If there are parts of scripts that you use that are frequently repeated instead of
```console
"kubectl -n {{k8s_namespace}} exec {{k8s_pod}} -c nginx  -- env > {{k8s_pod}}.env"
```
you can define shortcut
```console
"XKNE" : "kubectl -n {{k8s_namespace}} exec "
```consle
and use
```console
"{{XKNE}} -c nginx -- env > {{k8s_pod}}}.env"
```
and 

### sequence - how template fields are replacetd and in which order

Before executing of each line:
- if line is OP_ command it handled from case to case
- otherwise command is for execution so :
    * first shortcuts template fields are being resolved, 
    * then k8s_ template fields are being replaced
    * then podConverter mappings


### sequence.json - shortcuts - available shortcuts
you can list available shortcuts
```console
kubectl tks list shortcuts
$ kubectl tks list shortcuts
K - kubectl
KN - {{K}} -n {{k8s_namespace}}
KNE - {{KN}} exec {{k8s_pod}}
KNEC - {{KNE}} -c {{p2c}}
KNEC- - {{KNEC}} --
KCN - kubectl --context {{k8s_context}} -n {{k8s_namespace}}
KCNE - {{KCN}} exec {{k8s_pod}}
KCNEC - {{KCNE}} -c {{p2c}}
KCNEC- - {{KCNE}} -c {{p2c}} --
KCTL - kubectl --context {{k8s_context}} -n {{k8s_namespace}}
KCTL_EXEC - {{KCTL}} exec {{k8s_pod}} -c {{p2c}} --
KCTL_EXEC_IT - {{KCTL}} exec -it {{k8s_pod}} -c {{p2c}} --
KCTL_EXEC_IT_BASH - {{KCTL}} exec -it {{k8s_pod}} -c {{p2c}} -- /bin/bash -c
KCTL_LOGS_SWITCHES -  --prefix --timestamps --max-log-requests 100
KCTL_LOGS1 - {{KCTL}} logs -f {{KCTL_LOGS_SWITCHES}} {{k8s_pod}} -c {{p2c}}
KCTL_LOGS2 - {{KCTL}} logs -f {{KCTL_LOGS_SWITCHES}} {{k8s_pod}} -c {{p2cLog}}
KCTL_EXEC_BASH - {{KCTL}} exec {{k8s_pod}} -c {{p2c}} -- /bin/bash -c
```

## sequence.json - podConverter intro
While running tks in various cases you might have to run same script (list of commands) on various pods.
Some pods might have different pod name (pod that you are interested to jump and execute something in),
some might have different container name, or it might have differetn shell, for example not /bin/sh but /bin/bash.

To make script reusable on various kinds pods there is podConverter section of sequence.json
Below is a section that describe rules that would be used for mapping pod name to container name
```console
"podConverter" : {
    "p2c" : {
        "busybox" : "busybox.*",
        "nginx" : "nginx.*",
        "main" : ".*"
        },
```
This rule say: there is a pod to container mapping named p2c, and rules are:
- if podname match regexp "busybox.*" then podname should be busybox
- if podname match regexp "nginx.*" then podname should be nginx
- in any other case (".*" matches all) pod name should be main

now you can use {{p2c}} field template to call same script on both busybox and nginx kind of pods.
```console
"kubectl -n {{k8s_namespace}} exec {{k8s_pod}} -c {{p2c}}  -- env > {{k8s_pod}}.env"
```
Or you can use shortcuts and just focus on what you need to be done on pod exec side
```console
"{{KNEC-}} env > {{k8s_pod}}}.env"
```
Keep in mind that "env > {{k8s_pod}}.env" is being executed on local machine, as env is executed on
kubernetes pods side, but redirect is done on local machine.
If you like to have env copy on pod you should do 
```console
"{{KNEC-}} /bin/bash -c 'env > {{k8s_pod}}}.env'"
```
ie 
```console
"kubectl -n {{k8s_namespace}} exec {{k8s_pod}} -c {{p2c}}  -- /bin/sh `env > {{k8s_pod}}.env`"
```

### sequence.json - podConverter - adding second rule
In some cases containers that do print logs are not the same container so you can specify new podConverter section
```console
   "p2cLogs" : {
       "busybox" : "logs",
       "main" : ".*"
       }
```
and you can use {{p2cLogs}} field template in your scripts.


### sequence.json - podConverters in oneLiners
Now that you have sequences.json you can use podConverters in your OneLiners
```console
$ kubectl tks -n test-run start -l app=nginx  "kubectl -n {{k8s_namespace}} exec -t {{k8s_pod}} -c {{p2c}} -- env" -d
# No script kubectl -n {{k8s_namespace}} exec -t {{k8s_pod}} -c {{p2c}} -- env in sequence file /Users/none/.tks/sequences.json, assuming oneLiner
# No matching script kubectl -n {{k8s_namespace}} exec -t {{k8s_pod}} -c {{p2c}} -- env in conf file
nginx-sample1-6475dd48b7-bgtsx 0 0 kubectl -n test-run exec -t nginx-sample1-6475dd48b7-bgtsx -c nginx -- env
nginx-sample1-6475dd48b7-br6jc 1 0 kubectl -n test-run exec -t nginx-sample1-6475dd48b7-br6jc -c nginx -- env
nginx-sample1-6475dd48b7-n6mtg 2 0 kubectl -n test-run exec -t nginx-sample1-6475dd48b7-n6mtg -c nginx -- env
```

# advanced topics
## executing within kubectl exec on pod

Once tks tool is started, new session tmux session is created. Then for each pod new tmux window (and pane) is created.
First thing after creating windows for each pod is fetching prompt line. Prompt will be used to confirm that
previous command has been executed and that next command can be executed. 

After every command execution tks tool will wait for prompt to appear, so that it can execute next command.
There is a command that can canel this behaviour. After each command there is look ahead, if next step is
OP_NO_PROMPT_WAIT, tks will not wait for prompt to return, it will start executing next command as soon as it can.

There are cases where prompt changes, for example if you  kubectl exec to remote pod you will get prompt from pod.
Usually that prompt is pod name, but it can be anything. In such cases, in order to continue, and not to wait for prompt that was initially loaded (local host prompt) there is OP_REFRESH_PROMPT, this command will read prompt again.
For example:
```console
    "remote-exec": [
        "kubectl --context {{k8s_context}} -n {{k8s_namespace}} exec -it {{k8s_pod}} -c {{p2c}} -- /bin/bash",
        "{{OP_NO_PROMPT_WAIT}}",
        "{{OP_REFRESH_PROMPT}}",
        "sleep 20",
        "sleep 10",
        "hostname",
        "exit",
        "{{OP_NO_PROMPT_WAIT}}",
        "{{OP_REFRESH_PROMPT}}",
        "hostname"
    ]
```
this example will execute kubectl, but will not wait for prompt, it will load new prompt, then will execute remote (kubectl pod/container) sleep 20 (will wait for prompt), then sleep 10 (will wait for prompt), then it will exit without waiting for prompt, then it will load prompt again (as after exit we are back in local shell),  

There are cases where you might benefit from some sleep time, so there is OP_SLEEP, this command have an argument
```console
"{{OP_SLEEP}} 5"
```
this will sleep for 5 seconds - it will sleep on tks/tmux side.

OP_SLEEP could be used if you access remote node that can have slow response time, so refreshing prompt won't help,
but sleeping few seconds might help.

If you want to comment some operation there is OP_COMMENT, everything after this OP_COMMENT will be rendered and printed out. For example:
```console
"{{OP_COMMENT}} this command will do something on {{k8s_pod}}"
```
will print something like
```console
"#COMMENT: this command will do something on nginx-sample1-6475dd48b7-bgtsx"
```
As mentioned above if you use OP_TERMINATE at last execution step tmux session will be terminated (along with all shell terminals). OP_ATTACH will attach to tmux, so you can inspect manually in tmux.

OP_FINALLY is used to do some final (like aggregation or processing) job, once, on tks side. For example:
```console
   "env-nginx-simple": [
        "kubectl -n {{k8s_namespace}} exec {{k8s_pod}} -c nginx  -- env > {{k8s_pod}}.env",
        "{{OP_FINALLY}} tar czvf nginx.tgz *.env ; rm *.env"
    ],
```
This script will get environment from all needed pods, then finally it will make archive and remove env files.

OP_FINALLY, OP_TERMINATE, OP_ATTACH are end script commands. No more steps will be executed after those instructions.


# FAQ
 
## Q: What kubernetes arguments can be used ?
A: k8s_config, k8s_context, k8s_namespace, k8s_pod
```console
$ kubectl tks list kctl
Kubectl params:
 k8s_config
 k8s_context
 k8s_namespace
 k8s_pod
```


## Q: What other OP_ commands are available ?
A: 
```console
$ kubectl tks list control
Controls:
 OP_TERMINATE - Terminate tmux, script end
 OP_ATTACH - Attach tmux, script end
 OP_DETACH - Detach tmux, script end, default behavior
 OP_FINALLY - Finally execute, script end
 OP_EXECUTE - Execute line
 OP_INFO - Print info
 OP_COMMENT - Print comment, render
 OP_NO_PROMPT_WAIT - Do not wait for prompt for last command
 OP_SLEEP - Sleep for n seconds
 OP_REFRESH_PROMPT - Load new prompt
```

## Q: How to use same execution line for different pods container names, ie when -c container_name is not same ?
A: use podConverter section of sequences.json file (~/.tks/sequences.json)
```console
"podConverter" : {
    "p2c" : {
        "busybox" : "busybox.*",
        "nginx" : "nginx.*",
        "main" : ".*"
        },
```

then use {{p2c}} in kubectl command line.
You can have more than one item, so if you need different pod name to container mapper define
different set of rules.

