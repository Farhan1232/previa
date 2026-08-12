package data

import (
	"fmt"
	"time"

	"previa/pkg/models"
)

// ---------------------------------------------------------------------------
// Countries and currencies
// ---------------------------------------------------------------------------

var countries = []models.Country{
	{Code: "EE", Name: "Estonia", Flag: "🇪🇪", Currency: "EUR", Locale: "en", Lat: 59.4370, Lng: 24.7536, Zoom: 11,
		Cities: []string{"Tallinn", "Tartu", "Pärnu", "Narva", "Viimsi"}},
	{Code: "DE", Name: "Germany", Flag: "🇩🇪", Currency: "EUR", Locale: "de", Lat: 52.5200, Lng: 13.4050, Zoom: 11,
		Cities: []string{"Berlin", "Munich", "Hamburg", "Frankfurt", "Cologne"}},
	{Code: "ES", Name: "Spain", Flag: "🇪🇸", Currency: "EUR", Locale: "es", Lat: 41.3874, Lng: 2.1686, Zoom: 12,
		Cities: []string{"Barcelona", "Madrid", "Valencia", "Málaga", "Seville"}},
	{Code: "FI", Name: "Finland", Flag: "🇫🇮", Currency: "EUR", Locale: "en", Lat: 60.1699, Lng: 24.9384, Zoom: 11,
		Cities: []string{"Helsinki", "Espoo", "Tampere", "Turku", "Oulu"}},
	{Code: "PT", Name: "Portugal", Flag: "🇵🇹", Currency: "EUR", Locale: "en", Lat: 38.7223, Lng: -9.1393, Zoom: 12,
		Cities: []string{"Lisbon", "Porto", "Cascais", "Faro", "Sintra"}},
	{Code: "NL", Name: "Netherlands", Flag: "🇳🇱", Currency: "EUR", Locale: "en", Lat: 52.3676, Lng: 4.9041, Zoom: 12,
		Cities: []string{"Amsterdam", "Rotterdam", "Utrecht", "The Hague", "Eindhoven"}},
	{Code: "AT", Name: "Austria", Flag: "🇦🇹", Currency: "EUR", Locale: "de", Lat: 48.2082, Lng: 16.3738, Zoom: 12,
		Cities: []string{"Vienna", "Graz", "Salzburg", "Linz", "Innsbruck"}},
	// Euro like every other market for now: the client asked the frontend to
	// show one currency. The Currency field stays on the model and the symbol
	// table still knows Kč, so restoring a koruna market is this one value.
	{Code: "CZ", Name: "Czechia", Flag: "🇨🇿", Currency: "EUR", Locale: "en", Lat: 50.0755, Lng: 14.4378, Zoom: 12,
		Cities: []string{"Prague", "Brno", "Ostrava", "Plzeň", "Liberec"}},
}

func countryName(code string) string {
	for _, c := range countries {
		if c.Code == code {
			return c.Name
		}
	}
	return code
}

func currencyFor(code string) string {
	for _, c := range countries {
		if c.Code == code {
			return c.Currency
		}
	}
	return "EUR"
}

// ---------------------------------------------------------------------------
// Country-specific advertising banners
// ---------------------------------------------------------------------------

// Advertising slots, keyed by market and placement.
//
// The homepage strip is seeded Active:false. The client asked for it to be
// dismissed for now but kept editable, so the copy stays here and admin can
// switch it back on per country without anything being re-entered.
//
// The search strip is a separate, shorter placement with its own copy — it is
// not the homepage banner re-used, and it is live.
var banners = []models.Banner{
	{ID: "bn-ee", CountryCode: "EE", Sponsor: "Baltic Home Loans",
		Headline: "Fixed-rate home loans from 3.4% in Estonia",
		Body:     "Get a decision in principle in two working days. No arrangement fee on applications submitted before 30 September.",
		CTALabel: "Check your rate", CTAHref: "/help", Theme: "navy", Placement: "home", Active: false},
	{ID: "bn-de", CountryCode: "DE", Sponsor: "Hausbank Direkt",
		Headline: "Baufinanzierung with a 15-year fixed rate",
		Body:     "Compare offers from 340 German lenders in one application. Free consultation with a local advisor in Berlin, Munich and Hamburg.",
		CTALabel: "Compare offers", CTAHref: "/help", Theme: "slate", Placement: "home", Active: false},
	{ID: "bn-es", CountryCode: "ES", Sponsor: "Mediterráneo Hipotecas",
		Headline: "Mortgages for non-resident buyers in Spain",
		Body:     "Up to 70% financing for international buyers, with English-speaking advisors in Barcelona, Valencia and Málaga.",
		CTALabel: "Speak to an advisor", CTAHref: "/help", Theme: "gold", Placement: "home", Active: false},
	{ID: "bn-fi", CountryCode: "FI", Sponsor: "Pohjola Asuntolaina",
		Headline: "Housing loans with a repayment holiday",
		Body:     "Take up to twelve months without capital repayments when you buy your first home in Finland.",
		CTALabel: "See the terms", CTAHref: "/help", Theme: "navy", Placement: "home", Active: false},
	{ID: "bn-pt", CountryCode: "PT", Sponsor: "Atlântico Crédito Habitação",
		Headline: "Portuguese mortgages for residents and expatriates",
		Body:     "Fixed, mixed and variable rates with no early-repayment penalty on fixed-term products.",
		CTALabel: "Request a quote", CTAHref: "/help", Theme: "slate", Placement: "home", Active: false},
	{ID: "bn-nl", CountryCode: "NL", Sponsor: "Randstad Hypotheken",
		Headline: "Hypotheek advice for Amsterdam buyers",
		Body:     "Independent advice across 28 Dutch lenders, including options for self-employed applicants.",
		CTALabel: "Book a consultation", CTAHref: "/help", Theme: "navy", Placement: "home", Active: false},
	{ID: "bn-at", CountryCode: "AT", Sponsor: "Donau Wohnkredit",
		Headline: "Wohnbaufinanzierung across Austria",
		Body:     "Fixed rates held for 25 years, with subsidised options for energy-efficient properties.",
		CTALabel: "Calculate your loan", CTAHref: "/help", Theme: "slate", Placement: "home", Active: false},
	{ID: "bn-cz", CountryCode: "CZ", Sponsor: "Vltava Hypotéka",
		Headline: "Czech mortgages with a rate guarantee",
		Body:     "Lock today's rate for six months while you search for the right property in Prague or Brno.",
		CTALabel: "Lock your rate", CTAHref: "/help", Theme: "gold", Placement: "home", Active: false},

	// --- Search results strip -------------------------------------------------
	// Shorter than the homepage banner and separately worded, matching the
	// promoted-development strip the client referenced on kinnisvara24.
	{ID: "bs-ee", CountryCode: "EE", Sponsor: "Uusarendused",
		Headline: "New developments in Tallinn and Tartu",
		Body:     "Twelve projects taking reservations this autumn, from 137 000 €.",
		CTALabel: "See developments", CTAHref: "/developments", Theme: "gold",
		Placement: "search", Active: true},
	{ID: "bs-de", CountryCode: "DE", Sponsor: "Neubau Berlin",
		Headline: "Newly completed apartments in Berlin",
		Body:     "Move-in ready homes in Mitte, Pankow and Friedrichshain.",
		CTALabel: "See developments", CTAHref: "/developments", Theme: "navy",
		Placement: "search", Active: true},
	{ID: "bs-es", CountryCode: "ES", Sponsor: "Obra Nueva",
		Headline: "New build on the Catalan coast",
		Body:     "Sea-facing developments in Barcelona, Sitges and Badalona.",
		CTALabel: "See developments", CTAHref: "/developments", Theme: "gold",
		Placement: "search", Active: true},
	{ID: "bs-fi", CountryCode: "FI", Sponsor: "Uudiskohteet",
		Headline: "New homes in Helsinki and Espoo",
		Body:     "Kalasatama, Jätkäsaari and Otaniemi projects now selling.",
		CTALabel: "See developments", CTAHref: "/developments", Theme: "navy",
		Placement: "search", Active: true},
	{ID: "bs-pt", CountryCode: "PT", Sponsor: "Novos Empreendimentos",
		Headline: "New developments in Lisbon and Porto",
		Body:     "Riverside and city-centre projects with completion in 2027.",
		CTALabel: "See developments", CTAHref: "/developments", Theme: "slate",
		Placement: "search", Active: true},
	{ID: "bs-nl", CountryCode: "NL", Sponsor: "Nieuwbouw",
		Headline: "Nieuwbouw across the Randstad",
		Body:     "Amsterdam, Utrecht and Rotterdam projects open for registration.",
		CTALabel: "See developments", CTAHref: "/developments", Theme: "navy",
		Placement: "search", Active: true},
	{ID: "bs-at", CountryCode: "AT", Sponsor: "Neubauprojekte",
		Headline: "New apartments in Vienna",
		Body:     "Projects in Leopoldstadt, Favoriten and Donaustadt.",
		CTALabel: "See developments", CTAHref: "/developments", Theme: "gold",
		Placement: "search", Active: true},
	{ID: "bs-cz", CountryCode: "CZ", Sponsor: "Nové projekty",
		Headline: "New developments in Prague",
		Body:     "Karlín, Smíchov and Vinohrady projects taking reservations.",
		CTALabel: "See developments", CTAHref: "/developments", Theme: "slate",
		Placement: "search", Active: true},
}

