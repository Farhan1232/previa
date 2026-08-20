"""Generate the Previa globe mark.

Third pass, to the client's 17 August notes:

  · "the green earth is wrong shape. Take it from the carwelt logo, there are
    the continents exactly the right shape." The landmasses were four hand-drawn
    blobs. They are now real coastlines — coarse lon/lat outlines of Europe,
    Africa, Asia, the Americas and the larger islands — put through a genuine
    orthographic projection, so the face of the sphere is the face of the Earth
    seen from over the Previa markets rather than a decorative suggestion of one
  · "they have 2 borderlines, remove the borderlines. There is only the
    background color on the circle." The cream collar, the navy hairline and the
    warm sheen arc around each property disc are gone. A disc is now one flat
    colour
  · "this background color make blue red and green" — all three, one per disc.
    Which disc takes which is set in DISC_FILLS and chosen for contrast against
    what sits under it; see the note there
  · "The 3 houses are on the right place, these stay like they are." Placement,
    geometry and colours of the houses are untouched.

Geometry is computed here rather than hand-placed so the website mark and the
favicon are the same drawing. Two files carry the result and must stay in step:

    web/templates/components/logo.html   ids suffixed per size variant
    public/static/img/favicon.svg        ids suffixed "f", inside the tile

    python3 docs/logo_gen.py <id-suffix>     prints a complete standalone SVG

Placement lives in GX/GY/R and DISCS. Two numbers are load-bearing:
  · a disc radius near 0.32 × R is the balance point. Larger swallows the
    planet; smaller and the house inside stops reading at 34px
  · a disc distance of R + overhang − r puts a chosen amount of the disc past
    the rim. The first two are set to about a third out, which breaks the
    outline without costing the circle its read as a ball

The viewpoint lives in VIEW_LON/VIEW_LAT. It is over north-east Africa at
32°E, 17°N, the face that puts Europe — every market Previa operates in — high
and unobstructed, with Africa filling the middle, Arabia and India running off
to the right and the Atlantic breaking the western limb.
"""
import math

# --- Globe -----------------------------------------------------------------
GX, GY, R = 32.0, 32.4, 22.0

# The sub-observer point: the coordinate at the exact centre of the disc.
VIEW_LON, VIEW_LAT = 32.0, 17.0

# --- Property discs ---------------------------------------------------------
# (angle in degrees measured from the globe's centre, distance, radius)
# Distance is chosen so the first two straddle the rim: centre inside, roughly
# a third of the disc hanging past it. The third sits wholly within the globe.
DISCS = [
    (-135.0, 18.0, 7.0),   # upper left  — overhangs
    (-27.0, 18.4, 6.4),    # right       — overhangs
    (98.0, 12.0, 6.6),     # lower       — inside
]
# One flat colour per disc, named by the client on 18 August: "the house on the
# green background, this background make #7CFC00, the red background make
# #FF0000 and the blue make #FF00FF."
#
# Which disc takes which is unchanged — green sits on the sea at the upper left,
# red over the Arabian Sea, blue on Africa — because the client identified each
# disc by the colour already on it.
DISC_FILLS = ["#7CFC00", "#FF0000", "#FF00FF"]
ROOFS = ["#C25A3C", "#1F6F66", "#2A5C8A"]


def fmt(v):
    s = f"{v:.1f}".rstrip("0").rstrip(".")
    return s if s != "-0" else "0"


def disc_centres():
    out = []
    for ang, dist, r in DISCS:
        a = math.radians(ang)
        out.append((GX + dist * math.cos(a), GY + dist * math.sin(a), r))
    return out


