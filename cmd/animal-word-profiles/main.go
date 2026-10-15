package main

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type profile struct {
	Slug, File, URL, Species string
	Expected                 int
}

var profiles = []profile{
	{"nasonia", "animals-Nasonia.doc", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Nasonia.doc", "Nasonia vitripennis", 107},
	{"leafcutter-ant", "animals-leafcutter.ant.doc", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/leafcutter.ant.doc", "Atta cephalotes", 59},
	{"seed-harvester-ant", "animals-Pogonomyrmex.doc", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Pogonomyrmex.doc", "Pogonomyrmex barbatus", 115},
	{"argentine-ant", "animals-Linepithema.doc", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Linepithema.doc", "Linepithema humile", 178},
	{"aphid", "animals-aphid.doc", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/aphid.doc", "Acyrthosiphon pisum", 68},
	{"public-insect", "animals-public.insect.doc", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/public.insect.doc", "Public insect collection", 1744},
	{"spider-mite", "animals-Two.spotted.spider.mite.doc", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Two.spotted.spider.mite.doc", "Tetranychus urticae", 95},
	{"ixodes", "animals-ixodes.doc", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/ixodes.doc", "Ixodes scapularis", 338},
	{"daphnia", "animals-Daphnia.doc", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Daphnia.doc", "Daphnia pulex", 87},
	{"hydra", "animals-hydra.doc", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/hydra.doc", "Hydra magnipapillata", 27},
	{"trichoplax", "animals-trichoplax.doc", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/trichoplax.doc", "Trichoplax adhaerens", 107},
	{"sponge", "animals-spongeCYPs.doc", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/spongeCYPs.doc", "Reniera sp.", 19},
}
var par = regexp.MustCompile(`\([^)]*\)`)
var nums = regexp.MustCompile(`\d+`)
var safe = regexp.MustCompile(`[^A-Za-z0-9]+`)

type record struct {
	Block, Line                                int
	Header, Species, Symbol, ID, Sequence, Key string
	Status                                     []string
}

func main() {
	for _, p := range profiles {
		if e := run(p); e != nil {
			panic(fmt.Errorf("%s: %w", p.Slug, e))
		}
	}
}
func run(p profile) error {
	txt := filepath.Join("raw", strings.TrimSuffix(p.File, filepath.Ext(p.File))+".txt")
	rs, l, e := parse(p, txt)
	if e != nil {
		return e
	}
	dh, e := hash(filepath.Join("raw", p.File))
	if e != nil {
		return e
	}
	th, e := hash(txt)
	if e != nil {
		return e
	}
	if e = writeCSV(p, rs, dh, th); e != nil {
		return e
	}
	if e = writeAudit(p, rs, l, dh, th); e != nil {
		return e
	}
	fmt.Printf("%s: %d records\n", p.Species, len(rs))
	return nil
}
func parse(p profile, path string) ([]record, int, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, 0, e
	}
	ls := strings.Split(strings.ReplaceAll(strings.ReplaceAll(string(b), "\r\n", "\n"), "\r", "\n"), "\n")
	var st []int
	for i, l := range ls {
		if strings.HasPrefix(strings.TrimSpace(l), ">") {
			st = append(st, i)
		}
	}
	st = append(st, len(ls))
	var out []record
	for bi := 0; bi < len(st)-1; bi++ {
		a, z := st[bi], st[bi+1]
		h := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(ls[a]), ">"))
		sym, id, species := identity(p, h, bi+1)
		var sb strings.Builder
		if p.Slug == "public-insect" {
			sb.WriteString(publicInsectInlineProtein(h))
		}
		for _, l := range ls[a+1 : z] {
			q := proteinLine(l)
			sb.WriteString(q)
			if strings.HasSuffix(q, "*") {
				break
			}
		}
		s := sb.String()
		var status []string
		if s == "" {
			status = append(status, "source-sequence-missing")
		} else {
			if strings.HasSuffix(s, "*") {
				s = strings.TrimSuffix(s, "*")
				status = append(status, "terminal-stop-removed")
			}
			if strings.Contains(s, "*") {
				status = append(status, "internal-stop-retained")
			}
			if strings.ContainsAny(s, "Xx") {
				status = append(status, "ambiguous-X-retained")
			}
			if strings.Contains(s, "-") {
				status = append(status, "source-gap-retained")
			}
			if len(s) < 350 {
				status = append(status, "short-source-fragment")
			}
		}
		lo := strings.ToLower(h)
		for _, m := range []struct{ k, v string }{{"pseudo", "source-pseudogene"}, {"partial", "source-partial"}, {"fragment", "source-fragment"}, {"duplicate", "source-duplicate"}, {"allele", "source-allele"}, {"frameshift", "source-frameshift"}} {
			{
				if strings.Contains(lo, m.k) {
					status = append(status, m.v)
				}
			}
		}
		key := fmt.Sprintf("%s:block-%04d:line-%04d:%s", p.Slug, bi+1, a+1, strings.ToLower(safe.ReplaceAllString(id, "-")))
		out = append(out, record{bi + 1, a + 1, h, species, sym, id, s, key, status})
	}
	if len(out) != p.Expected {
		return nil, len(ls), fmt.Errorf("headers=%d want%d", len(out), p.Expected)
	}
	return out, len(ls), nil
}

func publicInsectInlineProtein(header string) string {
	fields := strings.Fields(header)
	for i, field := range fields {
		if !regexp.MustCompile(`^[A-Z]{1,4}_?\d{5,}(?:\.\d+)?$`).MatchString(field) {
			continue
		}
		var b strings.Builder
		for _, token := range fields[i+1:] {
			if !res(token) {
				return ""
			}
			b.WriteString(token)
		}
		if b.Len() >= 100 {
			return b.String()
		}
	}
	return ""
}
func identity(p profile, h string, n int) (string, string, string) {
	f := strings.Fields(h)
	first := ""
	if len(f) > 0 {
		first = strings.TrimLeft(f[0], "*&")
	}
	sym, id := first, first
	if sym == "" {
		sym = fmt.Sprintf("block-%04d", n)
		id = sym
	}
	species := p.Species
	if p.Slug == "daphnia" && strings.Contains(strings.ToLower(h), "daphnia magna") {
		species = "Daphnia magna"
	}
	if p.Slug == "public-insect" {
		species = publicInsectSpecies(h)
	}
	return sym, id, species
}

// publicInsectSpecies recognizes only complete binomials explicitly printed in
// this fixed historical resource.  The collection contains prose that also
// looks like two lowercase words, so an open-ended title-case regex would
// silently invent taxa.  Unlisted/abbreviated/ambiguous headers retain the
// aggregate collection label and their exact source header for provenance.
func publicInsectSpecies(header string) string {
	known := []string{
		"Acyrthosiphon pisum", "Aedes aegypti", "Aedes albopictus", "Anopheles albimanus",
		"Anopheles funestus", "Anopheles gambiae", "Anopheles minimus", "Antheraea mylitta",
		"Antheraea pernyi", "Antheraea yamamai", "Apis mellifera", "Bemisia tabaci",
		"Blattella germanica", "Blatella germanica", "Bombyx mandarina", "Bombyx mori",
		"Ceratitis capitata", "Chironomus tentans", "Cnaphalocrocis medinalis",
		"Coptotermes acinaciformis", "Culex pipiens pallens", "Culex pipiens",
		"Culex quinquefasciatus", "Danaus plexippus", "Dendroctonus ponderosae",
		"Depressaria pastinacella", "Diabrotica virgifera", "Diploptera punctata",
		"Drosophila acanthoptera", "Drosophila hydei", "Drosophila melanogaster",
		"Drosophila mettleri", "Drosophila nigrospiracula", "Drosophila simulans",
		"Epiphyas postvittana", "Heliconius erato", "Helicoverpa armigera", "Helicoverpa zea",
		"Heliothis virescens", "Hodotermopsis sjostedti", "Homalodisca vitripennis",
		"Ips paraconfusus", "Ips pini", "Laupala kohalensis", "Leptinotarsa decemlineata",
		"Liposcelis bostrychophila", "Locusta migratoria", "Lucilia cuprina",
		"Lutzomyia longipalpis", "Lygus lineolaris", "Mamestra brassicae",
		"Manduca sexta", "Mastotermes darwiniensis", "Mayetiola destructor",
		"Musca domestica", "Nasonia vitripennis", "Nilaparvata lugens",
		"Ochlerotatus sollicitans", "Ostrinia furnacalis", "Papilio canadensis",
		"Papilio glaucus", "Papilio polyxenes", "Papilio xuthus", "Pediculus humanus",
		"Philotrypesis pilosa", "Phyllopertha diversa", "Plutella xylostella",
		"Reticulitermes flavipes", "Solenopsis invicta", "Spodoptera exigua",
		"Spodoptera litura", "Spodoptera littoralis", "Tribolium castaneum",
		"Trichogramma cacoeciae", "Zygaena filipendulae",
	}
	lower := strings.ToLower(header)
	// Longest first prevents Culex pipiens from swallowing its explicit pallens form.
	sort.SliceStable(known, func(i, j int) bool { return len(known[i]) > len(known[j]) })
	for _, name := range known {
		if strings.Contains(lower, strings.ToLower(name)) {
			return name
		}
	}
	return "Public insect collection"
}
func proteinLine(l string) string {
	l = par.ReplaceAllString(l, " ")
	l = nums.ReplaceAllString(l, " ")
	l = strings.ReplaceAll(l, "&", " ")
	var b strings.Builder
	for _, t := range strings.Fields(l) {
		if !res(t) {
			return ""
		}
		b.WriteString(t)
	}
	return b.String()
}
func res(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r < 'A' || r > 'Z') && r != 'x' && r != '*' && r != '-' {
			return false
		}
	}
	return true
}
func hash(p string) (string, error) {
	b, e := os.ReadFile(p)
	if e != nil {
		return "", e
	}
	return fmt.Sprintf("%x", sha256.Sum256(b)), nil
}
func writeCSV(p profile, rs []record, dh, th string) error {
	path := filepath.Join("sources", "reviewed", "animals", p.Slug+".csv")
	if e := os.MkdirAll(filepath.Dir(path), 0o755); e != nil {
		return e
	}
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "normalized_file", "normalized_sha256", "source_block", "source_line", "source_header", "review_status"})
	for _, r := range rs {
		_ = w.Write([]string{"animals", r.Species, r.Symbol, r.ID, r.Key, p.URL, fmt.Sprintf("%s fixed Word profile block %d line %d; exact header=%s", p.Slug, r.Block, r.Line, r.Header), r.Sequence, p.File, dh, strings.TrimSuffix(p.File, filepath.Ext(p.File)) + ".txt", th, strconv.Itoa(r.Block), strconv.Itoa(r.Line), r.Header, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}
func writeAudit(p profile, rs []record, l int, dh, th string) error {
	c := map[string]int{}
	seq := 0
	for _, r := range rs {
		if r.Sequence != "" {
			seq++
		}
		for _, s := range r.Status {
			c[s]++
		}
	}
	ks := make([]string, 0, len(c))
	for k := range c {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	var b strings.Builder
	fmt.Fprintf(&b, "# Animal resource review: %s\n\n- File: `%s`\n- URL: %s\n- DOC SHA-256: `%s`\n- Text SHA-256: `%s`\n- Lines: `%d`\n- Explicit blocks: `%d`\n- Literal proteins: `%d`\n- Missing: `%d`\n- Fixed profile: `%s`\n- Review: `complete`\n\nEvery physical header remains source-distinct. Protein comes only from literal residue lines inside that block; coordinates/phase/prose are excluded and parsing stops at terminal stop. Pseudogenes, fragments, duplicates, alleles, X/x, gaps and internal stops remain. No DNA is translated and no external sequence is substituted.\n\n| Status | Records |\n|---|---:|\n", p.Species, p.File, p.URL, dh, th, l, len(rs), seq, len(rs)-seq, p.Slug)
	for _, k := range ks {
		fmt.Fprintf(&b, "| %s | %d |\n", k, c[k])
	}
	path := filepath.Join("docs", "animals", p.Slug+".md")
	if e := os.MkdirAll(filepath.Dir(path), 0o755); e != nil {
		return e
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
