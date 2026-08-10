package data

import (
	"fmt"
	"math"
	"strings"
	"time"

	"previa/pkg/models"
)

// propertySeed is a compact description of one listing. buildProperties()
// expands it into a full models.Property, deriving slugs, price per m², image
// sets, feature chips and timestamps so the seed table stays readable.
type propertySeed struct {
	title    string
	deal     models.DealType
	ptype    models.PropertyType
	country  string
	city     string
	district string
	address  string
	postal   string
	price    float64
	rooms    int
	beds     int
	baths    int
	area     float64
	land     float64
	floor    int
	floors   int
	year     int
	cond     models.Condition
	energy   string
	lat, lng float64
	broker   string
	flags    string // comma-separated: featured,furnished,parking,balcony,garden,elevator,terrace,sauna,seaview,new,verified,private
	imgs     []int  // indices into the property photo pool
	desc     string
	status   models.ListingStatus
	devID    string
	daysAgo  int
	views    int
}

// The seed table. Deliberately varied across markets, deal types, price bands
// and property categories so every filter combination returns something
// plausible and no screen looks empty.
var propertySeeds = []propertySeed{
	// ---- Estonia — Tallinn ------------------------------------------------
	{
		title: "Restored merchant's apartment in Tallinn Old Town", deal: models.DealSale, ptype: models.TypeApartment,
		country: "EE", city: "Tallinn", district: "Vanalinn", address: "Pikk 42", postal: "10133",
		price: 429000, rooms: 4, beds: 2, baths: 2, area: 118, floor: 3, floors: 4, year: 1897,
		cond: models.ConditionRenovated, energy: "D", lat: 59.4404, lng: 24.7453, broker: "br-01",
		flags: "featured,parking,balcony,verified", imgs: []int{1, 45, 55, 63, 71},
		desc: "A rare limestone-walled apartment on one of the Old Town's best-preserved merchant streets. The 2022 restoration kept the original beamed ceilings and Baltic pine floors while adding underfloor heating, triple glazing and a fully rebuilt service core. Two bedrooms face a quiet inner courtyard; the living room and kitchen open onto Pikk itself.",
		daysAgo: 3, views: 1842,
	},
	{
		title: "Sea-view penthouse in Kalamaja with private roof terrace", deal: models.DealSale, ptype: models.TypeApartment,
		country: "EE", city: "Tallinn", district: "Kalamaja", address: "Vabriku 21", postal: "10411",
		price: 615000, rooms: 5, beds: 3, baths: 2, area: 164, floor: 7, floors: 7, year: 2021,
		cond: models.ConditionNew, energy: "A", lat: 59.4453, lng: 24.7318, broker: "br-02",
		flags: "featured,parking,balcony,elevator,terrace,seaview,sauna,verified,new", imgs: []int{3, 46, 56, 64, 72, 2},
		desc: "The top floor of the Vabriku Kvartal conversion, with an 84 m² private terrace facing the bay. Floor-to-ceiling glazing runs the length of the living space, and the kitchen is a bespoke Finnish install in oak and honed granite. Includes two underground parking bays, a storage cage and a private electric sauna.",
		daysAgo: 8, views: 3120,
	},
	{
		title: "Bright two-room apartment near Telliskivi Creative City", deal: models.DealRent, ptype: models.TypeApartment,
		country: "EE", city: "Tallinn", district: "Põhja-Tallinn", address: "Telliskivi 60", postal: "10412",
		price: 780, rooms: 2, beds: 1, baths: 1, area: 54, floor: 4, floors: 6, year: 2018,
		cond: models.ConditionGood, energy: "B", lat: 59.4372, lng: 24.7278, broker: "br-01",
		flags: "furnished,balcony,elevator,parking", imgs: []int{47, 57, 65, 5},
		desc: "Fully furnished and available from the first of next month. The building sits directly behind the Telliskivi quarter, a five-minute walk from the Baltic Station market. Rent covers building maintenance and water; electricity and internet are metered separately.",
		daysAgo: 2, views: 894,
	},
	{
		title: "Family house with sauna in Nõmme pine district", deal: models.DealSale, ptype: models.TypeHouse,
		country: "EE", city: "Tallinn", district: "Nõmme", address: "Männiku tee 88", postal: "11214",
		price: 585000, rooms: 6, beds: 4, baths: 3, area: 232, land: 1240, floors: 2, year: 2009,
		cond: models.ConditionGood, energy: "C", lat: 59.3833, lng: 24.6833, broker: "br-03",
		flags: "parking,garden,sauna,terrace,verified", imgs: []int{13, 48, 58, 66, 73},
		desc: "Set back from the road on a mature pine plot in Nõmme's most sought-after pocket. Ground floor holds an open kitchen-dining room, a study and a wood-fired sauna suite; four bedrooms upstairs. Air-source heat pump installed 2021, roof replaced 2019.",
		daysAgo: 14, views: 2205,
	},
	{
		title: "Commercial ground floor unit on Rotermanni", deal: models.DealRent, ptype: models.TypeCommercial,
		country: "EE", city: "Tallinn", district: "Kesklinn", address: "Rotermanni 8", postal: "10111",
		price: 3400, rooms: 3, area: 186, floor: 0, floors: 5, year: 2014,
		cond: models.ConditionGood, energy: "B", lat: 59.4372, lng: 24.7565, broker: "br-02",
		flags: "parking,elevator,verified", imgs: []int{85, 86, 87},
		desc: "Corner retail unit with 11 metres of display frontage in the Rotermanni quarter, currently fitted as a showroom. Suitable for retail, hospitality or a client-facing office. Service charge €3.10/m² per month. Available on a five-year lease.",
		daysAgo: 21, views: 640,
	},
	{
		title: "Building plot with sea access in Viimsi", deal: models.DealSale, ptype: models.TypeLand,
		country: "EE", city: "Tallinn", district: "Viimsi", address: "Rohuneeme tee 142", postal: "74001",
		price: 189000, area: 0, land: 2870, year: 0,
		cond: models.ConditionGood, lat: 59.5183, lng: 24.7442, broker: "br-03",
		flags: "seaview", imgs: []int{91, 92},
		desc: "Detailed plan approved for a single dwelling of up to 320 m² across two storeys. Municipal water, sewerage and three-phase electricity are at the boundary. Sixty metres of shared shoreline access via the registered right of way.",
		daysAgo: 30, views: 512,
	},

	// ---- Germany — Berlin -------------------------------------------------
	{
		title: "Altbau apartment with stucco ceilings in Prenzlauer Berg", deal: models.DealSale, ptype: models.TypeApartment,
		country: "DE", city: "Berlin", district: "Prenzlauer Berg", address: "Kollwitzstraße 66", postal: "10435",
		price: 895000, rooms: 4, beds: 3, baths: 2, area: 141, floor: 2, floors: 5, year: 1904,
		cond: models.ConditionRenovated, energy: "C", lat: 52.5347, lng: 13.4184, broker: "br-04",
		flags: "featured,balcony,elevator,verified", imgs: []int{4, 49, 59, 67, 74},
		desc: "A classic Berlin Altbau one street from Kollwitzplatz, with 3.4-metre ceilings, original rosettes and restored double-wing doors. Renovated in 2020 with new electrics, oak parquet and a Bulthaup kitchen. South-facing balcony over a quiet courtyard; the building added a lift in 2017.",
		daysAgo: 5, views: 4110,
	},
	{
		title: "Loft conversion in a former Kreuzberg factory", deal: models.DealSale, ptype: models.TypeApartment,
		country: "DE", city: "Berlin", district: "Kreuzberg", address: "Ritterstraße 12", postal: "10969",
		price: 742000, rooms: 3, beds: 2, baths: 2, area: 156, floor: 4, floors: 5, year: 1928,
		cond: models.ConditionRenovated, energy: "D", lat: 52.5027, lng: 13.4069, broker: "br-05",
		flags: "elevator,terrace,parking", imgs: []int{50, 60, 68, 75, 6},
		desc: "Open-plan living across a single 156 m² floor plate with exposed steel trusses and six-metre industrial windows facing north-west. The mezzanine sleeping level was added under the 2019 permit. Freight lift opens directly into the apartment.",
		daysAgo: 11, views: 2680,
	},
	{
		title: "Modern three-room flat in Friedrichshain", deal: models.DealRent, ptype: models.TypeApartment,
		country: "DE", city: "Berlin", district: "Friedrichshain", address: "Boxhagener Straße 91", postal: "10245",
		price: 1650, rooms: 3, beds: 2, baths: 1, area: 82, floor: 3, floors: 6, year: 2016,
		cond: models.ConditionGood, energy: "B", lat: 52.5099, lng: 13.4590, broker: "br-04",
		flags: "furnished,balcony,elevator", imgs: []int{51, 61, 69, 8},
		desc: "Two streets from Boxhagener Platz and its Saturday market. Fitted kitchen, built-in wardrobes and a west-facing balcony. Kaltmiete €1,650 plus €280 Nebenkosten. Minimum twelve-month contract; SCHUFA report required.",
		daysAgo: 1, views: 1330,
	},
	{
		title: "Villa with garden studio in Grunewald", deal: models.DealSale, ptype: models.TypeVilla,
		country: "DE", city: "Berlin", district: "Grunewald", address: "Bismarckallee 24", postal: "14193",
		price: 2450000, rooms: 9, beds: 5, baths: 4, area: 412, land: 1850, floors: 3, year: 1921,
		cond: models.ConditionRenovated, energy: "C", lat: 52.4869, lng: 13.2646, broker: "br-05",
		flags: "featured,parking,garden,terrace,sauna,verified", imgs: []int{25, 52, 62, 70, 76, 26},
		desc: "A 1921 villa on a walled 1,850 m² plot, comprehensively modernised in 2018 while retaining the original staircase, panelling and leaded glazing. Five bedrooms, a library, a wine cellar and a separate 46 m² garden studio currently used as an architect's office. Double garage plus parking for four.",
		daysAgo: 19, views: 5940,
	},

	// ---- Spain — Barcelona ------------------------------------------------
	{
		title: "Eixample apartment with original hydraulic tiles", deal: models.DealSale, ptype: models.TypeApartment,
		country: "ES", city: "Barcelona", district: "Eixample", address: "Carrer de Mallorca 214", postal: "08008",
		price: 780000, rooms: 4, beds: 3, baths: 2, area: 132, floor: 2, floors: 6, year: 1912,
		cond: models.ConditionRenovated, energy: "D", lat: 41.3925, lng: 2.1567, broker: "br-06",
		flags: "featured,balcony,elevator,verified", imgs: []int{7, 53, 77, 9, 78},
		desc: "A first-floor principal in a Cerdà block between Passeig de Gràcia and Rambla de Catalunya. The 2021 refurbishment preserved the hydraulic floor tiles and moulded ceilings throughout the front rooms. Three balconies face the street; the kitchen and utility open onto the interior patio.",
		daysAgo: 6, views: 3870,
	},
	{
		title: "Sea-facing apartment in Barceloneta", deal: models.DealRent, ptype: models.TypeApartment,
		country: "ES", city: "Barcelona", district: "Barceloneta", address: "Passeig de Joan de Borbó 43", postal: "08003",
		price: 2100, rooms: 3, beds: 2, baths: 1, area: 76, floor: 5, floors: 8, year: 2005,
		cond: models.ConditionGood, energy: "C", lat: 41.3784, lng: 2.1899, broker: "br-07",
		flags: "furnished,balcony,elevator,seaview", imgs: []int{54, 79, 10, 80},
		desc: "Direct sea views from the living room and main bedroom, with the beach promenade below. Fully furnished including air conditioning in every room. Twelve-month contract, two months' deposit, community fees included in the rent.",
		daysAgo: 4, views: 2140,
	},
	{
		title: "Hillside villa with infinity pool in Sarrià", deal: models.DealSale, ptype: models.TypeVilla,
		country: "ES", city: "Barcelona", district: "Sarrià-Sant Gervasi", address: "Carrer de Bellesguard 18", postal: "08022",
		price: 3200000, rooms: 8, beds: 5, baths: 5, area: 465, land: 1120, floors: 3, year: 2015,
		cond: models.ConditionNew, energy: "A", lat: 41.4145, lng: 2.1290, broker: "br-06",
		flags: "featured,parking,garden,terrace,seaview,verified", imgs: []int{27, 81, 28, 82, 29},
		desc: "Built into the Collserola slope with uninterrupted views over the city to the sea. Three terraced levels, a 14-metre infinity pool and a fully glazed living floor that opens completely along its south elevation. Geothermal climate control, home automation, garaging for four cars.",
		daysAgo: 25, views: 7220,
	},
	{
		title: "Renovated studio in Gràcia", deal: models.DealRent, ptype: models.TypeApartment,
		country: "ES", city: "Barcelona", district: "Gràcia", address: "Carrer de Verdi 88", postal: "08012",
		price: 950, rooms: 1, beds: 1, baths: 1, area: 38, floor: 3, floors: 4, year: 1965,
		cond: models.ConditionRenovated, energy: "D", lat: 41.4050, lng: 2.1570, broker: "br-07",
		flags: "furnished,private", imgs: []int{83, 11, 84},
		desc: "A compact studio on Verdi, fully rebuilt in 2023 with a concealed kitchen, a walk-in shower room and generous built-in storage. Let directly by the owner. Suits a single professional or student; no agency commission.",
		daysAgo: 9, views: 1560,
	},

	// ---- Finland — Helsinki -----------------------------------------------
	{
		title: "Waterfront apartment in Jätkäsaari", deal: models.DealSale, ptype: models.TypeApartment,
		country: "FI", city: "Helsinki", district: "Jätkäsaari", address: "Länsisatamankatu 24", postal: "00220",
		price: 690000, rooms: 4, beds: 3, baths: 2, area: 106, floor: 8, floors: 12, year: 2020,
		cond: models.ConditionNew, energy: "A", lat: 60.1553, lng: 24.9186, broker: "br-08",
		flags: "featured,parking,balcony,elevator,sauna,seaview,verified,new", imgs: []int{12, 88, 30, 89},
		desc: "Eighth-floor corner apartment facing the western harbour, with a glazed 12 m² balcony usable through the winter. Building sauna, gym and residents' roof terrace. District heating, mechanical ventilation with heat recovery, EV-ready parking space in the garage.",
		daysAgo: 7, views: 2980,
	},
	{
		title: "Functionalist flat in Töölö with park outlook", deal: models.DealSale, ptype: models.TypeApartment,
		country: "FI", city: "Helsinki", district: "Töölö", address: "Museokatu 17", postal: "00100",
		price: 545000, rooms: 3, beds: 2, baths: 1, area: 88, floor: 4, floors: 6, year: 1938,
		cond: models.ConditionGood, energy: "D", lat: 60.1756, lng: 24.9200, broker: "br-09",
		flags: "balcony,elevator,verified", imgs: []int{90, 31, 93},
		desc: "A 1938 functionalist building one block from Hesperia Park. Original herringbone parquet and panelled doors throughout; kitchen and bathroom updated in 2019. Pipe renovation completed by the housing company in 2016 and fully paid.",
		daysAgo: 16, views: 1720,
	},
	{
		title: "Detached wooden house in Käpylä", deal: models.DealSale, ptype: models.TypeHouse,
		country: "FI", city: "Helsinki", district: "Käpylä", address: "Sampsantie 40", postal: "00610",
		price: 720000, rooms: 5, beds: 3, baths: 2, area: 148, land: 620, floors: 2, year: 1925,
		cond: models.ConditionRenovated, energy: "C", lat: 60.2158, lng: 24.9525, broker: "br-08",
		flags: "garden,sauna,parking,terrace", imgs: []int{14, 94, 32, 95},
		desc: "One of the protected 1920s garden-city timber houses in Käpylä, restored in 2020 under heritage supervision. Wood-fired sauna in the garden building, apple trees and a south-facing terrace. Ground-source heating installed 2020.",
		daysAgo: 22, views: 2450,
	},

	// ---- Portugal — Lisbon ------------------------------------------------
	{
		title: "Pombaline apartment with river views in Chiado", deal: models.DealSale, ptype: models.TypeApartment,
		country: "PT", city: "Lisbon", district: "Chiado", address: "Rua Garrett 58", postal: "1200-204",
		price: 850000, rooms: 4, beds: 2, baths: 2, area: 124, floor: 4, floors: 5, year: 1782,
		cond: models.ConditionRenovated, energy: "C", lat: 38.7108, lng: -9.1410, broker: "br-10",
		flags: "featured,elevator,terrace,seaview,verified", imgs: []int{15, 96, 33, 97},
		desc: "A top-floor apartment in an original Pombaline building on Rua Garrett, rebuilt behind the retained façade in 2019. The living room and both bedrooms look south over the rooftops to the Tejo. Lift added during the works; the terrace runs the width of the building.",
		daysAgo: 10, views: 5310,
	},
	{
		title: "Two-bedroom apartment in Príncipe Real", deal: models.DealRent, ptype: models.TypeApartment,
		country: "PT", city: "Lisbon", district: "Príncipe Real", address: "Rua da Escola Politécnica 92", postal: "1250-102",
		price: 1850, rooms: 3, beds: 2, baths: 2, area: 94, floor: 2, floors: 4, year: 1954,
		cond: models.ConditionRenovated, energy: "C", lat: 38.7168, lng: -9.1520, broker: "br-11",
		flags: "furnished,balcony,elevator", imgs: []int{98, 34, 99},
		desc: "Opposite the botanical garden, in a quiet 1950s building with a lift. Renovated in 2022 and let furnished, including appliances and air conditioning. Twelve-month minimum term; two months' deposit and proof of income required.",
		daysAgo: 3, views: 1980,
	},
	{
		title: "Coastal villa with pool in Cascais", deal: models.DealSale, ptype: models.TypeVilla,
		country: "PT", city: "Lisbon", district: "Cascais", address: "Avenida de Sintra 310", postal: "2750-000",
		price: 1980000, rooms: 7, beds: 4, baths: 4, area: 356, land: 1450, floors: 2, year: 2018,
		cond: models.ConditionNew, energy: "A", lat: 38.7000, lng: -9.4200, broker: "br-10",
		flags: "featured,parking,garden,terrace,seaview,verified", imgs: []int{35, 100, 36, 101},
		desc: "A contemporary villa fifteen minutes from Cascais marina, arranged as a single-storey living wing with a separate sleeping pavilion. Heated 12-metre pool, outdoor kitchen and mature Mediterranean planting on a walled plot. Solar thermal, photovoltaic array and a 9 kW car charger.",
		daysAgo: 28, views: 6640,
	},

	// ---- Netherlands — Amsterdam ------------------------------------------
	{
		title: "Canal house apartment on Prinsengracht", deal: models.DealSale, ptype: models.TypeApartment,
		country: "NL", city: "Amsterdam", district: "Jordaan", address: "Prinsengracht 452", postal: "1017 KX",
		price: 1150000, rooms: 4, beds: 2, baths: 2, area: 118, floor: 1, floors: 4, year: 1690,
		cond: models.ConditionRenovated, energy: "D", lat: 52.3650, lng: 4.8830, broker: "br-12",
		flags: "featured,verified", imgs: []int{16, 102, 37, 103},
		desc: "The bel-étage of a listed 17th-century canal house, with the original beamed ceiling and marble entrance hall intact. Renovated in 2021 to add a rear kitchen extension overlooking the garden. Two canal-facing rooms with the full-height sash windows characteristic of the period.",
		daysAgo: 13, views: 4880,
	},
	{
		title: "Family apartment in Amsterdam-Oost", deal: models.DealRent, ptype: models.TypeApartment,
		country: "NL", city: "Amsterdam", district: "Oost", address: "Javastraat 128", postal: "1095 CD",
		price: 2250, rooms: 4, beds: 3, baths: 1, area: 98, floor: 2, floors: 4, year: 2012,
		cond: models.ConditionGood, energy: "B", lat: 52.3640, lng: 4.9350, broker: "br-12",
		flags: "balcony,elevator,parking", imgs: []int{104, 38, 105},
		desc: "On the Javastraat with its food markets and cafés, a ten-minute cycle from the centre. Three bedrooms, a separate utility room and a sheltered balcony facing the garden side. Unfurnished; energy label B keeps running costs low.",
		daysAgo: 5, views: 2310,
	},

	// ---- Austria — Vienna --------------------------------------------------
	{
		title: "Ringstrasse apartment with ballroom proportions", deal: models.DealSale, ptype: models.TypeApartment,
		country: "AT", city: "Vienna", district: "Innere Stadt", address: "Schottenring 14", postal: "1010",
		price: 1680000, rooms: 6, beds: 3, baths: 3, area: 224, floor: 3, floors: 6, year: 1873,
		cond: models.ConditionRenovated, energy: "D", lat: 48.2150, lng: 16.3680, broker: "br-13",
		flags: "featured,elevator,balcony,verified", imgs: []int{17, 106, 39, 107},
		desc: "A Ringstrasse-era apartment with a 62 m² salon, 4.1-metre ceilings and the original parquet in every principal room. Restored in 2020 with new services routed behind retained cornices. Two bedrooms face the quiet courtyard; the salon and dining room overlook the Ring.",
		daysAgo: 17, views: 3990,
	},
	{
		title: "Modern flat near Naschmarkt", deal: models.DealRent, ptype: models.TypeApartment,
		country: "AT", city: "Vienna", district: "Mariahilf", address: "Gumpendorfer Straße 63", postal: "1060",
		price: 1420, rooms: 3, beds: 2, baths: 1, area: 78, floor: 5, floors: 7, year: 2019,
		cond: models.ConditionNew, energy: "A", lat: 48.1960, lng: 16.3540, broker: "br-13",
		flags: "furnished,balcony,elevator,new", imgs: []int{108, 40, 109},
		desc: "Five minutes from the Naschmarkt in a 2019 building with a lift and bicycle store. Fitted kitchen, built-in wardrobes and a south-west balcony. Betriebskosten €190 per month on top of the rent.",
		daysAgo: 2, views: 1120,
	},

	// ---- Czechia — Prague --------------------------------------------------
	{
		title: "Art Nouveau apartment in Vinohrady", deal: models.DealSale, ptype: models.TypeApartment,
		country: "CZ", city: "Prague", district: "Vinohrady", address: "Slezská 47", postal: "120 00",
		price: 520000, rooms: 4, beds: 2, baths: 2, area: 128, floor: 3, floors: 5, year: 1908,
		cond: models.ConditionRenovated, energy: "C", lat: 50.0755, lng: 14.4530, broker: "br-14",
		flags: "featured,elevator,balcony,verified", imgs: []int{18, 110, 41, 111},
		desc: "An Art Nouveau building two streets from Náměstí Míru, with the original stained stairwell glazing and ironwork intact. The apartment was renovated in 2021: new oak floors, a rebuilt bathroom and a kitchen opened to the dining room. French windows to a street-facing balcony.",
		daysAgo: 12, views: 2760,
	},
	{
		title: "Riverside apartment in Karlín", deal: models.DealRent, ptype: models.TypeApartment,
		country: "CZ", city: "Prague", district: "Karlín", address: "Sokolovská 112", postal: "186 00",
		price: 1180, rooms: 3, beds: 2, baths: 1, area: 84, floor: 6, floors: 8, year: 2017,
		cond: models.ConditionGood, energy: "B", lat: 50.0930, lng: 14.4500, broker: "br-14",
		flags: "furnished,balcony,elevator,parking", imgs: []int{112, 42, 113},
		desc: "In the regenerated Karlín district, a short walk from the river embankment and Křižíkova metro. Furnished to a good standard with a fitted kitchen and a balcony facing the courtyard. Parking in the underground garage available for an extra CZK 3,500 per month.",
		daysAgo: 6, views: 1440,
	},
	{
		title: "Office floor in Karlín business quarter", deal: models.DealRent, ptype: models.TypeCommercial,
		country: "CZ", city: "Prague", district: "Karlín", address: "Pobřežní 46", postal: "186 00",
		price: 4200, rooms: 8, area: 340, floor: 3, floors: 7, year: 2013,
		cond: models.ConditionGood, energy: "B", lat: 50.0918, lng: 14.4460, broker: "br-14",
		flags: "parking,elevator", imgs: []int{114, 115},
		desc: "An entire third floor delivered as open plan with four enclosed meeting rooms and a fitted kitchen. Raised floors, fan-coil climate control and 1:60 m² parking ratio in the basement garage. Available immediately on a minimum three-year lease.",
		daysAgo: 34, views: 380,
	},

	// ---- Additional stock for filter coverage -----------------------------
	{
		title: "Studio apartment in Tallinn city centre", deal: models.DealRent, ptype: models.TypeApartment,
		country: "EE", city: "Tallinn", district: "Kesklinn", address: "Tartu maantee 52", postal: "10115",
		price: 520, rooms: 1, beds: 1, baths: 1, area: 32, floor: 2, floors: 9, year: 1974,
		cond: models.ConditionSatisfying, energy: "E", lat: 59.4300, lng: 24.7650, broker: "br-01",
		flags: "furnished,elevator,private", imgs: []int{116, 43},
		desc: "A practical studio a few minutes from Ülemiste and the city centre, let directly by the owner. Kitchenette, shower room and a built-in wardrobe. Suitable for one person; rent includes the building maintenance fee.",
		daysAgo: 1, views: 720,
	},
	{
		title: "Semi-detached house in Berlin Zehlendorf", deal: models.DealSale, ptype: models.TypeHouse,
		country: "DE", city: "Berlin", district: "Zehlendorf", address: "Onkel-Tom-Straße 71", postal: "14169",
		price: 1120000, rooms: 6, beds: 4, baths: 2, area: 198, land: 540, floors: 2, year: 1962,
		cond: models.ConditionGood, energy: "D", lat: 52.4340, lng: 13.2530, broker: "br-05",
		flags: "parking,garden,terrace", imgs: []int{19, 117, 44},
		desc: "A well-kept 1960s semi on a quiet residential street in Zehlendorf, backing onto gardens. Four bedrooms across the upper floor, a converted attic study and a full basement. Gas condensing boiler renewed 2020; windows replaced 2015.",
		daysAgo: 20, views: 1890,
	},
	{
		title: "Garage and storage unit in Helsinki Kallio", deal: models.DealSale, ptype: models.TypeGarage,
		country: "FI", city: "Helsinki", district: "Kallio", address: "Vaasankatu 15", postal: "00500",
		price: 42000, area: 18, floors: 1, year: 1998,
		cond: models.ConditionGood, lat: 60.1860, lng: 24.9500, broker: "br-09",
		flags: "private", imgs: []int{118},
		desc: "A lock-up garage with power and a separate 4 m² storage cage in a residential block in Kallio. Suitable for a single vehicle plus seasonal storage. Housing company charge €38 per month.",
		daysAgo: 40, views: 210,
	},
	{
		title: "Agricultural land parcel outside Lisbon", deal: models.DealSale, ptype: models.TypeLand,
		country: "PT", city: "Lisbon", district: "Sintra", address: "Estrada de Colares 88", postal: "2710-000",
		price: 265000, land: 12400, year: 0,
		cond: models.ConditionGood, lat: 38.7990, lng: -9.3900, broker: "br-11",
		flags: "", imgs: []int{119, 120},
		desc: "A 1.24-hectare parcel on the Colares road with existing olive and vine planting. Classified as agricultural with an approved footprint for a 180 m² support building. Borehole in place; mains electricity 240 metres from the boundary.",
		daysAgo: 45, views: 430,
	},
	{
		title: "Renovated apartment in Barcelona Poblenou", deal: models.DealSale, ptype: models.TypeApartment,
		country: "ES", city: "Barcelona", district: "Poblenou", address: "Carrer de Pujades 172", postal: "08005",
		price: 465000, rooms: 3, beds: 2, baths: 2, area: 89, floor: 4, floors: 6, year: 1998,
		cond: models.ConditionRenovated, energy: "C", lat: 41.4010, lng: 2.1980, broker: "br-06",
		flags: "balcony,elevator,parking", imgs: []int{121, 20, 122},
		desc: "Ten minutes from Bogatell beach in the Poblenou grid, renovated throughout in 2023. Two double bedrooms, two full bathrooms and an open kitchen with a utility area. Includes a parking space in the building's garage.",
		daysAgo: 15, views: 2080,
	},
	{
		title: "Townhouse with courtyard in Lisbon Alfama", deal: models.DealSale, ptype: models.TypeHouse,
		country: "PT", city: "Lisbon", district: "Alfama", address: "Rua de São Miguel 34", postal: "1100-000",
		price: 720000, rooms: 5, beds: 3, baths: 3, area: 172, land: 95, floors: 3, year: 1890,
		cond: models.ConditionRenovated, energy: "C", lat: 38.7120, lng: -9.1290, broker: "br-10",
		flags: "terrace,garden,verified", imgs: []int{21, 123, 124},
		desc: "A narrow three-storey house rebuilt behind its historic façade in 2020, arranged with living space at street level, bedrooms above and a roof terrace looking across Alfama to the river. Private tiled courtyard at the rear. Air conditioning throughout.",
		daysAgo: 24, views: 3210,
	},
	{
		title: "Apartment in a new Helsinki development", deal: models.DealSale, ptype: models.TypeApartment,
		country: "FI", city: "Helsinki", district: "Kalasatama", address: "Kalasatamankatu 9", postal: "00580",
		price: 495000, rooms: 3, beds: 2, baths: 1, area: 72, floor: 14, floors: 21, year: 2026,
		cond: models.ConditionNew, energy: "A", lat: 60.1870, lng: 24.9800, broker: "br-08",
		flags: "balcony,elevator,parking,sauna,new,verified", imgs: []int{22, 125, 126},
		desc: "A fourteenth-floor apartment in the Kalasatama tower, completing next quarter. Glazed balcony facing south-east over the harbour, apartment sauna and a fitted kitchen with integrated appliances included in the price. Residents' gym, roof terrace and car-share scheme.",
		daysAgo: 4, views: 1650,
	},
	{
		title: "Warehouse conversion in Amsterdam Noord", deal: models.DealSale, ptype: models.TypeCommercial,
		country: "NL", city: "Amsterdam", district: "Noord", address: "Distelweg 88", postal: "1031 HD",
		price: 1450000, rooms: 6, area: 420, floors: 2, year: 1954,
		cond: models.ConditionRenovated, energy: "C", lat: 52.4010, lng: 4.8890, broker: "br-12",
		flags: "parking,terrace", imgs: []int{127, 23, 128},
		desc: "A former shipping warehouse on the Noord waterfront, converted in 2019 to creative studio space across two levels. Currently let to three tenants on rolling contracts. Sold with vacant possession available from January, or as an investment with tenants in place.",
		daysAgo: 38, views: 990,
	},
	{
		title: "Attic apartment in Vienna Neubau", deal: models.DealSale, ptype: models.TypeApartment,
		country: "AT", city: "Vienna", district: "Neubau", address: "Lindengasse 40", postal: "1070",
		price: 725000, rooms: 4, beds: 2, baths: 2, area: 112, floor: 5, floors: 5, year: 2011,
		cond: models.ConditionGood, energy: "B", lat: 48.2010, lng: 16.3500, broker: "br-13",
		flags: "terrace,elevator,verified", imgs: []int{24, 129, 130},
		desc: "A Dachgeschoss conversion in the seventh district with a 38 m² terrace facing south. Double-height living space under exposed rafters, two bedrooms on the upper level and a separate study. Lift serves the apartment directly.",
		daysAgo: 18, views: 2340,
	},
}

