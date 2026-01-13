package internal

import (
	"fmt"
	"os"
)

type NodeLayout interface {
	GetNodeID() NodeID
}

type NodeMmap[T NodeLayout] struct {
	*StructMmap[T]
}

func NewNodeReader[T NodeLayout](file *os.File) (*NodeMmap[T], error) {
	sf, err := NewStructMmap[T](file)
	if err != nil {
		return nil, err
	}
	return &NodeMmap[T]{StructMmap: sf}, nil
}

func (nf *NodeMmap[T]) FindByID(id NodeID, info *NodeSetInfo) (*T, error) {
	// binary search with interpolation
	lowOffset := info.StartOffset
	targetIdx := id.Index()
	lowIdx := info.StartIndex
	highOffset := lowOffset + info.Count - 1
	highIdx := info.EndIndex
	for lowOffset <= highOffset {
		if targetIdx < lowIdx || targetIdx > highIdx {
			return nil, fmt.Errorf("node ID %s not present", id.String())
		}
		// If nodes are contiguous in this range, compute offset directly
		if highIdx-lowIdx == highOffset-lowOffset {
			targetOffset := lowOffset + (targetIdx - lowIdx)
			return &nf.items[targetOffset], nil
		}
		// Interpolation search: estimate position based on target's relative position in index range
		var mid uint32
		if highIdx > lowIdx {
			// Estimate where target should be based on its position in the index range
			fraction := float64(targetIdx-lowIdx) / float64(highIdx-lowIdx)
			mid = lowOffset + uint32(fraction*float64(highOffset-lowOffset))
			// Ensure mid stays within bounds
			if mid < lowOffset {
				mid = lowOffset
			} else if mid > highOffset {
				mid = highOffset
			}
		} else {
			// When indices converge, use simple midpoint
			mid = (lowOffset + highOffset) / 2
		}
		midNode := &nf.items[mid]
		midIdx := (*midNode).GetNodeID().Index()
		if midIdx == targetIdx {
			return midNode, nil
		} else if midIdx < targetIdx {
			lowOffset = mid + 1
			lowIdx = midIdx + 1
		} else {
			highOffset = mid - 1
			highIdx = midIdx - 1
		}
	}
	return nil, fmt.Errorf("node ID %s not found", id.String())
}
