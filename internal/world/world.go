package world

// Port represents a named location on the map.
type Port struct {
	Name string
	Lon  float64
	Lat  float64
}

// Route represents a shipping route as a sequence of waypoints.
type Route struct {
	Waypoints [][2]float64 // each element is [lon, lat]
}

// World holds all game-state entities (ports, routes, etc.).
type World struct {
	Ports  []Port
	Routes []Route
}

// NewDefault returns a World pre-populated with sample ports and routes.
func NewDefault() *World {
	return &World{
		Ports: []Port{
			{"New York", -74.0, 40.7},
			{"Rio de Janeiro", -43.2, -22.9},
			{"London", -0.1, 51.5},
			{"Cape Town", 18.4, -33.9},
			{"Shanghai", 121.5, 31.2},
			{"Sydney", 151.2, -33.9},
			{"San Francisco", -122.4, 37.8},
			{"Tokyo", 139.7, 35.7},
			{"Singapore", 103.9, 1.3},
			{"Dubai", 55.3, 25.3},
		},
		Routes: []Route{
			{Waypoints: [][2]float64{
				{-0.1, 51.5}, {-9.1, 38.7}, {-16.9, 28.1}, {-43.2, -22.9},
			}},
		},
	}
}