// ---------------------------------------------------------------------------
// Agencies and brokers
// ---------------------------------------------------------------------------

var agencies = []models.Agency{
	{ID: "ag-01", Slug: "kadaka-kinnisvara", Name: "Kadaka Kinnisvara", CountryCode: "EE", City: "Tallinn",
		Address: "Roseni 10, 10111 Tallinn", Phone: "+372 640 1180", Email: "info@kadaka.example",
		Website: "kadaka.example", BrokerCount: 24, ListingCount: 186, Founded: 2004, IsVerified: true,
		Description: "An Estonian agency specialising in Tallinn's historic districts and the Viimsi peninsula. Kadaka has handled Old Town transactions since 2004 and maintains a dedicated restoration advisory team for protected buildings."},
	{ID: "ag-02", Slug: "hauptstadt-immobilien", Name: "Hauptstadt Immobilien", CountryCode: "DE", City: "Berlin",
		Address: "Kurfürstendamm 194, 10707 Berlin", Phone: "+49 30 8871 4400", Email: "kontakt@hauptstadt.example",
		Website: "hauptstadt.example", BrokerCount: 41, ListingCount: 312, Founded: 1998, IsVerified: true,
		Description: "Berlin's Altbau specialist, covering Mitte, Prenzlauer Berg, Kreuzberg and the western residential districts. Hauptstadt advises on Denkmalschutz obligations and Milieuschutz restrictions as standard."},
	{ID: "ag-03", Slug: "mediterrania-propietats", Name: "Mediterrània Propietats", CountryCode: "ES", City: "Barcelona",
		Address: "Carrer de Balmes 152, 08008 Barcelona", Phone: "+34 932 40 71 00", Email: "hola@mediterrania.example",
		Website: "mediterrania.example", BrokerCount: 33, ListingCount: 245, Founded: 2009, IsVerified: true,
		Description: "A Barcelona agency working across the Eixample grid, Gràcia and the upper Sarrià districts, with a dedicated desk for international buyers and NIE assistance."},
	{ID: "ag-04", Slug: "pohjola-koti", Name: "Pohjola Koti", CountryCode: "FI", City: "Helsinki",
		Address: "Aleksanterinkatu 17, 00100 Helsinki", Phone: "+358 9 6220 340", Email: "myynti@pohjolakoti.example",
		Website: "pohjolakoti.example", BrokerCount: 19, ListingCount: 128, Founded: 2011, IsVerified: true,
		Description: "Helsinki agency focused on the waterfront regeneration districts — Jätkäsaari, Kalasatama and Hernesaari — alongside the classic functionalist stock in Töölö and Käpylä."},
	{ID: "ag-05", Slug: "tejo-properties", Name: "Tejo Properties", CountryCode: "PT", City: "Lisbon",
		Address: "Avenida da Liberdade 110, 1250-146 Lisboa", Phone: "+351 21 340 8800", Email: "hello@tejo.example",
		Website: "tejo.example", BrokerCount: 27, ListingCount: 203, Founded: 2013, IsVerified: true,
		Description: "Lisbon and Cascais agency covering Pombaline restoration in the historic centre and new coastal development along the Estoril line. Advises on the Portuguese residency framework."},
	{ID: "ag-06", Slug: "grachten-makelaars", Name: "Grachten Makelaars", CountryCode: "NL", City: "Amsterdam",
		Address: "Herengracht 458, 1017 CA Amsterdam", Phone: "+31 20 620 9900", Email: "info@grachten.example",
		Website: "grachten.example", BrokerCount: 16, ListingCount: 94, Founded: 1992, IsVerified: true,
		Description: "An Amsterdam agency working within the canal ring since 1992, with particular expertise in listed monument transactions and the associated maintenance obligations."},
	{ID: "ag-07", Slug: "ringhaus-wien", Name: "Ringhaus Wien", CountryCode: "AT", City: "Vienna",
		Address: "Schottengasse 4, 1010 Wien", Phone: "+43 1 533 6100", Email: "office@ringhaus.example",
		Website: "ringhaus.example", BrokerCount: 21, ListingCount: 117, Founded: 2006, IsVerified: true,
		Description: "Vienna agency handling Gründerzeit and Ringstrasse-era apartments in the inner districts, plus Dachgeschoss conversions in Neubau and Mariahilf."},
	{ID: "ag-08", Slug: "vltava-reality", Name: "Vltava Reality", CountryCode: "CZ", City: "Prague",
		Address: "Vinohradská 33, 120 00 Praha 2", Phone: "+420 222 510 400", Email: "info@vltava.example",
		Website: "vltava.example", BrokerCount: 23, ListingCount: 156, Founded: 2008, IsVerified: true,
		Description: "Prague agency covering Vinohrady, Karlín and Smíchov, with a commercial division handling office and retail leasing in the Karlín business quarter."},
}

