package cidrtree

import (
	"fmt"
	"io"
	"strings"
)

// String returns a hierarchical tree diagram of the ordered CIDRs as string, just a wrapper for [Tree.Fprint].
func (t Table[V]) String() string {
	w := new(strings.Builder)
	_ = t.Fprint(w)
	return w.String()
}

// Fprint writes an ordered CIDR tree diagram to w. If w is nil, Fprint panics.
//
// The order from top to bottom is in ascending order of the start address
// and the subtree structure is determined by the CIDRs coverage.
func (t Table[V]) Fprint(w io.Writer) error {
	if err := t.root4.fprint(w); err != nil {
		return err
	}
	if err := t.root6.fprint(w); err != nil {
		return err
	}
	return nil
}

func (n *node[V]) fprint(w io.Writer) error {
	if n == nil {
		return nil
	}

	// pcm = parent-child-mapping
	var pcm parentChildsMap[V]

	// init map
	pcm.pcMap = make(map[*node[V]][]*node[V])

	pcm = n.buildParentChildsMap(pcm)

	if len(pcm.pcMap) == 0 {
		return nil
	}

	// start symbol
	if _, err := fmt.Fprint(w, "▼\n"); err != nil {
		return err
	}

	// start recursion with root and empty padding
	var root *node[V]
	return root.walkAndStringify(w, pcm, "")
}

func (n *node[V]) walkAndStringify(w io.Writer, pcm parentChildsMap[V], pad string) error {
	// the prefix (pad + glyphe) is already printed on the line on upper level
	if n != nil {
		if _, err := fmt.Fprintf(w, "%v (%v)\n", n.cidr, n.value); err != nil {
			return err
		}
	}

	glyphe := "├─ "
	spacer := "│  "

	// dereference child-slice for clearer code
	childs := pcm.pcMap[n]

	// for all childs do, but ...
	for i, child := range childs {
		// ... treat last child special
		if i == len(childs)-1 {
			glyphe = "└─ "
			spacer = "   "
		}
		// print prefix for next cidr
		if _, err := fmt.Fprint(w, pad+glyphe); err != nil {
			return err
		}

		// recdescent down
		if err := child.walkAndStringify(w, pcm, pad+spacer); err != nil {
			return err
		}
	}

	return nil
}

// parentChildsMap, needed for hierarchical tree printing, this is not BST printing!
//
// CIDR tree, parent->childs relation printed. A parent CIDR covers a child CIDR.
type parentChildsMap[T any] struct {
	pcMap map[*node[T]][]*node[T] // parent -> []child map
	stack []*node[T]              // just needed for the algo
}

// buildParentChildsMap, in-order traversal
func (n *node[V]) buildParentChildsMap(pcm parentChildsMap[V]) parentChildsMap[V] {
	if n == nil {
		return pcm
	}

	// in-order traversal, left tree
	pcm = n.left.buildParentChildsMap(pcm)

	// detect parent-child-mapping for this node
	pcm = n.pcmForNode(pcm)

	// in-order traversal, right tree
	return n.right.buildParentChildsMap(pcm)
}

// pcmForNode, find parent in stack, remove cidrs from stack, put this cidr on stack.
func (n *node[V]) pcmForNode(pcm parentChildsMap[V]) parentChildsMap[V] {
	// if this cidr is covered by a prev cidr on stack
	for j := len(pcm.stack) - 1; j >= 0; j-- {
		that := pcm.stack[j]
		if that.cidr.Contains(n.cidr.Addr()) {
			// cidr in node j is parent to cidr
			pcm.pcMap[that] = append(pcm.pcMap[that], n)
			break
		}

		// Remember: sort order of CIDRs is lower-left, superset to the left:
		// if this cidr wasn't covered by j, remove node at j from stack
		pcm.stack = pcm.stack[:j]
	}

	// stack is emptied, no cidr on stack covers current cidr
	if len(pcm.stack) == 0 {
		// parent is root
		pcm.pcMap[nil] = append(pcm.pcMap[nil], n)
	}

	// put current node on stack for next node
	pcm.stack = append(pcm.stack, n)

	return pcm
}
