# 3D Board Visualization

**Priority**: HIGH - Replace current flat SVG board with 3D isometric/popping tiles

## Overview

Replace the current 2D SVG board rendering with a 3D visualization library that creates depth and "popping" hex tiles similar to low-poly game aesthetics.

## Library Choice

Use **React Three Fiber** (R3F) - React renderer for Three.js. It's the most mature React-friendly 3D library.

Install: `npm install three @react-three/fiber @react-three/drei`

## Acceptance Criteria

### Visual Requirements

- [ ] Hex tiles rendered as 3D extruded hexagonal prisms with varying heights
- [ ] Each resource type has distinct color AND height:
  - Wood (forest): Green, tall (trees)
  - Brick (hills): Red/brown, medium
  - Sheep (pasture): Light green, low/flat
  - Wheat (fields): Yellow/gold, low with texture
  - Ore (mountains): Gray, tallest
  - Desert: Sandy beige, flat
- [ ] Tiles have slight gaps between them for visual separation
- [ ] Soft shadows cast by taller tiles onto shorter ones
- [ ] Isometric camera angle (roughly 45-60 degrees)

### Interactive Requirements

- [ ] Camera can be rotated with mouse drag (OrbitControls)
- [ ] Zoom in/out with scroll wheel
- [ ] Vertices rendered as 3D spheres at hex corners (clickable)
- [ ] Edges rendered as 3D cylinders/roads between vertices (clickable)
- [ ] Hover effects: tiles glow/highlight, vertices pulse
- [ ] Click handlers work same as current 2D board (placement callbacks)

### Data Attributes

- [ ] All interactive elements retain `data-cy` attributes for Playwright
- [ ] Playwright can still select vertices/edges by attribute

### Performance

- [ ] Smooth 60fps on mid-range hardware
- [ ] Use instanced meshes for repeated geometry (hex tiles)
- [ ] Lazy load Three.js bundle (code split)

## Technical Approach

### File Structure
