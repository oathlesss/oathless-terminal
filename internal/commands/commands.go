package commands

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// Response is the structured output from a command execution.
type Response struct {
	Output string `json:"output"`
	Type   string `json:"type"` // "text", "error", "html", "clear"
}

// Registry maps command names to their handler functions.
type Registry struct {
	commands map[string]func(args []string) Response
}

// New creates a new command registry with all built-in commands.
func New() *Registry {
	r := &Registry{commands: make(map[string]func(args []string) Response)}
	r.registerBuiltins()
	return r
}

// Execute runs a command by name and returns its response.
func (r *Registry) Execute(name string, args []string) Response {
	cmd, ok := r.commands[name]
	if !ok {
		return Response{
			Output: fmt.Sprintf("command not found: %s\nType 'help' for available commands.", name),
			Type:   "error",
		}
	}
	return cmd(args)
}

// Available returns a sorted list of command names.
// ExecuteWithInput runs a command with stdin-like input (for pipes).
func (r *Registry) ExecuteWithInput(name string, args []string, input string) Response {
	switch name {
	case "grep":
		return r.grepWithInput(args, input)
	default:
		return r.Execute(name, args)
	}
}

func (r *Registry) Available() []string {
	names := make([]string, 0, len(r.commands))
	for k := range r.commands {
		names = append(names, k)
	}
	return names
}

func (r *Registry) register(name string, fn func(args []string) Response) {
	r.commands[name] = fn
}

func (r *Registry) registerBuiltins() {
	r.register("help", r.help)
	r.register("about", r.about)
	r.register("whoami", r.whoami)
	r.register("projects", r.projects)
	r.register("contact", r.contact)
	r.register("clear", r.clear)
	r.register("date", r.date)
	r.register("echo", r.echo)
	r.register("neofetch", r.neofetch)
	r.register("banner", r.banner)
	r.register("social", r.social)
	r.register("theme", r.theme)
	r.register("ls", r.ls)
	r.register("cat", r.cat)
	r.register("history", r.history)
	r.register("uptime", r.uptime)
	r.register("grep", r.grep)
	r.register("hostname", r.hostname)
	r.register("matrix", r.matrix)
}

func (r *Registry) help(args []string) Response {
	return Response{
		Output: `available commands:
──────────────────────
  about       about me
  banner      display the oathless banner
  cat         display file contents
  clear       clear the terminal
  contact     how to reach me
  date        current date and time
  echo        echo text back
  grep        filter output by pattern
  help        show this help
  history     show command history
  hostname    show server hostname
  ls          list available pages
  matrix      follow the white rabbit...
  neofetch    system information (parody)
  projects    projects i've worked on
  social      social links
  theme       change color theme
  uptime      server uptime
  whoami      who am i
──────────────────────
piping: command | grep pattern
type any command to get started.`,
		Type: "text",
	}
}

func (r *Registry) about(args []string) Response {
	return Response{
		Output: `┌───────────────────────────────────────────┐
│                                           │
│   hi, i'm ruben.                          │
│                                           │
│   i build things — homelabs, mods,        │
│   games, tools, and whatever else         │
│   catches my interest.                    │
│                                           │
│   this site runs on an optiplex 3070      │
│   micro in the netherlands, behind        │
│   caddy, inside docker, built with        │
│   go and vue.                             │
│                                           │
│   type 'projects' to see what i've done.  │
│                                           │
└───────────────────────────────────────────┘`,
		Type: "text",
	}
}

func (r *Registry) whoami(args []string) Response {
	return Response{
		Output: "oathless",
		Type:   "text",
	}
}

func (r *Registry) projects(args []string) Response {
	return Response{
		Output: `projects:
──────────────────────
  coming soon.
──────────────────────`,
		Type: "text",
	}
}

func (r *Registry) contact(args []string) Response {
	return Response{
		Output: `contact:
──────────────────────
  email    rubenhesselink@pm.me
  github   github.com/oathlesss
  discord  @oathless
──────────────────────`,
		Type: "text",
	}
}

func (r *Registry) clear(args []string) Response {
	return Response{Type: "clear"}
}

func (r *Registry) date(args []string) Response {
	now := time.Now()
	return Response{
		Output: now.Format("Mon Jan 2 15:04:05 MST 2006"),
		Type:   "text",
	}
}

func (r *Registry) echo(args []string) Response {
	return Response{
		Output: strings.Join(args, " "),
		Type:   "text",
	}
}

