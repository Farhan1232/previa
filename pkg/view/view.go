// Package view renders Go html/template sets.
//
// Templates come from an fs.FS — normally the embedded copy in package web, so
// the binary is self-contained and behaves the same whether it runs as a
// long-lived server or as a serverless function. Passing os.DirFS instead
// enables live reloading during development.
package view

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"html/template"
	"net/http"
	"path"
	"sync"
)

// Engine compiles and renders template sets.
type Engine struct {
	fsys  fs.FS
	dev   bool
	funcs template.FuncMap

	mu    sync.RWMutex
	cache map[string]*template.Template
}

// New builds an engine over fsys, which must contain layouts/, components/,
// partials/ and pages/ directories.
//
// When dev is true every render recompiles, so template edits appear without a
// restart. When false the whole set is compiled once up front, turning a broken
// template into a startup failure rather than a visitor-facing 500.
func New(fsys fs.FS, dev bool) (*Engine, error) {
	e := &Engine{
		fsys:  fsys,
		dev:   dev,
		funcs: Funcs(),
		cache: map[string]*template.Template{},
	}
	if !dev {
		if err := e.warm(); err != nil {
			return nil, err
		}
	}
	return e, nil
}

// shared returns the layout, component and partial files every page includes.
func (e *Engine) shared() ([]string, error) {
	var out []string
	for _, sub := range []string{"layouts", "components", "partials"} {
		matches, err := fs.Glob(e.fsys, path.Join(sub, "*.html"))
		if err != nil {
			return nil, err
		}
		out = append(out, matches...)
	}
	return out, nil
}

// warm compiles every page so template errors surface at boot.
func (e *Engine) warm() error {
	pages, err := e.pageFiles()
	if err != nil {
		return err
	}
	for _, name := range pages {
		if _, err := e.compile(name); err != nil {
			return fmt.Errorf("template %s: %w", name, err)
		}
	}
	return nil
}

// pageFiles lists page template names relative to pages/, without extension.
func (e *Engine) pageFiles() ([]string, error) {
	var out []string
	err := fs.WalkDir(e.fsys, "pages", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path.Ext(p) != ".html" {
			return nil
		}
		rel := p[len("pages/") : len(p)-len(".html")]
		out = append(out, rel)
		return nil
	})
	return out, err
}

// compile builds one page's template set and caches it.
func (e *Engine) compile(page string) (*template.Template, error) {
	shared, err := e.shared()
	if err != nil {
		return nil, err
	}
	pagePath := path.Join("pages", page+".html")
	files := append([]string{pagePath}, shared...)

	t, err := template.New(path.Base(pagePath)).Funcs(e.funcs).ParseFS(e.fsys, files...)
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	e.cache[page] = t
	e.mu.Unlock()
	return t, nil
}

// lookup returns a cached set, recompiling in dev mode.
func (e *Engine) lookup(page string) (*template.Template, error) {
	if e.dev {
		return e.compile(page)
	}
	e.mu.RLock()
	t, ok := e.cache[page]
	e.mu.RUnlock()
	if ok {
		return t, nil
	}
	return e.compile(page)
}

// Render writes a full page using the base layout.
func (e *Engine) Render(w http.ResponseWriter, status int, page string, data any) {
	e.render(w, status, page, "base", data)
}

// RenderPartial writes a single named template — used for HTMX fragment
// responses, which must not carry the layout.
func (e *Engine) RenderPartial(w http.ResponseWriter, status int, page, block string, data any) {
	e.render(w, status, page, block, data)
}

func (e *Engine) render(w http.ResponseWriter, status int, page, block string, data any) {
	t, err := e.lookup(page)
	if err != nil {
		e.fail(w, fmt.Errorf("compile %s: %w", page, err))
		return
	}

	// Render to a buffer first so a template error mid-write cannot emit a
	// half-finished page on top of an already-sent 200.
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, block, data); err != nil {
		e.fail(w, fmt.Errorf("execute %s/%s: %w", page, block, err))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = io.Copy(w, &buf)
}

func (e *Engine) fail(w http.ResponseWriter, err error) {
	if e.dev {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("template error:", err)
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}
