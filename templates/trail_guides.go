package templates

// Short editorial notes, checked against the linked park authorities on
// 2026-09-12. Keep changing fees, hours and closure notices at the source.
type trailGuide struct {
	Summary string
	Source  string
	URL     string
}

var trailGuides = map[string]trailGuide{
	"amelia-earhart-park": {
		"Amelia Earhart Park has woodland mountain bike routes for beginner, intermediate and advanced riders. Miami-Dade County publishes maintenance notices on its park page, so check those alongside community reports.",
		"Miami-Dade County: Amelia Earhart Park",
		"https://www.miamidade.gov/global/recreation/park/amelia-earhart-park.page",
	},
	"dyer-park": {
		"Dyer Park offers both perimeter singletrack and a separate hill trail with climbs and descents on a former landfill. Palm Beach County lists these as mountain bike routes, distinct from the park's shared bicycle paths.",
		"Palm Beach County: Dyer mountain bike trails",
		"https://discover.pbc.gov/parks/Amenities/Bicycling.aspx",
	},
	"halpatiokee-regional-park": {
		"Halpatiokee's trails pass through a preserve with pine flatwoods, oak hammock and scrub near the St. Lucie River. Martin County provides a trail map and riding rules, including directional signs and shared-trail crossings.",
		"Martin County: Halpatiokee Regional Park",
		"https://www.martin.fl.us/Halpatiokee",
	},
	"jonathan-dickinson-state-park": {
		"The Camp Murphy Off-Road Bicycle Trail System is Jonathan Dickinson's mountain biking network. Its loops range from beginner routes to expert black-diamond trails; the park also has paved cycling routes.",
		"Florida State Parks: Jonathan Dickinson riding information",
		"https://www.floridastateparks.org/parks-and-trails/jonathan-dickinson-state-park/experiences-amenities",
	},
	"markham-park": {
		"Markham Park offers mountain bike trails with an adaptive and beginner loop as well as its broader off-road network. Broward County's trail information explains access-card requirements for riders.",
		"Broward County: mountain bike trail information (PDF)",
		"https://www.broward.org/Parks/Support/Documents/EmergencyManagementandPublicSafetySectionMarch2026.pdf",
	},
	"oleta-river-state-park": {
		"Oleta River State Park has novice routes as well as more challenging intermediate mountain bike trails. Paved cycling is also available, so choose the route type that matches your bike and experience.",
		"Florida State Parks: Oleta River riding information",
		"https://www.floridastateparks.org/parks-and-trails/oleta-river-state-park/experiences-amenities",
	},
	"pinehurst-park": {
		"The Pinehurst mountain bike trail is part of Okeeheelee Park. Palm Beach County places the trailhead east of Pinehurst Drive, south of Forest Hill Boulevard, and describes intermediate singletrack with beginner bypasses.",
		"Palm Beach County: Pinehurst mountain bike trail",
		"https://discover.pbc.gov/parks/Amenities/Bicycling.aspx",
	},
	"quiet-waters-park": {
		"Quiet Waters is one of Broward County's dedicated mountain biking parks. County-approved volunteers help maintain its trail network, and riders need a current mountain bike access card.",
		"Broward County: mountain bike trail information (PDF)",
		"https://www.broward.org/Parks/Support/Documents/EmergencyManagementandPublicSafetySectionMarch2026.pdf",
	},
	"riverbend-park": {
		"Riverbend offers shared hiking and bicycling on compacted shell-rock paths near the Loxahatchee River. These are multi-use recreational routes; expect walkers and choose a pace that leaves room for other visitors.",
		"Palm Beach County: Riverbend hiking and bicycling trails",
		"https://discover.pbc.gov/parks/Riverbend/Hiking-Bicycling.aspx",
	},
	"virginia-key": {
		"Virginia Key's North Point trails are in the City of Miami. The mountain bike network was developed with Virginia Key Bicycle Club volunteers and includes routes for different skill levels near the waterfront.",
		"City of Miami: Virginia Key Beach North Point Park",
		"https://www.miami.gov/Parks-Public-Places/Parks-Directory/Virginia-Key-Beach-North-Point-Park",
	},
	"west-delray-regional-park": {
		"West Delray's mountain bike singletrack is in the western part of the park. Palm Beach County describes it as an intermediate route and provides site rules alongside its other regional mountain biking locations.",
		"Palm Beach County: West Delray mountain bike trail",
		"https://discover.pbc.gov/parks/Amenities/Bicycling.aspx",
	},
}
