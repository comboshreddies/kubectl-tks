##   Tmux Kubectl Scripts - tks
# tks: kubectl plugin - or standalone tool
## A plugin for executing scripts on pods within tmux windows:

Tks plugin runs multiple execution scripts (sequences) in multiple tmux windows.
Each window runs a script on one pod. You can decide if you want to attach and inspect
executions per each pod, or you want to just execute and exit tmux.

As a good practice you should not run too much CLI exec things on pods/containers, but if you do,
and if you do frequently and on many different pods namespaces, clusters, then
this tool might be helpful.

With this tool you can:
1) run oneliners and (more powerful) scripts with template engine for kubernetes
```console
"exec {{pod}} -c {{p2c}}  -- /bin/sh `env > {{k8s_pod}}.env`"
```
tks will automatically change pod name and find correct p2c mapping for container for each execution step

2) you can select by namespace, label, then filter specific pods, like:
```console
kubectl tks --context minikube -n test-run start tcpdump-all -l app=nginx
kubectl tks --context minikube -n test-run start tcpdump-all -p busybox1,busybox2
kubectl tks --context minikube -n test-run start tcpdump-all -l app=busybox -p busybox1,busybox2
```

3) You can decide would you like to terminate, attach tmux or do some command at the and of execution

4) You can run scripts in:
* synchronous mode - so each script line step is done on all pods, and once all are complete (prompt is returned) tks will go to next step, add -s switch on start command
* asynchronous mode (default) - each pod runs its own set of sequences (no -s switch on start command)
 
With this tool you can attach tmux and take a look, or you can gather results of executions.
If you have some actions that you frequently do, like gathering some info from pods, you
can add a sequence that fits your needs, and next time you can run it fast. Tmux solves
problem of having multiple outputs of executions on same terminal, as each tmux screen/window
is dedicated to one pod execution, so you have terminal for each pod, do ctrl+b+n and
go to the next tmux-pod window.


## Installation

1) go build
2) copy kubectl-tks to your bin path (/usr/local/bin for example)

Additional step, not needed but useful, can be done later
3) mkdir  ~/.tks
4) copy sequences.json to ~/.tks/


# purpose
One can easily make a shell script to execute some command on each pod/container in parallel, then gather results.
Executing on more than a few pods might fail on some pods. Having output of many pod executions on one screen is not practical. You might need to take a look at each execution on your own, one pod at the time. You might want to keep a terminal session for each pod that you've connected to. Using tmux could help.


# intro

To practice first steps create a deployment

```console
kubectl create ns test-run
kubectl -n test-run apply -f k8s_yamls/sample_deploy1.yaml
```

then verify is everything up
```console
kubectl -n test-run get pod
kubectl -n test-run get pod -l app=nginx
```
![s_00_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_00_tks.svg?raw=true)


## One-liner example


One-liner example is nice way to start but real power comes with scripts that are tailored for purpose.

One-liners are designed to be short, but they might not be easily readable. Power
of one-liners come with prepared scripts and shortcuts, but will start with empty state
(so no ~/.tks/sequences.json for a start, we will add those later)
here is simple one liner that executes env for each pod in nginx container
We will use start command, and pods will be selected with -l app=nginx

```console
kubectl tks -n test-run start -l app=nginx  "_ exec {{pod}} -c nginx -- env"
```
![s_01_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_01_tks.svg?raw=true)

Attaching to tmux shows execution results for each pod - each pod one tmux window/pane.

Explanation:

Command start will try to find a script within ~/.tks/sequences.json
as there is no script named "_ exec {{pod}} -c nginx -- env", it
will assume argument is not a script name but one-liner

Underscore (ie _) shortcut will be replaced with same kubectl kuberntes parameters as tks was called out
(--kubeconfig,--context, --namespace ie -n), so in this case _ will be replaced with:
"kubectl -n test-run"

For each pod new tmux window will be created, and script:

kubectl -n test-run exec {{pod}} -c nginx -- env

will be executed for each pod, replacing {{pod}} with specific pod name

New tmux session OneLiner--test-run is created with base window, then new window is created for each pod.
Underscore (ie _ ) sign is used to repeat same --context --namespace --kubeconfig parameters as tks was called out,
so in this case "_ exec {{k8s_pod}} -c nginx -- env"  _ is replaced with kubectl -n test-run.