type brokerSeed struct {
	id, name, title, agency, country, city string
	phone, email                           string
	langs, specs                           []string
	rating                                 float64
	reviews, listings, sold, years         int
	promoted                               bool
	bio                                    string
}

var brokerSeeds = []brokerSeed{
	{"br-01", "Kadri Tamm", "Senior broker, level 7", "ag-01", "EE", "Tallinn", "+372 5123 4471", "kadri.tamm@kadaka.example",
		[]string{"Estonian", "English", "Finnish"}, []string{"Old Town", "Historic restoration", "Apartments"},
		4.9, 87, 14, 212, 15, true,
		"Kadri has worked Tallinn's Old Town and Kesklinn market since 2011 and holds the level 7 professional certificate. She advises on the restoration obligations that come with protected buildings and works regularly with Finnish and Swedish buyers."},
	{"br-02", "Marten Sepp", "Broker, level 6", "ag-01", "EE", "Tallinn", "+372 5220 9038", "marten.sepp@kadaka.example",
		[]string{"Estonian", "English", "Russian"}, []string{"New developments", "Waterfront", "Investment"},
		4.7, 54, 11, 148, 9, true,
		"Marten covers the Kalamaja and Põhja-Tallinn regeneration districts and handles most of Kadaka's new-development instructions. He advises private landlords on yield and rental positioning."},
	{"br-03", "Liis Kask", "Broker, level 5", "ag-01", "EE", "Tallinn", "+372 5661 2290", "liis.kask@kadaka.example",
		[]string{"Estonian", "English"}, []string{"Family homes", "Nõmme", "Land"},
		4.8, 41, 9, 96, 7, false,
		"Liis specialises in detached houses and building plots in Nõmme, Viimsi and the western suburbs, working mostly with families relocating from central Tallinn."},

	{"br-04", "Jonas Weber", "Immobilienmakler (IHK)", "ag-02", "DE", "Berlin", "+49 172 884 2210", "j.weber@hauptstadt.example",
		[]string{"German", "English"}, []string{"Altbau", "Prenzlauer Berg", "Condominiums"},
		4.8, 132, 18, 284, 14, true,
		"Jonas handles Altbau apartment sales across Prenzlauer Berg, Mitte and Friedrichshain. He is IHK-certified and advises buyers on Milieuschutz restrictions and the conversion rules that apply in protected areas."},
	{"br-05", "Annika Hoffmann", "Senior Immobilienmaklerin", "ag-02", "DE", "Berlin", "+49 170 553 8841", "a.hoffmann@hauptstadt.example",
		[]string{"German", "English", "French"}, []string{"Houses", "Grunewald", "Prime residential"},
		4.9, 96, 12, 176, 18, true,
		"Annika leads Hauptstadt's prime residential desk, covering Grunewald, Dahlem and Zehlendorf. She has handled the firm's highest-value transactions for the past six years and works discreetly on off-market instructions."},

	{"br-06", "Marc Puig", "Agent immobiliari (API)", "ag-03", "ES", "Barcelona", "+34 649 220 118", "marc.puig@mediterrania.example",
		[]string{"Catalan", "Spanish", "English"}, []string{"Eixample", "Modernista", "Prime residential"},
		4.8, 118, 16, 231, 13, true,
		"Marc is an API-registered agent covering the Eixample and Sarrià districts. He specialises in Modernista buildings and advises international buyers on the Spanish purchase process end to end."},
	{"br-07", "Elena Serra", "Agent immobiliària", "ag-03", "ES", "Barcelona", "+34 655 907 340", "elena.serra@mediterrania.example",
		[]string{"Catalan", "Spanish", "English", "Italian"}, []string{"Rentals", "Gràcia", "Barceloneta"},
		4.6, 74, 13, 142, 8, true,
		"Elena runs Mediterrània's rental division across Gràcia, Barceloneta and Poblenou, placing both long-term residents and relocating professionals."},

	{"br-08", "Aino Virtanen", "Kiinteistönvälittäjä LKV", "ag-04", "FI", "Helsinki", "+358 40 552 1180", "aino.virtanen@pohjolakoti.example",
		[]string{"Finnish", "English", "Swedish"}, []string{"Waterfront", "New developments", "Jätkäsaari"},
		4.9, 68, 10, 134, 11, true,
		"Aino is an LKV-qualified broker covering Helsinki's waterfront regeneration areas. She handles most of Pohjola Koti's Jätkäsaari and Kalasatama instructions and advises on housing-company charges and pipe-renovation liabilities."},
	{"br-09", "Mikko Laine", "Kiinteistönvälittäjä LKV", "ag-04", "FI", "Helsinki", "+358 50 331 7729", "mikko.laine@pohjolakoti.example",
		[]string{"Finnish", "English"}, []string{"Töölö", "Functionalist", "Houses"},
		4.7, 52, 8, 108, 12, false,
		"Mikko covers the classic Helsinki districts — Töölö, Kallio and Käpylä — with a focus on 1920s–1950s stock and the renovation programmes those housing companies are running."},

	{"br-10", "Rui Almeida", "Consultor imobiliário (AMI)", "ag-05", "PT", "Lisbon", "+351 912 448 007", "rui.almeida@tejo.example",
		[]string{"Portuguese", "English", "Spanish"}, []string{"Chiado", "Restoration", "Coastal houses"},
		4.9, 104, 15, 198, 12, true,
		"Rui holds an AMI licence and covers central Lisbon and the Cascais coast. He specialises in Pombaline restoration projects and advises international buyers on residency and tax registration."},
	{"br-11", "Sofia Costa", "Consultora imobiliária", "ag-05", "PT", "Lisbon", "+351 934 771 260", "sofia.costa@tejo.example",
		[]string{"Portuguese", "English", "French"}, []string{"Rentals", "Príncipe Real", "Land"},
		4.7, 61, 11, 124, 7, true,
		"Sofia manages Tejo's lettings book across central Lisbon and handles land and development-site instructions in the Sintra and Colares area."},

	{"br-12", "Daan Visser", "Register-makelaar NVM", "ag-06", "NL", "Amsterdam", "+31 6 2244 8890", "daan.visser@grachten.example",
		[]string{"Dutch", "English", "German"}, []string{"Canal ring", "Monuments", "Commercial"},
		4.8, 79, 9, 156, 16, true,
		"Daan is an NVM-registered broker working inside the Amsterdam canal ring. He advises on rijksmonument obligations and handles the firm's commercial conversions in Noord."},

	{"br-13", "Lukas Bauer", "Immobilienmakler", "ag-07", "AT", "Vienna", "+43 664 220 4417", "lukas.bauer@ringhaus.example",
		[]string{"German", "English"}, []string{"Innere Stadt", "Gründerzeit", "Dachgeschoss"},
		4.8, 71, 12, 143, 11, true,
		"Lukas covers Vienna's first, sixth and seventh districts, handling Ringstrasse-era apartments and rooftop conversions. He advises buyers on the Austrian Wohnungseigentum framework."},

	{"br-14", "Petra Novák", "Realitní makléřka", "ag-08", "CZ", "Prague", "+420 776 330 128", "petra.novak@vltava.example",
		[]string{"Czech", "English", "German"}, []string{"Vinohrady", "Karlín", "Commercial leasing"},
		4.7, 88, 14, 167, 10, true,
		"Petra works across Vinohrady and Karlín on both residential sales and office leasing. She handles Vltava's commercial instructions in the Karlín business quarter."},
}