def house(cx, cy, r):
    """The house silhouette, its roof, and its door, sized to the disc."""
    apex = (cx, cy - 0.80 * r)
    eave = 0.64 * r
    eave_y = cy - 0.10 * r
    wall = 0.455 * r
    base = cy + 0.70 * r
    body = (
        f"M{fmt(cx - wall)} {fmt(base)} L{fmt(cx - wall)} {fmt(eave_y)} "
        f"L{fmt(cx - eave)} {fmt(eave_y)} L{fmt(apex[0])} {fmt(apex[1])} "
        f"L{fmt(cx + eave)} {fmt(eave_y)} L{fmt(cx + wall)} {fmt(eave_y)} "
        f"L{fmt(cx + wall)} {fmt(base)} Z"
    )
    roof = (
        f"M{fmt(cx - eave)} {fmt(eave_y)} L{fmt(apex[0])} {fmt(apex[1])} "
        f"L{fmt(cx + eave)} {fmt(eave_y)} Z"
    )
    door = (
        f"M{fmt(cx - 0.16 * r)} {fmt(base)} L{fmt(cx - 0.16 * r)} {fmt(cy + 0.24 * r)} "
        f"L{fmt(cx + 0.16 * r)} {fmt(cy + 0.24 * r)} L{fmt(cx + 0.16 * r)} {fmt(base)} Z"
    )
    return body, roof, door


