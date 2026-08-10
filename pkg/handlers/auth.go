package handlers

import (
	"net/http"
	"strings"
)

// AuthData drives the authentication screens. All authentication in this
// milestone is mocked: no credential is verified, stored or transmitted.
type AuthData struct {
	Mode    string // "login" | "register" | "forgot"
	Errors  map[string]string
	Values  map[string]string
	Success bool
	Message string
}

// Login renders the sign-in page.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	pd := h.base(r, "", "Log in — Previa", "Access your saved properties, searches and listings.")
	pd.Meta.NoIndex = true
	pd.Data = AuthData{Mode: "login", Errors: map[string]string{}, Values: map[string]string{}}
	h.View.Render(w, http.StatusOK, "auth/login", pd)
}

// LoginSubmit validates the form and mocks a successful sign-in.
func (h *Handler) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	errs := map[string]string{}
	if email == "" {
		errs["email"] = "Enter your email address."
	} else if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		errs["email"] = "Enter a valid email address."
	}
	if password == "" {
		errs["password"] = "Enter your password."
	} else if len(password) < 8 {
		errs["password"] = "Your password is at least 8 characters."
	}

	if len(errs) > 0 {
		pd := h.base(r, "", "Log in — Previa", "Access your saved properties, searches and listings.")
		pd.Meta.NoIndex = true
		pd.Data = AuthData{Mode: "login", Errors: errs, Values: map[string]string{"email": email}}
		h.View.Render(w, http.StatusUnprocessableEntity, "auth/login", pd)
		return
	}

	// Mock session cookie. It carries no identity and grants nothing.
	http.SetCookie(w, &http.Cookie{
		Name: "previa_demo_session", Value: "1", Path: "/",
		MaxAge: 60 * 60 * 24, HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Register renders the sign-up page.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	pd := h.base(r, "", "Create an account — Previa",
		"Create a free Previa account to save properties, store searches and publish listings.")
	pd.Meta.NoIndex = true
	pd.Data = AuthData{Mode: "register", Errors: map[string]string{}, Values: map[string]string{}}
	h.View.Render(w, http.StatusOK, "auth/register", pd)
}

// RegisterSubmit validates the form and mocks account creation.
func (h *Handler) RegisterSubmit(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	errs := map[string]string{}
	if name == "" {
		errs["name"] = "Enter your name."
	}
	if email == "" {
		errs["email"] = "Enter your email address."
	} else if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		errs["email"] = "Enter a valid email address."
	}
	switch {
	case password == "":
		errs["password"] = "Choose a password."
	case len(password) < 8:
		errs["password"] = "Use at least 8 characters."
	case strings.ToLower(password) == password:
		errs["password"] = "Include at least one capital letter."
	}

	confirm := r.FormValue("password_confirm")
	if errs["password"] == "" {
		if confirm == "" {
			errs["password_confirm"] = "Re-enter your password."
		} else if confirm != password {
			errs["password_confirm"] = "The two passwords don't match."
		}
	}

	if r.FormValue("terms") == "" {
		errs["terms"] = "You need to accept the terms to continue."
	}

	if len(errs) > 0 {
		pd := h.base(r, "", "Create an account — Previa", "Create a free Previa account.")
		pd.Meta.NoIndex = true
		pd.Data = AuthData{Mode: "register", Errors: errs,
			Values: map[string]string{"name": name, "email": email}}
		h.View.Render(w, http.StatusUnprocessableEntity, "auth/register", pd)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// ForgotPassword renders the reset-request page.
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	pd := h.base(r, "", "Reset your password — Previa", "Request a password reset link for your Previa account.")
	pd.Meta.NoIndex = true
	pd.Data = AuthData{Mode: "forgot", Errors: map[string]string{}, Values: map[string]string{}}
	h.View.Render(w, http.StatusOK, "auth/forgot-password", pd)
}

// ForgotPasswordSubmit mocks sending a reset link. No email is sent.
func (h *Handler) ForgotPasswordSubmit(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	email := strings.TrimSpace(r.FormValue("email"))

	pd := h.base(r, "", "Reset your password — Previa", "Request a password reset link.")
	pd.Meta.NoIndex = true

	if email == "" || !strings.Contains(email, "@") {
		pd.Data = AuthData{Mode: "forgot",
			Errors: map[string]string{"email": "Enter a valid email address."},
			Values: map[string]string{"email": email}}
		h.View.Render(w, http.StatusUnprocessableEntity, "auth/forgot-password", pd)
		return
	}

	// The wording deliberately does not confirm whether the address is
	// registered — that would let anyone enumerate accounts.
	pd.Data = AuthData{Mode: "forgot", Success: true, Errors: map[string]string{},
		Values:  map[string]string{"email": email},
		Message: "If an account exists for " + email + ", a reset link is on its way."}
	h.View.Render(w, http.StatusOK, "auth/forgot-password", pd)
}

// Logout clears the demo cookie.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "previa_demo_session", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
