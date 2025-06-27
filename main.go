package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/wikilink"
)

type Link struct {
	SourceId string
	TargetId string
}

type Category struct {
	Name string
	Id   int
}

type Node struct {
	Id       string
	Name     string
	Category int
	Value    string
	Links    []string
}

type Graph struct {
	Nodes      []Node
	Links      []Link
	Categories []Category
}

type GraphString struct {
	Nodes      []string
	Links      []string
	Categories []string
}

func main() {
	paths := getNotePaths()
	if len(paths) == 0 {
		return
	}

	f, err := os.Create("graph.html")
	if err != nil {
		log.Fatalf("failed to create graph: %v", err)
	}
	defer f.Close()

	graphData := parseNotes(paths)
	graphString := graphData.toGraphStrings()
	htmlGraph := graphString.generateHTML()

	if err := base(htmlGraph).Render(context.Background(), f); err != nil {
		log.Fatalf("failed to write output file: %v", err)
	}
}

func getNotePaths() []string {
	var paths []string
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("error accessing path %q: %v", path, err)
			return nil
		}
		if !d.IsDir() && isMarkdownFile(path) {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		log.Printf("error walking directory: %v", err)
	}
	return paths
}

func isMarkdownFile(file string) bool {
	return filepath.Ext(file) == ".md"
}

func parseNotes(paths []string) Graph {
	var links []Link
	var nodes []Node
	categoryMap := make(map[string]int)

	for _, path := range paths {
		node, ok := newNode(path, categoryMap)
		if !ok {
			continue
		}
		nodes = append(nodes, node)

		for _, link := range node.Links {
			links = append(links, Link{SourceId: node.Id, TargetId: link + ".md"})
		}
	}

	return Graph{
		Nodes:      nodes,
		Links:      links,
		Categories: sortCategories(categoryMap),
	}
}

func sortCategories(categoryMap map[string]int) []Category {
	var categories []Category
	for name, id := range categoryMap {
		categories = append(categories, Category{Name: name, Id: id})
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Id < categories[j].Id
	})
	return categories
}

func newNode(path string, categoryMap map[string]int) (Node, bool) {
	source, err := os.ReadFile(path)
	if err != nil {
		log.Printf("error reading file %s: %v", path, err)
		return Node{}, false
	}

	md := goldmark.New(
		goldmark.WithExtensions(&wikilink.Extender{}),
	)
	reader := text.NewReader(source)
	parser := md.Parser()

	node := extractContent(parser.Parse(reader), source)
	node.Id = path

	category := filepath.Dir(path)
	if _, exists := categoryMap[category]; !exists {
		categoryMap[category] = len(categoryMap)
	}
	node.Category = categoryMap[category]

	return node, true
}

func extractContent(doc ast.Node, source []byte) Node {
	var node Node
	var header string
	var valueLines []string

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := n.(type) {
		case *ast.Text:
			valueLines = append(valueLines, string(n.Segment.Value(source)))

		case *ast.Heading:
			if n.Level == 1 && header == "" {
				var headerText string
				for c := n.FirstChild(); c != nil; c = c.NextSibling() {
					if text, ok := c.(*ast.Text); ok {
						headerText += string(text.Segment.Value(source))
					}
				}
				header = headerText
			}

		case *wikilink.Node:
			node.Links = append(node.Links, string(n.Target))
		}

		return ast.WalkContinue, nil
	})

	node.Value = strings.Join(valueLines, "\n")
	node.Name = header
	return node
}

func (g Graph) toGraphStrings() GraphString {
	return GraphString{
		Nodes:      nodesToStrings(g.Nodes),
		Links:      linksToStrings(g.Links),
		Categories: categoriesToStrings(g.Categories),
	}
}

func nodesToStrings(nodes []Node) []string {
	var result []string
	for _, node := range nodes {
		sanitized := strings.ReplaceAll(node.Value, "\n", "\\n")
		nodeStr := fmt.Sprintf(`{ id: "%s", name: "%s", category: %d, value: "%s" }`,
			node.Id, node.Name, node.Category, sanitized)
		result = append(result, nodeStr)
	}
	return result
}

func linksToStrings(links []Link) []string {
	var result []string
	for _, link := range links {
		result = append(result, fmt.Sprintf(`{ source: "%s", target: "%s" }`, link.SourceId, link.TargetId))
	}
	return result
}

func categoriesToStrings(categories []Category) []string {
	var result []string
	for _, category := range categories {
		result = append(result, fmt.Sprintf(`{ name: "%s" }`, category.Name))
	}
	return result
}

func (g GraphString) generateHTML() string {
	return fmt.Sprintf(`
<script type="text/javascript">
	var myChart = echarts.init(document.getElementById('main'));

	var graph = {
		nodes: [%s],
		links: [%s],
		categories: [%s]
	};

	graph.nodes.forEach(function (node) {
		node.symbolSize = 8;
	});

	var option = {
		color: ['#b4637a', '#f6c177', '#ea9a97', '#3e8fb0', '#9ccfd8', '#c4a7e7'],
		tooltip: {
			backgroundColor: '#1f2937',
			borderColor: '#374151',
			textStyle: {
				color: '#ffffff',
				fontSize: 14,
				fontFamily: 'Inclusive Sans, sans-serif'
			},
			formatter: function (params) {
				return params.data.value ? params.data.value.replace(/\\n/g, '<br>') : '';
			}
		},
		legend: [{
			data: graph.categories.map(a => a.name),
			textStyle: {
				color: '#ffffff',
				fontSize: 14,
				fontFamily: 'Inclusive Sans, sans-serif'
			}
		}],
		series: [{
			type: 'graph',
			layout: 'force',
			data: graph.nodes,
			links: graph.links,
			categories: graph.categories,
			roam: true,
			label: {
				show: true,
				position: 'right',
				color: '#ffffff',
				fontSize: 12,
				fontFamily: 'Inclusive Sans, sans-serif',
				formatter: function(params) {
					return params.data.name;
				}
			},
			emphasis: {
				focus: 'adjacency',
				blurScope: 'coordinateSystem'
			},
			force: {
				repulsion: 300,
				gravity: 0.1,
				edgeLength: 100,
				layoutAnimation: true
			},
			zoom: 1.5,
			center: ['50%%', '50%%']
		}]
	};

	myChart.setOption(option);

	window.addEventListener('resize', function () {
		myChart.resize();
	});
</script>
`,
		strings.Join(g.Nodes, ", "),
		strings.Join(g.Links, ", "),
		strings.Join(g.Categories, ", "),
	)
}