# --- Coastlines -------------------------------------------------------------
#
# Longitude/latitude rings, one per landmass, traced from a world outline at
# roughly 300km resolution. Coarse on purpose: the mark is drawn at 34px in the
# header, where anything finer turns to noise. What matters — and what the
# client asked for — is that the silhouettes are the real ones, so the boot of
# Italy, the horn of Africa, the Iberian block and the Indian triangle are all
# where a reader expects them.
COASTS = {
    # Europe: Iberia round to Scandinavia, with the Mediterranean coast, the
    # Balkans, the Black Sea and the Baltic drawn in.
    "eurasia": [
        (-9.5, 43.8), (-8.9, 41.9), (-9.5, 38.7), (-8.9, 37.0), (-6.0, 36.0),
        (-2.2, 36.7), (0.7, 40.8), (3.3, 41.9), (3.0, 43.5), (5.0, 43.3),
        (8.9, 44.4), (10.3, 43.0), (12.4, 41.3), (15.6, 40.0), (18.5, 40.1),
        (16.2, 41.9), (13.5, 45.7), (15.3, 44.2), (18.7, 42.9), (19.4, 41.4),
        (23.0, 40.0), (23.7, 37.9), (26.5, 39.0), (28.9, 41.0), (35.0, 42.0),
        (41.5, 41.5), (39.0, 43.5), (37.5, 45.2), (33.5, 46.0), (31.0, 46.6),
        (28.8, 45.2), (28.0, 43.4), (26.0, 41.0), (29.0, 40.5), (36.0, 36.2),
        (35.6, 34.5), (34.3, 31.3), (32.5, 31.1), (33.9, 35.0), (35.5, 36.6),
        (36.5, 33.0), (34.9, 29.5), (38.5, 24.0), (43.0, 12.8), (45.0, 12.0),
        (48.5, 14.0), (54.5, 17.0), (58.6, 20.6), (59.8, 22.5), (56.4, 26.5),
        (52.0, 27.0), (48.5, 30.0), (49.5, 29.0), (56.0, 25.5), (61.5, 25.2),
        (66.5, 25.4), (68.5, 23.8), (72.6, 21.4), (72.9, 18.9), (73.5, 15.9),
        (76.0, 10.0), (77.6, 8.1), (80.3, 13.1), (82.0, 16.8), (86.5, 20.7),
        (89.0, 21.9), (91.8, 22.0), (94.2, 18.0), (97.6, 16.5), (98.5, 10.5),
        (100.3, 5.5), (103.5, 1.3), (104.5, 8.6), (106.5, 10.5), (109.5, 15.0),
        (108.0, 18.5), (110.5, 21.0), (114.0, 22.5), (117.5, 23.6), (120.0, 27.0),
        (121.8, 30.9), (121.0, 34.5), (119.0, 37.5), (122.0, 39.5), (125.5, 39.8),
        (126.5, 37.5), (129.0, 35.2), (129.5, 38.5), (128.5, 41.5), (131.0, 43.5),
        (135.0, 48.5), (140.5, 53.0), (142.5, 59.0), (146.0, 59.5), (152.0, 59.0),
        (155.0, 56.5), (162.0, 58.0), (163.5, 61.0), (170.0, 63.0), (172.0, 66.0),
        (179.5, 66.5), (176.0, 69.0), (168.0, 69.8), (160.0, 70.0), (152.0, 72.0),
        (142.0, 73.5), (132.0, 74.0), (124.0, 73.0), (112.0, 76.5), (105.0, 77.5),
        (95.0, 76.5), (88.0, 75.5), (78.0, 73.0), (70.0, 73.0), (60.0, 70.5),
        (55.0, 68.5), (52.0, 69.5), (44.0, 68.0), (40.0, 66.5), (36.0, 65.0),
        (33.0, 66.5), (30.0, 65.0), (31.5, 62.5), (28.5, 60.5), (24.0, 60.0),
        (21.5, 63.0), (24.0, 65.8), (21.0, 65.9), (17.5, 62.5), (18.5, 60.0),
        (16.5, 56.5), (12.5, 56.0), (10.5, 57.7), (8.1, 56.5), (8.5, 54.5),
        (4.8, 53.0), (3.2, 51.4), (0.0, 49.5), (-1.5, 49.8), (-4.8, 48.5),
        (-1.5, 46.2), (-1.2, 44.0), (-1.8, 43.4), (-4.5, 43.6), (-9.5, 43.8),
    ],
    # Scandinavia's western coast is part of the ring above; Norway's fjord
    # edge is traced here as its own line so the peninsula keeps its shape.
    "norway": [
        (4.9, 58.1), (5.2, 60.4), (7.5, 63.0), (11.0, 64.0), (14.0, 66.5),
        (16.0, 68.3), (20.5, 70.0), (25.0, 71.1), (28.5, 71.0), (31.0, 70.0),
        (30.0, 65.0), (24.0, 65.8), (21.5, 63.0), (17.5, 62.5), (11.0, 59.0),
        (8.0, 58.0), (4.9, 58.1),
    ],
    "britain": [
        (-5.0, 50.1), (-3.0, 51.5), (-4.5, 53.3), (-3.0, 54.9), (-5.8, 55.5),
        (-5.0, 58.6), (-3.0, 58.6), (-1.8, 57.5), (-2.0, 56.4), (0.3, 54.0),
        (1.7, 52.7), (1.0, 51.4), (-1.0, 50.7), (-5.0, 50.1),
    ],
    "ireland": [
        (-10.4, 51.6), (-9.0, 53.3), (-9.9, 54.3), (-7.3, 55.4), (-5.5, 54.6),
        (-6.0, 53.2), (-6.6, 52.2), (-8.5, 51.5), (-10.4, 51.6),
    ],
    "iceland": [
        (-24.5, 65.5), (-22.0, 66.5), (-16.0, 66.5), (-14.5, 65.0),
        (-18.0, 63.4), (-22.5, 63.8), (-24.5, 65.5),
    ],
    # Africa: Gibraltar down the Atlantic side, round the Cape, up past the
    # Horn and back along the Red Sea and the Maghreb.
    "africa": [
        (-5.4, 35.9), (-9.8, 31.4), (-13.2, 27.7), (-16.5, 21.3), (-17.1, 16.0),
        (-16.0, 12.4), (-13.6, 9.5), (-11.5, 7.7), (-7.5, 4.4), (-3.0, 5.1),
        (1.2, 6.1), (5.6, 4.3), (8.5, 4.4), (9.4, 2.0), (9.3, -0.6),
        (11.8, -3.0), (12.2, -5.8), (13.4, -8.7), (11.7, -15.8), (14.5, -22.5),
        (17.2, -28.5), (18.4, -32.9), (20.0, -34.8), (25.6, -34.0), (30.0, -31.0),
        (32.9, -26.0), (35.5, -21.0), (40.5, -16.0), (40.5, -11.0), (39.3, -6.8),
        (41.8, -2.0), (43.5, 2.0), (48.5, 5.5), (51.4, 10.4), (48.0, 11.5),
        (43.5, 11.0), (43.0, 12.7), (38.5, 17.5), (37.2, 21.0), (35.5, 24.0),
        (34.2, 27.5), (32.6, 29.9), (31.0, 31.4), (25.0, 32.0), (19.5, 30.5),
        (15.5, 32.4), (11.0, 33.5), (10.5, 36.9), (8.0, 37.1), (3.0, 36.8),
        (-1.0, 35.7), (-5.4, 35.9),
    ],
    "madagascar": [
        (49.4, -12.1), (50.5, -15.5), (49.5, -17.5), (48.5, -21.5), (47.0, -25.2),
        (45.2, -25.6), (43.3, -22.0), (43.4, -17.5), (46.0, -15.5), (48.0, -13.5),
        (49.4, -12.1),
    ],
    # North America: the Alaskan corner, the Arctic edge, the eastern seaboard,
    # the Gulf and Mexico, then back up the Pacific coast.
    "n-america": [
        (-166.0, 66.0), (-161.0, 70.5), (-148.0, 70.5), (-133.0, 69.5),
        (-124.0, 70.0), (-114.0, 68.5), (-105.0, 68.5), (-95.0, 68.5),
        (-85.0, 70.0), (-78.0, 73.0), (-68.0, 70.0), (-64.0, 60.0),
        (-56.5, 53.5), (-59.0, 47.5), (-64.5, 47.0), (-67.0, 44.8),
        (-70.5, 41.5), (-74.0, 40.5), (-76.0, 37.0), (-78.5, 33.8),
        (-81.5, 31.0), (-80.1, 25.2), (-82.7, 27.8), (-84.5, 30.0),
        (-89.0, 29.0), (-94.0, 29.5), (-97.4, 25.9), (-97.0, 21.0),
        (-94.5, 18.2), (-91.0, 18.7), (-88.0, 21.5), (-87.0, 18.0),
        (-83.5, 15.0), (-83.0, 9.5), (-79.5, 9.0), (-77.5, 8.0),
        (-82.5, 8.0), (-87.0, 13.0), (-92.0, 14.5), (-96.5, 15.7),
        (-101.0, 17.5), (-105.5, 20.5), (-109.5, 23.0), (-112.5, 27.0),
        (-114.5, 31.0), (-117.2, 32.5), (-121.0, 35.5), (-124.2, 40.5),
        (-124.5, 47.5), (-131.0, 53.5), (-136.0, 58.5), (-146.0, 60.5),
        (-152.0, 58.5), (-158.0, 56.0), (-162.0, 60.0), (-166.0, 66.0),
    ],
    "greenland": [
        (-45.0, 60.0), (-52.0, 66.0), (-55.0, 70.0), (-58.0, 75.5),
        (-63.0, 80.0), (-45.0, 83.0), (-25.0, 82.0), (-18.0, 76.0),
        (-22.0, 70.0), (-32.0, 66.0), (-40.0, 62.0), (-45.0, 60.0),
    ],
    # South America: Panama down the Pacific side to Cape Horn, back up the
    # Atlantic past the Brazilian bulge.
    "s-america": [
        (-77.5, 8.0), (-79.0, 2.0), (-80.9, -2.2), (-81.3, -6.0),
        (-77.0, -12.0), (-71.5, -17.5), (-70.4, -23.4), (-71.5, -30.0),
        (-73.0, -37.0), (-75.0, -43.0), (-74.5, -48.5), (-70.0, -53.5),
        (-67.0, -55.5), (-65.0, -52.0), (-68.0, -48.0), (-65.0, -43.0),
        (-62.5, -40.0), (-57.5, -35.5), (-53.5, -33.0), (-48.5, -25.5),
        (-44.0, -23.0), (-39.0, -18.0), (-37.0, -11.0), (-35.0, -8.0),
        (-38.5, -4.0), (-44.0, -2.5), (-50.0, 0.0), (-51.5, 4.0),
        (-57.0, 6.0), (-61.5, 8.5), (-66.0, 10.7), (-71.5, 11.8),
        (-74.5, 11.0), (-77.0, 8.5), (-77.5, 8.0),
    ],
    # Australia, plus the two large islands north and east of it.
    "australia": [
        (113.5, -22.0), (114.0, -26.0), (115.0, -31.5), (117.5, -35.1),
        (123.0, -34.0), (129.0, -31.7), (134.0, -32.5), (137.5, -35.5),
        (140.0, -38.0), (145.5, -38.9), (150.0, -37.5), (153.0, -32.0),
        (153.5, -28.0), (149.5, -22.5), (146.0, -19.0), (145.0, -15.0),
        (142.5, -10.7), (140.0, -17.5), (136.5, -12.0), (132.5, -11.0),
        (129.0, -14.5), (125.0, -14.0), (122.0, -18.0), (117.0, -20.5),
        (113.5, -22.0),
    ],
    "new-guinea": [
        (131.0, -1.0), (135.0, -2.0), (140.0, -2.5), (145.0, -4.5),
        (150.5, -9.5), (147.0, -8.5), (143.0, -8.0), (138.0, -8.5),
        (133.0, -4.5), (131.0, -1.0),
    ],
    "borneo": [
        (109.5, 1.8), (113.0, 3.2), (117.0, 4.2), (119.0, 1.0),
        (117.5, -3.5), (114.5, -3.5), (110.0, -3.0), (108.8, -1.0),
        (109.5, 1.8),
    ],
    "sumatra-java": [
        (95.3, 5.6), (98.5, 3.5), (102.0, 0.0), (105.5, -5.9), (114.0, -8.5),
        (114.5, -7.0), (106.0, -6.0), (102.0, -3.0), (98.0, 1.5), (95.3, 5.6),
    ],
    "japan": [
        (130.0, 31.5), (132.0, 34.0), (135.5, 34.0), (139.5, 35.0),
        (141.5, 38.5), (141.0, 41.5), (145.5, 43.5), (144.0, 45.5),
        (140.0, 43.0), (139.0, 39.0), (135.0, 35.5), (132.0, 33.5),
        (130.0, 31.5),
    ],
    "sri-lanka": [
        (79.9, 9.6), (81.2, 8.5), (81.9, 7.0), (80.4, 5.9), (79.7, 7.5),
        (79.9, 9.6),
    ],
    "new-zealand": [
        (172.7, -34.4), (174.5, -36.0), (178.5, -37.5), (177.0, -39.5),
        (174.9, -41.3), (172.0, -40.5), (170.5, -43.0), (166.5, -46.0),
        (168.5, -46.7), (172.0, -43.5), (174.0, -41.5), (175.5, -39.5),
        (174.0, -35.5), (172.7, -34.4),
    ],
}


