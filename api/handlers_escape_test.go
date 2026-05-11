package api

import "testing"

func TestValidateEulerianPath_AllowsContinuationFromEitherEndOfFirstEdge(t *testing.T) {
	graph := escapeGraph{
		Vertices: []escapeVertex{{}, {}, {}},
		Edges: []escapeEdge{
			{ID: "e0", From: 0, To: 1},
			{ID: "e1", From: 0, To: 2},
		},
	}

	isValid, isComplete, _ := validateEulerianPath(graph, []string{"e0", "e1"})
	if !isValid {
		t.Fatalf("expected path to be valid")
	}
	if !isComplete {
		t.Fatalf("expected path to be complete")
	}
}

func TestValidateEulerianPath_RejectsDisconnectedContinuation(t *testing.T) {
	graph := escapeGraph{
		Vertices: []escapeVertex{{}, {}, {}, {}},
		Edges: []escapeEdge{
			{ID: "e0", From: 0, To: 1},
			{ID: "e1", From: 2, To: 3},
		},
	}

	isValid, isComplete, _ := validateEulerianPath(graph, []string{"e0", "e1"})
	if isValid {
		t.Fatalf("expected path to be invalid")
	}
	if isComplete {
		t.Fatalf("expected disconnected path to be incomplete")
	}
}

func TestValidateEulerianPath_ReturnsCompleteForFullValidEulerPath(t *testing.T) {
	graph := escapeGraph{
		Vertices: []escapeVertex{{}, {}, {}, {}},
		Edges: []escapeEdge{
			{ID: "e0", From: 0, To: 1},
			{ID: "e1", From: 1, To: 2},
			{ID: "e2", From: 2, To: 3},
		},
	}

	isValid, isComplete, _ := validateEulerianPath(graph, []string{"e0", "e1", "e2"})
	if !isValid {
		t.Fatalf("expected path to be valid")
	}
	if !isComplete {
		t.Fatalf("expected path to be complete")
	}
}

func TestValidateEulerianPath_RejectsRepeatedEdge(t *testing.T) {
	graph := escapeGraph{
		Vertices: []escapeVertex{{}, {}, {}},
		Edges: []escapeEdge{
			{ID: "e0", From: 0, To: 1},
			{ID: "e1", From: 1, To: 2},
		},
	}

	isValid, isComplete, _ := validateEulerianPath(graph, []string{"e0", "e0"})
	if isValid {
		t.Fatalf("expected repeated edge path to be invalid")
	}
	if isComplete {
		t.Fatalf("expected repeated edge path to be incomplete")
	}
}
