package manager

import (
    "math/rand"
)

//представляет координаты вершины
type Vertex struct {
    X int `json:"x"`
    Y int `json:"y"`
}

//представляет ребро между двумя вершинами
type Edge struct {
    From int `json:"from"`
    To   int `json:"to"`
}

//представляет данные графа для передачи на фронт
type GraphData struct {
    VerticesCount int      `json:"vertices_count"`
    EdgesCount    int      `json:"edges_count"`
    Vertices      []Vertex `json:"vertices"`
    Edges         []Edge   `json:"edges"`
}


//создаёт связный граф с verticesCount вершинами
func GenerateConnectedGraph(verticesCount int) *GraphData {
    edges := make([]Edge, 0)
    
    // оставное дерево
    // соединяем каждую новую вершину со случайной из уже добавленных
    for i := 1; i < verticesCount; i++ {
        from := i
        to := rand.Intn(i) // случайная вершина из уже добавленных 
        edges = append(edges, Edge{From: from, To: to})
    }
    
    // случайно добавляем дополнительные рёбра
    // от 0 до verticesCount-1
    extraEdgesCount := rand.Intn(verticesCount)
    for i := 0; i < extraEdgesCount; i++ {
        from := rand.Intn(verticesCount)
        to := rand.Intn(verticesCount)
        if from == to {
            continue // пропускаем петли
        }
        // не существует ли уже такое ребро?
        if !edgeExists(edges, from, to) {
            edges = append(edges, Edge{From: from, To: to})
        }
    }
    
    // координаты
    vertices := make([]Vertex, verticesCount)
    for i := 0; i < verticesCount; i++ {
        vertices[i] = Vertex{
            X: rand.Intn(1001),
            Y: rand.Intn(1001),
        }
    }
    
    return &GraphData{
        VerticesCount: verticesCount,
        EdgesCount:    len(edges),
        Vertices:      vertices,
        Edges:         edges,
    }
}

// проверяет, есть ли уже ребро в списке
func edgeExists(edges []Edge, from, to int) bool {
    for _, e := range edges {
        if (e.From == from && e.To == to) || (e.From == to && e.To == from) {
            return true
        }
    }
    return false
}


