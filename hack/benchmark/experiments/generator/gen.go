package main

import (
	"io/ioutil"

	"sigs.k8s.io/yaml"

	"istio.io/tools/isotope/convert/pkg/graph"
	"istio.io/tools/isotope/convert/pkg/graphviz"
)

const (
  TOPOLOGY_YAML = "topology.yaml"
  GRAPHVIZ_OUT = "topology.dot"
)

func createGraphviz(serviceGraph graph.ServiceGraph) {
  dotLang, err := graphviz.ServiceGraphToDotLanguage(serviceGraph)
  if err != nil {
    panic(err)
  }

  err = ioutil.WriteFile(GRAPHVIZ_OUT, []byte(dotLang), 0o644)
  if err != nil {
    panic(err)
  }
}

func main() {
  yamlContents, err := ioutil.ReadFile(TOPOLOGY_YAML)
  if err != nil {
    panic(err)
  }

  var serviceGraph graph.ServiceGraph
  err = yaml.Unmarshal(yamlContents, &serviceGraph)
  if err != nil {
    panic(err)
  }

  // Create graphviz
  createGraphviz(serviceGraph)
}