func buildBrokers() []models.Broker {
	out := make([]models.Broker, 0, len(brokerSeeds))
	for _, s := range brokerSeeds {
		out = append(out, models.Broker{
			ID: s.id, Slug: slugify(s.name), Name: s.name, Title: s.title,
			AgencyID: s.agency, AgencyName: agencyName(s.agency),
			Photo: "/static/img/brokers/" + s.id + ".jpg",
			Phone: s.phone, Email: s.email,
			CountryCode: s.country, City: s.city,
			Languages: s.langs, Specialties: s.specs,
			Bio:    s.bio,
			Rating: s.rating, Reviews: s.reviews,
			ActiveListings: s.listings, SoldCount: s.sold, YearsActive: s.years,
			IsPromoted: s.promoted, IsVerified: true,
		})
	}
	return out
}

func agencyName(id string) string {
	for _, a := range agencies {
		if a.ID == id {
			return a.Name
		}
	}
	return ""
}

func agencyForBroker(brokerID string) string {
	for _, b := range brokerSeeds {
		if b.id == brokerID {
			return b.agency
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// Developments
// ---------------------------------------------------------------------------

func buildDevelopments(now time.Time) []models.Development {
	img := func(n int) models.Image {
		return models.Image{URL: fmt.Sprintf("/static/img/developments/d%02d.jpg", n), Width: 1400, Height: 900}
	}
	pimg := func(n int) models.Image {
		return models.Image{URL: fmt.Sprintf("/static/img/properties/p%03d.jpg", n), Width: 1200, Height: 900}
	}

	devs := []models.Development{
		{ID: "dv-01", Slug: "kalaranna-kvartal", Name: "Kalaranna Kvartal", Developer: "Nordvest Arendus",
			CountryCode: "EE", City: "Tallinn", District: "Kalamaja", Address: "Kalaranna 8, 10415 Tallinn",
			Coords:    models.Coordinates{Lat: 59.4462, Lng: 24.7381},
			PriceFrom: models.Money{Amount: 219000, Currency: "EUR"}, AreaFrom: 38, AreaTo: 142,
			TotalUnits: 186, AvailableUnits: 41, Floors: 7, CompletionQuarter: "Q2 2027", EnergyRating: "A",
			Description: "Six buildings on the Kalamaja seafront, arranged around a courtyard that opens directly onto the coastal promenade. Every apartment has a glazed balcony or terrace, and the two seaward buildings have unobstructed views across the bay. Ground floors are given over to cafés, a food market and a childcare centre.",
			Cover:       img(1), Images: []models.Image{img(1), pimg(2), pimg(66), pimg(74)}},

		{ID: "dv-02", Slug: "spreebogen-residenz", Name: "Spreebogen Residenz", Developer: "Hauptstadt Projekt GmbH",
			CountryCode: "DE", City: "Berlin", District: "Moabit", Address: "Alt-Moabit 102, 10559 Berlin",
			Coords:    models.Coordinates{Lat: 52.5250, Lng: 13.3420},
			PriceFrom: models.Money{Amount: 398000, Currency: "EUR"}, AreaFrom: 46, AreaTo: 168,
			TotalUnits: 240, AvailableUnits: 63, Floors: 9, CompletionQuarter: "Q4 2027", EnergyRating: "A",
			Description: "A riverside development on the Spree bend in Moabit, ten minutes from the Hauptbahnhof. The scheme includes a private residents' garden along the water, an underground garage with EV charging throughout, and a concierge desk staffed on weekdays.",
			Cover:       img(2), Images: []models.Image{img(2), pimg(4), pimg(67), pimg(75)}},

		{ID: "dv-03", Slug: "diagonal-mar-vistes", Name: "Diagonal Mar Vistes", Developer: "Mediterrània Desenvolupament",
			CountryCode: "ES", City: "Barcelona", District: "Sant Martí", Address: "Avinguda Diagonal 640, 08017 Barcelona",
			Coords:    models.Coordinates{Lat: 41.4090, Lng: 2.2160},
			PriceFrom: models.Money{Amount: 445000, Currency: "EUR"}, AreaFrom: 62, AreaTo: 195,
			TotalUnits: 154, AvailableUnits: 28, Floors: 14, CompletionQuarter: "Q1 2027", EnergyRating: "A",
			Description: "Two towers close to the Diagonal Mar park and the beach, with a shared rooftop pool and gym on the eleventh floor. Apartments from the eighth floor upward have sea views; all units include a terrace of at least twelve square metres.",
			Cover:       img(3), Images: []models.Image{img(3), pimg(7), pimg(68), pimg(77)}},

		{ID: "dv-04", Slug: "kalasatama-tornit", Name: "Kalasatama Tornit", Developer: "Pohjola Rakennus",
			CountryCode: "FI", City: "Helsinki", District: "Kalasatama", Address: "Kalasatamankatu 9, 00580 Helsinki",
			Coords:    models.Coordinates{Lat: 60.1870, Lng: 24.9800},
			PriceFrom: models.Money{Amount: 289000, Currency: "EUR"}, AreaFrom: 32, AreaTo: 118,
			TotalUnits: 312, AvailableUnits: 74, Floors: 21, CompletionQuarter: "Q3 2026", EnergyRating: "A",
			Description: "A residential tower above the Kalasatama metro station and shopping centre, with a residents' sauna suite and roof terrace on the top floor. Apartments include a fitted kitchen, and most have a glazed balcony usable through the winter months.",
			Cover:       img(4), Images: []models.Image{img(4), pimg(22), pimg(69), pimg(88)}},

		{ID: "dv-05", Slug: "estoril-jardins", Name: "Estoril Jardins", Developer: "Tejo Desenvolvimento",
			CountryCode: "PT", City: "Lisbon", District: "Cascais", Address: "Avenida Marginal 220, 2765-247 Estoril",
			Coords:    models.Coordinates{Lat: 38.7050, Lng: -9.3960},
			PriceFrom: models.Money{Amount: 520000, Currency: "EUR"}, AreaFrom: 78, AreaTo: 240,
			TotalUnits: 96, AvailableUnits: 33, Floors: 5, CompletionQuarter: "Q2 2027", EnergyRating: "A",
			Description: "Low-rise apartments and duplexes set in two hectares of Mediterranean gardens between Estoril and Cascais, a short walk from the coastal path. Two outdoor pools, a residents' gym and covered parking for every unit.",
			Cover:       img(5), Images: []models.Image{img(5), pimg(35), pimg(70), pimg(96)}},

		{ID: "dv-06", Slug: "houthaven-kade", Name: "Houthaven Kade", Developer: "Grachten Ontwikkeling",
			CountryCode: "NL", City: "Amsterdam", District: "Houthaven", Address: "Houthavenkade 40, 1013 Amsterdam",
			Coords:    models.Coordinates{Lat: 52.3930, Lng: 4.8760},
			PriceFrom: models.Money{Amount: 465000, Currency: "EUR"}, AreaFrom: 54, AreaTo: 156,
			TotalUnits: 128, AvailableUnits: 19, Floors: 8, CompletionQuarter: "Q4 2026", EnergyRating: "A",
			Description: "Waterfront apartments on the Houthaven islands, built to an energy-neutral standard with heat pumps and rooftop photovoltaics. Residents have private moorings and a bicycle store with direct access to the western ring route.",
			Cover:       img(6), Images: []models.Image{img(6), pimg(16), pimg(71), pimg(102)}},

		{ID: "dv-07", Slug: "donaupark-wohnen", Name: "Donaupark Wohnen", Developer: "Ringhaus Projektbau",
			CountryCode: "AT", City: "Vienna", District: "Donaustadt", Address: "Wagramer Straße 88, 1220 Wien",
			Coords:    models.Coordinates{Lat: 48.2400, Lng: 16.4300},
			PriceFrom: models.Money{Amount: 335000, Currency: "EUR"}, AreaFrom: 44, AreaTo: 132,
			TotalUnits: 204, AvailableUnits: 88, Floors: 11, CompletionQuarter: "Q1 2028", EnergyRating: "A",
			Description: "A development on the edge of the Donaupark with direct access to the river island and the U1 line. The scheme includes a residents' courtyard, communal roof gardens on each building and a car-sharing fleet included in the service charge.",
			Cover:       img(7), Images: []models.Image{img(7), pimg(17), pimg(72), pimg(106)}},

		{ID: "dv-08", Slug: "karlin-nabrezi", Name: "Karlín Nábřeží", Developer: "Vltava Development",
			CountryCode: "CZ", City: "Prague", District: "Karlín", Address: "Rohanské nábřeží 12, 186 00 Praha 8",
			Coords:    models.Coordinates{Lat: 50.0950, Lng: 14.4560},
			PriceFrom: models.Money{Amount: 268000, Currency: "EUR"}, AreaFrom: 40, AreaTo: 128,
			TotalUnits: 168, AvailableUnits: 52, Floors: 8, CompletionQuarter: "Q3 2027", EnergyRating: "B",
			Description: "Apartments along the Vltava embankment in Karlín, next to the new river park. The ground floors hold retail and a co-working space reserved for residents, and the upper floors step back to create private roof terraces.",
			Cover:       img(8), Images: []models.Image{img(8), pimg(18), pimg(73), pimg(110)}},
	}

	for i := range devs {
		devs[i].Features = []models.Feature{
			{Key: "parking", Label: "Underground parking", Icon: "car"},
			{Key: "energy", Label: "Energy class " + devs[i].EnergyRating, Icon: "bolt"},
			{Key: "elevator", Label: "Lift access", Icon: "elevator"},
			{Key: "green", Label: "Landscaped grounds", Icon: "leaf"},
		}
	}
	return devs
}

// ---------------------------------------------------------------------------
// Articles
// ---------------------------------------------------------------------------

func buildArticles(now time.Time) []models.Article {
	type as struct {
		slug, title, cat, author, role string
		excerpt                        string
		body                           []string
		mins, days, img                int
		featured                       bool
	}
	seeds := []as{
		{"what-energy-class-actually-costs-you", "What an energy class actually costs you each year", "Buying guides",
			"Kadri Tamm", "Senior broker, Kadaka Kinnisvara",
			"Two apartments of the same size can differ by more than €1,400 a year in running costs. Here is how to read an energy certificate before you make an offer.",
			[]string{
				"The energy label on a listing is the single most under-read number in a property advertisement. Buyers scan the price, the area and the number of rooms, then skip past a letter that determines what the property will cost to run for as long as they own it.",
				"The certificate expresses the building's calculated energy demand per square metre per year. A class A apartment in Tallinn typically lands under 100 kWh/m², while a class E building from the 1970s can exceed 250. On a 90 m² apartment, that difference works out at roughly €1,400 a year at current district-heating prices.",
				"What matters more than the letter itself is why the building earned it. A poor rating caused by single glazing is a solvable problem with a known cost. A poor rating caused by uninsulated concrete panel construction is a much larger undertaking, and in a multi-apartment building it is not a decision you make alone.",
				"Before you offer, ask for the housing company's renovation plan alongside the certificate. If a facade insulation programme is scheduled, the cost will land on you as the new owner, and it should be reflected in what you pay today.",
			}, 6, 2, 1, true},

		{"buying-in-spain-as-a-non-resident", "Buying in Spain as a non-resident: the paperwork, in order", "International",
			"Marc Puig", "Agent immobiliari, Mediterrània Propietats",
			"An NIE, a Spanish bank account and a clear reservation contract. The sequence matters more than most buyers expect.",
			[]string{
				"Most delays in a Spanish purchase are administrative rather than financial. Buyers arrive with financing arranged and then lose six weeks to documents that could have been started before they ever viewed a property.",
				"The first requirement is an NIE, the foreigner's identification number. Every party to a Spanish property transaction needs one, and it is required before the deed can be signed. Applications can be made at a Spanish consulate in your home country, which is usually faster than applying once you arrive.",
				"Second is a Spanish bank account. Notaries expect the purchase funds to move through the Spanish banking system, and utilities and community fees will be direct-debited from it afterwards.",
				"Third, read the reservation contract carefully. It typically takes the property off the market for a fixed period against a deposit of a few thousand euros. What varies between agencies is whether that deposit is refundable if a survey turns up a problem. Ask before you sign, not after.",
				"Budget between 10% and 14% above the purchase price for transfer tax, notary fees, registry fees and legal costs. In Catalonia the transfer tax alone runs to 10% on most resale purchases.",
			}, 8, 5, 2, true},

		{"the-rental-yield-question", "Rental yield: the number most landlords calculate wrongly", "Investment",
			"Marten Sepp", "Broker, Kadaka Kinnisvara",
			"Gross yield flatters almost every listing. Net yield, calculated honestly, is what determines whether a rental actually works.",
			[]string{
				"Gross yield is annual rent divided by purchase price. It is easy to calculate, which is why it appears in every advertisement, and it is close to useless as a decision tool.",
				"Net yield subtracts what the property actually costs you to hold: the housing-company charge, insurance, income tax on the rent, maintenance and — the line most first-time landlords omit — vacancy.",
				"A realistic vacancy assumption for a well-positioned city apartment is three to four weeks a year. That alone removes roughly 7% from your gross figure. Maintenance on an older building runs at about 1% of the property value annually once you average across a decade.",
				"Run the numbers on a property advertised at 6% gross and you will typically land somewhere between 3.4% and 4.1% net. That is not a bad return, but it is a very different investment case from the one on the listing.",
			}, 7, 9, 3, false},

		{"viewing-checklist-forty-minutes", "The forty-minute viewing: what to check, in what order", "Buying guides",
			"Annika Hoffmann", "Senior broker, Hauptstadt Immobilien",
			"Most viewings last under half an hour and cover the wrong things. A structured walkthrough surfaces the expensive problems first.",
			[]string{
				"Buyers tend to spend their viewing time on the things that are easiest to change — kitchen finishes, wall colours, the state of the flooring — and almost no time on the things that are expensive or impossible to change.",
				"Start outside. Look at the facade, the roofline and the guttering, and check whether the render is sound at ground level. Water damage announces itself on the outside of a building long before it appears inside.",
				"Inside, open windows and check how they close. Run every tap and flush every toilet. Look under sinks and inside cupboards on external walls, where damp shows first. Check the consumer unit — a modern board with RCD protection tells you the electrics have been touched this century.",
				"Only then look at the rooms as rooms. By that point you know what the property will cost you beyond the asking price, and you can judge whether the layout is worth it.",
				"Ask for the building's maintenance history and the minutes of the last two owners' meetings. In an apartment building, the decisions already taken by your future neighbours are decisions you inherit.",
			}, 9, 12, 4, true},

		{"lisbon-restoration-what-you-inherit", "Restoring a Pombaline building: what you actually inherit", "Renovation",
			"Rui Almeida", "Consultor imobiliário, Tejo Properties",
			"Lisbon's historic centre rewards careful restoration and punishes optimistic budgets. A structural survey is not optional.",
			[]string{
				"The Pombaline buildings that fill Lisbon's Baixa and Chiado were built after the 1755 earthquake using a timber cage — the gaiola — designed to flex rather than collapse. Two and a half centuries later, the condition of that frame is the single variable that determines what a restoration will cost.",
				"Where the timber is sound, restoration is largely a question of services, finishes and thermal upgrade. Where it has been compromised by damp, insect damage or the removal of members during earlier alterations, you are looking at structural work that requires municipal approval and specialist contractors.",
				"Commission a structural survey specifically of the gaiola before exchanging. A general condition report will not tell you what you need to know, and the difference between the two outcomes can be several hundred thousand euros on a mid-sized building.",
				"Factor in time as well as money. Works in protected buildings in the historic centre require approval from the municipal heritage department, and that process runs to months rather than weeks.",
			}, 8, 16, 5, false},

		{"what-a-broker-actually-does", "What a broker actually does for the fee", "Selling guides",
			"Aino Virtanen", "Broker, Pohjola Koti",
			"Sellers increasingly ask what the commission buys. It is a fair question, and the honest answer is more specific than 'marketing'.",
			[]string{
				"A commission of two to four per cent is a substantial sum, and sellers are right to ask what it covers. Photography and a portal listing are the visible part, and they are also the cheapest part.",
				"The work that determines the outcome is pricing. Setting an asking price is not a matter of averaging recent sales; it is a judgement about where your property sits within a band, how much of the market it addresses at each price point, and how long you can afford to wait.",
				"The second piece is qualification. A viewing schedule filled with buyers who cannot finance the purchase costs you weeks, and weeks on the market cost you negotiating position. Checking financing before booking a viewing is unglamorous and it is most of the value.",
				"The third is the negotiation itself, and specifically what happens after an offer is accepted. Surveys turn up problems. Lenders revise valuations. Chains break. Managing that period without the sale collapsing is the part sellers never see and the part they are actually paying for.",
			}, 6, 21, 6, false},

		{"berlin-milieuschutz-explained", "Milieuschutz, explained for buyers", "Regulation",
			"Jonas Weber", "Immobilienmakler, Hauptstadt Immobilien",
			"Large parts of central Berlin sit inside social-preservation areas. If you plan to renovate or convert, the rules change what is possible.",
			[]string{
				"Milieuschutz — formally soziale Erhaltungssatzung — is a designation applied to districts where the city wants to limit displacement of existing residents. Substantial parts of Prenzlauer Berg, Friedrichshain, Kreuzberg and Neukölln fall inside such areas.",
				"For a buyer, the practical effect is that certain works require approval that would be routine elsewhere. Merging two apartments, adding a second bathroom, installing a lift, or upgrading beyond a defined standard can all require permission, and permission is not automatic.",
				"The designation also gives the district a pre-emption right on sales, and imposes a conversion ban that restricts splitting rental buildings into individually owned apartments.",
				"None of this makes a property a poor purchase. It does mean that a renovation budget built on assumptions from outside the area may not survive contact with the district office. Check the designation before you offer, not after.",
			}, 7, 26, 7, false},

		{"first-apartment-deposit-maths", "The deposit maths nobody explains to first-time buyers", "Buying guides",
			"Elena Serra", "Agent immobiliària, Mediterrània Propietats",
			"The deposit is not the barrier most buyers think it is. The costs stacked on top of it usually are.",
			[]string{
				"First-time buyers plan for the deposit and are then caught out by everything that sits alongside it. The deposit is the visible number; the transaction costs are the one that determines whether the purchase is affordable.",
				"Depending on the market, transfer taxes, notary fees, registration and legal costs add between 6% and 15% to the purchase price. In Spain that is typically 10–14%. In Portugal, IMT plus stamp duty and notary costs land around 7–8%. In Estonia the equivalent is closer to 1%.",
				"Lenders assess affordability on the loan, not on your total outlay, which means it is entirely possible to be approved for a mortgage you cannot actually complete on.",
				"Build the full cost stack before you start viewing, and view within it. It is a considerably better experience than discovering the gap after an offer has been accepted.",
			}, 5, 31, 8, false},

		{"heat-pumps-in-older-buildings", "Heat pumps in older buildings: when they work", "Renovation",
			"Mikko Laine", "Broker, Pohjola Koti",
			"An air-source heat pump in a poorly insulated building is an expensive way to be cold. Sequence the work correctly.",
			[]string{
				"Heat pumps are efficient at moving heat, not at generating it. In a well-insulated building they deliver three or four units of heat per unit of electricity. In a leaky one, that ratio collapses and the system runs on its immersion backup through the coldest months — precisely when you need it.",
				"The order of works matters more than the equipment specification. Airtightness and insulation come first, emitters second, heat source third. Reversing that sequence is the most common and most expensive mistake in domestic retrofit.",
				"Emitters are the piece most often overlooked. Heat pumps run at lower flow temperatures than gas or oil boilers, which means existing radiators are frequently undersized. Underfloor heating suits them best; oversized radiators are the usual compromise.",
				"Done in the right order, the result is a building that costs meaningfully less to heat. Done in the wrong order, it is a well-reviewed appliance struggling against a building that was never prepared for it.",
			}, 7, 38, 9, false},

		{"reading-a-floor-plan", "How to read a floor plan properly", "Buying guides",
			"Liis Kask", "Broker, Kadaka Kinnisvara",
			"Floor plans are the most information-dense part of a listing and the part buyers spend the least time on.",
			[]string{
				"A photograph shows you a corner of a room chosen by someone whose interest is in selling it. A floor plan shows you the whole property and cannot flatter it. It is the most honest document in the listing.",
				"Start with circulation. Trace the route from the entrance to each room. Count how many times you pass through one room to reach another — bedrooms accessed through living space are much less usable than the area suggests.",
				"Then look at wall thickness. Thick walls on the plan are usually structural, which tells you which alterations are straightforward and which involve an engineer.",
				"Check window positions against the compass marker. A living room with windows on one elevation only will be noticeably darker than the photographs imply, particularly on a north-facing frontage.",
				"Finally, measure the furniture you already own against the plan. A room that measures well can still fail to take a three-metre sofa and a dining table at the same time.",
			}, 6, 44, 10, false},
	}

	out := make([]models.Article, 0, len(seeds))
	for i, s := range seeds {
		out = append(out, models.Article{
			ID:      fmt.Sprintf("ar-%02d", i+1),
			Slug:    s.slug,
			Title:   s.title,
			Excerpt: s.excerpt,
			Body:    s.body,
			Cover: models.Image{
				URL: fmt.Sprintf("/static/img/articles/a%02d.jpg", s.img),
				Alt: s.title, Width: 1200, Height: 800,
			},
			Category:    s.cat,
			Tags:        []string{s.cat},
			Author:      s.author,
			AuthorRole:  s.role,
			AuthorPhoto: authorPhoto(s.author),
			ReadMinutes: s.mins,
			PublishedAt: now.AddDate(0, 0, -s.days),
			IsFeatured:  s.featured,
		})
	}
	return out
}

func authorPhoto(name string) string {
	for _, b := range brokerSeeds {
		if b.name == name {
			return "/static/img/brokers/" + b.id + ".jpg"
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// Packages, testimonials
// ---------------------------------------------------------------------------

var packages = []models.Package{
	{ID: "pk-basic", Name: "Basic", Tagline: "For a single private listing",
		Price: models.Money{Amount: 19, Currency: "EUR"}, DurationDays: 30, PhotoLimit: 10, BumpCount: 0, IsEnabled: true,
		Features: []string{"30 days online", "Up to 10 photos", "Appears in search and on the map", "Email enquiries from buyers"}},
	{ID: "pk-standard", Name: "Standard", Tagline: "Most private sellers choose this",
		Price: models.Money{Amount: 39, Currency: "EUR"}, DurationDays: 60, PhotoLimit: 25, BumpCount: 2, IsPopular: true, IsEnabled: true,
		Features: []string{"60 days online", "Up to 25 photos", "Floor plan upload", "2 bumps to the top of results", "Highlighted in the results list", "Basic performance statistics"}},
	{ID: "pk-premium", Name: "Premium", Tagline: "Maximum exposure for one property",
		Price: models.Money{Amount: 89, Currency: "EUR"}, DurationDays: 90, PhotoLimit: 50, BumpCount: 6, IsPremium: true, IsEnabled: true,
		Features: []string{"90 days online", "Up to 50 photos and video", "Featured placement on the homepage", "Gold Featured badge in every result", "6 bumps to the top of results", "Priority position on the map", "Full performance statistics", "Social media promotion"}},
	{ID: "pk-agency", Name: "Agency", Tagline: "For brokers and agencies",
		Price: models.Money{Amount: 249, Currency: "EUR"}, DurationDays: 30, PhotoLimit: 50, BumpCount: 20, IsEnabled: true,
		Features: []string{"Up to 40 active listings", "Agency profile page", "Broker profiles for your team", "Promoted broker placement in one country", "Bulk upload and XML feed", "Dedicated account manager", "Monthly performance reporting"}},
}

var testimonials = []models.Testimonial{
	{Name: "Helena Mägi", Role: "Bought an apartment", City: "Tallinn", Stars: 5,
		Quote: "I searched from Stockholm for four months. Being able to filter by energy class and see everything on the map meant I only flew over for three viewings instead of a dozen."},
	{Name: "Tobias Lang", Role: "Sold a family home", City: "Berlin", Stars: 5,
		Quote: "The listing wizard took about twenty minutes, including the photographs. We had eleven enquiries in the first week and accepted an offer on day nine."},
	{Name: "Clara Ferrer", Role: "Rented an apartment", City: "Barcelona", Stars: 5,
		Quote: "Saved searches did the work for me. I got an alert the morning a flat in Gràcia went live, viewed it that afternoon and signed the contract two days later."},
}

// ---------------------------------------------------------------------------
// Account activity
// ---------------------------------------------------------------------------

func buildSavedSearches(now time.Time) []models.SavedSearch {
	return []models.SavedSearch{
		{ID: "ss-01", Name: "Two-bed apartments in Kalamaja", Query: "deal=sale&country=EE&city=Tallinn&type=apartment&bedrooms=2&price_max=400000",
			Summary: "For sale · Tallinn · Apartment · 2+ bedrooms · up to €400 000", Deal: models.DealSale,
			AlertsOn: true, Frequency: "instant", NewMatches: 3, ResultCount: 18, CreatedAt: now.AddDate(0, 0, -22)},
		{ID: "ss-02", Name: "Berlin Altbau under €900k", Query: "deal=sale&country=DE&city=Berlin&type=apartment&price_max=900000&condition=renovated",
			Summary: "For sale · Berlin · Apartment · Renovated · up to €900 000", Deal: models.DealSale,
			AlertsOn: true, Frequency: "daily", NewMatches: 0, ResultCount: 26, CreatedAt: now.AddDate(0, 0, -40)},
		{ID: "ss-03", Name: "Rentals near the beach, Barcelona", Query: "deal=rent&country=ES&city=Barcelona&price_max=2200&furnished=1",
			Summary: "For rent · Barcelona · Furnished · up to €2 200/month", Deal: models.DealRent,
			AlertsOn: false, Frequency: "weekly", NewMatches: 7, ResultCount: 31, CreatedAt: now.AddDate(0, 0, -9)},
		{ID: "ss-04", Name: "Waterfront, Helsinki", Query: "deal=sale&country=FI&city=Helsinki&type=apartment&price_min=450000",
			Summary: "For sale · Helsinki · Apartment · from €450 000", Deal: models.DealSale,
			AlertsOn: true, Frequency: "daily", NewMatches: 1, ResultCount: 12, CreatedAt: now.AddDate(0, 0, -4)},
	}
}

func buildNotifications(now time.Time) []models.Notification {
	return []models.Notification{
		{ID: "nt-01", Type: "match", Icon: "search", Title: "3 new matches for “Two-bed apartments in Kalamaja”",
			Body: "Three listings published in the last 24 hours match this saved search.", Link: "/saved-searches",
			IsRead: false, CreatedAt: now.Add(-42 * time.Minute)},
		{ID: "nt-02", Type: "message", Icon: "mail", Title: "Kadri Tamm replied to your enquiry",
			Body: "Regarding “Restored merchant's apartment in Tallinn Old Town” — viewing slots are open on Thursday afternoon.", Link: "/notifications",
			IsRead: false, CreatedAt: now.Add(-5 * time.Hour)},
		{ID: "nt-03", Type: "listing", Icon: "home", Title: "Your listing is now live",
			Body: "“Bright two-room apartment near Telliskivi Creative City” passed review and is visible in search results.", Link: "/my-listings",
			IsRead: false, CreatedAt: now.Add(-26 * time.Hour)},
		{ID: "nt-04", Type: "payment", Icon: "card", Title: "Payment received — Standard package",
			Body: "Invoice PRV-2026-0418 for €39.00 has been paid. Your listing runs until 9 October 2026.", Link: "/billing",
			IsRead: true, CreatedAt: now.AddDate(0, 0, -3)},
		{ID: "nt-05", Type: "system", Icon: "info", Title: "A property you saved has reduced in price",
			Body: "“Semi-detached house in Berlin Zehlendorf” is now €1 120 000, down from €1 165 000.", Link: "/favourites",
			IsRead: true, CreatedAt: now.AddDate(0, 0, -6)},
		{ID: "nt-06", Type: "match", Icon: "search", Title: "7 new matches for “Rentals near the beach, Barcelona”",
			Body: "Seven listings matching this search were published this week.", Link: "/saved-searches",
			IsRead: true, CreatedAt: now.AddDate(0, 0, -8)},
	}
}

func buildDrafts(now time.Time) []models.Draft {
	return []models.Draft{
		{ID: "df-01", Title: "Apartment on Vabaduse väljak", Deal: models.DealSale, Type: models.TypeApartment,
			City: "Tallinn", Step: 9, TotalSteps: 14, Completion: 64,
			CoverImage: "/static/img/properties/p012.jpg", UpdatedAt: now.Add(-3 * time.Hour)},
		{ID: "df-02", Title: "Untitled listing", Deal: models.DealRent, Type: models.TypeApartment,
			City: "Tartu", Step: 3, TotalSteps: 14, Completion: 21,
			CoverImage: "", UpdatedAt: now.AddDate(0, 0, -2)},
		{ID: "df-03", Title: "Summer cottage, Pärnu county", Deal: models.DealSale, Type: models.TypeHouse,
			City: "Pärnu", Step: 12, TotalSteps: 14, Completion: 86,
			CoverImage: "/static/img/properties/p019.jpg", UpdatedAt: now.AddDate(0, 0, -5)},
	}
}

func buildPayments(now time.Time) []models.Payment {
	return []models.Payment{
		{ID: "pm-01", InvoiceNo: "PRV-2026-0418", PackageName: "Standard", PropertyTitle: "Bright two-room apartment near Telliskivi Creative City",
			Amount: models.Money{Amount: 39, Currency: "EUR"}, Method: "stripe", Status: models.PaymentPaid, CreatedAt: now.AddDate(0, 0, -3)},
		{ID: "pm-02", InvoiceNo: "PRV-2026-0377", PackageName: "Premium", PropertyTitle: "Sea-view penthouse in Kalamaja with private roof terrace",
			Amount: models.Money{Amount: 89, Currency: "EUR"}, Method: "paypal", Status: models.PaymentPaid, CreatedAt: now.AddDate(0, 0, -31)},
		{ID: "pm-03", InvoiceNo: "PRV-2026-0341", PackageName: "Standard", PropertyTitle: "Studio apartment in Tallinn city centre",
			Amount: models.Money{Amount: 39, Currency: "EUR"}, Method: "paysera", Status: models.PaymentFailed, CreatedAt: now.AddDate(0, 0, -48)},
		{ID: "pm-04", InvoiceNo: "PRV-2026-0298", PackageName: "Basic", PropertyTitle: "Building plot with sea access in Viimsi",
			Amount: models.Money{Amount: 19, Currency: "EUR"}, Method: "stripe", Status: models.PaymentRefunded, CreatedAt: now.AddDate(0, 0, -66)},
		{ID: "pm-05", InvoiceNo: "PRV-2026-0255", PackageName: "Standard", PropertyTitle: "Family house with sauna in Nõmme pine district",
			Amount: models.Money{Amount: 39, Currency: "EUR"}, Method: "stripe", Status: models.PaymentPaid, CreatedAt: now.AddDate(0, 0, -84)},
	}
}
