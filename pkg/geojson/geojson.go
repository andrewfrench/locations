package geojson

type GeoJson struct {
	BuiltAt  int64     `json:"builtAt"`
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Type       string     `json:"type"`
	Properties Properties `json:"properties,omitempty"`
	Geometry   Geometry   `json:"geometry,omitempty"`
}

type Properties struct {
}

type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float32 `json:"coordinates"`
}
