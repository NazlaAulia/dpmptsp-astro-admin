# @dpmptsp/ui

Shared frontend primitives.

## Why there are no components in here

SPEC.md §11.7 says a shared package is for what is genuinely shared, and §11.1
says not to build an abstraction before there is a real need. When the apps were
split, **nothing** was actually shared between the public site and the admin
panel: the only import crossing the boundary was the login page pulling in the
public layout, which was an accident and was deleted.

The two surfaces also have opposite designs — the public site is Tailwind with
gradient hero sections, the admin panel is scoped `<style>` blocks with its own
spacing scale and a dark-theme toggle the public site has no equivalent of.
A shared `Button` would have to satisfy both and would end up satisfying
neither.

So this package holds only the things that are genuinely common and genuinely
pure: no framework, no DOM, no data access, no dependencies. That also makes
them the only directly unit-testable part of the frontend.

Add a component here when a second real consumer exists — not in anticipation
of one.
