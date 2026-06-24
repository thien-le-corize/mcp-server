package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Config struct {
	NginxErrorLog  string   `json:"nginx_error_log"`
	NginxAccessLog string   `json:"nginx_access_log"`
	LogBasePath    string   `json:"log_base_path"`
	AllowedPaths   []string `json:"allowed_log_paths"`
	Path           string   `json:"path"`
}

var config Config

func loadConfig() {
	data, err := os.ReadFile("/opt/tt-mcp/config.json")
	if err != nil {
		data, err = os.ReadFile("config.json")
		if err != nil {
			config = Config{
				NginxErrorLog:  "/var/log/nginx/error.log",
				NginxAccessLog: "/var/log/nginx/access.log",
				LogBasePath:    "/home",
				AllowedPaths:   []string{"/var/log/", "/home/"},
			}
			return
		}
	}
	json.Unmarshal(data, &config)
}

func envPath() []string {
	if config.Path != "" {
		return append(os.Environ(), "PATH="+config.Path)
	}
	return os.Environ()
}

func run(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Env = envPath()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("error: %s\n%s", err, string(out))
	}
	return string(out)
}

func shell(cmd string) string {
	c := exec.Command("bash", "-c", cmd)
	c.Env = envPath()
	out, err := c.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("error: %s\n%s", err, string(out))
	}
	return string(out)
}

func toolResult(text string) any {
	return map[string]any{"content": []TextContent{{Type: "text", Text: text}}}
}

func toolError(text string) any {
	return map[string]any{"content": []TextContent{{Type: "text", Text: text}}, "isError": true}
}

type P map[string]string

func schemaEmpty() any {
	return map[string]any{"type": "object", "properties": map[string]any{}}
}

func schema(props P, required []string) any {
	p := map[string]any{}
	for k, v := range props {
		p[k] = map[string]string{"type": v}
	}
	s := map[string]any{"type": "object", "properties": p}
	if required != nil {
		s["required"] = required
	}
	return s
}

var tools = []Tool{
	// System
	{"get_memory_usage", "Get system memory and swap usage", schemaEmpty()},
	{"get_disk_usage", "Get disk space usage", schemaEmpty()},
	{"get_top_processes", "Get top processes by resource usage", schema(P{"sort": "string", "count": "number"}, nil)},
	{"get_disk_io", "Get disk I/O stats from sar (sysstat)", schema(P{"date": "string", "time_filter": "string"}, []string{"date"})},
	{"get_sar", "Run sar command with any flag (-q load, -r memory, -u cpu, -n network, -d disk)", schema(P{"flag": "string", "date": "string", "time_filter": "string"}, []string{"flag", "date"})},
	// PM2
	{"get_pm2_status", "Get PM2 process list and status", schemaEmpty()},
	{"get_pm2_logs", "Get recent PM2 logs", schema(P{"name": "string", "lines": "number"}, []string{"name"})},
	// Docker
	{"get_docker_ps", "List Docker containers", schemaEmpty()},
	{"get_docker_logs", "Get Docker container logs", schema(P{"container": "string", "lines": "number"}, []string{"container"})},
	// Nginx basic
	{"get_nginx_errors", "Get recent Nginx error log", schema(P{"lines": "number"}, nil)},
	{"get_nginx_access", "Get recent Nginx access log", schema(P{"lines": "number"}, nil)},
	// Nginx traffic analysis
	{"count_requests_in_timerange", "Count total requests in a time range from nginx access logs", schema(P{"date": "string", "start_hour": "string", "start_min": "string", "end_hour": "string", "end_min": "string", "log_suffix": "string"}, []string{"date", "start_hour", "start_min", "end_hour", "end_min"})},
	{"count_requests_per_site", "Count requests per website in a time range", schema(P{"date": "string", "start_hour": "string", "start_min": "string", "end_hour": "string", "end_min": "string", "log_suffix": "string"}, []string{"date", "start_hour", "start_min", "end_hour", "end_min"})},
	{"count_requests_per_minute", "Count requests per minute in a time range", schema(P{"date": "string", "start_hour": "string", "start_min": "string", "end_hour": "string", "end_min": "string", "log_suffix": "string"}, []string{"date", "start_hour", "start_min", "end_hour", "end_min"})},
	{"top_ips_in_timerange", "Top IPs by request count in a time range", schema(P{"date": "string", "start_hour": "string", "start_min": "string", "end_hour": "string", "end_min": "string", "limit": "number", "log_suffix": "string"}, []string{"date", "start_hour", "start_min", "end_hour", "end_min"})},
	{"top_urls_in_timerange", "Top requested URLs in a time range", schema(P{"date": "string", "start_hour": "string", "start_min": "string", "end_hour": "string", "end_min": "string", "limit": "number", "log_suffix": "string"}, []string{"date", "start_hour", "start_min", "end_hour", "end_min"})},
	{"analyze_ip", "Analyze a specific IP: which sites, URLs, user-agents, status codes, timeline", schema(P{"ip": "string", "date": "string", "log_suffix": "string"}, []string{"ip"})},
	{"grep_requests", "Grep nginx access logs with custom pattern in a time range", schema(P{"pattern": "string", "date": "string", "start_hour": "string", "start_min": "string", "end_hour": "string", "end_min": "string", "log_suffix": "string"}, []string{"pattern"})},
	// PM2 advanced
	{"get_pm2_error_logs", "Get PM2 error logs for a process", schema(P{"name": "string", "lines": "number"}, []string{"name"})},
	// OOM investigation
	{"investigate_oom", "Find OOM killer events: when, which process was killed, memory state", schema(P{"since": "string", "lines": "number"}, nil)},
	{"get_pm2_restarts", "Show PM2 processes with restart counts and uptime to detect OOM crashes", schemaEmpty()},
	{"get_dmesg_oom", "Get kernel OOM killer messages from dmesg", schema(P{"lines": "number"}, nil)},
	// Journal
	{"get_journal", "Get systemd journal logs for a unit", schema(P{"unit": "string", "lines": "number", "since": "string"}, []string{"unit"})},
	// Generic log
	{"get_app_log", "Read log file (restricted to allowed paths)", schema(P{"path": "string", "lines": "number"}, []string{"path"})},
	// Incident
	{"investigate_incident", "Quick incident overview: memory, disk, pm2, docker, nginx errors", schemaEmpty()},
}

