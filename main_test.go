package main

import "testing"

func TestParseInput(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		defaultURL string
		wantURL    string
		wantID     string
	}{
		{
			name:       "trace ID only",
			input:      "abc123def456",
			defaultURL: "http://localhost:16686",
			wantURL:    "http://localhost:16686",
			wantID:     "abc123def456",
		},
		{
			name:       "full jaeger URL localhost",
			input:      "http://localhost:16686/trace/abc123def456",
			defaultURL: "http://localhost:16686",
			wantURL:    "http://localhost:16686",
			wantID:     "abc123def456",
		},
		{
			name:       "full jaeger URL different host",
			input:      "http://jaeger:16686/trace/xyz789",
			defaultURL: "http://localhost:16686",
			wantURL:    "http://jaeger:16686",
			wantID:     "xyz789",
		},
		{
			name:       "https URL",
			input:      "https://jaeger.example.com:443/trace/test123",
			defaultURL: "http://localhost:16686",
			wantURL:    "https://jaeger.example.com:443",
			wantID:     "test123",
		},
		{
			name:       "URL without port",
			input:      "http://jaeger/trace/mytraceID",
			defaultURL: "http://localhost:16686",
			wantURL:    "http://jaeger",
			wantID:     "mytraceID",
		},
		{
			name:       "trace ID with hyphens",
			input:      "abc-123-def-456",
			defaultURL: "http://localhost:16686",
			wantURL:    "http://localhost:16686",
			wantID:     "abc-123-def-456",
		},
		{
			name:       "URL with trailing slash",
			input:      "http://localhost:16686/trace/abc123/",
			defaultURL: "http://localhost:16686",
			wantURL:    "http://localhost:16686",
			wantID:     "abc123",
		},
		{
			name:       "URL with query parameters",
			input:      "http://localhost:16686/trace/abc123?uiEmbed=v0",
			defaultURL: "http://localhost:16686",
			wantURL:    "http://localhost:16686",
			wantID:     "abc123",
		},
		{
			name:       "malformed URL returns as trace ID",
			input:      "://invalid",
			defaultURL: "http://localhost:16686",
			wantURL:    "http://localhost:16686",
			wantID:     "://invalid",
		},
		{
			name:       "URL without trace path",
			input:      "http://localhost:16686/something/else",
			defaultURL: "http://localhost:16686",
			wantURL:    "http://localhost:16686",
			wantID:     "http://localhost:16686/something/else",
		},
		{
			name:       "empty input",
			input:      "",
			defaultURL: "http://localhost:16686",
			wantURL:    "http://localhost:16686",
			wantID:     "",
		},
		{
			name:       "custom default URL",
			input:      "abc123",
			defaultURL: "http://jaeger-prod:16686",
			wantURL:    "http://jaeger-prod:16686",
			wantID:     "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotID := parseInput(tt.input, tt.defaultURL)
			if gotURL != tt.wantURL {
				t.Errorf("parseInput() gotURL = %v, want %v", gotURL, tt.wantURL)
			}
			if gotID != tt.wantID {
				t.Errorf("parseInput() gotID = %v, want %v", gotID, tt.wantID)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		// Microsecond range (<1000us)
		{
			name:     "zero microseconds",
			input:    0,
			expected: "0us",
		},
		{
			name:     "single microsecond",
			input:    1,
			expected: "1us",
		},
		{
			name:     "small microseconds",
			input:    500,
			expected: "500us",
		},
		{
			name:     "boundary just below milliseconds",
			input:    999,
			expected: "999us",
		},
		// Millisecond range (1000-999999us)
		{
			name:     "exactly one millisecond",
			input:    1000,
			expected: "1.00ms",
		},
		{
			name:     "fractional milliseconds",
			input:    1500,
			expected: "1.50ms",
		},
		{
			name:     "large milliseconds with decimal",
			input:    150500,
			expected: "150.50ms",
		},
		{
			name:     "round milliseconds",
			input:    123000,
			expected: "123.00ms",
		},
		{
			name:     "boundary just below seconds",
			input:    999999,
			expected: "1000.00ms",
		},
		// Second range (>=1000000us)
		{
			name:     "exactly one second",
			input:    1000000,
			expected: "1.00s",
		},
		{
			name:     "fractional seconds",
			input:    2500000,
			expected: "2.50s",
		},
		{
			name:     "large seconds",
			input:    10123456,
			expected: "10.12s",
		},
		{
			name:     "very large duration",
			input:    123456789,
			expected: "123.46s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.input)
			if got != tt.expected {
				t.Errorf("formatDuration(%d) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSpanNode_hasError(t *testing.T) {
	tests := []struct {
		name     string
		tags     []tag
		hasError bool
	}{
		{
			name: "otel.status_code ERROR",
			tags: []tag{
				{Key: "otel.status_code", Value: "ERROR"},
			},
			hasError: true,
		},
		{
			name: "error tag true",
			tags: []tag{
				{Key: "error", Value: true},
			},
			hasError: true,
		},
		{
			name: "error tag false",
			tags: []tag{
				{Key: "error", Value: false},
			},
			hasError: false,
		},
		{
			name: "error tag non-bool",
			tags: []tag{
				{Key: "error", Value: "true"},
			},
			hasError: false,
		},
		{
			name: "otel.status_code OK",
			tags: []tag{
				{Key: "otel.status_code", Value: "OK"},
			},
			hasError: false,
		},
		{
			name:     "no error tags",
			tags:     []tag{{Key: "http.method", Value: "GET"}},
			hasError: false,
		},
		{
			name:     "empty tags",
			tags:     []tag{},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &spanNode{
				span: span{Tags: tt.tags},
			}
			if got := n.hasError(); got != tt.hasError {
				t.Errorf("hasError() = %v, want %v", got, tt.hasError)
			}
		})
	}
}

func TestSpanNode_matchesSelf(t *testing.T) {
	tests := []struct {
		name    string
		node    *spanNode
		cfg     *config
		matches bool
	}{
		{
			name: "minDuration filter - span below threshold",
			node: &spanNode{
				span:    span{Duration: 50_000}, // 50ms
				service: "test-service",
			},
			cfg: &config{
				minDuration: 100000000, // 100ms in nanoseconds
			},
			matches: false,
		},
		{
			name: "minDuration filter - span at threshold",
			node: &spanNode{
				span:    span{Duration: 100_000}, // 100ms
				service: "test-service",
			},
			cfg: &config{
				minDuration: 100000000, // 100ms in nanoseconds
			},
			matches: true,
		},
		{
			name: "minDuration filter - span above threshold",
			node: &spanNode{
				span:    span{Duration: 200_000}, // 200ms
				service: "test-service",
			},
			cfg: &config{
				minDuration: 100000000, // 100ms in nanoseconds
			},
			matches: true,
		},
		{
			name: "minDuration filter - zero duration disabled",
			node: &spanNode{
				span:    span{Duration: 1},
				service: "test-service",
			},
			cfg: &config{
				minDuration: 0,
			},
			matches: true,
		},
		{
			name: "errorsOnly filter - span with error",
			node: &spanNode{
				span: span{
					Duration: 100_000,
					Tags: []tag{
						{Key: "otel.status_code", Value: "ERROR"},
					},
				},
				service: "test-service",
			},
			cfg: &config{
				errorsOnly: true,
			},
			matches: true,
		},
		{
			name: "errorsOnly filter - span without error",
			node: &spanNode{
				span: span{
					Duration: 100_000,
					Tags:     []tag{},
				},
				service: "test-service",
			},
			cfg: &config{
				errorsOnly: true,
			},
			matches: false,
		},
		{
			name: "errorsOnly filter - disabled",
			node: &spanNode{
				span: span{
					Duration: 100_000,
					Tags:     []tag{},
				},
				service: "test-service",
			},
			cfg: &config{
				errorsOnly: false,
			},
			matches: true,
		},
		{
			name: "service filter - matching service",
			node: &spanNode{
				span:    span{Duration: 100_000},
				service: "my-service",
			},
			cfg: &config{
				service: "my-service",
			},
			matches: true,
		},
		{
			name: "service filter - non-matching service",
			node: &spanNode{
				span:    span{Duration: 100_000},
				service: "other-service",
			},
			cfg: &config{
				service: "my-service",
			},
			matches: false,
		},
		{
			name: "service filter - empty string disabled",
			node: &spanNode{
				span:    span{Duration: 100_000},
				service: "any-service",
			},
			cfg: &config{
				service: "",
			},
			matches: true,
		},
		{
			name: "multiple filters - all pass",
			node: &spanNode{
				span: span{
					Duration: 200_000, // 200ms
					Tags: []tag{
						{Key: "error", Value: true},
					},
				},
				service: "my-service",
			},
			cfg: &config{
				minDuration: 100000000, // 100ms in nanoseconds
				errorsOnly:  true,
				service:     "my-service",
			},
			matches: true,
		},
		{
			name: "multiple filters - duration fails",
			node: &spanNode{
				span: span{
					Duration: 50_000, // 50ms
					Tags: []tag{
						{Key: "error", Value: true},
					},
				},
				service: "my-service",
			},
			cfg: &config{
				minDuration: 100000000, // 100ms in nanoseconds
				errorsOnly:  true,
				service:     "my-service",
			},
			matches: false,
		},
		{
			name: "multiple filters - error fails",
			node: &spanNode{
				span: span{
					Duration: 200_000,
					Tags:     []tag{},
				},
				service: "my-service",
			},
			cfg: &config{
				minDuration: 100000000, // 100ms in nanoseconds
				errorsOnly:  true,
				service:     "my-service",
			},
			matches: false,
		},
		{
			name: "multiple filters - service fails",
			node: &spanNode{
				span: span{
					Duration: 200_000,
					Tags: []tag{
						{Key: "error", Value: true},
					},
				},
				service: "other-service",
			},
			cfg: &config{
				minDuration: 100000000, // 100ms in nanoseconds
				errorsOnly:  true,
				service:     "my-service",
			},
			matches: false,
		},
		{
			name: "no filters",
			node: &spanNode{
				span:    span{Duration: 1},
				service: "any-service",
			},
			cfg:     &config{},
			matches: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.matchesSelf(tt.cfg); got != tt.matches {
				t.Errorf("matchesSelf() = %v, want %v", got, tt.matches)
			}
		})
	}
}

func TestBuildTree_SingleRootSpan(t *testing.T) {
	tr := trace{
		TraceID: "trace1",
		Spans: []span{
			{
				SpanID:        "span1",
				OperationName: "root",
				StartTime:     1000,
				Duration:      100,
				ProcessID:     "p1",
			},
		},
		Processes: map[string]process{
			"p1": {ServiceName: "service1"},
		},
	}

	roots, startTime := buildTree(tr)

	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if startTime != 1000 {
		t.Errorf("expected startTime 1000, got %d", startTime)
	}
	if roots[0].span.SpanID != "span1" {
		t.Errorf("expected root span ID 'span1', got %s", roots[0].span.SpanID)
	}
	if roots[0].service != "service1" {
		t.Errorf("expected service 'service1', got %s", roots[0].service)
	}
	if len(roots[0].children) != 0 {
		t.Errorf("expected no children, got %d", len(roots[0].children))
	}
}

func TestBuildTree_ParentChildRelationship(t *testing.T) {
	tr := trace{
		TraceID: "trace1",
		Spans: []span{
			{
				SpanID:        "span1",
				OperationName: "root",
				StartTime:     1000,
				Duration:      200,
				ProcessID:     "p1",
			},
			{
				SpanID:        "span2",
				OperationName: "child",
				References: []reference{
					{RefType: "CHILD_OF", SpanID: "span1"},
				},
				StartTime: 1050,
				Duration:  50,
				ProcessID: "p1",
			},
		},
		Processes: map[string]process{
			"p1": {ServiceName: "service1"},
		},
	}

	roots, startTime := buildTree(tr)

	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if startTime != 1000 {
		t.Errorf("expected startTime 1000, got %d", startTime)
	}
	if roots[0].span.SpanID != "span1" {
		t.Errorf("expected root span ID 'span1', got %s", roots[0].span.SpanID)
	}
	if len(roots[0].children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(roots[0].children))
	}
	if roots[0].children[0].span.SpanID != "span2" {
		t.Errorf("expected child span ID 'span2', got %s", roots[0].children[0].span.SpanID)
	}
}

func TestBuildTree_MultipleRoots(t *testing.T) {
	tr := trace{
		TraceID: "trace1",
		Spans: []span{
			{
				SpanID:        "span1",
				OperationName: "root1",
				StartTime:     1000,
				Duration:      100,
				ProcessID:     "p1",
			},
			{
				SpanID:        "span2",
				OperationName: "root2",
				StartTime:     2000,
				Duration:      100,
				ProcessID:     "p1",
			},
		},
		Processes: map[string]process{
			"p1": {ServiceName: "service1"},
		},
	}

	roots, startTime := buildTree(tr)

	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
	if startTime != 1000 {
		t.Errorf("expected startTime 1000, got %d", startTime)
	}
}

func TestBuildTree_MultiLevelNesting(t *testing.T) {
	tr := trace{
		TraceID: "trace1",
		Spans: []span{
			{
				SpanID:        "span1",
				OperationName: "root",
				StartTime:     1000,
				Duration:      300,
				ProcessID:     "p1",
			},
			{
				SpanID:        "span2",
				OperationName: "child1",
				References: []reference{
					{RefType: "CHILD_OF", SpanID: "span1"},
				},
				StartTime: 1050,
				Duration:  150,
				ProcessID: "p1",
			},
			{
				SpanID:        "span3",
				OperationName: "grandchild",
				References: []reference{
					{RefType: "CHILD_OF", SpanID: "span2"},
				},
				StartTime: 1100,
				Duration:  50,
				ProcessID: "p1",
			},
			{
				SpanID:        "span4",
				OperationName: "child2",
				References: []reference{
					{RefType: "CHILD_OF", SpanID: "span1"},
				},
				StartTime: 1250,
				Duration:  30,
				ProcessID: "p1",
			},
		},
		Processes: map[string]process{
			"p1": {ServiceName: "service1"},
		},
	}

	roots, startTime := buildTree(tr)

	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if startTime != 1000 {
		t.Errorf("expected startTime 1000, got %d", startTime)
	}

	root := roots[0]
	if root.span.SpanID != "span1" {
		t.Errorf("expected root span ID 'span1', got %s", root.span.SpanID)
	}
	if len(root.children) != 2 {
		t.Fatalf("expected 2 children on root, got %d", len(root.children))
	}

	// Verify sorted by startTime
	if root.children[0].span.SpanID != "span2" {
		t.Errorf("expected first child 'span2', got %s", root.children[0].span.SpanID)
	}
	if root.children[1].span.SpanID != "span4" {
		t.Errorf("expected second child 'span4', got %s", root.children[1].span.SpanID)
	}

	// Verify grandchild
	if len(root.children[0].children) != 1 {
		t.Fatalf("expected 1 grandchild, got %d", len(root.children[0].children))
	}
	if root.children[0].children[0].span.SpanID != "span3" {
		t.Errorf("expected grandchild 'span3', got %s", root.children[0].children[0].span.SpanID)
	}
}

func TestBuildTree_OrphanSpanWithMissingParent(t *testing.T) {
	tr := trace{
		TraceID: "trace1",
		Spans: []span{
			{
				SpanID:        "span1",
				OperationName: "orphan",
				References: []reference{
					{RefType: "CHILD_OF", SpanID: "nonexistent"},
				},
				StartTime: 1000,
				Duration:  100,
				ProcessID: "p1",
			},
		},
		Processes: map[string]process{
			"p1": {ServiceName: "service1"},
		},
	}

	roots, startTime := buildTree(tr)

	if len(roots) != 1 {
		t.Fatalf("expected 1 root (orphan), got %d", len(roots))
	}
	if startTime != 1000 {
		t.Errorf("expected startTime 1000, got %d", startTime)
	}
	if roots[0].span.SpanID != "span1" {
		t.Errorf("expected orphan span ID 'span1', got %s", roots[0].span.SpanID)
	}
}

func TestBuildTree_SortedByStartTime(t *testing.T) {
	tr := trace{
		TraceID: "trace1",
		Spans: []span{
			{
				SpanID:        "span3",
				OperationName: "third",
				StartTime:     3000,
				Duration:      100,
				ProcessID:     "p1",
			},
			{
				SpanID:        "span1",
				OperationName: "first",
				StartTime:     1000,
				Duration:      100,
				ProcessID:     "p1",
			},
			{
				SpanID:        "span2",
				OperationName: "second",
				StartTime:     2000,
				Duration:      100,
				ProcessID:     "p1",
			},
		},
		Processes: map[string]process{
			"p1": {ServiceName: "service1"},
		},
	}

	roots, startTime := buildTree(tr)

	if len(roots) != 3 {
		t.Fatalf("expected 3 roots, got %d", len(roots))
	}
	if startTime != 1000 {
		t.Errorf("expected startTime 1000 (earliest), got %d", startTime)
	}

	// Verify roots are sorted
	if roots[0].span.StartTime != 1000 {
		t.Errorf("expected first root startTime 1000, got %d", roots[0].span.StartTime)
	}
	if roots[1].span.StartTime != 2000 {
		t.Errorf("expected second root startTime 2000, got %d", roots[1].span.StartTime)
	}
	if roots[2].span.StartTime != 3000 {
		t.Errorf("expected third root startTime 3000, got %d", roots[2].span.StartTime)
	}
}

func TestBuildTree_ChildrenSortedByStartTime(t *testing.T) {
	tr := trace{
		TraceID: "trace1",
		Spans: []span{
			{
				SpanID:        "span1",
				OperationName: "root",
				StartTime:     1000,
				Duration:      500,
				ProcessID:     "p1",
			},
			{
				SpanID:        "span3",
				OperationName: "child3",
				References: []reference{
					{RefType: "CHILD_OF", SpanID: "span1"},
				},
				StartTime: 1300,
				Duration:  50,
				ProcessID: "p1",
			},
			{
				SpanID:        "span2",
				OperationName: "child2",
				References: []reference{
					{RefType: "CHILD_OF", SpanID: "span1"},
				},
				StartTime: 1200,
				Duration:  50,
				ProcessID: "p1",
			},
			{
				SpanID:        "span4",
				OperationName: "child1",
				References: []reference{
					{RefType: "CHILD_OF", SpanID: "span1"},
				},
				StartTime: 1100,
				Duration:  50,
				ProcessID: "p1",
			},
		},
		Processes: map[string]process{
			"p1": {ServiceName: "service1"},
		},
	}

	roots, _ := buildTree(tr)

	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if len(roots[0].children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(roots[0].children))
	}

	// Verify children are sorted by startTime
	children := roots[0].children
	if children[0].span.StartTime != 1100 {
		t.Errorf("expected first child startTime 1100, got %d", children[0].span.StartTime)
	}
	if children[1].span.StartTime != 1200 {
		t.Errorf("expected second child startTime 1200, got %d", children[1].span.StartTime)
	}
	if children[2].span.StartTime != 1300 {
		t.Errorf("expected third child startTime 1300, got %d", children[2].span.StartTime)
	}
}

func TestBuildTree_MultipleServices(t *testing.T) {
	tr := trace{
		TraceID: "trace1",
		Spans: []span{
			{
				SpanID:        "span1",
				OperationName: "root",
				StartTime:     1000,
				Duration:      200,
				ProcessID:     "p1",
			},
			{
				SpanID:        "span2",
				OperationName: "child",
				References: []reference{
					{RefType: "CHILD_OF", SpanID: "span1"},
				},
				StartTime: 1050,
				Duration:  100,
				ProcessID: "p2",
			},
		},
		Processes: map[string]process{
			"p1": {ServiceName: "frontend"},
			"p2": {ServiceName: "backend"},
		},
	}

	roots, _ := buildTree(tr)

	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if roots[0].service != "frontend" {
		t.Errorf("expected root service 'frontend', got %s", roots[0].service)
	}
	if len(roots[0].children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(roots[0].children))
	}
	if roots[0].children[0].service != "backend" {
		t.Errorf("expected child service 'backend', got %s", roots[0].children[0].service)
	}
}

