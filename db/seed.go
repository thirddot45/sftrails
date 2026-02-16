package db

import "database/sql"

func SeedTrails(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM trails").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	trails := []struct {
		name, location, city, description string
		lat, lng                          float64
	}{
		{"Markham Park", "Sunrise", "Sunrise", "Technical singletrack with rock gardens and log features", 26.1256, -80.3584},
		{"Oleta River State Park", "North Miami Beach", "North Miami Beach", "Flowy trails through mangroves with bay views", 25.9207, -80.1412},
		{"Virginia Key", "Key Biscayne", "Key Biscayne", "Tight singletrack through coastal hardwood hammock", 25.7380, -80.1526},
		{"Quiet Waters Park", "Deerfield Beach", "Deerfield Beach", "Beginner-friendly flat trails with fun features", 26.3048, -80.1268},
		{"Halpatiokee Regional Park", "Stuart", "Stuart", "Natural terrain trails through pine flatwoods", 27.1312, -80.4062},
		{"Amelia Earhart Park", "Hialeah", "Hialeah", "Short but fun loops with tabletops and berms", 25.9223, -80.2917},
		{"Jonathan Dickinson State Park", "Hobe Sound", "Hobe Sound", "Sandy singletrack through scrub and pine forest", 27.0184, -80.1125},
		{"Riverbend Park", "Jupiter", "Jupiter", "Flowy trails along the Loxahatchee River", 26.9518, -80.1605},
		{"Dyer Park", "West Palm Beach", "West Palm Beach", "Fast and flowy with nice elevation changes", 26.6358, -80.1530},
		{"Pinehurst Park", "Greenacres", "Greenacres", "Compact trail system with good mix of terrain", 26.6174, -80.1586},
		{"West Delray Regional Park", "Delray Beach", "Delray Beach", "Well-maintained trails with progressive features", 26.4472, -80.1516},
		{"Tree Tops Park", "Davie", "Davie", "Shaded trails through tropical hardwood hammock", 26.0585, -80.2706},
	}

	stmt, err := db.Prepare(`INSERT INTO trails (name, location, city, description, latitude, longitude) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range trails {
		if _, err := stmt.Exec(t.name, t.location, t.city, t.description, t.lat, t.lng); err != nil {
			return err
		}
	}
	return nil
}