def _unit(lon, lat):
    """A lon/lat pair as a unit vector, with +Z through the sub-observer point."""
    a, b = math.radians(lon), math.radians(lat)
    return (math.cos(b) * math.cos(a), math.cos(b) * math.sin(a), math.sin(b))


def _rotate(v):
    """Rotate the globe so VIEW_LON/VIEW_LAT faces the viewer."""
    x, y, z = v
    l = math.radians(VIEW_LON)
    p = math.radians(VIEW_LAT)
    # Spin to the view longitude, then tilt to the view latitude.
    x1 = x * math.cos(l) + y * math.sin(l)
    y1 = -x * math.sin(l) + y * math.cos(l)
    z1 = z
    # After the spin, x1 points at the viewer. Tilt about the y axis.
    x2 = x1 * math.cos(p) + z1 * math.sin(p)
    z2 = -x1 * math.sin(p) + z1 * math.cos(p)
    return (x2, y1, z2)


def _slerp(a, b, t):
    """A point t of the way along the great circle from a to b."""
    dot = max(-1.0, min(1.0, sum(u * v for u, v in zip(a, b))))
    ang = math.acos(dot)
    if ang < 1e-9:
        return a
    s = math.sin(ang)
    ca, cb = math.sin((1 - t) * ang) / s, math.sin(t * ang) / s
    return tuple(ca * u + cb * v for u, v in zip(a, b))