## one-liner leaves tmux session open

Second run of same one-liner will try to open same session name and (if you have not terminated
previous tmux session) it will fail with error. One-liners might have overlapping session-names and you 
can't open tmux with same session name.

Tks tool creates tmux session based on kubectl context, kubectl namespace, and script name.
One liners are considered as unnamed script names, and for one-liners script name is always "OneLiner".

Let's try to run another one-liner on same namespace and context (no explicit context) and check the error:
```console
kubectl tks -n test-run start -l app=nginx  "_ exec {{k8s_pod}} -c nginx -- env"
```
![s_02_1_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_02_1_tks.svg?raw=true)
```
# Unable to read conf file /Users/none/.tks/sequences.json, assuming oneLiner
# unable to open sequence json file /Users/none/.tks/sequences.json
# there is already session with this name (OneLiner--test-run), exiting
```
you can use
```console
tmux ls
```
to check what sessions are running.

If you want to use different tmux session name, use start command with -S <your session name>.
This way you can explicitly name your tmux sessions names, so they won't overlap.


## one-liner with previous session removal

You can always remove previous session with tmux kill-session .
If you want to start and if same tmux session name is present remove previous session, use the -T option.

Flag -T will terminate the previous session of the same name as one that is starting . Session Names are
generated by concatenating sequence-name (for one-liners it's OneLiner), kubernetes context (if specified) and
kubernetes namespace (if specified). 

Default behaviour of the tks plugin is to leave the tmux session in a detached state,
so you can inspect results of execution.

```console
kubectl tks -n test-run start -l app=nginx  "_ exec -t {{k8s_pod}} -c nginx -- env" -T
```

![s_02_2_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_02_2_tks.svg?raw=true)
Below you will see how to auto terminate session within script/one-liner.
You might have something valuable within tmux and we do not want to
delete previous session as default behaviour.

## list of available kubernets related template fields

You can always check what kuberntes template variable fields are available with
```console
kubectl tks list kctl
```
![s_07_1_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_07_1_tks.svg?raw=true)


## one-liner with more kubernetes related templated fields

Tks used k8s_pod as template variable field for each pod that was there in selected set of pods (via label or -p ).
We can use k8s_namespace template field to specify namespace (so you don't have to repeat test-run).
There are also k8s_config and k8s_context template fields that would be filled if specified within tks
command line parameters - if they are not specified they will be empty strings.

here is example of extened k8s variable names
```console
kubectl tks -n test-run start -l app=nginx  "kubectl -n {{k8s_namespace}} exec -t {{k8s_pod}} -c nginx -- env" -T
```
![s_07_2_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_07_2_tks.svg?raw=true)

There are shorcuts for kuberenetes related templated fields (as shown with tks list kctl above).
Those are: cnf for kubeconfig, ctx for context , nsp for namespace, and pod for pods.

here is example of short k8s variable names, they work the same
```console
kubectl tks -n test-run start -l app=nginx  "kubectl -n {{nsp}} exec -t {{pod}} -c nginx -- env" -T
```
![s_07_3_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_07_3_tks.svg?raw=true)

It's simpler to use _ instead of kubectl --context {{k8s_context}} -n {{k8s_namespace}} exec -t {{k8s_pod}}
but there are cases where you might need to more explicit. 

here is an example of _ start line shortcut with shortened {{pod}} kctl variable:
```console
kubectl tks -n test-run start -l app=nginx  "_ exec {{pod}} -c nginx -- env ; echo {{pod}}" -T -d
```
![s_07_4_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_07_4_tks.svg?raw=true)

here is an example that shows that _ will act differently if more kubectl arguments (in this case context) are passed:
```console
kubectl tks --context minikube -n test-run start -l app=nginx  "_ exec {{pod}} -c nginx -- env ; echo {{pod}}" -T -d
```
![s_07_5_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_07_5_tks.svg?raw=true)


## one-liner with dry run mode

If you want to show what will be executed and not to execute for real there is -d flag

```console
kubectl tks -n test-run start -l app=nginx  "_ exec {{k8s_pod}} -c nginx -- env" -d
```
![s_03_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_03_tks.svg?raw=true)

Tks in dryRun mode will render all known template variables types
first shortuts, then internal _ (as _ can be used in shortcuts too), then controls and
then run time available kubernetes/kctl - it will contact kubernetes and ask for pods,
to be able to fill them in desired command line.


## one-liner with more than one executions

You can specify more than one command in One-Liner execution. Commands are separated by ';' sign.
```console
kubectl tks -n test-run start -l app=nginx  "_ exec {{k8s_pod}} -c nginx -- env;echo {{k8s_pod}}" -T
```
![s_04_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_04_tks.svg?raw=true)

You can attach to tmux to inspect execution, and switch between tmux terminal windows.


## tks tmux control operations

To be able to better manage tmux session there are specific control operations that are built in in tks.
You can check them with

```console
kubectl tks list control
```
![s_05_1_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_05_1_tks.svg?raw=true)

We will cover just few most important below.


## one-liner with comment control instruction

Comment control instruction is used by specifying {{OP_COMMENT}}
All content left of OP_COMMENT tag will be rendered with template variables (shortcuts, kctl or podMap),
and will be displayed back to tmux screen.

```console
kubectl tks -n test-run start -l app=nginx  "_ exec {{pod}} -c nginx -- env ;{{OP_COMMENT}} doing env on pod {{pod}}" -T
```
![s_05_2_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_05_2_tks.svg?raw=true)


## one-liner with terminating tmux session control instruction 

If you do not want to manually terminate tmux session every time you run script or oneliner
you can add OP_TERMINATE. Tks will terminate tks tmux session (closing all windows) once all pods
have all script steps (commands) executed. OP_TERMINATE is final command, no other commands will be
processes after.

```console
kubectl tks -n test-run start -l app=nginx  "_ exec {{pod}} -c nginx -- env ;{{OP_TERMINATE}}" -T
```
![s_05_3_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_05_3_tks.svg?raw=true)

Now at the end of execution tks will terminate session, so next time you might not need -T.


## one-liner with attaching to tmux session

You can use {{OP_ATTACH}} to attach to tmux session (windows) at the end of script.
```console
kubectl tks -n test-run start -l app=nginx  "_ exec {{pod}} -c nginx -- env ;{{OP_ATTACH}}" -T
```
![s_05_4_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_05_4_tks.svg?raw=true)

OP_ATTACH is also final command, no other commands will be executed after this instruction is reached.
With attach you can get overview of what has been executed or continue to execute commands in terminals.


## tks supports sync and async mode

By default, with no additional switches tks will work in async mode.
If you add -s switch it will run in sync mode.

In async each script per pod will run separately at their own pace
In async mode some pod-scripts might have completed while others are still progressing
In sync mode each step is executed on all pods, prompt line (that confirms command returned) 
is waited for all pods, then next step is executed.

In sync mode all pod should execute first command, then second can be executed.
Before running your script you can check with dry mode how it will behave with sync or async mode.
```console
tks -n test-run start -l app=nginx "echo ABCD {{pod}}; echo XYZW {{pod}}" -T -d -s
```
![s_06_1_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_06_1_tks.svg?raw=true)
Dry run sync mode tries to show that first step will be executed on all pods,
then second (so there is grouping with /|\ characters)

In async mode each pod runs script instructions indenpendently.
Same as previous example but without -s (so in async mode) will look like
```console
tks -n test-run start -l app=nginx "echo ABCD {{pod}}; echo XYZW {{pod}}" -T -d
```
![s_06_2_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_06_2_tks.svg?raw=true)
Dry run in async mode shows grouping by pod, but each pod execution will run in parallel.

here is example of running in sync mode (so no dry run)
```console
tks -n test-run start -l app=nginx 'X=$(($RANDOM % 20));sleep $X;echo ABCD;sleep $X;echo XYZW {{pod}}' -q -T -s
```
![s_06_2_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_06_2_tks.svg?raw=true)
and same but in async (default) mode
```console
tks -n test-run start -l app=nginx 'X=$(($RANDOM % 20));sleep $X;echo ABCD;sleep $X;echo XYZW {{pod}}' -q -T 
```
![s_06_2_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_06_2_tks.svg?raw=true)
Output of async mode shows nature of async, some pods complete last line before others, there is no strict ordering.
If some pod execution is slow or blocked that won't stop other pod executions.

Keep in mind that if for any reason (for example pod termination/deletion) one pod does not return prompt,
no further steps will be executed. In sync mode no more steps will be exectued on any pod. In async mode
only pod that that didn't return prompt will be blocked, others will continue running, as they run separately.

In both cases tks will keep hanging on - so you might check what went wrong with tmux attach.


## tks list scripts and get more info

### tks and tks list

kubectl-tks can be used independently as tks binary (just link or copy). tks is a kubectl plugin
but plugins are standalone applications to, so instead of kubectl tks I've used tks in examples below

here is an output of tks list
```console
tks list
```
![s_10_1_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_10_1_tks.svg?raw=true)

### tks list kctl 

We've already seen this one, it will show tks available kubectl params used for template fields.
Those kctl params are internal to tks.

let's repeat
```console
tks list kctl 
```
![s_10_2_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_10_2_tks.svg?raw=true)


### tks list control

To see what integrated control operations are available in tks you can run.

```console
tks list control
```
![s_10_3_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_10_3_tks.svg?raw=true)

OP_TERMINATE, OP_ATTACH and OP_FINALLY are terminal commands, no other instructions will be processed after this
command is reached.
OP_INFO is used as short description for scripts, if OP_INFO is specified as first step in script sequence it will
be used for help line. See below for tks list scripts. Scripts without OP_INFO on first line will not show help.

### tks scripts

Config file sequences.json contains section for scripts. Those scripts can be listed with following:
```console
tks list scripts
```
![s_10_4_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_10_4_tks.svg?raw=true)

This command shows list of available scripts than one can run (start command) with tks.

You can copy or modify this sequences.json file, but keep in mind that if json format is broken tks will not be able
to load content from it.

More info about each script can be found out by tks info <script name>, will be explained below.

If you want to use some other sequence.json config file you can always use -f other_sequence_file.json 
For example one can have separate sequence.json for apt-based containers, and separate for yum-based containers,
and separate for apk, but you can keep them in one file using podMap - more about that below.


### tks shortcuts

Shortcuts are also defined in sequences.json . Shortcuts are here to help you write most frequently repeated parts
of script instructions.

```console
tks list shortcuts
```

![s_10_5_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_10_5_tks.svg?raw=true)

In execution shortcuts are resolved multiple times, so in one shortcut you can refer to other shortcut.


### tks podMap

PodMap section of sequences.json file contains set of regexp rules executed on pod-name ( ie {{pod}}), and if
regexp matches it returns value on left.

Let's see details:
```console
tks list podMap
```
![s_10_6_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_10_6_tks.svg?raw=true)

We will learn more about scripts, shortcuts, podMaps below in that goes in details on sequence.json content .


## using sequence file for storing complex scripts

If you are tired of creating One-Liners you have to remember, you could copy sequence.json to ~/.tks directory (as mentioned in installation step 3 and 4 above) and start using existing or write your own by editing sequences.json file.


## sequence.json contents

sequence.json file contains 3 sections:
- scripts
- shortcuts
- podMap

We've seen those sections with tks list examples, above.
(tks list can show two more items (kctl,control) that are not within sequences.json file, those are integrated)

### sequence.json - scripts section 

Scripts section defines script name and list of actions.
For example:
```jsonl
"scripts" : {
    "env-nginx-simple": [
        "{{OP_INFO}} execute env on each pod and put to pod.env file",
        "kubectl -n {{k8s_namespace} exec {{k8s_pod}} -c nginx  -- env > {{k8s_pod}}.env",
        "cat {{k8s_pod}}.env"
    ],
```
Scripts are working same way as OneLiners but you can call them by name (env-nginx-simple) instead of typing long OneLiners. Each line is new command that will be executed in one step of a script. OP_INFO is help tag - it is shown in tks list scripts as short hlep and explanation of a script.

Important difference between one-liners and scripts:

In scripts you can use ; sign as separator but it will not check control ({{OP_}}) instructions there, it will be considered as regular shell command terminator/separator. Note that {{OP_}} operators and _ will be resolved or checked only
on the begining of script line.
 
In one-liners usage of ; will split one line to multiple commands, so in one-liners tks will frist split one-line
to multiple lines by ; character, then consider each line as regular script line, and thus process {{OP_}} or _
at the beginning (ie ;_ or ;{{OP_ATTACH}} will be processed).

Currently there is no way to escape ; in one-liners ie every ; is considered as a place to split one line to
multiple command lines. If you need more lines with ; character, you should write scripts, not one-liners.


### sequence.json - more info 
Instead of looking with editor or jq in sequences.json you can get details of a script with
```console
tks info env-nginx-simple
```
![s_08_2_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_08_2_tks.svg?raw=true)

Also you can try to expand shortcuts in script (but this env-nginx-simple does not have shortcuts)
```console
tks info env-nginx-simple -x
```
![s_08_2_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_08_2_tks.svg?raw=true)
Info can not expand kctl (k8s_) parameters as those are available run time (same goes for podMaps), 
but it will expand shorcuts.

To check how this cript will look at execution time we can try dry run:
```console
kubectl tks -n test-run start env-nginx-simple -l app=nginx -T -d
```
![s_09_4_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_09_4_tks.svg?raw=true)

and then we can run the script
```console
kubectl tks -n test-run start env-nginx-simple -l app=nginx -T
```
![s_09_5_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_09_5_tks.svg?raw=true)

You can check files in local directory, there should be nginx-sample1-*.env files for each pod
Also you can attach tmux, as it's left running.


### sequence.json - scripts - scripts with OP_ commands 

In sequence.json scripts there is :
```jsonl
    "env-nginx-simple-t": [
        "{{OP_INFO}} execute env on each pod and put to pod.env file, terminate tmux",
        "kubectl -n {{k8s_namespace}} exec {{k8s_pod}} -c nginx  -- env > {{k8s_pod}}.env",
        "cat {{k8s_pod}}.env",
        "{{OP_TERMINATE}}"
    ],
```
If you run this one, OP_TERMINATE will instruct tks to terminate the tmux session, you will get only env files.
If you have reliable set of commands that you're not interested to inspect, attach, and you're only interested 
in results scripts provide (usually you can store output in local files, or copy files) then use OP_TERMINATE.

If you want to use some other sequence.json config file you can always use -f other_sequence_file.json 
For example I do have separate sequence.json for apt-based containers, and separate for yum-based containers.

Feel free to adjust and modify sripts section to fit your needs.


### sequence.json - shortcuts

If you get tired of writing a full kubectl line every time you can use shortcuts.

If there are parts of scripts that you use frequently, instead of writing same every time
```console
"kubectl -n {{k8s_namespace}} exec {{k8s_pod}} -c nginx  -- env > {{k8s_pod}}.env"
```
you can define shortcut, in shortcuts section of json sequence.json file
```jsonl
"XKNE" : "kubectl -n {{k8s_namespace}} exec "
```
and following as a line for execution one-liner or a script
```console
"{{XKNE}} -c nginx -- env > {{k8s_pod}}}.env"
```
explanation of naming: eXample Kubectl with Namespace Exec - shortcut XKNE

here is an example of expanding (via dry run) one-liner that uses shortcut ECB from sequences.json
```console
kubectl tks -n test-run start -l app=nginx "{{ECB}} date" -T -d
```
![s_09_5_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_09_5_tks.svg?raw=true)

Shortcuts are making one-liners (and scripts) more expressive, fast to type. Let's try to get in details
of how ECV resolution is being done:
* ECB is shortcut for {{EC}} /bin/bash -c
* EC is shortcut for _ exec {{pod}} -c {{p2c}} --
* _ is dynamic shortcut (tks implemented) that repeats kubectl options (context, namespace, config)
* p2c is podMap that maps every podname like 'nginx.*' to 'nginx', that is how from {{p2c}} tks got nginx


example for specific pod nginx-sample1-59d677c5cb-flpft:
kubectl -n test-run exec nginx-sample1-59d677c5cb-flpft -c nginx -- /bin/bash -c  date

explanation of naming: 
EC - Exec to Container
ECB - Exec to Conatiner Bash
Excessive shortened naming are used so one-liners and script lines look short 
and fit for this documentation. Use more reasonable ones for your purpose.
Name p2c for the podMap is shortcut for pod to container.

### sequence - how template fields are replaced and in which order

Before executing of each line:
- if line is OP_ command it handled from case to case
- otherwise command is for execution so :
    * first shortcuts template fields are being resolved, 
    * then strating _ is resolved
    * then k8s_ template fields are being replaced
    * then podMap mappings


### sequence.json - shortcuts - 

We've seen list of shortcuts above, tks list shortcuts. 
To expand only shortcuts for scripts use info -x.
To expand shortcuts use one-liners in dry-run mode


## sequence.json - podMap details

While running tks in some cases you might have to run the same script (sequence of commands) on various pods.
Some pods might have different pod name (pod that you are interested to jump and execute something in),
some might have different container name, or it might have different shell, for example not /bin/sh but /bin/bash.

To make script reusable on various kinds pods there is podMap section of sequence.json
Below is a section that describe rules that would be used for mapping pod name to container name
```jsonl
"podMap" : {
    "p2c" : [ 
        { "busybox" : "busybox.*"},
        {"nginx" : "nginx.*"},
        {"main" : ".*"}
        ],
```

This rule say: there is a pod to container mapping named p2c, and rules are:
- if the podname match regex "busybox.*" then returned value should be busybox
- if the podname match regex "nginx.*" then returned value should be nginx
- in any other case (".*" matches all) returned value should be main
explained naming: p2c stands for pod to container as it is used to map podname to containers

Now you can use {{p2c}} field template to call the same script on both busybox and nginx pods.
```console
"_ exec {{k8s_pod}} -c {{p2c}}  -- env > {{k8s_pod}}.env"
```
Or you can use shortcuts that uses {{p2c}} and just focus on what you need to be done on pod exec side
```console
"{{EC}} env > {{k8s_pod}}}.env"
```
EC is resolved/expanded to '_ exec {{pod}} -c {{p2c}} --'

Keep in mind that "env > {{k8s_pod}}.env" is being executed on local machine, as env is executed on
kubernetes pods side, but redirects are done on local machine.
If you like to have env output recorded to pod local file you should do
```console
"{{EC}} /bin/bash -c 'env > {{k8s_pod}}}.env'"
``` 
ie
```console
"kubectl -n {{k8s_namespace}} exec {{k8s_pod}} -c {{p2c}}  -- /bin/sh 'env > {{k8s_pod}}.env'"
```


### sequence.json - podMap - adding second rule

In some cases containers that do print logs are not the same container as one doing main job so you can specify new podMap section:
```jsonl
   "p2cLogs" : [ 
       {"logger" : "busybox.*"},
       {"main" : ".*"}
       ]
```
and you can use {{p2cLogs}} field template in your scripts.
Here we assume that those pods that match podname 'busybox.*' do have separate container named logger that
is used to produce logs.


### sequence.json - podMaps in oneLiners

You can't manage podMaps from cli, you have to rely on sequences.json
Once you have sequences.json you can use podMaps in your OneLiners

```console
kubectl tks -n test-run start -l app=nginx  "_ exec -t {{pod}} -c {{p2c}} -- env" -d
```
![s_15_1_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_15_1_tks.svg?raw=true)

In FAQ section you can find example of mapping pod-name to shell (bash or sh), mapping pod name to run apt or apk or yum,
or mapping pod name to package name (as different linux distributions might have different name for same package).

podMaps are great way to cover more cases in your shortcuts and scripts, so same scripts can run in various cases.

# advanced topics

## executing within kubectl exec on pod - remote-exec script 
Once the tks tool is started, a new tmux session is created. Then for each pod a new tmux window (and pane) is created.
First thing after creating windows for each pod is fetching the prompt line. Prompt will be used to confirm that
the previous command has been executed.

After every command execution tks tool will wait for a prompt to appear, so that it can execute the next command.
There is a command that can cancel this behaviour. After each exec command there is look ahead, if next step is
OP_NO_PROMPT_WAIT, tks will not wait for prompt to return, and tks will start executing (sending) the next command as soon as it can.

There are cases where prompt changes, for example if you kubectl exec to remote pod (or ssh to remote host) you will get prompt from pod.

Usually that prompt is a pod name, but it can be anything. In such cases, waiting for old prompt line would not make progress, so in order to continue, and not to wait for a prompt that was initially loaded (local host prompt) there is OP_REFRESH_PROMPT, this command will read the prompt again.

For example:

```jsonl
"remote-exec": [
  "{{OP_INFO}} example of attaching to remote host with different prompt",
  "{{EICBS}}",
  "{{OP_NO_PROMPT_WAIT}}",
  "{{OP_REFRESH_PROMPT}}",
  "date",
  "hostname",
  "exit",
  "{{OP_NO_PROMPT_WAIT}}",
  "{{OP_REFRESH_PROMPT}}",
  "hostname"
]
```
this example will execute kubectl, but will not wait for prompt, it will load new prompt, then will execute remote (kubectl pod/container) date (will wait for prompt), then will execute hostname (will wait for prompt), then it will exit without waiting for prompt, then it will load prompt again (as after exit we are back in local shell), and execute hostname
on local host (one that we run tks on)

```console
kubectl tks -n test-run start remote-exec -l app=nginx -q
```
![s_14_2_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_14_2_tks.svg?raw=true)


## capturing files on pods then copy then to local

Another nice script in sequences.json is tcpdump-all

here is dry run
```console
tks -n test-run start -l app=nginx tcpdump-all -d
```

and here is execution
```console
kubectl tks -n test-run start tcpdump-all -l app=nginx -q
```
![s_14_1_tks.svg](https://github.com/comboshreddies/kubectl-tks/blob/main/scripts/printouts/recorded/s_14_1_tks.svg?raw=true)


## more info on control operations

There are cases where you might benefit from some sleep time.
Tks supports OP_SLEEP, this command have an argument
```console
"{{OP_SLEEP}} 5"
```
this will sleep for 5 seconds - it will sleep on the tks/tmux side.

OP_SLEEP could be used if you access remote node that can have slow response time, so refreshing prompt won't help, but sleeping a few seconds might help.


If you want to comment some operation there is OP_COMMENT, everything after this OP_COMMENT will be rendered and printed out. For example:
```console
"{{OP_COMMENT}} this command will do something on {{k8s_pod}}"
```
will print something like
```console
"#COMMENT: this command will do something on nginx-sample1-6475dd48b7-bgtsx"
```
Comment can be used as a verification of what is rendered/executed in a run time.


As mentioned above if you use OP_TERMINATE after the last execution step tmux session will be terminated (along with all shell terminals). OP_ATTACH will attach to tmux, so you can inspect manually in tmux.

OP_FINALLY is used to do some final (like aggregation or processing) job, once, on tks side. For example:
```console
   "env-nginx-simple": [
        "kubectl -n {{k8s_namespace}} exec {{k8s_pod}} -c nginx  -- env > {{k8s_pod}}.env",
        "{{OP_FINALLY}} tar czvf nginx.tgz *.env ; rm *.env"
    ],
```
This script will get the environment from all needed pods, then finally it will archive and remove env files.

OP_FINALLY, OP_TERMINATE, OP_ATTACH are terminating script commands. No more steps will be executed after those instructions.


# FAQ
 
## Q: What kubernetes arguments can be used ?
A: k8s_config, k8s_context, k8s_namespace, k8s_pod
```console
kubectl tks list kctl
```
```
Kubectl params:
 k8s_config or short cnf
 k8s_context or short ctx
 k8s_namespace or short nsp
 k8s_pod or short pod
```

## Q: What other OP_ commands are available ?
A: 
```console
kubectl tks list control
```
```
Controls:
 OP_TERMINATE - _T - Terminate tmux, script end
 OP_ATTACH - _A - Attach tmux, script end
 OP_DETACH - _D - Detach tmux, script end, default behavior
 OP_FINALLY - _F - Finally execute, script end
 OP_EXECUTE - _E - Execute line, no need to specify, default behaviour
 OP_INFO - _I - Print info
 OP_COMMENT - _C - Print comment, render
 OP_NO_PROMPT_WAIT - _N - Do not wait for prompt for last command
 OP_SLEEP - _S - Sleep for n seconds
 OP_REFRESH_PROMPT - _R - Load new prompt
```

## Q: How to use the same execution line for different pod container names, i.e. when -c container_name is not the same ?
A: use podMap section of sequences.json file (~/.tks/sequences.json)

```jsonl
"podMap" : {
    "p2c" : [ 
        {"busybox" : "busybox.*"},
        {"nginx" : "nginx.*"},
        {"main" : ".*"}
        ],
```
then use {{p2c}} in the kubectl command line.
You can have more than one podMap item, so if you need different pod name to container mapper define
different set of rules.

here is an example of dry run showing how p2c converts podname to correct container name
```console
kubectl tks -n test-run start "_ exec {{pod}} -c {{p2c}} -- env" -d
```

```
#Sripts loaded from sequence.file
#PodMap loaded from sequence.file
#Shortcuts loaded from sequence.file
# No matching script _ exec {{pod}} -c {{p2c}} -- env in conf file
# assuming oneLiner
busybox1-7f7f64dd8d-6gxqp 0 0 kubectl -n test-run exec busybox1-7f7f64dd8d-6gxqp -c busybox -- env
busybox1-7f7f64dd8d-hxfdn 1 0 kubectl -n test-run exec busybox1-7f7f64dd8d-hxfdn -c busybox -- env
busybox1-7f7f64dd8d-pfmvv 2 0 kubectl -n test-run exec busybox1-7f7f64dd8d-pfmvv -c busybox -- env
nginx-sample1-6475dd48b7-bgtsx 3 0 kubectl -n test-run exec nginx-sample1-6475dd48b7-bgtsx -c nginx -- env
nginx-sample1-6475dd48b7-br6jc 4 0 kubectl -n test-run exec nginx-sample1-6475dd48b7-br6jc -c nginx -- env
nginx-sample1-6475dd48b7-n6mtg 5 0 kubectl -n test-run exec nginx-sample1-6475dd48b7-n6mtg -c nginx -- env
nginx-sample2-5ffd775bc4-5lkpx 6 0 kubectl -n test-run exec nginx-sample2-5ffd775bc4-5lkpx -c nginx -- env
nginx-sample2-5ffd775bc4-9br96 7 0 kubectl -n test-run exec nginx-sample2-5ffd775bc4-9br96 -c nginx -- env
nginx-sample2-5ffd775bc4-tcqn9 8 0 kubectl -n test-run exec nginx-sample2-5ffd775bc4-tcqn9 -c nginx -- env
```
Above shows there are 3 deployments nginx-sample1, nginx-sample2 and busybox1, and -c (container) parameter is
busybox for busybox pods and nginx for nginx pods.

## Q: How to use same script for different types of pods with different shells or different package managers?
A: Create podMap section (for example for ldap2 package that has different names in alpine and debian) like:
```console
'p2inst': [ 
   { "apt install -y" : "debian-pod-name.*"},
   { "apk add" : "alpine-pod-name.*"},
   { "yum install -y" : "centos-pod-name.*"}
],
"p2sh" : [
   {"/bin/sh" : "busybox.*"},
   {"/bin/bash" : ".*"}
],
"p2LdapPack" : [
   {"libldap2-dev" : "debian.*"},
   {"openldap-dev" : "alpine.*"}
],
```
then use in scripts or one-liners like
```console
"kubectl --context {{ctx}} -n {{nsp}} exec {{pod}} -c {{p2c}} -- {{p2sh}} -c '{{p2inst}} {{p2LdapPack}}'"
```
or with shorcuts
```console
"{{EC}} {{p2sh}} -c '{{p2inst}} {{p2Ldapappck}}'"
```
or you can make whole line a single shortcut .
put in shortucts:
```
"INSTALL" : "{{EC}} {{p2sh}} -c '{{p2inst}}",
"INSTALL_LDAP" : "{{INSTALL}} {{p2Ldapappck}}'"
```
and now you can run weather you have yum,dep/apt, or apk distro on container
```console
kubectl tks -n test-run start -l app=nginx "{{INSTALL_LDAP}}
```

# Best practice:
- use OP_INFO as a first line of a script as it will be used as a help line for script

- do always run time limited (timeout) and execution limited commands (like -c in tcpdump or ping), otherwise
your execution might be left running on a pod for a very long time, and affect normal pod state

- if you are running something without limitations, create sequence that could terminate such executions
for example if you are running tcpdump, do create sequence to kill any tcpdump that might be left running

- if you need longer set of sequences, then frequent kubectl exec sequence is not optimal. You have two
options. 
One is to create local script, then copy script to pod/container, run script, copy back results - in 
total 3 kubectl actions: copy script, exec script, copy results. 
Second is to interactivelly exec to kube pod (kubect exec -it ) and then request {{OP_NO_PROMPT_WAIT}} and then {{OP_REFRESH_PROMPT}}. Those commands will instruct tks not to wait for prompt, and to load new prompt line in tks state.
In this way you can keep kubectl interactive session open and continue to send instructions to pod/container.