// Build time-range regex for nginx log format: DD/Mon/YYYY:HH:MM
func buildTimeRegex(date, startH, startM, endH, endM string) string {
	// Generate regex matching minutes from startH:startM to endH:endM
	// For simplicity, build a grep -E pattern
	sh, _ := strconv.Atoi(startH)
	sm, _ := strconv.Atoi(startM)
	eh, _ := strconv.Atoi(endH)
	em, _ := strconv.Atoi(endM)

	var parts []string
	for h := sh; h <= eh; h++ {
		mStart := 0
		mEnd := 59
		if h == sh {
			mStart = sm
		}
		if h == eh {
			mEnd = em
		}
		for m := mStart; m <= mEnd; m++ {
			parts = append(parts, fmt.Sprintf("%s:%02d:%02d", date, h, m))
		}
	}
	// Build alternation in groups of efficiency
	if len(parts) == 0 {
		return date + ":" + startH + ":" + startM
	}
	// Use simpler regex approach
	return strings.Join(parts, "|")
}

func logGlob(suffix string) string {
	base := config.LogBasePath
	if suffix == "" {
		return base + "/*.log"
	}
	return base + "/*" + suffix + "*.log"
}

func timeGrep(date, startH, startM, endH, endM, logSuffix string) string {
	// Build efficient grep pattern
	regex := buildTimeRegex(date, startH, startM, endH, endM)
	glob := logGlob(logSuffix)
	return fmt.Sprintf(`grep -hE "(%s)" %s`, regex, glob)
}

func getParam(params json.RawMessage, key string, def string) string {
	var m map[string]any
	json.Unmarshal(params, &m)
	if m == nil {
		return def
	}
	var args map[string]any
	if a, ok := m["arguments"]; ok {
		if am, ok := a.(map[string]any); ok {
			args = am
		}
	}
	if args == nil {
		return def
	}
	if v, ok := args[key]; ok {
		switch t := v.(type) {
		case string:
			return t
		case float64:
			return strconv.Itoa(int(t))
		}
	}
	return def
}

func getToolName(params json.RawMessage) string {
	var m map[string]any
	json.Unmarshal(params, &m)
	if m == nil {
		return ""
	}
	if n, ok := m["name"].(string); ok {
		return n
	}
	return ""
}