def _project(v):
    """Orthographic screen position for a rotated unit vector.

    A point on the far side of the globe is pushed just outside the rim in the
    same direction, keeping the outline's direction of travel intact. The whole
    drawing is clipped to the globe, so those excursions are invisible and the
    visible part of a coastline that runs over the horizon — the Americas at
    the western limb, Japan at the eastern one — is cut exactly at the edge
    instead of being chorded across the face.
    """
    fx, fy, fz = v[0], v[1], v[2]
    rho = math.hypot(fy, fz)
    if fx < 0:
        if rho < 1e-6:
            return (GX, GY)  # the antipode: direction is undefined, park it
        k = (1.04 + (1.0 - rho) * 0.55) / rho
        fy, fz = fy * k, fz * k
    return (GX + fy * R, GY - fz * R)


def coast_paths():
    """Every visible coastline as an SVG path, projected onto the globe.

    Three things keep the output small enough to inline in a template that is
    served on every page:

      · a landmass with no point on the near side is dropped outright. On this
        face that removes New Zealand, Borneo, Java and most of the Pacific
        rim, none of which could be drawn anyway
      · an edge is only subdivided when it runs near or across the horizon,
        where a straight chord would visibly cut the corner. Well inside the
        face, a great circle and a straight line differ by less than a tenth of
        a unit at this radius
      · a projected point within 0.14 units of the last one kept is dropped,
        which is below what a 64px drawing can resolve
    """
    out = []
    for name, ring in COASTS.items():
        verts = [_rotate(_unit(lon, lat)) for lon, lat in ring]
        if all(v[0] < 0 for v in verts):
            continue  # entirely behind the globe on this face

        pts = []
        for i in range(len(verts)):
            a, b = verts[i], verts[(i + 1) % len(verts)]
            near_limb = a[0] < 0.32 or b[0] < 0.32
            steps = 6 if near_limb else 1
            for s in range(steps):
                p = _project(_slerp(a, b, s / steps) if steps > 1 else a)
                if pts and math.hypot(p[0] - pts[-1][0], p[1] - pts[-1][1]) < 0.14:
                    continue
                pts.append(p)
        # Collapse runs that are wholly outside the rim. They are clipped away
        # at render time, so all that is needed there is enough of a detour to
        # keep the ring's winding intact — the intermediate points are bytes
        # spent drawing something nobody sees.
        kept = []
        for i, p in enumerate(pts):
            outside = math.hypot(p[0] - GX, p[1] - GY) > R * 1.01
            prev_out = kept and math.hypot(kept[-1][0] - GX, kept[-1][1] - GY) > R * 1.01
            nxt = pts[(i + 1) % len(pts)]
            next_out = math.hypot(nxt[0] - GX, nxt[1] - GY) > R * 1.01
            if outside and prev_out and next_out:
                continue
            kept.append(p)
        if len(kept) < 3:
            continue
        d = "M" + " L".join(f"{fmt(x)} {fmt(y)}" for x, y in kept) + " Z"
        out.append((name, d))
    return out


