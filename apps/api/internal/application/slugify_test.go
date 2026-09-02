package application

import "testing"

// These expectations are copied from apps/web/tests/slugify.test.mjs, which
// pins the behaviour of the JavaScript implementation article URLs were
// generated with.
//
// They must not be "improved". Every one of these strings may already be a live
// URL, and a nicer slug is a 404.
func TestSlugifyMatchesTheJavaScriptImplementation(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Kunjungan Wisatawan Kapal Pesiar", "kunjungan-wisatawan-kapal-pesiar"},
		{"  Spasi   Ganda  ", "spasi-ganda"},
		{"Investasi, Tahun 2024!", "investasi-tahun-2024"},
		{"Tanda_Bawah", "tanda_bawah"},
		{"", ""},

		// --- the quirks ---

		// An en dash between spaces: both spaces become hyphens, then the dash
		// itself is stripped, leaving two. The legacy dump contains 20 en
		// dashes, so this shape exists in real URLs.
		{"Perizinan Berusaha – OSS", "perizinan-berusaha--oss"},

		// Accented letters are deleted outright rather than folded. "É" is not
		// in \w, so it disappears; transliterating to "e" would produce a
		// different slug and break the link.
		{"Éksport Impor", "ksport-impor"},

		// Same double-hyphen shape via an ampersand.
		{"DPMPTSP & Pelayanan", "dpmptsp--pelayanan"},
	}

	for _, c := range cases {
		if got := Slugify(c.in); got != c.want {
			t.Errorf("Slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
