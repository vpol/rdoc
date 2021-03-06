package rdoc

import (
	"errors"
	"fmt"
	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/gpestana/rdoc/clock"
	op "github.com/gpestana/rdoc/operation"
	"log"
)

type Doc struct {
	Id               string
	Clock            clock.Clock
	OperationsId     []string
	Head             *Node
	OperationsBuffer []op.Operation
}

// Returns a new rdoc data structure. It receives an ID which must be
// unique in the context of the network.
func Init(id string) *Doc {
	headNode := newNode(nil)
	c := clock.New([]byte(id))
	return &Doc{
		Id:               id,
		Clock:            c,
		OperationsId:     []string{},
		Head:             headNode,
		OperationsBuffer: []op.Operation{},
	}
}

func (d *Doc) ApplyRemoteOperation(o op.Operation) (*Doc, error) {
	// if operation has been applied already, skip
	if containsId(d.OperationsId, o.ID) {
		return d, nil
	}
	// if operation dependencies havent been all applied in the document, buffer
	// the operation
	missingOp := diff(o.Deps, d.OperationsId)
	if len(missingOp) != 0 {
		d.OperationsBuffer = append(d.OperationsBuffer, o)
		return d, nil
	}
	return d.ApplyOperation(o)
}

func (d *Doc) ApplyOperation(o op.Operation) (*Doc, error) {
	nPtr, travNodes, createdNodes := d.traverse(o.Cursor)

	// updates dependencies of traversed and created nodes
	var deps []*Node
	deps = append(deps, travNodes...)
	deps = append(deps, createdNodes...)
	for _, n := range deps {
		n.AddDependency(o.ID)
	}

	//TODO: let's assume the Mutate() call never fails for now.
	//TODO: how to rollback side effects of traverse if Mutate() fails?
	err := nPtr.Mutate(o)
	if err != nil {
		return d, err
	}

	d.OperationsId = append(d.OperationsId, o.ID)
	return d, nil
}

// Traverses the document form root element to the node indicated by the cursor
// input. When a path does not exist in the current document, create the node
// and link it to the document.
// The traverse function returns a pointer to the last node, a list of pointers
// of nodes traversed and a list of pointers of nodes created
func (d *Doc) traverse(cursor op.Cursor) (*Node, []*Node, []*Node) {
	var nPtr *Node
	var travNodes []*Node
	var createdNodes []*Node

	nPtr = d.Head

	for _, c := range cursor.Path {
		k := c.Get()
		switch c.Type() {
		case op.MapT:
			nif, exists := nPtr.hmap.Get(k.(string))
			if !exists {
				nn := newNode(k)
				_ = nPtr.link(op.MapT, nn)
				nPtr = nn
				travNodes = append(travNodes, nPtr)
				createdNodes = append(createdNodes, nPtr)
				continue
			}
			nPtr = nif.(*Node)
			travNodes = append(travNodes, nPtr)

		case op.ListT:
			nif, exists := nPtr.list.Get(k.(int))
			if !exists {
				nn := newNode(k)
				_ = nPtr.link(op.ListT, nn)
				nPtr = nn
				travNodes = append(travNodes, nPtr)
				createdNodes = append(createdNodes, nPtr)
				continue
			}
			nPtr = nif.(*Node)
			travNodes = append(travNodes, nPtr)
		}
	}
	return nPtr, travNodes, createdNodes
}

func (d Doc) String() string {
	ids := fmt.Sprintf("ID: %v; ClockId: %v", d.Id, d.Clock)
	ops := fmt.Sprintf("Operations: applied: %v, buffered: %v", d.OperationsId, d.OperationsBuffer)
	node := fmt.Sprintf("Head: %v", d.Head)
	return fmt.Sprintf("%v\n%v\n%v\n", ids, ops, node)
}

type Node struct {
	key  interface{}
	deps []string
	hmap *hashmap.Map
	list *arraylist.List
	reg  *hashmap.Map
}

func newNode(key interface{}) *Node {
	return &Node{
		key:  key,
		deps: []string{},
		hmap: hashmap.New(),
		list: arraylist.New(),
		reg:  hashmap.New(),
	}
}

func (n *Node) GetList() *arraylist.List {
	return n.list
}

func (n *Node) GetMap() *hashmap.Map {
	return n.hmap
}

func (n *Node) GetReg() *hashmap.Map {
	return n.reg
}

// applies operation mutation to the node
// note: assumes that mutation never fails for now
func (n *Node) Mutate(o op.Operation) error {
	mut := o.Mutation
	var err error

	// 1) remove nodes if type of mutation is type Delete or Assign
	switch mut.Typ {
	case op.Delete:
		// delete and return
		children := n.allChildren()
		clearNodes(children, o.Deps)
		return nil
	case op.Assign:
		// delete and proceed
		children := n.allChildren()
		clearNodes(children, o.Deps)
	}

	// 2) modify node if mutation is type Insert or Assign
	// TODO: refactor this for node.Add(...) which will be implemented at a node
	// level
	switch mut.Key.(type) {
	case int:
		// list
		nn := newNode(mut.Key)
		nn.reg.Put(o.ID, mut.Value)
		n.list.Insert(mut.Key.(int), nn)
	case string:
		// map
		log.Println("mutating a map", mut.Key, o.ID, mut.Value)
		nn := newNode(mut.Key)
		nn.reg.Put(o.ID, mut.Value)
		n.hmap.Put(mut.Key.(string), nn)
	case nil:
		// register
		n.reg.Put(o.ID, mut.Value)
	default:
		return errors.New("Key type is not valid")
	}

	return err
}

// appends new dependency to Node
func (n *Node) AddDependency(d string) {
	// TODO: should check if dep is valid with clock.Clock primitves?
	n.deps = append(n.deps, d)
}

// Links a node to the current node. The new node is linked depending on the
// type of linking required. It can be of type MapT, ListT or RegT.
func (n *Node) link(linkType int, node *Node) error {
	switch linkType {
	case op.MapT:
		key, ok := node.key.(string)
		if !ok {
			return errors.New("Map key must be string")
		}
		n.hmap.Put(key, node)

	case op.ListT:
		key, ok := node.key.(int)
		if !ok {
			return errors.New("List key must be an int")
		}
		n.list.Insert(key, node)

	case op.RegT:
		log.Println("linking RegT")
	default:
		return errors.New("linking type not correct")
	}

	return nil
}

// Returns all subsequent nodes from a particular Node
func (n *Node) allChildren() []*Node {
	var children []*Node
	var tmp []*Node
	tmp = append(tmp, directChildren(n)...)

	for {
		if len(tmp) == 0 {
			break
		}
		nextTmp := tmp[:1]
		tmp = tmp[1:]

		c := nextTmp[0]
		tmp = append(tmp, directChildren(c)...)
		children = append(children, c)
	}

	return children
}