def graticule():
    """Equator and one meridian, drawn as the projection actually places them.

    Two lines, not a net. The coastlines are what say "planet" now; the
    graticule is only there to curve the surface, and at 34px a full net over
    green landmasses turns the whole face to hatching.
    """
    rows = []
    for lons, lats in (
        ([l for l in range(-180, 181, 10)], None),  # the equator
        (None, [l for l in range(-90, 91, 10)]),    # the prime meridian
    ):
        if lats is None:
            ring = [(lon, 0.0) for lon in lons]
        else:
            ring = [(0.0, lat) for lat in lats]
        seg, runs = [], []
        for lon, lat in ring:
            v = _rotate(_unit(lon, lat))
            if v[0] < 0:  # behind the globe — break the line rather than wrap it
                if len(seg) > 1:
                    runs.append(seg)
                seg = []
                continue
            seg.append(_project(v))
        if len(seg) > 1:
            runs.append(seg)
        for run in runs:
            rows.append("M" + " L".join(f"{fmt(x)} {fmt(y)}" for x, y in run))
    return rows


def body(uid):
    """The drawing itself, shared by the logo template and the favicon."""
    d = disc_centres()
    L = []
    a = L.append

    a(f'<circle cx="{fmt(GX)}" cy="{fmt(GY)}" r="{fmt(R)}" fill="url(#pvS{uid})"/>')
    a(f'<g clip-path="url(#pvC{uid})">')
    a(f'  <g fill="url(#pvN{uid})">')
    for name, path in coast_paths():
        a(f'    <path d="{path}"/>')
    a("  </g>")
    a(f'  <g fill="none" stroke="url(#pvL{uid})" stroke-width=".8">')
    for row in graticule():
        a(f'    <path d="{row}"/>')
    a("  </g>")
    for cx, cy, r in d:
        a(
            f'  <ellipse cx="{fmt(cx + 1.1)}" cy="{fmt(cy + 1.5)}" rx="{fmt(r * 1.16)}" '
            f'ry="{fmt(r * 1.1)}" fill="url(#pvD{uid})"/>'
        )
    a(
        f'  <ellipse cx="{fmt(GX - R * 0.42)}" cy="{fmt(GY - R * 0.52)}" rx="{fmt(R * 0.46)}" '
        f'ry="{fmt(R * 0.30)}" fill="url(#pvH{uid})" '
        f'transform="rotate(-26 {fmt(GX - R * 0.42)} {fmt(GY - R * 0.52)})"/>'
    )
    a("</g>")
    # No rim and no limb darkening. "Around the earth there is a borderline,
    # remove this completely, so the earth stretches now till to the end (do not
    # make it smaller)" — so the sea's own gradient is the outermost thing in
    # the drawing and it runs to r exactly. The sphere still reads as a ball
    # because pvS is lit from the upper left; what is gone is the ring, not the
    # roundness.

    # Three discs, one flat colour each and nothing around them. The collar,
    # the hairline and the sheen arc that used to ring every disc are what the
    # client called "2 borderlines"; a disc is now only its background colour.
    a('<g class="logo__badges">')
    for (cx, cy, r), fill in zip(d, DISC_FILLS):
        a(f'  <circle cx="{fmt(cx)}" cy="{fmt(cy)}" r="{fmt(r)}" fill="{fill}"/>')
    a("</g>")

    a('<g class="logo__houses" fill="#FAF4E6">')
    for cx, cy, r in d:
        a(f'  <path d="{house(cx, cy, r)[0]}"/>')
    a("</g>")
    for (cx, cy, r), roof in zip(d, ROOFS):
        a(f'<path d="{house(cx, cy, r)[1]}" fill="{roof}"/>')
    a('<g fill="#123A56">')
    for cx, cy, r in d:
        a(f'  <path d="{house(cx, cy, r)[2]}"/>')
    a("</g>")
    a(
        '<g fill="none" stroke="#0A2438" stroke-opacity=".85" stroke-width=".75" '
        'stroke-linejoin="round">'
    )
    for cx, cy, r in d:
        a(f'  <path d="{house(cx, cy, r)[0]}"/>')
    a("</g>")
    return "\n".join(L)


