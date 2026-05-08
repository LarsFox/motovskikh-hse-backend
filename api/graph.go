package api

import (
	"math/rand/v2"
	"net/http"
)

type graphVertex struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type graphEdge struct {
	From int `json:"from"`
	To   int `json:"to"`
}


func (m *Manager) hndlrGraph(w http.ResponseWriter, r *http.Request) {
	const (
		vertexCount = 12
		edgeCount   = 15
		maxCoord    = 1000
	)

	vertices := make([]graphVertex, 0, vertexCount)
	for range vertexCount {
		vertices = append(vertices, graphVertex{
			X: rand.IntN(maxCoord + 1),
			Y: rand.IntN(maxCoord + 1),
		})
	}

	edges := make([]graphEdge, 0, edgeCount)
	for len(edges) < edgeCount {
		from := rand.IntN(vertexCount)
		to := rand.IntN(vertexCount)
		if from == to {
			continue
		}
		edges = append(edges, graphEdge{From: from, To: to})
	}

	m.send(w, map[string]any{
		"vertices": vertices,
		"edges":    edges,
	})
}

