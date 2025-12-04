package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

var version = "dev"

type config struct {
	jaegerURL    string
	jsonOutput   bool
	minDuration  time.Duration
	errorsOnly   bool
	service      string
	maxDepth     int
	relativeTime bool
}

type traceResponse struct {
	Data []trace `json:"data"`
}

type trace struct {
	TraceID   string             `json:"traceID"`
	Spans     []span             `json:"spans"`
	Processes map[string]process `json:"processes"`
}

type span struct {
	SpanID        string      `json:"spanID"`
	OperationName string      `json:"operationName"`
	References    []reference `json:"references"`
	StartTime     int64       `json:"startTime"`
	Duration      int64       `json:"duration"`
	ProcessID     string      `json:"processID"`
	Tags          []tag       `json:"tags"`
}

type reference struct {
	RefType string `json:"refType"`
	SpanID  string `json:"spanID"`
}

type process struct {
	ServiceName string `json:"serviceName"`
}

type tag struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

type spanNode struct {
	span     span
	service  string
	children []*spanNode
}

func (n *spanNode) matchesFilter(cfg *config) bool {
	if n.matchesSelf(cfg) {
		return true
	}
	for _, child := range n.children {
		if child.matchesFilter(cfg) {
			return true
		}
	}
	return false
}

func (n *spanNode) matchesSelf(cfg *config) bool {
	if cfg.minDuration > 0 && time.Duration(n.span.Duration)*time.Microsecond < cfg.minDuration {
		return false
	}
	if cfg.errorsOnly && !n.hasError() {
		return false
	}
	if cfg.service != "" && n.service != cfg.service {
		return false
	}
	return true
}

func (n *spanNode) hasError() bool {
	for _, t := range n.span.Tags {
		if t.Key == "otel.status_code" && t.Value == "ERROR" {
			return true
		}
		if t.Key == "error" {
			if v, ok := t.Value.(bool); ok && v {
				return true
			}
		}
	}
	return false
}

func main() {
	cfg := &config{jaegerURL: "http://localhost:16686"}
	showVersion := false

	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.StringVar(&cfg.jaegerURL, "url", cfg.jaegerURL, "Jaeger URL")
	flag.BoolVar(&cfg.jsonOutput, "json", false, "output verbose JSON with all tags")
	flag.DurationVar(
		&cfg.minDuration,
		"min-duration",
		0,
		"only show spans with duration >= this value (e.g. 100ms, 1s)",
	)
	flag.BoolVar(&cfg.errorsOnly, "error", false, "only show error spans and their ancestors")
	flag.StringVar(&cfg.service, "service", "", "only show spans from this service")
	flag.IntVar(&cfg.maxDepth, "depth", 0, "limit tree depth (0 = unlimited)")
	flag.BoolVar(&cfg.relativeTime, "relative", false, "show timestamps relative to trace start")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `jtree - display Jaeger traces in a hierarchical view

Usage:
  jtree [flags] <trace-id>

Examples:
  jtree abc123def456
  jtree http://localhost:16686/trace/abc123def456
  jtree -json abc123def456
  jtree -url http://jaeger:16686 abc123def456

Flags:
`)
		flag.PrintDefaults()
	}
	flag.Parse()

	if showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	baseURL, traceID := parseInput(flag.Arg(0), cfg.jaegerURL)
	cfg.jaegerURL = baseURL

	resp, err := http.Get(fmt.Sprintf("%s/api/traces/%s", cfg.jaegerURL, traceID))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch trace: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "jaeger returned status %d\n", resp.StatusCode)
		os.Exit(1)
	}

	var traceResp traceResponse
	if err := json.NewDecoder(resp.Body).Decode(&traceResp); err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode response: %v\n", err)
		os.Exit(1)
	}

	if len(traceResp.Data) == 0 {
		fmt.Fprintf(os.Stderr, "no trace found with ID %s\n", traceID)
		os.Exit(1)
	}

	t := traceResp.Data[0]
	roots, startTime := buildTree(t)
	printRoots(roots, startTime, cfg)
}

func buildTree(t trace) ([]*spanNode, int64) {
	var startTime int64
	spanMap := make(map[string]*spanNode)

	for _, s := range t.Spans {
		service := ""
		if p, ok := t.Processes[s.ProcessID]; ok {
			service = p.ServiceName
		}
		spanMap[s.SpanID] = &spanNode{span: s, service: service}
		if startTime == 0 || s.StartTime < startTime {
			startTime = s.StartTime
		}
	}

	var roots []*spanNode
	for _, s := range t.Spans {
		node := spanMap[s.SpanID]
		parentID := getParentID(s)
		if parentID != "" {
			if parent, ok := spanMap[parentID]; ok {
				parent.children = append(parent.children, node)
				continue
			}
		}
		roots = append(roots, node)
	}

	sortNodes(roots)
	for _, node := range spanMap {
		sortNodes(node.children)
	}

	return roots, startTime
}

func getParentID(s span) string {
	for _, ref := range s.References {
		if ref.RefType == "CHILD_OF" {
			return ref.SpanID
		}
	}
	return ""
}

func sortNodes(nodes []*spanNode) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].span.StartTime < nodes[j].span.StartTime
	})
}

func printRoots(roots []*spanNode, startTime int64, cfg *config) {
	for _, root := range roots {
		printNode(root, 0, startTime, cfg)
	}
}

func printNode(node *spanNode, depth int, startTime int64, cfg *config) {
	if !node.matchesFilter(cfg) {
		return
	}
	if cfg.maxDepth > 0 && depth >= cfg.maxDepth {
		return
	}

	indent := strings.Repeat("  ", depth)
	duration := formatDuration(node.span.Duration)

	if cfg.jsonOutput {
		tags := make(map[string]any)
		for _, t := range node.span.Tags {
			tags[t.Key] = t.Value
		}
		out := map[string]any{
			"span_id":  node.span.SpanID,
			"service":  node.service,
			"duration": duration,
			"tags":     tags,
		}
		jsonBytes, _ := json.Marshal(out)
		fmt.Printf("%s%s %s\n", indent, node.span.OperationName, string(jsonBytes))
	} else {
		var timeStr string
		if cfg.relativeTime {
			offset := node.span.StartTime - startTime
			timeStr = "+" + formatDuration(offset)
		} else {
			timeStr = time.UnixMicro(node.span.StartTime).Format("15:04:05.000")
		}
		fmt.Printf("%s%s [%s] %s %s\n", indent, node.span.OperationName, node.service, timeStr, duration)
	}

	for _, child := range node.children {
		printNode(child, depth+1, startTime, cfg)
	}
}

func formatDuration(us int64) string {
	if us < 1000 {
		return fmt.Sprintf("%dus", us)
	}
	if us < 1000000 {
		return fmt.Sprintf("%.2fms", float64(us)/1000)
	}
	return fmt.Sprintf("%.2fs", float64(us)/1000000)
}

func parseInput(input, defaultURL string) (baseURL, traceID string) {
	u, err := url.Parse(input)
	if err != nil || u.Scheme == "" {
		return defaultURL, input
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "trace" {
		return fmt.Sprintf("%s://%s", u.Scheme, u.Host), parts[1]
	}

	return defaultURL, input
}