def mark_bbox():
    """The outer box of the whole drawing: the globe plus the discs' overhang.

    Two of the three discs hang past the rim, so the drawing is wider than the
    planet — 44.8 across against the globe's 44 — and a favicon that sized
    itself on the circle alone would cut a slice off the red one.
    """
    xs, ys = [GX - R, GX + R], [GY - R, GY + R]
    for cx, cy, r in disc_centres():
        xs += [cx - r, cx + r]
        ys += [cy - r, cy + r]
    return min(xs), min(ys), max(xs), max(ys)


def favicon_transform(size=64.0):
    """The transform that fills the favicon tile with the mark.

    The client's 18 August note: "the favicon we do that this planet with the
    houses in a max size, what can be, so in the favicon area the left and
    right borders of the planet and up and down borders touch the favicon area
    borders."

    So the drawing is scaled to the tile instead of being inset in it — the .88
    margin that used to keep the overhanging discs off the corners is gone. The
    box that is fitted is the drawing's, not the circle's: fitting the circle
    would put the planet's edge exactly on all four borders but push the red
    disc 1.2 units past the right one, and a disc sliced flat against the tile
    edge reads as a rendering fault. Fitting the drawing puts the planet on the
    left border, within half a unit of the top and bottom ones, and the red
    disc exactly on the right — the planet fills the tile, and nothing is cut.
    """
    x0, y0, x1, y1 = mark_bbox()
    k = size / max(x1 - x0, y1 - y0)
    cx, cy = (x0 + x1) / 2, (y0 + y1) / 2
    return k, cx, cy