// buildProperties expands the seed table into full domain objects.
func buildProperties(now time.Time, pool []string) []models.Property {
	out := make([]models.Property, 0, len(propertySeeds))

	for i, s := range propertySeeds {
		flags := flagSet(s.flags)

		p := models.Property{
			ID:          fmt.Sprintf("pr-%03d", i+1),
			Title:       s.title,
			Deal:        s.deal,
			Type:        s.ptype,
			Status:      s.status,
			CountryCode: s.country,
			Country:     countryName(s.country),
			City:        s.city,
			District:    s.district,
			Address:     s.address,
			PostalCode:  s.postal,
			Coords:      models.Coordinates{Lat: s.lat, Lng: s.lng},
			Precision:   models.PrecisionExact,
			Rooms:       s.rooms,
			Bedrooms:    s.beds,
			Bathrooms:   s.baths,
			Area:        s.area,
			LandArea:    s.land,
			Floor:       s.floor,
			TotalFloors: s.floors,
			BuildYear:   s.year,
			Condition:   s.cond,
			EnergyRating: s.energy,
			Description: s.desc,
			BrokerID:    s.broker,
			DevelopmentID: s.devID,
			Views:       s.views,
			Saves:       s.views / 14,
			CreatedAt:   now.AddDate(0, 0, -s.daysAgo),
			UpdatedAt:   now.AddDate(0, 0, -s.daysAgo/2),
			ExpiresAt:   now.AddDate(0, 0, 60-s.daysAgo),
		}

		if p.Status == "" {
			p.Status = models.StatusActive
		}
		p.Price = models.Money{Amount: s.price, Currency: currencyFor(s.country)}
		if s.deal == models.DealRent {
			p.RentPeriod = "month"
		}
		if s.area > 0 && s.deal == models.DealSale {
			p.PricePerM2 = math.Round(s.price / s.area)
		}

		p.Furnished = flags["furnished"]
		p.Parking = flags["parking"]
		p.Balcony = flags["balcony"]
		p.Garden = flags["garden"]
		p.Elevator = flags["elevator"]
		p.Terrace = flags["terrace"]
		p.Sauna = flags["sauna"]
		p.SeaView = flags["seaview"]
		p.IsFeatured = flags["featured"]
		p.IsVerified = flags["verified"]
		p.IsNewDevelopment = flags["new"]

		if flags["private"] {
			p.SellerKind = models.SellerPrivate
			p.BrokerID = ""
		} else {
			p.SellerKind = models.SellerBroker
			p.AgencyID = agencyForBroker(s.broker)
		}

		p.Slug = slugify(s.title) + "-" + p.ID
		p.Images = buildGallery(pool, s.imgs, p)
		p.Features = featuresFor(p)
		p.Highlights = highlightsFor(p)

		out = append(out, p)
	}
	return out
}

