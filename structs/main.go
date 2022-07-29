package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

type DependencyLoader interface {
	GetDependencyGraph(options ...DependencyLoaderOption) (*DependencyGraph, error)
}

var _ DependencyLoader = (*FileDependencyGraph)(nil)

type FileDependencyGraph struct {
	// FilePath location on disk where the dependency graph is located
	FilePath string
}

// DependencyGraph is the directed graph of dependent applications
type DependencyGraph struct {
	Graph []DependencyList `json:"graph"`
}

// DependencyList contains all the dependencies for a given application
// dependencies can be both applications and generalized providers
type DependencyList struct {
	Name         string   `json:"name"`
	Dependencies []string `json:"dependencies"`
}

// GetApplicationList returns the complete list of applications required, based on the initial list
// basically, perform a breadth first search in the directed graph, and find all nodes accessible from the initial ones
// nodes or vertices can be used interchangeably
// first transform the dependency graph into a more suitable map of maps, then perform the BFS
// we presume all data structures were validated before receiving them, so we don't need to check again
func (d *DependencyGraph) GetApplicationList(initialList []string) ([]string, error) {
	var graphMap = map[string]map[string]bool{}

	// graphMap has vertices on the first level, and edges on the second level
	// classic Directed Graph data structure
	for _, v := range d.Graph {
		g := map[string]bool{}
		graphMap[v.Name] = g

		for _, app := range v.Dependencies {
			g[app] = true
		}
	}

	return bfs(graphMap, initialList)
}

// bfs performs the breadth first search of the graph, starting with the initial vertex list found in the initialList
// there is no built-in queue in go, but we'll use the initialList provided as input
func bfs(graph map[string]map[string]bool, initialList []string) ([]string, error) {
	var (
		visited = map[string]bool{}
		ret     []string
	)

	//if len(initialList) == 0 {
	//	return []string{}, nil
	//}

	for _, vertex := range initialList {
		visited[vertex] = true
	}

	for len(initialList) != 0 {
		crtVertex := initialList[0]
		initialList = initialList[1:]

		ret = append(ret, crtVertex)

		// check if the current vertex has edges originating from it
		// iterate over the edges and update the lists as required
		if edges, ok := graph[crtVertex]; ok {
			for v := range edges {
				if _, ok := visited[v]; !ok {
					initialList = append(initialList, v)
					visited[v] = true
				}
			}
		}
	}

	return ret, nil
}

// DependencyLoaderOption defines a functional option for DependencyLoader
type DependencyLoaderOption func(*DependencyGraph) error

// WithUniqueApplicationNames returns a validator for application name uniqueness
func WithUniqueApplicationNames() DependencyLoaderOption {
	return func(graph *DependencyGraph) error {
		var nameSet = map[string]bool{}
		for _, app := range graph.Graph {
			if _, ok := nameSet[app.Name]; ok {
				return fmt.Errorf("duplicate application name in dependency graph: %s", app.Name)
			}

			nameSet[app.Name] = true
		}

		log.WithFields(log.Fields{"applications": nameSet}).Debug("validated application name uniqueness")

		return nil
	}
}

// WithValidatedDependencies returns a validator for dependency names matching a list of known providers
// go doesn't have a built-in "contains" functionality like Python. For now, we'll transform the list into a set and check for existence in O(1)
// Alternatively, we can adopt the O(n) search, iterating over the list for every dependency, but remove the initial memory allocation for the hash map
func WithValidatedDependencies(providers []fmt.Stringer) DependencyLoaderOption {
	return func(graph *DependencyGraph) error {
		var providerSet = map[string]bool{}
		for _, p := range providers {
			providerSet[p.String()] = true
		}

		log.WithField("providers", providerSet).Debugf("validating that dependencies are part of providers set")

		// iterate over all nodes, and then over all dependencies for each of the nodes
		// validate that the dependencies are part of the provider list
		for _, app := range graph.Graph {
			for _, dependency := range app.Dependencies {
				if _, ok := providerSet[dependency]; !ok {
					return fmt.Errorf("dependency not found in the provider set: %s", dependency)
				}
			}
		}

		return nil
	}
}

// GetDependencyGraph returns the DependencyGraph object, applying all the loader options
// functional options can be passed to perform validations on the loaded graph
// an error is returned if any of the validators fail
func (f *FileDependencyGraph) GetDependencyGraph(options ...DependencyLoaderOption) (*DependencyGraph, error) {
	var err error

	var ret *DependencyGraph

	log.Debugf("loading file from path %s", f.FilePath)

	ret, err = loadFromFile(f.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load from file %s: %w", f.FilePath, err)
	}

	log.Infof("loaded data from file from path %s", f.FilePath)

	for _, opt := range options {
		if err := opt(ret); err != nil {
			return nil, fmt.Errorf("validation for dependency graph failed: %w", err)
		}
	}

	return ret, nil
}

// loadFromFile takes a filename and deserializes the contents into a DependencyGraph object
func loadFromFile(filename string) (*DependencyGraph, error) {
	log.Debugf("loading dependency graph from file: %s", filename)

	graphBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read from file %s: %w", filename, err)
	}

	graph, err := load(graphBytes)
	if err != nil {
		return nil, err
	}

	log.Infof("loaded dependecy graph from file: %s", filename)

	return graph, nil
}

// load takes a byte slice and deserializes the contents into a DependencyGraph object.
// Encapsulates deserialization without assuming the source is a file.
func load(data []byte) (*DependencyGraph, error) {
	rawGraph := DependencyGraph{}

	// if there's no data passed as input, return the empty object instead of failing
	if len(data) == 0 {
		return &rawGraph, nil
	}

	err := json.Unmarshal(data, &rawGraph)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall data: %w", err)
	}

	return &rawGraph, nil
}

func main() {
	fmt.Printf("hello world")
	l := FileDependencyGraph{
		FilePath: "deps.json",
	}
	g, _ := l.GetDependencyGraph()
	fmt.Printf("%#v\n", g)
	list, _ := g.GetApplicationList([]string{"orchestrator"})
	fmt.Printf("%#v\n", list)
}