def defs(uid):
    return f"""<radialGradient id="pvS{uid}" cx=".34" cy=".26" r=".86">
  <stop offset="0" stop-color="#FFEFB8"/>
  <stop offset=".40" stop-color="#F8CE55"/>
  <stop offset=".78" stop-color="#EDAE30"/>
  <stop offset="1" stop-color="#D89A22"/>
</radialGradient>
<linearGradient id="pvN{uid}" x1=".15" y1="0" x2=".85" y2="1">
  <stop offset="0" stop-color="#48A863"/>
  <stop offset="1" stop-color="#22794A"/>
</linearGradient>
<linearGradient id="pvL{uid}" x1=".12" y1=".08" x2=".88" y2=".92">
  <stop offset="0" stop-color="#0C4A2C" stop-opacity=".34"/>
  <stop offset="1" stop-color="#0C4A2C" stop-opacity=".14"/>
</linearGradient>
<radialGradient id="pvD{uid}">
  <stop offset=".38" stop-color="#5A3708" stop-opacity=".34"/>
  <stop offset="1" stop-color="#5A3708" stop-opacity="0"/>
</radialGradient>
<radialGradient id="pvH{uid}">
  <stop offset="0" stop-color="#fff" stop-opacity=".34"/>
  <stop offset="1" stop-color="#fff" stop-opacity="0"/>
</radialGradient>
<clipPath id="pvC{uid}"><circle cx="{fmt(GX)}" cy="{fmt(GY)}" r="{fmt(R)}"/></clipPath>"""


# The favicon tile. Flat #8B008B, named by the client on 18 August: "around the
# favicon we draw rounded rectangle like it is now, and this background make
# #cc00cc", and then, after seeing that first magenta in a tab strip, "the
# favicon backgroun lets try: #8B008B". Flat rather than the lit gradient it
# had, because a tile that is also shaded competes with the planet sitting on
# it — and the client named one colour, not two.
TILE_FILL = "#8B008B"
TILE_RADIUS = 14.0


def favicon(uid="f"):
    """The complete favicon: the tile, then the mark scaled to fill it."""
    k, cx, cy = favicon_transform()
    L = [
        '<svg viewBox="0 0 64 64" xmlns="http://www.w3.org/2000/svg">',
        "  <defs>",
    ]
    L += ["    " + row for row in defs(uid).splitlines()]
    L += [
        "  </defs>",
        f'  <rect width="64" height="64" rx="{fmt(TILE_RADIUS)}" fill="{TILE_FILL}"/>',
        f'  <g transform="translate(32 32) scale({k:.4f}) translate({fmt(-cx)} {fmt(-cy)})">',
    ]
    L += ["    " + row for row in body(uid).splitlines()]
    L += ["  </g>", "</svg>"]
    return "\n".join(L)


if __name__ == "__main__":
    import sys

    args = sys.argv[1:]
    if args and args[0] == "--favicon":
        print(favicon(args[1] if len(args) > 1 else "f"))
    else:
        uid = args[0] if args else "f"
        print('<svg viewBox="0 0 64 64" xmlns="http://www.w3.org/2000/svg">')
        print("<defs>")
        print(defs(uid))
        print("</defs>")
        print(body(uid))
        print("</svg>")
