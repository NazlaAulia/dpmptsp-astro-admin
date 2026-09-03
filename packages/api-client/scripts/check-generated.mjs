// Fails if src/generated/ is out of date with apps/api/openapi.yaml.
//
// The build reads the committed generated file, so a spec change that was never
// regenerated — or a spec that no longer parses — otherwise passes unnoticed.

import { execFileSync } from "node:child_process";
import { mkdtempSync, readFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const committed = "src/generated/schema.d.ts";
const tmp = mkdtempSync(join(tmpdir(), "apiclient-"));
const fresh = join(tmp, "schema.d.ts");

try {
  execFileSync("npx", ["openapi-typescript", "../../apps/api/openapi.yaml", "-o", fresh], {
    stdio: ["ignore", "ignore", "inherit"],
  });
} catch {
  console.error("openapi.yaml could not be compiled — see the error above.");
  process.exit(1);
}

const a = readFileSync(committed, "utf8");
const b = readFileSync(fresh, "utf8");
rmSync(tmp, { recursive: true, force: true });

if (a !== b) {
  console.error(
    "src/generated/ is stale. Run: pnpm --filter @dpmptsp/api-client generate"
  );
  process.exit(1);
}
console.log("generated types are up to date");
