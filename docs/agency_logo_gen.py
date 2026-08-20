"""Generate the seeded agencies' company logos.

The client asked for a company logo alongside the profile picture: "the brokers
might want to use their company logo as well … make 'your company logo' where
the brokers can upload their compay logo … in the single ad page it will be
displayed under 'view listings of this broker'".

Uploads are mocked in this build, so the eight seeded agencies need marks to
stand in for what a real broker would upload — otherwise the block that shows
them can never be seen, and neither the client nor the backend team can tell
whether it works.

Each mark is a monogram tile beside the agency's name, in that agency's own
colour: the shape a small business actually uses, and legible at the 160 × 48
the seller box gives it. SVG rather than raster so one file serves the seller
box, the broker profile and the directory without a variant per size, and so
the file stays under a kilobyte.

    python3 docs/agency_logo_gen.py        writes public/static/img/agencies/

Colours are literal. A logo is somebody else's brand; it does not follow the
Previa theme, and it has to hold on the white card it is drawn on in either
theme, which is why every pairing below is dark-on-light.
"""
import pathlib

OUT = pathlib.Path("public/static/img/agencies")

# (slug, display name, monogram, tile colour, word colour)
AGENCIES = [
    ("kadaka-kinnisvara", "Kadaka", "KK", "#1F5E48", "#123A2C"),
    ("hauptstadt-immobilien", "Hauptstadt", "HI", "#1B3A63", "#14263F"),
    ("mediterrania-propietats", "Mediterrània", "MP", "#C0532B", "#7A3319"),
    ("pohjola-koti", "Pohjola Koti", "PK", "#1B5E86", "#123F5A"),
    ("tejo-properties", "Tejo", "TP", "#0E7C6B", "#0A5045"),
    ("grachten-makelaars", "Grachten", "GM", "#8A3A2E", "#5C261E"),
    ("ringhaus-wien", "Ringhaus", "RW", "#6B4E9E", "#453268"),
    ("vltava-reality", "Vltava", "VR", "#2A5C8A", "#1B3C5A"),
]

# The canvas the seller box and the broker profile both reserve for a logo.
W, H = 200, 56
TILE = 40


def esc(s):
    return s.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")


def logo(name, mono, tile, word):
    """A monogram tile and a wordmark, vertically centred on the canvas.

    The name is drawn at a font size that keeps the longest of them inside the
    canvas: textLength with lengthAdjust="spacingAndGlyphs" makes that a
    guarantee rather than an estimate, since the font a viewer has is not
    something an SVG can rely on.
    """
    x = TILE + 14
    avail = W - x - 8
    return f"""<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {W} {H}" width="{W}" height="{H}"
     role="img" aria-label="{esc(name)}">
  <rect x="0" y="{(H - TILE) // 2}" width="{TILE}" height="{TILE}" rx="10" fill="{tile}"/>
  <text x="{TILE / 2}" y="{H / 2}" text-anchor="middle" dominant-baseline="central"
        font-family="Manrope, ui-sans-serif, system-ui, sans-serif" font-size="16"
        font-weight="800" fill="#fff" letter-spacing="0.5">{mono}</text>
  <text x="{x}" y="{H / 2}" dominant-baseline="central"
        font-family="Manrope, ui-sans-serif, system-ui, sans-serif" font-size="17"
        font-weight="700" fill="{word}"
        textLength="{min(avail, len(name) * 10)}" lengthAdjust="spacingAndGlyphs">{esc(name)}</text>
</svg>
"""


if __name__ == "__main__":
    OUT.mkdir(parents=True, exist_ok=True)
    for slug, name, mono, tile, word in AGENCIES:
        path = OUT / f"{slug}.svg"
        path.write_text(logo(name, mono, tile, word))
        print(path)