func isPathAllowed(path string) bool {
	for _, p := range config.AllowedPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func handleTool(name string, params json.RawMessage) any {
	lines := getParam(params, "lines", "100")

	switch name {
	// === System ===
	case "get_memory_usage":
		return toolResult(run("free", "-h"))
	case "get_disk_usage":
		return toolResult(run("df", "-h"))
	case "get_top_processes":
		sort := getParam(params, "sort", "cpu")
		count := getParam(params, "count", "20")
		key := "-pcpu"
		if sort == "mem" {
			key = "-pmem"
		}
		out := run("ps", "aux", "--sort", key)
		n, _ := strconv.Atoi(count)
		ll := strings.Split(out, "\n")
		if len(ll) > n+1 {
			ll = ll[:n+1]
		}
		return toolResult(strings.Join(ll, "\n"))
	case "get_disk_io":
		date := getParam(params, "date", "")
		tf := getParam(params, "time_filter", "")
		saFile := fmt.Sprintf("/var/log/sysstat/sa%s", date)
		cmd := fmt.Sprintf("sar -d -f %s", saFile)
		if tf != "" {
			cmd += " | egrep '" + tf + "'"
		}
		return toolResult(shell(cmd))

	case "get_sar":
		flag := getParam(params, "flag", "-q")
		date := getParam(params, "date", "")
		tf := getParam(params, "time_filter", "")
		// Only allow safe sar flags
		allowed := map[string]bool{"-q": true, "-r": true, "-u": true, "-d": true, "-n DEV": true, "-b": true, "-w": true, "-S": true}
		if !allowed[flag] {
			return toolError("allowed flags: -q (load), -r (mem), -u (cpu), -d (disk), -n DEV (net), -b (io), -w (context switch), -S (swap)")
		}
		saFile := fmt.Sprintf("/var/log/sysstat/sa%s", date)
		cmd := fmt.Sprintf("sar %s -f %s", flag, saFile)
		if tf != "" {
			cmd += " | egrep '" + tf + "'"
		}
		return toolResult(shell(cmd))

	// === PM2 ===
	case "get_pm2_status":
		return toolResult(run("pm2", "jlist"))
	case "get_pm2_logs":
		n := getParam(params, "name", "")
		if n == "" {
			return toolError("name is required")
		}
		return toolResult(run("pm2", "logs", n, "--nostream", "--lines", lines))

	// === Docker ===
	case "get_docker_ps":
		return toolResult(run("docker", "ps", "-a", "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}"))
	case "get_docker_logs":
		c := getParam(params, "container", "")
		if c == "" {
			return toolError("container is required")
		}
		return toolResult(run("docker", "logs", "--tail", lines, c))

	// === Nginx basic ===
	case "get_nginx_errors":
		return toolResult(run("tail", "-n", lines, config.NginxErrorLog))
	case "get_nginx_access":
		return toolResult(run("tail", "-n", lines, config.NginxAccessLog))

	// === Nginx traffic analysis ===
	case "count_requests_in_timerange":
		date := getParam(params, "date", "")
		sh := getParam(params, "start_hour", "")
		sm := getParam(params, "start_min", "")
		eh := getParam(params, "end_hour", "")
		em := getParam(params, "end_min", "")
		ls := getParam(params, "log_suffix", "")
		cmd := timeGrep(date, sh, sm, eh, em, ls) + " | wc -l"
		return toolResult(shell(cmd))

	case "count_requests_per_site":
		date := getParam(params, "date", "")
		sh := getParam(params, "start_hour", "")
		sm := getParam(params, "start_min", "")
		eh := getParam(params, "end_hour", "")
		em := getParam(params, "end_min", "")
		ls := getParam(params, "log_suffix", "")
		// Use grep -R to keep filename
		regex := buildTimeRegex(date, sh, sm, eh, em)
		glob := logGlob(ls)
		cmd := fmt.Sprintf(`grep -RE "(%s)" %s | cut -d: -f1 | sort | uniq -c | sort -nr`, regex, glob)
		return toolResult(shell(cmd))

	case "count_requests_per_minute":
		date := getParam(params, "date", "")
		sh := getParam(params, "start_hour", "")
		sm := getParam(params, "start_min", "")
		eh := getParam(params, "end_hour", "")
		em := getParam(params, "end_min", "")
		ls := getParam(params, "log_suffix", "")
		cmd := timeGrep(date, sh, sm, eh, em, ls) + ` | sed -E 's/.*\[([0-9]{2}\/[A-Za-z]+\/[0-9]{4}:[0-9]{2}:[0-9]{2}).*/\1/' | sort | uniq -c`
		return toolResult(shell(cmd))

	case "top_ips_in_timerange":
		date := getParam(params, "date", "")
		sh := getParam(params, "start_hour", "")
		sm := getParam(params, "start_min", "")
		eh := getParam(params, "end_hour", "")
		em := getParam(params, "end_min", "")
		limit := getParam(params, "limit", "30")
		ls := getParam(params, "log_suffix", "")
		cmd := timeGrep(date, sh, sm, eh, em, ls) + fmt.Sprintf(` | awk '{print $1}' | sort | uniq -c | sort -nr | head -%s`, limit)
		return toolResult(shell(cmd))

	case "top_urls_in_timerange":
		date := getParam(params, "date", "")
		sh := getParam(params, "start_hour", "")
		sm := getParam(params, "start_min", "")
		eh := getParam(params, "end_hour", "")
		em := getParam(params, "end_min", "")
		limit := getParam(params, "limit", "50")
		ls := getParam(params, "log_suffix", "")
		cmd := timeGrep(date, sh, sm, eh, em, ls) + fmt.Sprintf(` | awk '{print $7}' | sort | uniq -c | sort -nr | head -%s`, limit)
		return toolResult(shell(cmd))

	case "analyze_ip":
		ip := getParam(params, "ip", "")
		if ip == "" {
			return toolError("ip is required")
		}
		date := getParam(params, "date", "")
		ls := getParam(params, "log_suffix", "")
		glob := logGlob(ls)
		if date == "" {
			glob = logGlob("")
		}

		var b strings.Builder
		b.WriteString("=== SITES ===\n")
		b.WriteString(shell(fmt.Sprintf(`grep -R "%s" %s | cut -d: -f1 | sort | uniq -c | sort -nr`, ip, glob)))
		b.WriteString("\n=== TOP URLs ===\n")
		b.WriteString(shell(fmt.Sprintf(`grep -Rh "%s" %s | awk '{print $7}' | sort | uniq -c | sort -nr | head -30`, ip, glob)))
		b.WriteString("\n=== STATUS CODES ===\n")
		b.WriteString(shell(fmt.Sprintf(`grep -Rh "%s" %s | awk '{print $9}' | sort | uniq -c | sort -nr`, ip, glob)))
		b.WriteString("\n=== USER AGENTS ===\n")
		b.WriteString(shell(fmt.Sprintf(`grep -Rh "%s" %s | awk -F'"' '{print $6}' | sort | uniq -c | sort -nr | head -10`, ip, glob)))
		b.WriteString("\n=== TIMELINE (per minute) ===\n")
		b.WriteString(shell(fmt.Sprintf(`grep -Rh "%s" %s | sed -E 's/.*\[([0-9]{2}\/[A-Za-z]+\/[0-9]{4}:[0-9]{2}:[0-9]{2}).*/\1/' | sort | uniq -c`, ip, glob)))
		return toolResult(b.String())

	case "grep_requests":
		pattern := getParam(params, "pattern", "")
		if pattern == "" {
			return toolError("pattern is required")
		}
		date := getParam(params, "date", "")
		sh := getParam(params, "start_hour", "")
		sm := getParam(params, "start_min", "")
		eh := getParam(params, "end_hour", "")
		em := getParam(params, "end_min", "")
		ls := getParam(params, "log_suffix", "")
		if sh != "" && eh != "" {
			cmd := timeGrep(date, sh, sm, eh, em, ls) + fmt.Sprintf(` | egrep "%s"`, pattern)
			return toolResult(shell(cmd))
		}
		glob := logGlob(ls)
		cmd := fmt.Sprintf(`grep -Rh "%s" %s | head -200`, pattern, glob)
		return toolResult(shell(cmd))

	// === Journal ===
	case "get_journal":
		unit := getParam(params, "unit", "")
		if unit == "" {
			return toolError("unit is required")
		}
		args := []string{"-u", unit, "-n", lines, "--no-pager"}
		if since := getParam(params, "since", ""); since != "" {
			args = append(args, "--since", since)
		}
		return toolResult(run("journalctl", args...))

	// === PM2 advanced ===
	case "get_pm2_error_logs":
		n := getParam(params, "name", "")
		if n == "" {
			return toolError("name is required")
		}
		// PM2 error logs are at ~/.pm2/logs/<name>-error.log
		cmd := fmt.Sprintf(`tail -n %s ~/.pm2/logs/%s-error.log 2>/dev/null || pm2 logs %s --err --nostream --lines %s`, lines, n, n, lines)
		return toolResult(shell(cmd))

	// === OOM Investigation ===
	case "investigate_oom":
		since := getParam(params, "since", "24 hours ago")
		var b strings.Builder
		b.WriteString("=== KERNEL OOM KILLER (dmesg) ===\n")
		b.WriteString(shell(`dmesg -T | grep -i "out of memory\|oom-kill\|killed process" | tail -50`))
		b.WriteString("\n\n=== JOURNAL OOM EVENTS ===\n")
		b.WriteString(shell(fmt.Sprintf(`journalctl --since "%s" --no-pager | grep -i "out of memory\|oom-kill\|killed process" | tail -50`, since)))
		b.WriteString("\n\n=== PM2 RESTART HISTORY ===\n")
		b.WriteString(shell(`pm2 jlist 2>/dev/null | python3 -c "
import sys,json
procs=json.load(sys.stdin)
for p in procs:
  e=p.get('pm2_env',{})
  if e.get('restart_time',0)>0:
    print(f\"{p['name']}: restarts={e['restart_time']} status={e['status']} uptime_since={e.get('pm_uptime','?')}\")
" 2>/dev/null || pm2 list`))
		b.WriteString("\n\n=== CURRENT MEMORY ===\n")
		b.WriteString(run("free", "-h"))
		return toolResult(b.String())

	case "get_pm2_restarts":
		out := shell(`pm2 jlist 2>/dev/null | python3 -c "
import sys,json,datetime
procs=json.load(sys.stdin)
for p in sorted(procs, key=lambda x: x.get('pm2_env',{}).get('restart_time',0), reverse=True):
  e=p.get('pm2_env',{})
  up=datetime.datetime.fromtimestamp(e.get('pm_uptime',0)/1000).strftime('%Y-%m-%d %H:%M:%S') if e.get('pm_uptime') else '?'
  print(f\"{p['name']:20s} restarts={e.get('restart_time',0):4d}  status={e.get('status'):10s}  up_since={up}\")
" 2>/dev/null || pm2 list`)
		return toolResult(out)

	case "get_dmesg_oom":
		n := getParam(params, "lines", "50")
		return toolResult(shell(fmt.Sprintf(`dmesg -T | grep -i "out of memory\|oom-kill\|killed process\|memory cgroup" | tail -%s`, n)))


	// === Generic log ===
	case "get_app_log":
		path := getParam(params, "path", "")
		if path == "" {
			return toolError("path is required")
		}
		if !isPathAllowed(path) {
			return toolError("access denied. Allowed: " + strings.Join(config.AllowedPaths, ", "))
		}
		return toolResult(run("tail", "-n", lines, path))

	// === Incident ===
	case "investigate_incident":
		var b strings.Builder
		b.WriteString("=== MEMORY ===\n")
		b.WriteString(run("free", "-h"))
		b.WriteString("\n\n=== DISK ===\n")
		b.WriteString(run("df", "-h"))
		b.WriteString("\n\n=== PM2 ===\n")
		b.WriteString(run("pm2", "jlist"))
		b.WriteString("\n\n=== DOCKER ===\n")
		b.WriteString(run("docker", "ps", "-a", "--format", "table {{.Names}}\t{{.Status}}"))
		b.WriteString("\n\n=== NGINX ERRORS (last 20) ===\n")
		b.WriteString(run("tail", "-n", "20", config.NginxErrorLog))
		return toolResult(b.String())
	}
	return toolError("unknown tool: " + name)
}

func handle(req Request) Response {
	switch req.Method {
	case "initialize":
		return Response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":   map[string]any{"tools": map[string]any{}},
			"serverInfo":     map[string]any{"name": "tt-mcp", "version": "1.0.0"},
		}}
	case "notifications/initialized":
		return Response{}
	case "tools/list":
		return Response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"tools": tools}}
	case "tools/call":
		name := getToolName(req.Params)
		result := handleTool(name, req.Params)
		return Response{JSONRPC: "2.0", ID: req.ID, Result: result}
	default:
		return Response{JSONRPC: "2.0", ID: req.ID, Error: &Error{Code: -32601, Message: "method not found"}}
	}
}

func main() {
	loadConfig()
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	writer := bufio.NewWriter(os.Stdout)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}
		resp := handle(req)
		if resp.JSONRPC == "" {
			continue
		}
		data, _ := json.Marshal(resp)
		fmt.Fprintf(writer, "%s\n", data)
		writer.Flush()
	}
}
