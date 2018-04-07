package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Map struct {
	Node
	KV []KV
}

// Returns a new empty Map<T>
func NewMap() *Map {
	return &Map{}
}

// Adds a new empty entry to the Map
func (m *Map) Add(key string) {
	kv := KV{
		Key:   key,
		Value: nil,
	}
	m.KV = append(m.KV, kv)
}

// Get a node value from input key. If the key does not exist in the map,
// creates a new KV in which key is the input passed and value is empty
func (m Map) Get(key string) CRDT {
	for _, kv := range m.KV {
		if kv.Key == key {
			return kv.Value
		}
	}
	// If key does not exist, create one, add it to the map and returns new empty
	// value
	nKv := KV{
		Key: key,
	}
	return nKv.Value
}

// Gets all keys from Map<T>
func (m *Map) Keys() []string {
	keys := []string{}
	for _, kv := range m.KV {
		keys = append(keys, kv.Key)
	}
	return keys
}

// Gets all values from Map<T>
func (m *Map) Values() []CRDT {
	vals := []CRDT{}
	for _, kv := range m.KV {
		vals = append(vals, kv.Value)
	}
	return vals
}

func (m Map) AddOpPresence(id string) {}
func (m Map) RmOpPresence(id string)  {}

func (m Map) MarshalJSON() ([]byte, error) {
	repr := map[string]CRDT{}
	for _, kv := range m.KV {
		k, v := kv.getKV()
		repr[k] = v
	}
	b, err := json.Marshal(repr)
	return b, err
}

func (m Map) String() string {
	if len(m.KV) == 0 {
		return fmt.Sprintf("{}")
	}

	out := []string{}
	for _, kv := range m.KV {
		out = append(out, kv.String())
	}
	return fmt.Sprintf("[%v]", strings.Join(out, ","))
}

type KV struct {
	Key   string
	Value CRDT
}

func (kv KV) getKV() (string, CRDT) {
	return kv.Key, kv.Value
}

func (kv KV) String() string {
	return fmt.Sprintf("{%v:%v}", kv.Key, kv.Value)
}

func (kv KV) MarshalJSON() ([]byte, error) {
	fmt.Println("KV.MarshalJSON called")
	return []byte{}, nil
}
