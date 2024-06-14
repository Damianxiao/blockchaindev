package trie

import (
	"blockchaindev/crypto/sha3"
	"blockchaindev/kvstore"
	"bytes"
	"encoding/hex"
	"hash"
	"math/big"
	"sort"
	"strings"
)

// if state is nil
var emptyRoot = hash.BigToHash(big.NewInt(0))

// trie struct
type ITrie interface {
	Store(key, value []byte) error

	Root() hash.Hash
	// put a key get a value of account
	Load(key []byte) ([]byte, error)
}

// state is the entrance of trie , provide the root of trie
type State struct {
	root *TrieNode
	db   kvstore.KVDatabase
}

// trie node struct
type TrieNode struct {
	// key []byte
	Path     string
	Leaf     bool
	value    hash.Hash
	Children Children
}

// define a new name for []Child
type Children []Child

type Child struct {
	Path string
	Hash hash.Hash
}

func (Children Children) Len() int {
	return len(Children)
}

func (Children Children) Less(i, j int) bool {
	return Children[i].Path < Children[j].Path
}

func (Children Children) Swap(i, j int) {
	Children[i], Children[j] = Children[j], Children[i]
}

func NewTrieNode() *TrieNode {
	return &TrieNode{}
}

// create a new trieTree or retrieve the trieTree from db
func NewState(db kvstore.KVDatabase, root hash.Hash) *State {
	if bytes.Equal(root[:], emptyRoot[:]) {
		return &State{
			db:   db,
			root: NewTrieNode(),
		}
	} else {
		value, err := db.Get(root[:])
		if err != nil {
			return nil
		}
		node, err := TrieNodeFromBytes(value)
		if err != nil {
			panic(err)
		}
		return &State{
			db:   db,
			root: node,
		}
	}
}

// get node by bytes
func TrieNodeFromBytes(data []byte) *TrieNode {
	var node TrieNode
	err := rlp.DecodeBytes(data, &node)
	if err != nil {
		return nil, err
	}
	return &node, err
}

func NewTrieNode() *TrieNode {
	return &TrieNode{}
}

func (node *TrieNode) sort() {
	sort.Sort(node.Children)
}

func (node TrieNode) Bytes() []byte {
	return rlp.EncodeToBytes(node)
}

func (node TrieNode) Hash() hash.Hash {
	data := node.Bytes()
	return sha3.Keccak256(data)
}

func (state State) Root() hash.Hash {
	return state.root.Hash()
}

func (state *State) Store(key, value []byte) error {
	// encode to string find the path
	path := hex.EncodeToString(key)
	paths, hashes := state.FindParents(path)

	hash := sha3.Sha3(value)
	state.db.Put(hash[:], value)
	// update all parents

	return nil
}

func (state State) TrieNodeFromHash(hash hash.Hash) *TrieNode {
	data, err := state.db.Get(hash[:])
	if err != nil {
		return nil
	}
	return TrieNodeFromBytes(data)
}

func (state *State) SaveTrieNode(node TrieNode) {
	hash := node.Hash()
	state.db.Put(hash[:], node.Bytes())
}

func (state *State) UpdateParents(path string, hash hash.Hash, paths []string, hashes []hash.Hash) {
	// connect the whole paths to one string
	prefix := strings.Join(paths, "")
	depth := len(paths)

	// if the path is the same as the prefix , prove the path is complete match a single node
	if strings.EqualFold(path, prefix) {
		// update, because the path is the same , we do update , not insert
		node := state.TrieNodeFromHash(hashes[depth-1]) // get the leaf
		node.value = hash
		state.SaveTrieNode(*node)
		childHash := node.Hash()
		// from the last second node to the root
		for i := depth - 2; i >= 0; i-- {
			node := state.TrieNodeFromHash(hashes[i])
			// update the children collection
			for _, child := range node.Children {
				if child.Path == paths[i] {
					child.Hash = childHash
					state.SaveTrieNode(*node)
					// this Hash() should calculate the new hash of the node include the new child
					childHash = node.Hash()
					path = child.Path
					break
				}
			}
			if i == 0 {
				state.root = node
			}
			// modify the value of the node
			node.value = childHash
			state.SaveTrieNode(*node)
			childHash = node.Hash()
		}
	} else {
		// the node is not exist ,do insert. use hashes get the lastnode of the new node
		lastNode := state.TrieNodeFromHash(hashes[depth-1])

		if len(lastNode.Path) != len(paths[depth-1]) {
			// need fork
			prefix := strings.Join(paths, "")
			// slice the part of the path
			leafPath := path[len(prefix):]
			node := NewTrieNode()
			node.Path = leafPath
			node.Leaf = true
			node.value = hash
			// save to db
			state.SaveTrieNode(*node)
			// update the lastNode to add the new child
			lastNode.Children = append(lastNode.Children, Child{Path: leafPath, Hash: node.Hash()})
			// sort the children
			lastNode.sort()
			SaveTrieNode(*lastNode)
			// update the parents
			childPath := lastNode.Path
			childHash := lastNode.Hash()
			for i := depth - 2; i >= 0; i-- {
				node := state.TrieNodeFromHash(hashes[i])
				for _,child := range node.Children {
					if child.Path == childPath {
						child.Hash = childHash
						state.SaveTrieNode(*node)
						childHash = node.Hash()
						childPath = node.Path
						break
					}
				}
				if i == 0 {
					state.root = node
				}
		} else {
			//insert
		}

	}
}

// return the array of paths and hashes
func (state State) FindParents(path string) ([]string, []hash.Hash) {
	// from the root to the leaf
	current := state.root
	paths, hashes := make([]string, 0), make([]hash.Hash, 0)
	paths = append(paths, "")
	hashes = append(hashes, emptyRoot)
	prefix := ""
	for {
		flag := false
		for _, child := range current.Children {
			tmp := prefix + child.Path
			length := lengthPrefix(prefix, tmp)
			if length == len(prefix) {
				prefix += child.Path
				paths = append(paths, child.Path)
				hashes = append(hashes, child.Hash)
				flag = true
				data, _ := state.db.Get(child.Hash[:])
				current = TrieNodeFromBytes(data)
				break
			} else if length > len(prefix) {
				// match partially
				prefix = prefix[:length]
				paths = append(paths, child.Path[:length-len(prefix)])
				hashes = append(hashes, child.Hash)
				return paths, hashes
			}

		}
		if !flag {
			break
		}
	}

	return paths, hashes
}

func lengthPrefix(s1, s2 string) int {
	length := len(s1)
	if length < len(s2) {
		length = len(s2)
	}
	for i := 0; i < length; i++ {
		if s1[i] != s2[i] {
			return i
		}
	}
	return length
}
