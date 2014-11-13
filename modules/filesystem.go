package modules


// File System Node for directory trees
type FsNode struct{
    Nodes []*FsNode
    Name string
    Hash string
}