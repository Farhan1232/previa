package handlers

import (
	"html/template"
	"net/http"

	"previa/pkg/models"
)

// PropertyDetailData is the detail-page payload.
type PropertyDetailData struct {
	Property    models.Property
	Broker      models.Broker
	HasBroker   bool
	Agency      models.Agency
	HasAgency   bool
	Similar     []models.Property
	Favourite   bool
	Development models.Development
	HasDev      bool
	DetailMap   template.JS
}

// PropertyDetail renders one listing.
func (h *Handler) PropertyDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	p, ok := h.Store.Properties.BySlug(ctx, r.PathValue("slug"))
	if !ok {
		h.NotFound(w, r)
		return
	}

	pd := h.base(r, "search", p.Title+" — Previa", truncateText(p.Description, 155))
	pd.Meta.OGType = "article"
	pd.NeedsMap = true
	if img := p.PrimaryImage(); img.URL != "" {
		pd.Meta.OGImage = img.URL
	}

	dd := PropertyDetailData{
		Property:  p,
		Similar:   h.Store.Properties.Similar(ctx, p, 4),
		Favourite: h.Store.Account.IsFavourite(ctx, p.ID),
		// A single-marker map centred on the listing itself.
		DetailMap: buildMapConfig([]models.Property{p}, nil, pd.Country, h.Cfg.MapsKey),
	}
	if p.BrokerID != "" {
		dd.Broker, dd.HasBroker = h.Store.Brokers.ByID(ctx, p.BrokerID)
	}
	if p.AgencyID != "" {
		for _, a := range h.Store.Brokers.Agencies(ctx) {
			if a.ID == p.AgencyID {
				dd.Agency, dd.HasAgency = a, true
			}
		}
	}
	if p.DevelopmentID != "" {
		for _, d := range h.Store.Content.Developments(ctx, "") {
			if d.ID == p.DevelopmentID {
				dd.Development, dd.HasDev = d, true
			}
		}
	}

	pd.Data = dd
	h.View.Render(w, http.StatusOK, "property-detail", pd)
}

// DevelopmentDetailData is the development-page payload.
type DevelopmentDetailData struct {
	Development models.Development
	Units       []models.Property
	Similar     []models.Development
	DevMap      template.JS
}

// DevelopmentDetail renders one development.
func (h *Handler) DevelopmentDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	d, ok := h.Store.Content.DevelopmentBySlug(ctx, r.PathValue("slug"))
	if !ok {
		h.NotFound(w, r)
		return
	}

	pd := h.base(r, "developments", d.Name+" — New development in "+d.City+" — Previa",
		truncateText(d.Description, 155))
	pd.Meta.OGImage = d.Cover.URL
	pd.NeedsMap = true

	var similar []models.Development
	for _, o := range h.Store.Content.Developments(ctx, d.CountryCode) {
		if o.ID != d.ID {
			similar = append(similar, o)
		}
	}

	// The development itself is the map's only marker.
	marker := models.Property{
		ID: d.ID, Title: d.Name, Coords: d.Coords, City: d.City,
		Price: d.PriceFrom, Slug: d.Slug, Images: d.Images,
	}

	pd.Data = DevelopmentDetailData{
		Development: d,
		Units:       h.Store.Properties.ByDevelopment(ctx, d.ID),
		Similar:     take(similar, 3),
		DevMap:      buildMapConfig([]models.Property{marker}, nil, pd.Country, h.Cfg.MapsKey),
	}
	h.View.Render(w, http.StatusOK, "development-detail", pd)
}

// BrokerProfileData is the broker-page payload.
type BrokerProfileData struct {
	Broker    models.Broker
	Agency    models.Agency
	HasAgency bool
	Listings  []models.Property
}

// BrokerProfile renders one broker.
func (h *Handler) BrokerProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	b, ok := h.Store.Brokers.BySlug(ctx, r.PathValue("slug"))
	if !ok {
		h.NotFound(w, r)
		return
	}

	pd := h.base(r, "brokers", b.Name+" — "+b.Title+" — Previa", truncateText(b.Bio, 155))

	bd := BrokerProfileData{
		Broker:   b,
		Listings: h.Store.Properties.ByBroker(ctx, b.ID, 12),
	}
	for _, a := range h.Store.Brokers.Agencies(ctx) {
		if a.ID == b.AgencyID {
			bd.Agency, bd.HasAgency = a, true
		}
	}

	pd.Data = bd
	h.View.Render(w, http.StatusOK, "broker-profile", pd)
}

// AgencyProfileData is the agency-page payload.
type AgencyProfileData struct {
	Agency   models.Agency
	Brokers  []models.Broker
	Listings []models.Property
}

// AgencyProfile renders one agency.
func (h *Handler) AgencyProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	a, ok := h.Store.Brokers.AgencyBySlug(ctx, r.PathValue("slug"))
	if !ok {
		h.NotFound(w, r)
		return
	}

	pd := h.base(r, "brokers", a.Name+" — Real-estate agency in "+a.City+" — Previa",
		truncateText(a.Description, 155))

	brokers := h.Store.Brokers.BrokersByAgency(ctx, a.ID)
	var listings []models.Property
	for _, b := range brokers {
		listings = append(listings, h.Store.Properties.ByBroker(ctx, b.ID, 4)...)
	}

	pd.Data = AgencyProfileData{Agency: a, Brokers: brokers, Listings: take(listings, 8)}
	h.View.Render(w, http.StatusOK, "agency-profile", pd)
}

// ArticleDetailData is the article-page payload.
type ArticleDetailData struct {
	Article models.Article
	Related []models.Article
}

// ArticleDetail renders one article.
func (h *Handler) ArticleDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	a, ok := h.Store.Content.ArticleBySlug(ctx, r.PathValue("slug"))
	if !ok {
		h.NotFound(w, r)
		return
	}

	pd := h.base(r, "articles", a.Title+" — Previa", a.Excerpt)
	pd.Meta.OGType = "article"
	pd.Meta.OGImage = a.Cover.URL

	pd.Data = ArticleDetailData{Article: a, Related: h.Store.Content.RelatedArticles(ctx, a, 3)}
	h.View.Render(w, http.StatusOK, "article-detail", pd)
}

func truncateText(s string, n int) string {
	if len(s) <= n {
		return s
	}
	cut := s[:n]
	for i := len(cut) - 1; i > n/2; i-- {
		if cut[i] == ' ' {
			return cut[:i] + "…"
		}
	}
	return cut + "…"
}
