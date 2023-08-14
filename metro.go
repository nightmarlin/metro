package metro

import (
	"context"
	"fmt"
	"sync"
)

/*
A rail network is an ordered Graph
A track segment is an Edge of the Graph
A station is a special Node of the Graph
A rail line is an ordered list of Edges in the Graph, beginning at a station and
  ending at a station, passing through an arbitrary number of nodes, some of
  which may also be stations
A train is the indexing element of such a traversal
*/

// A RailNetworkMap defines the graph of a rail network.
type RailNetworkMap struct {
	Segments    map[string]TrackSegment  // Segments is a map of TrackSegment.Name to TrackSegment
	Connections []TrackSegmentConnection // Connections defines the edges of the graph.
}

// A TrackSegment defines information about the piece of track.
type TrackSegment struct {
	Name string
}

// A TrackSegmentConnection defines the connection between two TrackSegment s.
// The following constraints must hold:
//
//  1. 1 <= len(In) <= 2
//  2. 1 <= len(Out) <= 2
//  2. 2 <= ( len(In) + len(Out) ) <= 3
type TrackSegmentConnection struct {
	In  []string
	Out []string
}

type RouteDirection int

func (rd RouteDirection) Next() RouteDirection {
	switch rd {
	case OutboundDirection:
		return InboundDirection
	default:
		return OutboundDirection
	}
}

const (
	OutboundDirection = iota
	InboundDirection
)

type RailLine struct {
	Name         string
	StationNames []string // The Station s the RailLine passes through (used in future for automatic route navigation)

	Route map[RouteDirection][]string // The ordered list of TrackSegment s the RailLine passes through on its RouteDirection journey
}

func (rl RailLine) NextSegment(train Train) (Train, error) {
	if train.LineName != rl.Name {
		return Train{}, ErrWrongLine
	}

	var nextIDX int

	for sIDX, segmentName := range rl.Route[train.Direction] {
		if segmentName == train.CurrentSegmentName {
			nextIDX = sIDX + 1
		}
	}
	if nextIDX == 0 {
		return Train{}, ErrNotFound
	}

	if nextIDX >= len(rl.Route[train.Direction]) {
		train.Direction = train.Direction.Next()
		nextIDX = 0
	}
	// recheck as route may not exist and we don't want to panic
	if nextIDX >= len(rl.Route[train.Direction]) {
		return Train{}, ErrInvalidRoute
	}

	train.CurrentSegmentName = rl.Route[train.Direction][nextIDX]
	return train, nil
}

type Train struct {
	Name               string
	CurrentSegmentName string
	LineName           string
	Direction          RouteDirection
}

type Platform struct {
	Name        string
	SegmentName string
}

type Station struct {
	Name      string
	Platforms []Platform
}

type Metro struct {
	mux sync.RWMutex

	Map RailNetworkMap

	Stations []Station
	Lines    []RailLine
	Trains   []Train
}

func (m *Metro) getLineByName(name string) (RailLine, error) {
	for _, line := range m.Lines {
		if line.Name == name {
			return line, nil
		}
	}
	return RailLine{}, ErrNotFound
}

func (m *Metro) Tick(ctx context.Context) error {
	defer m.mux.Unlock()
	m.mux.Lock()

	for trainIDX, train := range m.Trains {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := m.getLineByName(train.LineName)
		if err != nil {
			return fmt.Errorf("getting line %q for train %q: %w", train.LineName, train.Name, err)
		}

		newTrain, err := line.NextSegment(train)
		if err != nil {
			return fmt.Errorf("traversing line %q for train %q: %w", line.Name, train.Name, err)
		}

		m.Trains[trainIDX] = newTrain
	}

	return nil
}