func (r *Registry) neofetch(args []string) Response {
	return Response{
		Output: fmt.Sprintf(`         .o0K0o.         oathless@oathless.dev
        '0KWXKKK0;        ────────────────────
       .KN..xWXKKd        os      %s
       lX'  kMNxX:        kernel  %s
       KM.  kMNkNc        shell   %s
      .XM0kkKN0KXl        uptime  %s
      lWWNWWWNWWNx
      XWWWWWWWWWWW0
     'WWWWWWWWWWWWWo
    .NWWWWWWWWWWWWWW.
    kMWWWWWWWWWWWWWW0
   '0WWWWWWWWWWWWWWWK,
   cNWWWWWWWWWWWWWWWO.
  .kWMMMMMWWWWWWWWNk'.
  lNMMMMMMMWWWWWWWKc
 'OMMMMMMMMWWWWWNk,
 dNMMMMMMMMMWWWMK;
'OMMMMMMMMMMMWMK:
,WMMMMMMMMMMMNk,
KMMMMMMMNkdl;.
WMMMMW0c.
MMMM0'
0Mx
l.`, runtime.GOOS, runtime.Version(), "bash", formatUptime()),
		Type: "text",
	}
}

func (r *Registry) banner(args []string) Response {
	return Response{
		Output: `
 ██████╗  █████╗ ████████╗██╗  ██╗██╗     ███████╗███████╗███████╗
██╔═══██╗██╔══██╗╚══██╔══╝██║  ██║██║     ██╔════╝██╔════╝██╔════╝
██║   ██║███████║   ██║   ███████║██║     █████╗  ███████╗███████╗
██║   ██║██╔══██║   ██║   ██╔══██║██║     ██╔══╝  ╚════██║╚════██║
╚██████╔╝██║  ██║   ██║   ██║  ██║███████╗███████╗███████║███████║
 ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝╚══════╝╚══════╝╚══════╝╚══════╝

   welcome to my terminal. type 'help' to get started.`,
		Type: "text",
	}
}

func (r *Registry) social(args []string) Response {
	return Response{
		Output: `social:
──────────────────────
  github   https://github.com/oathlesss
  discord  @oathless
──────────────────────`,
		Type: "text",
	}
}

func (r *Registry) theme(args []string) Response {
	if len(args) == 0 {
		return Response{
			Output: `available themes:
  rose-pine (current)  •  green  •  amber  •  matrix
usage: theme <name>`,
			Type: "text",
		}
	}
	theme := strings.ToLower(args[0])
	valid := map[string]bool{"rose-pine": true, "green": true, "amber": true, "matrix": true}
	if !valid[theme] {
		return Response{
			Output: fmt.Sprintf("unknown theme '%s'. available: rose-pine, green, amber, matrix", theme),
			Type:   "error",
		}
	}
	return Response{
		Output: fmt.Sprintf("theme changed to %s", theme),
		Type:   fmt.Sprintf("theme:%s", theme),
	}
}

func (r *Registry) ls(args []string) Response {
	return Response{
		Output: `about.txt  banner.txt  contact.txt  projects.txt  social.txt`,
		Type:   "text",
	}
}

func (r *Registry) cat(args []string) Response {
	if len(args) == 0 {
		return Response{Output: "usage: cat <file>", Type: "error"}
	}

	files := map[string]string{
		"about.txt":    r.about(nil).Output,
		"banner.txt":   r.banner(nil).Output,
		"contact.txt":  r.contact(nil).Output,
		"projects.txt": r.projects(nil).Output,
		"social.txt":   r.social(nil).Output,
		"help.txt":     r.help(nil).Output,
	}

	content, ok := files[args[0]]
	if !ok {
		return Response{
			Output: fmt.Sprintf("cat: %s: no such file", args[0]),
			Type:   "error",
		}
	}
	return Response{Output: content, Type: "text"}
}

func (r *Registry) history(args []string) Response {
	return Response{
		Output: "(history is local to your browser session)",
		Type:   "text",
	}
}

func (r *Registry) uptime(args []string) Response {
	return Response{
		Output: "up " + formatUptime(),
		Type:   "text",
	}
}

func (r *Registry) hostname(args []string) Response {
	return Response{
		Output: "oathless.dev",
		Type:   "text",
	}
}

func (r *Registry) grep(args []string) Response {
	return Response{
		Output: "usage: grep <pattern>\npipe output through grep: command | grep pattern",
		Type:   "error",
	}
}

func (r *Registry) matrix(args []string) Response {
	return Response{
		Output: "follow the white rabbit...",
		Type:   "matrix",
	}
}

func formatUptime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "unavailable"
	}
	parts := strings.Fields(string(data))
	if len(parts) == 0 {
		return "unavailable"
	}
	var seconds float64
	fmt.Sscanf(parts[0], "%f", &seconds)

	days := int(seconds) / 86400
	hours := (int(seconds) % 86400) / 3600
	minutes := (int(seconds) % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func (r *Registry) grepWithInput(args []string, input string) Response {
	if len(args) == 0 {
		return Response{Output: "usage: grep <pattern>", Type: "error"}
	}

	pattern := strings.ToLower(args[0])
	lines := strings.Split(input, "\n")
	var matched []string

	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), pattern) {
			matched = append(matched, line)
		}
	}

	if len(matched) == 0 {
		return Response{Output: "(no matches)", Type: "text"}
	}

	return Response{Output: strings.Join(matched, "\n"), Type: "text"}
}