// flagSet turns "featured,parking" into a lookup map.
func flagSet(s string) map[string]bool {
	m := map[string]bool{}
	for _, f := range strings.Split(s, ",") {
		if f = strings.TrimSpace(f); f != "" {
			m[f] = true
		}
	}
	return m
}

// The photo pool is laid out by subject so galleries can be assembled
// correctly per property type:
//
//	p001–p060  building exteriors
//	p061–p130  interiors (living, kitchen, bedroom, bathroom, office)
//	p131–p140  land and open plots
const (
	poolExteriorEnd = 60
	poolInteriorEnd = 130
)

// buildGallery assembles a plausible photo set for one listing.
//
// A cover image has to match the listing: a plot of land must not lead with a
// bathroom, and an apartment must not lead with one either. The seed's index
// list is used only as a deterministic offset, so galleries stay stable between
// runs while always drawing from the right subject range.
func buildGallery(pool []string, seed []int, p models.Property) []models.Image {
	if len(pool) == 0 {
		return nil
	}

	pick := func(lo, hi, n int) []string {
		lo-- // 1-based photo numbers to 0-based slice indices; hi stays exclusive
		if lo < 0 {
			lo = 0
		}
		if hi > len(pool) {
			hi = len(pool)
		}
		if lo >= hi {
			return nil
		}
		span := hi - lo
		offset := 0
		if len(seed) > 0 {
			offset = seed[0]
		}
		var out []string
		for i := 0; i < n; i++ {
			out = append(out, pool[lo+((offset+i*3)%span)])
		}
		return out
	}

	var urls []string
	switch p.Type {
	case models.TypeLand:
		// Plots: open land only, no rooms.
		urls = pick(poolInteriorEnd+1, len(pool), 3)
	case models.TypeGarage:
		urls = pick(1, poolExteriorEnd, 2)
	case models.TypeCommercial:
		// One exterior, then office interiors.
		urls = append(pick(1, poolExteriorEnd, 1), pick(poolInteriorEnd-24, poolInteriorEnd, 3)...)
	default:
		// Residential: exterior cover, then a spread of rooms.
		urls = append(pick(1, poolExteriorEnd, 1), pick(poolExteriorEnd+1, poolInteriorEnd, 4)...)
	}

	seen := map[string]bool{}
	out := make([]models.Image, 0, len(urls))
	for _, u := range urls {
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, models.Image{URL: u, Alt: p.Title, Width: 1200, Height: 900})
	}
	return out
}