func TestBuildTree_IgnoreNonChildOfReferences(t *testing.T) {
	tr := trace{
		TraceID: "trace1",
		Spans: []span{
			{
				SpanID:        "span1",
				OperationName: "root",
				StartTime:     1000,
				Duration:      100,
				ProcessID:     "p1",
			},
			{
				SpanID:        "span2",
				OperationName: "sibling",
				References: []reference{
					{RefType: "FOLLOWS_FROM", SpanID: "span1"},
				},
				StartTime: 1100,
				Duration:  50,
				ProcessID: "p1",
			},
		},
		Processes: map[string]process{
			"p1": {ServiceName: "service1"},
		},
	}

	roots, _ := buildTree(tr)

	// FOLLOWS_FROM should not create parent-child relationship
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots (FOLLOWS_FROM ignored), got %d", len(roots))
	}
	if len(roots[0].children) != 0 {
		t.Errorf("expected no children on first root, got %d", len(roots[0].children))
	}
	if len(roots[1].children) != 0 {
		t.Errorf("expected no children on second root, got %d", len(roots[1].children))
	}
}

func TestBuildTree_EmptyTrace(t *testing.T) {
	tr := trace{
		TraceID:   "trace1",
		Spans:     []span{},
		Processes: map[string]process{},
	}

	roots, startTime := buildTree(tr)

	if len(roots) != 0 {
		t.Fatalf("expected 0 roots, got %d", len(roots))
	}
	if startTime != 0 {
		t.Errorf("expected startTime 0, got %d", startTime)
	}
}

