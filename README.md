<img src="demo/logo.png" align="left" width="100" />

# noteGraph

noteGraph is a simple CLI tool written in Go that renders your Markdown notes into an interactive HTML graph.
It visualizes connections via Wikilinks and mirrors the hierarchical directory structure of your notes.

## 📂 Demo Preview

```
├── cats
│   ├── behaviour-territory.md
│   └── companions.md
├── dogs
│   ├── domesticated.md
│   └── intelligence.md
├── spiders
│   ├── ecological-controllers.md
│   └── spider-silk.md
└── graph.html
```
## ✨ Example Output

graph view:

![demo image](https://github.com/hendr1km/noteGraph/blob/main/demo/screen0.png)

Hovering over a node displays the note content:

![demo image](https://github.com/hendr1km/noteGraph/blob/main/demo/screen1.png)


## 🚀 Usage

To use this tool, simply clone the repository and build the binary using Go:
```
go build -o graph
```
You can then either:
Move the graph binary into your `~/go/bin folder` and run it globally using the `graph` command, or 
Run it directly from your local notes directory using:
```
./graph
```
This command will generate a graph.html file in the current directory, containing a visual representation of your notes with Wikilink connections and folder-based categorization.
