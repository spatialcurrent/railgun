package geo

func TileToBoundingBox(z int, x int, y int) []float64 {
	w := TileToLongitude(x, z)
	e := TileToLongitude(x+1, z)
	s := TileToLatitude(y+1, z)
	n := TileToLatitude(y, z)
	return []float64{w, s, e, n}
}