func TestBuildTree_MissingProcessID(t *testing.T) {
	tr := trace{
		TraceID: "trace1",
		Spans: []span{
			{
				SpanID:        "span1",
				OperationName: "root",
				StartTime:     1000,
				Duration:      100,
				ProcessID:     "p999",
			},
		},
		Processes: map[string]process{
			"p1": {ServiceName: "service1"},
		},
	}

	roots, _ := buildTree(tr)

	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	// Service should be empty string when ProcessID not found
	if roots[0].service != "" {
		t.Errorf("expected empty service, got %s", roots[0].service)
	}
}

func TestSpanNode_matchesFilter(t *testing.T) {
	tests := []struct {
		name    string
		node    *spanNode
		cfg     *config
		matches bool
	}{
		{
			name: "span matches self",
			node: &spanNode{
				span:    span{Duration: 200_000},
				service: "test-service",
			},
			cfg: &config{
				minDuration: 100000000, // 100ms in nanoseconds
			},
			matches: true,
		},
		{
			name: "span does not match self, no children",
			node: &spanNode{
				span:     span{Duration: 50_000},
				service:  "test-service",
				children: []*spanNode{},
			},
			cfg: &config{
				minDuration: 100000000, // 100ms in nanoseconds
			},
			matches: false,
		},
		{
			name: "parent does not match self, but child matches",
			node: &spanNode{
				span:    span{Duration: 50_000},
				service: "parent-service",
				children: []*spanNode{
					{
						span:    span{Duration: 200_000},
						service: "child-service",
					},
				},
			},
			cfg: &config{
				minDuration: 100000000, // 100ms in nanoseconds
			},
			matches: true,
		},
		{
			name: "parent does not match, grandchild matches",
			node: &spanNode{
				span:    span{Duration: 50_000},
				service: "parent-service",
				children: []*spanNode{
					{
						span:    span{Duration: 60_000},
						service: "child-service",
						children: []*spanNode{
							{
								span:    span{Duration: 200_000},
								service: "grandchild-service",
							},
						},
					},
				},
			},
			cfg: &config{
				minDuration: 100000000, // 100ms in nanoseconds
			},
			matches: true,
		},
		{
			name: "parent has error, child does not match service filter",
			node: &spanNode{
				span: span{
					Duration: 100_000,
					Tags: []tag{
						{Key: "error", Value: true},
					},
				},
				service: "target-service",
				children: []*spanNode{
					{
						span:    span{Duration: 100_000},
						service: "other-service",
					},
				},
			},
			cfg: &config{
				service: "target-service",
			},
			matches: true,
		},
		{
			name: "parent does not match service, child matches service",
			node: &spanNode{
				span:    span{Duration: 100_000},
				service: "other-service",
				children: []*spanNode{
					{
						span:    span{Duration: 100_000},
						service: "target-service",
					},
				},
			},
			cfg: &config{
				service: "target-service",
			},
			matches: true,
		},
		{
			name: "parent without error, child has error (errorsOnly filter)",
			node: &spanNode{
				span:    span{Duration: 100_000, Tags: []tag{}},
				service: "parent-service",
				children: []*spanNode{
					{
						span: span{
							Duration: 100_000,
							Tags: []tag{
								{Key: "otel.status_code", Value: "ERROR"},
							},
						},
						service: "child-service",
					},
				},
			},
			cfg: &config{
				errorsOnly: true,
			},
			matches: true,
		},
		{
			name: "neither parent nor children match",
			node: &spanNode{
				span:    span{Duration: 50_000},
				service: "parent-service",
				children: []*spanNode{
					{
						span:    span{Duration: 60_000},
						service: "child-service",
					},
					{
						span:    span{Duration: 70_000},
						service: "child-service-2",
					},
				},
			},
			cfg: &config{
				minDuration: 100000000, // 100ms in nanoseconds
			},
			matches: false,
		},
		{
			name: "one child matches among multiple children",
			node: &spanNode{
				span:    span{Duration: 50_000},
				service: "parent-service",
				children: []*spanNode{
					{
						span:    span{Duration: 60_000},
						service: "child-service-1",
					},
					{
						span:    span{Duration: 200_000},
						service: "child-service-2",
					},
					{
						span:    span{Duration: 70_000},
						service: "child-service-3",
					},
				},
			},
			cfg: &config{
				minDuration: 100000000, // 100ms in nanoseconds
			},
			matches: true,
		},
		{
			name: "complex tree - deep match",
			node: &spanNode{
				span:    span{Duration: 10_000},
				service: "root",
				children: []*spanNode{
					{
						span:    span{Duration: 20_000},
						service: "level1-a",
						children: []*spanNode{
							{
								span:    span{Duration: 30_000},
								service: "level2-a",
							},
						},
					},
					{
						span:    span{Duration: 25_000},
						service: "level1-b",
						children: []*spanNode{
							{
								span:    span{Duration: 35_000},
								service: "level2-b",
								children: []*spanNode{
									{
										span:    span{Duration: 200_000},
										service: "level3-b",
									},
								},
							},
						},
					},
				},
			},
			cfg: &config{
				minDuration: 100000000, // 100ms in nanoseconds
			},
			matches: true,
		},
		{
			name: "no filters - always matches",
			node: &spanNode{
				span:    span{Duration: 1},
				service: "any-service",
			},
			cfg:     &config{},
			matches: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.matchesFilter(tt.cfg); got != tt.matches {
				t.Errorf("matchesFilter() = %v, want %v", got, tt.matches)
			}
		})
	}
}
