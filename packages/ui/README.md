# @dpmptsp/ui

Shared frontend primitives.

## No components

The public site and the admin panel share no markup: they use different design
systems, and nothing crossed the boundary when they were split.

This package therefore holds only pure helpers — no framework, no DOM, no data
access, no dependencies. Add a component when a second consumer actually needs
one.
