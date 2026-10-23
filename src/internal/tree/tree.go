package tree

import "fmt"

type Entry string

const (
	File Entry = "file"
	Dir        = "directory"
)

type entry struct {
	Type Entry
	Name string
	Hash string
}

type Node struct {
	Children []*Node
	Entry    *entry
	Parent   *Node //TODO Either make GetPath() if needed or remove this
}

func (n *Node) IsDir() bool {
	return n.Entry.Type == Dir
}

func New(root string) *Node {
	e := &entry{
		Type: Dir,
		Name: root,
	}

	return &Node{
		Entry: e,
	}
}

func (n *Node) AddChildren(node ...*Node) {
	for _, val := range node {
		val.Parent = n
		n.Children = append(n.Children, val)
	}

	//n.Children = append(n.Children, node...) - If not implementing GetPath()
}

func (n *Node) AddHash(hash string) {
	n.Entry.Hash = hash
}

func Serialize(node *Node) []byte {
	var serial []byte
	serial = fmt.Appendf(serial, "%s %s\x00%s", node.Entry.Type, node.Entry.Name, node.Entry.Hash)
	return serial
}
