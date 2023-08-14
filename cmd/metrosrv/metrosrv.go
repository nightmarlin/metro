package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/nightmarlin/metro"
)

func main() {
	start := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	log := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug},
		),
	)

	intChan := make(chan os.Signal)
	defer close(intChan)
	signal.Notify(intChan, os.Interrupt)
	go func() {
		for range intChan {
			cancel()
		}
	}()

	log.LogAttrs(
		ctx,
		slog.LevelInfo,
		"starting engine, use ctrl+c to exit",
		slog.Any("starting state", testMetro),
	)

	defer log.LogAttrs(
		ctx,
		slog.LevelInfo,
		"execution complete",
		slog.Duration("executed_in", time.Now().Sub(start)),
	)

	tmr := time.NewTimer(time.Second)
	for range tmr.C {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := testMetro.Tick(ctx); err != nil {
			log.LogAttrs(ctx, slog.LevelError, "tick handler failed!", slog.String("error", err.Error()))
			return
		}

		log.LogAttrs(
			ctx,
			slog.LevelInfo,
			"ticked!",
			slog.String("train_location", testMetro.Trains[0].CurrentSegmentName),
		)

		tmr.Reset(time.Second)
	}
}

var testMetro = &metro.Metro{
	Map: metro.RailNetworkMap{
		Segments: map[string]metro.TrackSegment{
			"segment_0": {Name: "segment_0"},
			"segment_1": {Name: "segment_1"},
			"segment_2": {Name: "segment_2"},

			"segment_3": {Name: "segment_0"},
			"segment_4": {Name: "segment_1"},
			"segment_5": {Name: "segment_2"},
		},

		Connections: []metro.TrackSegmentConnection{
			// outbound
			oneToOneSC("segment_0", "segment_1"),
			oneToOneSC("segment_1", "segment_2"),

			// inbound
			oneToOneSC("segment_5", "segment_4"),
			oneToOneSC("segment_4", "segment_3"),
		},
	},

	Stations: []metro.Station{
		{
			Name: "station_0",
			Platforms: []metro.Platform{
				{Name: "station_0_platform_0", SegmentName: "segment_0"},
				{Name: "station_0_platform_1", SegmentName: "segment_3"},
			},
		},
		{
			Name: "station_1",
			Platforms: []metro.Platform{
				{Name: "station_1_platform_0", SegmentName: "segment_2"},
				{Name: "station_1_platform_1", SegmentName: "segment_5"},
			},
		},
	},

	Lines: []metro.RailLine{
		{
			Name:         "urmi",
			StationNames: []string{"station_0", "station_1"},
			Route: map[metro.RouteDirection][]string{
				metro.OutboundDirection: {"segment_0", "segment_1", "segment_2"}, // todo: pathfinding
				metro.InboundDirection:  {"segment_5", "segment_4", "segment_3"}, // todo: pathfinding
			},
		},
	},

	Trains: []metro.Train{
		{
			Name:               "train_0",
			LineName:           "urmi",
			CurrentSegmentName: "segment_0",
			Direction:          metro.OutboundDirection,
		},
	},
}

func oneToOneSC(in, out string) metro.TrackSegmentConnection {
	return metro.TrackSegmentConnection{In: []string{in}, Out: []string{out}}
}
func twoToOneSC(in1, in2, out string) metro.TrackSegmentConnection {
	return metro.TrackSegmentConnection{In: []string{in1, in2}, Out: []string{out}}
}
func oneToTwoSC(in, out1, out2 string) metro.TrackSegmentConnection {
	return metro.TrackSegmentConnection{In: []string{in}, Out: []string{out1, out2}}
}
