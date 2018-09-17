package geo

func TileToBoundingBox(z int, x int, y int) []float64 {
	e := TileToLongitude(x+1, z)
	w := TileToLongitude(x, z)
	s := TileToLatitude(y+1, z)
	n := TileToLongitude(y, z)
	return []float64{w, s, e, n}
}