// featuresFor derives the amenity chip list shown on the detail page.
func featuresFor(p models.Property) []models.Feature {
	var f []models.Feature
	add := func(key, label, icon string, on bool) {
		if on {
			f = append(f, models.Feature{Key: key, Label: label, Icon: icon})
		}
	}
	add("parking", "Parking", "car", p.Parking)
	add("balcony", "Balcony", "sun", p.Balcony)
	add("terrace", "Terrace", "sun", p.Terrace)
	add("garden", "Garden", "leaf", p.Garden)
	add("elevator", "Elevator", "elevator", p.Elevator)
	add("sauna", "Sauna", "flame", p.Sauna)
	add("seaview", "Sea view", "waves", p.SeaView)
	add("furnished", "Furnished", "sofa", p.Furnished)
	if p.EnergyRating != "" {
		add("energy", "Energy class "+p.EnergyRating, "bolt", true)
	}
	if p.BuildYear > 0 {
		add("year", fmt.Sprintf("Built %d", p.BuildYear), "calendar", true)
	}
	return f
}

// highlightsFor produces the short bullet list under the detail-page title.
func highlightsFor(p models.Property) []string {
	var h []string
	if p.Condition == models.ConditionNew {
		h = append(h, "Newly built and never occupied")
	}
	if p.Condition == models.ConditionRenovated {
		h = append(h, "Fully renovated")
	}
	if p.SeaView {
		h = append(h, "Direct sea or water views")
	}
	if p.Terrace {
		h = append(h, "Private terrace")
	}
	if p.EnergyRating == "A" {
		h = append(h, "Energy class A — low running costs")
	}
	if p.Parking {
		h = append(h, "Parking included")
	}
	if len(h) == 0 {
		h = append(h, "Available to view this week")
	}
	return h
}

// slugify makes a URL-safe slug from a title.
func slugify(s string) string {
	s = strings.ToLower(s)
	repl := strings.NewReplacer(
		"ä", "a", "ö", "o", "ü", "u", "õ", "o", "å", "a", "é", "e", "è", "e",
		"ã", "a", "ç", "c", "í", "i", "á", "a", "ó", "o", "ú", "u", "ř", "r",
		"š", "s", "ž", "z", "ě", "e", "č", "c", "ň", "n", "ď", "d", "ť", "t",
		"'", "", "’", "", "—", "-", "–", "-",
	)
	s = repl.Replace(s)

	var b strings.Builder
	lastDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
