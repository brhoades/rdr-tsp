package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"

	log "github.com/inconshreveable/log15"
	"github.com/spf13/cobra"
)

type Node struct {
	Day  int
	X    float64
	Y    float64
	Name string
}

type Graph struct {
	Nodes []*Node
}

type ProblemGraph struct {
	Graphs      [3]*Graph
	NodesByName map[string]([]*Node)
}

func distance(a *Node, b *Node) float64 {
	return math.Abs(a.X-b.X) + math.Abs(a.Y-b.Y)
}

func getGraph(filename string) (*ProblemGraph, error) {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	r := csv.NewReader(bytes.NewReader(data))
	p := ProblemGraph{
		NodesByName: make(map[string]([]*Node)),
	}
	for i := 0; i < 3; i++ {
		p.Graphs[i] = &Graph{}
	}

	records, err := r.ReadAll()

	if err != nil {
		return nil, err
	}

	for i, r := range records {
		var err error
		var n Node

		n.Name = r[0]
		if n.X, err = strconv.ParseFloat(r[1], 64); err != nil {
			return nil, fmt.Errorf(`Error parsing X coordinate of "%s" on line %d`, r[1], i)
		}
		if n.Y, err = strconv.ParseFloat(r[2], 64); err != nil {
			return nil, fmt.Errorf(`Error parsing Y coordinate of "%s" on line %d`, r[2], i)
		}
		if n.Day, err = strconv.Atoi(r[3]); err != nil {
			return nil, fmt.Errorf(`Error parsing day "%s" on line %d`, r[3], i)
		}

		p.NodesByName[n.Name] = append(p.NodesByName[n.Name], &n)
		p.Graphs[n.Day].Nodes = append(p.Graphs[n.Day].Nodes, &n)
	}

	return &p, nil
}

func main() {
	(&cobra.Command{
		Use:   "rdrtsp <path to csv file> <start>",
		Short: "Calculate the shortest path to visit provided points from the specified starting location.",
		Run: func(cmd *cobra.Command, args []string) {
			p, err := getGraph(args[0])

			if err != nil {
				log.Error("Error parsing CSV.", err)
			}

			fmt.Printf("Loaded %d locations, day 1: %d, day 2: %d, day 3: %d.\n", len(p.NodesByName), len(p.Graphs[0].Nodes), len(p.Graphs[1].Nodes), len(p.Graphs[2].Nodes))
			start := p.NodesByName[args[1]][0]
			remainder := filterNodes([]*Node{start}, p.Graphs[0].Nodes)
			log.Debug("Start nodes", "start", start)
			path, _ := findPath(p.NodesByName[args[1]], remainder, p, 0, 100000000)

			for _, p := range path {
				fmt.Println(p.Name)
			}
		},
		Args: cobra.ExactArgs(2),
	}).Execute()
}

func findPath(visited []*Node, options []*Node, p *ProblemGraph, day int, lastBest float64) ([]*Node, float64) {
	// log.Debug("CALL", "visited", len(visited), "options", len(options))
	if len(visited)/18 >= day+1 && day != 2 {
		log.Debug("Extra day...", "day", day)
		day++
		options = filterNodes(visited, p.Graphs[day].Nodes)
	}

	var bestPath []*Node
	var bestWeight float64 = 10000000

	if len(options) == 1 {
		visited = append(visited, options[0])
		dist := totalDistance(visited)

		// log.Debug("One option left, bottomed out.", "totalDistance", dist)
		return visited, dist
	}

	var nVisited []*Node

	for _, next := range options {
		nVisited = append(visited[:], next)
		if len(nVisited) > 1 {
			// log.Debug("Evaluating next node", "nVisited", nVisited[len(nVisited)-1])
		}

		_, weight := findPath(nVisited, filterNodes(nVisited, options), p, day, bestWeight)
		if weight < bestWeight {
			bestWeight = weight
			bestPath = nVisited
		}
	}

	if bestWeight > lastBest {
		return bestPath, bestWeight
	}

	// log.Debug("Found best for path.", "bestPath", bestPath, "bestWeight", bestWeight)
	visited = bestPath
	options = filterNodes(options, nVisited)

	return visited, totalDistance(visited)
}

func totalDistance(path []*Node) float64 {
	if len(path) < 2 {
		return 0.0
	}
	sum := 0.0

	for i, x := range path[:len(path)-1] {
		sum += distance(x, path[i+1])
	}

	return sum
}

func filterNodes(remove []*Node, from []*Node) []*Node {
	ret := make([]*Node, 0, len(from))

	for _, f := range from {
		found := false

		for _, r := range remove {
			if f.Name == r.Name {
				found = true
				break
			}
		}

		if !found {
			ret = append(ret, f)
		}
	}

	return ret
}
