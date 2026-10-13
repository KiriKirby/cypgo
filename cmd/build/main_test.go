package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestReviewedPlantResourceCountsAndRepresentativeSequences(t *testing.T) {
	rows, err := readReviewedSources("../../sources")
	if err != nil {
		t.Fatal(err)
	}
	type totals struct{ records, sequences int }
	got := map[string]totals{}
	byKey := map[string]record{}
	for _, row := range rows {
		if row.Category != "plants" {
			continue
		}
		value := got[row.Species]
		value.records++
		if row.Sequence != "" {
			value.sequences++
		}
		got[row.Species] = value
		byKey[row.RecordKey] = row
	}
	want := map[string]totals{
		"potato":                {920, 899},
		"Capsicum annuum":       {617, 617},
		"Thellungiella parvula": {207, 207},
		"Eutrema salsugineum":   {226, 226},
		"Capsella grandiflora":  {220, 220},
		"Capsella rubella":      {245, 245},
		"Brassica rapa":         {377, 377},
		"Boechera stricta":      {227, 227},
		"Carica papaya":         {182, 182},
		"Mimulus guttatus":      {368, 368},
		"Theobroma cacao":       {336, 336},
		"Gossypium raimondii":   {449, 449},
		"Eucalyptus grandis":    {762, 762},
		"Cucumis sativus":       {229, 229},
		"Citrulus lanatus":      {233, 233},
		"Prunus persica":        {625, 625},
		"Prunus mume":           {282, 282},
		"Malus domestica":       {331, 331},
		"Fragaria ananassa":     {206, 206},
		"Fragaria vesca":        {331, 331},
		"Cannabis sativa":       {356, 356},
		"Glycine max":           {171, 115},
		"Medicago truncatula":   {349, 349},
		"Lotus japonicus":       {246, 246},
		"Cicer arietinum":       {211, 211},
		"Cajanus cajanifolius":  {291, 291},
		"Jatropha curcas":       {481, 481},
		"Ricinus communis":      {263, 263},
		"Manihot esculenta":     {336, 336},
		"Linum usitatissimum":   {468, 468},
		"Vitis vinifera":        {672, 672},
		"Nelumbo nucifera":      {364, 364},
		"Aquilegia coerulea":    {1023, 1023},
		"Oryza sativa":          {427, 15},
	}
	for species, expected := range want {
		if got[species] != expected {
			t.Errorf("%s totals=%+v want %+v", species, got[species], expected)
		}
	}
	var totalRecords, totalSequences int
	for _, value := range got {
		totalRecords += value.records
		totalSequences += value.sequences
	}
	if totalRecords != 21048 || totalSequences != 19897 {
		t.Fatalf("reviewed plant totals=%d records/%d sequences, want 21048/19897", totalRecords, totalSequences)
	}
	checks := map[string]struct {
		id, symbol, prefix, suffix string
		length                     int
	}{
		"thellungiella-parvula:row-0002:Thelparv1-1.3647":                                    {"Thelparv1-1.3647", "CYP51G1", "MELDSENKLL", "VRYKRRQLS", 488},
		"thellungiella-parvula:row-0105:Thelparv3-6.460":                                     {"Thelparv3-6.460", "CYP84A25", "MESLSSQTLN", "PNTRLICPD", 519},
		"thellungiella-parvula:row-0208:Thelparv5-6.70":                                      {"Thelparv5-6.70", "CYP735A2", "MMVVLVKCVL", "VQLILKPLDS", 517},
		"eutrema-salsugineum:row-0002:Eutrsals5.11644":                                       {"Eutrsals5.11644", "CYP51G1", "MEMDSENKLL", "VRYKRRQLS", 488},
		"eutrema-salsugineum:row-0115:Eutrsals11.1015":                                       {"Eutrsals11.1015", "CYP84A25", "MESLLSQTLN", "NTRLICPVST", 521},
		"eutrema-salsugineum:row-0227:Eutrsals9.9885":                                        {"Eutrsals9.9885", "CYP735A2", "MMVVLVKYVL", "VQLILKPLDS", 518},
		"capsella-grandiflora:row-0002:Capsgran568.245":                                      {"Capsgran568.245", "CYP51G1", "MELDSDNKLL", "MVRYKRRQLS", 488},
		"capsella-grandiflora:row-0112:Capsgran1383.277":                                     {"Capsgran1383.277", "CYP84A28", "MDCLLCSPIL", "PSYRLNCPML", 510},
		"capsella-grandiflora:row-0221:Capsgran6164.55":                                      {"Capsgran6164.55", "CYP735A2", "MMFTLVLKYI", "VQLVLKPLDS", 512},
		"capsella-rubella:row-0002:Capsrube1.4039":                                           {"Capsrube1.4039", "CYP51G1", "MELDSDNKLL", "MVRYKRRQLS", 488},
		"capsella-rubella:row-0124:Capsrube1.225":                                            {"Capsrube1.225", "CYP86A4", "MEISNVMLLV", "VYLNTGVAMV", 542},
		"capsella-rubella:row-0246:Capsrube2.8371":                                           {"Capsrube2.8371", "CYP735A2", "MMFTLVLKYI", "VQLVLKPLDS", 512},
		"brassica-rapa:row-0002:BrasrapaChr06.4366":                                          {"BrasrapaChr06.4366", "CYP51G1", "MDLDSDKNLL", "MVRYKRRQLS", 488},
		"brassica-rapa:row-0190:BrasrapaChr01.811":                                           {"BrasrapaChr01.811", "CYP81H1", "MEETKLEKEK", "VASELLLRIQ", 525},
		"brassica-rapa:row-0378:BrasrapaChr02.9859":                                          {"BrasrapaChr02.9859", "CYP735A2", "MMFVSVKYAL", "VQLILKPLDS", 517},
		"boechera-stricta:row-0002:Boecstri13671.1564":                                       {"Boecstri13671.1564", "CYP51G1", "MELNSENKLL", "MVRYKRRQLS", 488},
		"boechera-stricta:row-0115:Boecstri26326.1829":                                       {"Boecstri26326.1829", "CYP82F1", "MKLFLLSDLF", "GPSFKLGIIP", 383},
		"boechera-stricta:row-0228:Boecstri26959.1471":                                       {"Boecstri26959.1471", "CYP735A2", "MMVTLVLKYV", "GVQLILKPLD", 511},
		"carica-papaya:cyp51g1:block-0001":                                                   {"CYP51G1", "CYP51G1", "MDVDNKFFNM", "MVRYKRRQIA", 484},
		"carica-papaya:cyp87a10:block-0105":                                                  {"CYP87A10", "CYP87A10", "MWGVCTGTVI", "MEKDTTNRTI", 477},
		"carica-papaya:cyp736a16:block-0225":                                                 {"CYP736A16", "CYP736A16", "MDWIWSLFSF", "CAIPSFRLKI", 498},
		"mimulus-guttatus:row-0002:Mimugutt2.513":                                            {"Mimugutt2.513", "CYP51G1", "MESVESTIAN", "YKRRALLPSH", 493},
		"mimulus-guttatus:row-0185:Mimugutt5.4318":                                           {"Mimugutt5.4318", "CYP86B7", "MNSSSSPSSL", "EIEKFLKLQS", 543},
		"mimulus-guttatus:row-0369:Mimugutt15.202":                                           {"Mimugutt15.202", "CYP76AH10", "MIVKVEYLKA", "MSFFSSFSKI", 456},
		"theobroma-cacao:row-0002:Theocaca2.12599":                                           {"Theocaca2.12599", "CYP51G1", "MEVDDKFLNM", "RYKRRQLSVN", 486},
		"theobroma-cacao:row-0169:Theocaca4.19207":                                           {"Theocaca4.19207", "CYP85A1", "MALFMVILGL", "NGLHIRVSSY", 465},
		"theobroma-cacao:row-0337:Theocaca2.40273":                                           {"Theocaca2.40273", "CYP749A9", "MTTLGNPVII", "VGGLHIYINN", 487},
		"gossypium-raimondii:row-0002:GossraimChr05.57164":                                   {"GossraimChr05.57164", "CYP51G1", "MEADDKFLNM", "RYKRRQLSVE", 486},
		"gossypium-raimondii:row-0226:GossraimChr08.52938":                                   {"GossraimChr08.52938", "CYP87A37", "MLVLMCIGTL", "IQLLEKTRME", 474},
		"gossypium-raimondii:row-0450:GossraimChr13.18698":                                   {"GossraimChr13.18698", "CYP749A9", "MGDPIMILST", "HSLKSEEGPS", 519},
		"eucalyptus-grandis:row-0002:Eucagran2.60235":                                        {"Eucagran2.60235", "CYP51G1", "MDTDNKLFNV", "VRYKRRQLSV", 485},
		"eucalyptus-grandis:row-0382:Eucagran4.35942":                                        {"Eucagran4.35942", "CYP88A20", "MELEVAASVA", "QIRKTPXXXX", 487},
		"eucalyptus-grandis:row-0767:Eucagran2.22888":                                        {"Eucagran2.22888", "CYP714E16", "MNINLTSCSI", "SLQNTCTLLL", 412},
		"cucumis-sativus:row-0002:Cucusati00919.1591":                                        {"Cucusati00919.1591", "CYP51G1", "MEPDNIKFFN", "RYKRRKLSVS", 487},
		"cucumis-sativus:row-0116:Cucusati03487.421":                                         {"Cucusati03487.421", "CYP87D19", "MSTPVTGLVA", "GIHVSFSAKA", 475},
		"cucumis-sativus:row-0230:Cucusati02653.1189":                                        {"Cucusati02653.1189", "CYP749A47", "MDVLSKYIII", "LQVILRSISN", 520},
		"citrulus-lanatus:row-0001:Citrlana13966.2":                                          {"Citrlana13966.2", "CYP51G1", "MEPDNKFLNV", "RYKRRKLSVS", 486},
		"citrulus-lanatus:row-0117:Citrlana00824.53":                                         {"Citrlana00824.53", "CYP84A", "MTFHFPSMDF", "PYKRVVCSLD", 523},
		"citrulus-lanatus:row-0233:Citrlana09587.6":                                          {"Citrlana09587.6", "CYP749A", "MNSKNKEAII", "LHLILHSISN", 530},
		"prunus-persica-updated:row-0002:Prunpers1.21767":                                    {"Prunpers1.21767", "CYP51G1", "MDVDNKLFSV", "RYKRRELSFD", 486},
		"prunus-persica-updated:row-0161:Prunpers2.25113":                                    {"Prunpers2.25113", "CYP89A", "LIGSFLWLQK", "LQAHLSPRVK", 485},
		"prunus-persica-updated:row-0320:Prunpers1.28002":                                    {"Prunpers1.28002", "CYP749A", "MSWFRDPVIT", "SGLLNLLCIS", 508},
		"prunus-persica-old:block-0001:ppa004078m":                                           {"ppa004078m", "", "MEVSTALMIL", "KGVKMVAGIA", 531},
		"prunus-persica-old:block-0154:ppa015200m":                                           {"ppa015200m", "", "AFSLFLYSLL", "TPRLPAQLYE", 517},
		"prunus-persica-old:block-0306:ppa005500m":                                           {"ppa005500m", "", "MFKGQTKNLP", "LPIKIRPRKL", 457},
		"prunus-mume:row-0002:Prunmume007362126.1.236":                                       {"Prunmume007362126.1.236", "CYP51G1", "MDVDNKLFSV", "RYKRRELSFD", 486},
		"prunus-mume:row-0140:Prunmume007362363.1.141":                                       {"Prunmume007362363.1.141", "CYP89A", "METWFLILAA", "LQAHLSPRVK", 515},
		"prunus-mume:row-0283:Prunmume007362430.1.83":                                        {"Prunmume007362430.1.83", "CYP", "MINDQGHIDA", "NCHRGMSWIF", 341},
		"malus-domestica:row-0002:Maludome13.10293":                                          {"Maludome13.10293", "CYP51G", "MDMDNKLFSV", "RYKRRELSVE", 486},
		"malus-domestica:row-0167:Maludome6.19040":                                           {"Maludome6.19040", "CYP89A", "METWFLIFIA", "PLQAHVIPRI", 518},
		"malus-domestica:row-0332:Maludome13.7961":                                           {"Maludome13.7961", "CYP", "MSYASCCLFA", "CKVVKLVSIS", 422},
		"fragaria-ananassa:row-0002:Fraganan_rscf00000033.1.21":                              {"Fraganan_rscf00000033.1.21", "CYP51G1", "MDVDTKLLDG", "RYKRRELSAN", 487},
		"fragaria-ananassa:row-0104:Fraganan_rscf00001030.1.4":                               {"Fraganan_rscf00001030.1.4", "CYP92A", "MELVSTSHVL", "VEPRLPIHLY", 513},
		"fragaria-ananassa:row-0207:Fraganan_icon00003587_a.1.1rev":                          {"Fraganan_icon00003587_a.1.1rev", "CYP88A", "TRYGQKGISK", "QNPEFYQKAK", 251},
		"fragaria-vesca:row-0002:Fragvesc4.13647":                                            {"Fragvesc4.13647", "CYP71AH", "MSSYLSMSMV", "LCLAATPVYL", 506},
		"fragaria-vesca:row-0168:Fragvesc5.3625":                                             {"Fragvesc5.3625", "CYP93A", "METDFQGYII", "ARLSPFPSID", 514},
		"fragaria-vesca:row-0332:Fragvesc3.12935":                                            {"Fragvesc3.12935", "CYP", "MEAILLDKLG", "QAQKVESILC", 422},
		"cannabis-sativa:row-0002:Cannsati32097775.1":                                        {"Cannsati32097775.1", "CYP51G1", "MEVDNKMYSV", "RYKRRELSVN", 486},
		"cannabis-sativa:row-0179:Cannsati92824.3":                                           {"Cannsati92824.3", "CYP81", "MEQEQLYTAL", "IMNSIFLKAH", 490},
		"cannabis-sativa:row-0357:Cannsati114114.1":                                          {"Cannsati114114.1", "CYP", "LSWTRPPMAR", "ALPLVWTPAA", 261},
		"soybean:block-0001:CYP51G1":                                                         {"CYP51G1", "CYP51G1", "MEIDSRFLNT", "MVVGVKGKVM", 475},
		"soybean:block-0086:CYP85A13":                                                        {"CYP85A13", "CYP85A13", "MALLMTIVVG", "NGLHIRVTSY", 464},
		"soybean:block-0171:CYP736A":                                                         {"CYP736A", "CYP736A", "DHLTYRSAIA", "LISFGSGRRG", 124},
		"medicago-truncatula:piece-0001:CYP51G1":                                             {"CYP51G1", "CYP51G1", "MNVFDGNKFL", "YKRRELSVNQ", 489},
		"medicago-truncatula:piece-0188:CR321277.1":                                          {"CR321277.1", "", "IDHIFVSIQD", "DMSESFGFTV", 195},
		"medicago-truncatula:piece-0371:CYP711A12":                                           {"CYP711A12", "CYP711A12", "MVFMDLEWLF", "SVIKRTEMSC", 541},
		"lotus-japonicus:row-0002:LotujapoCM0846.228":                                        {"LotujapoCM0846.228", "CYP51G1", "MEIIDGGGNK", "YKRRVLSANQ", 491},
		"lotus-japonicus:row-0124:LotujapoCM1089.101":                                        {"LotujapoCM1089.101", "CYP83E2", "MVLPILLVLC", "NATWICSKND", 510},
		"lotus-japonicus:row-0247:LotujapoCM0241.538":                                        {"LotujapoCM0241.538", "CYP736A", "MLPPALAIPA", "IPTYRLINEG", 497},
		"cicer-arietinum:row-0002:Cicearie00316.166":                                         {"Cicearie00316.166", "CYP51G1", "MDVFDGNKFL", "YKRRQLSVDQ", 489},
		"cicer-arietinum:row-0107:Cicearie05416.5":                                           {"Cicearie05416.5", "CYP85A", "MAFIMEIVGV", "NGLHIRISSN", 464},
		"cicer-arietinum:row-0212:Cicearie00884.10":                                          {"Cicearie00884.10", "CYP", "MVEVIVISSV", "GLQIILWYWE", 417},
		"cajanus-cajanifolius:row-0002:Cajacaja03.1045":                                      {"Cajacaja03.1045", "CYP51G", "MEVDGRFLST", "YKRKELSVNP", 487},
		"cajanus-cajanifolius:row-0147:Cajacaja126288.1":                                     {"Cajacaja126288.1", "CYP82L", "VLKAERPVLK", "RLPLHLYETL", 288},
		"cajanus-cajanifolius:row-0292:Cajacaja000201.48":                                    {"Cajacaja000201.48", "CYP", "MELHWVWMSV", "LIASRSTHYI", 480},
		"jatropha-curcas:block-0001:cyp51g1":                                                 {"CYP51G1", "CYP51G1", "MAAENNFLNM", "RYKRRELSVD", 486},
		"jatropha-curcas:block-0253:jcca0270761.10":                                          {"JcCA0270761.10", "CYP82C", "MDSSLQLIAI", "PRLPAELYSC", 525},
		"jatropha-curcas:block-0517:jccb0103981.10":                                          {"JcCB0103981.10", "CYP727B23", "MSSPCKLPNK", "SEIVFVRRSS", 559},
		"ricinus-communis:block-0001:cyp51g1":                                                {"CYP51G1", "CYP51G1", "MDSDNNLMNV", "RYKRRELSVN", 486},
		"ricinus-communis:block-0132:xp_002524040.1":                                         {"XP_002524040.1", "CYP89A98", "METWFLIIVT", "PRFKRNNHNI", 516},
		"ricinus-communis:block-0263:xp_002525551.1":                                         {"XP_002525551.1", "CYP727B23", "MENSRMLLKD", "QEIVFVKRSS", 537},
		"manihot-esculenta:row-0002:Maniescu00510.82":                                        {"Maniescu00510.82", "CYP51G", "MDSDNKLLNM", "RYKRRKLSVD", 486},
		"manihot-esculenta:row-0170:Maniescu06512.602":                                       {"Maniescu06512.602", "CYP84A", "MEALLQALQP", "VPNPRLLCPL", 514},
		"manihot-esculenta:row-0337:Maniescu05162.61":                                        {"Maniescu05162.61", "CYP82C31", "MLIDVYNVCV", "TTAFYIKALV", 427},
		"linum-usitatissimum:row-0002:Linuusit731.8":                                         {"Linuusit731.8", "CYP51G", "MDFDQLKLMN", "YKRRVLVADD", 489},
		"linum-usitatissimum:row-0236:Linuusit231.525":                                       {"Linuusit231.525", "CYP85A", "MALLIILAFI", "NGMHLRIYKY", 459},
		"linum-usitatissimum:row-0469:Linuusit641.547":                                       {"Linuusit641.547", "CYP", "MKISICQLHV", "TKKIIITNVN", 269},
		"vitis-vinifera:block-0001:caap02000072.1":                                           {"CAAP02000072.1", "CYP51G6", "MDVDNKFF", "VMVRYKRRVLPVD", 486},
		"vitis-vinifera:block-0348:caap02005229":                                             {"CAAP02005229", "CYP82D10", "MYFLLQYLNITT", "VLISPRLSSCSLYN", 526},
		"vitis-vinifera:block-0702:caap02005006.1":                                           {"CAAP02005006.1", "CYP736A27", "MAVWTWTA", "HLVAIPTYRLRQ", 492},
		"nelumbo-nucifera:block-0001:maker-scaffold_5-snap-gene-84.25-mrna-1":                {"maker-scaffold_5-snap-gene-84.25-mRNA-1", "CYP51G1a", "MDLKENKF", "KGKVMVRYKRRRLSVE", 487},
		"nelumbo-nucifera:block-0183:cyp89a84p":                                              {"CYP89A84P", "CYP89A84P", "LTKKEIVS", "SHEVTEDVTIDGYLVP", 101},
		"nelumbo-nucifera:block-0366:augustus_masked-scaffold_1-processed-gene-119.5-mrna-1": {"augustus_masked-scaffold_1-processed-gene-119.5-mRNA-1", "CYP736A100P", "VVVIVSVL", "LLVIPTFRLKNNFGSF", 501},
		"aquilegia-coerulea-current:row-0002:Aquicoer8.821":                                  {"Aquicoer8.821", "CYP51G1", "MDMENTTQ", "MVRFKRRQLSID", 496},
		"aquilegia-coerulea-current:row-0552:Aquicoer116.257":                                {"Aquicoer116.257", "CYP865A", "MGGFDLVS", "PVHGAHIIFQNL", 472},
		"aquilegia-coerulea-old:block-0001:22061829":                                         {"22061829", "CYP51G1", "MDMENTTQ", "MVRFKRRQLSID", 496},
		"aquilegia-coerulea-old:block-0472:22037341":                                         {"22037341", "CYP749", "MKIEATKT", "FQIEKLKVCLLL", 112},
	}
	for key, expected := range checks {
		row, ok := byKey[key]
		if !ok {
			t.Errorf("missing representative %s", key)
			continue
		}
		if row.ID != expected.id || row.Symbol != expected.symbol || len(row.Sequence) != expected.length || !strings.HasPrefix(row.Sequence, expected.prefix) || !strings.HasSuffix(row.Sequence, expected.suffix) {
			t.Errorf("representative %s changed: id=%q symbol=%q length=%d", key, row.ID, row.Symbol, len(row.Sequence))
		}
	}
}

func TestPlantRecordAuditUsesReviewStatusNotBestHitTextForEST(t *testing.T) {
	dir := t.TempDir()
	manifest := filepath.Join(dir, "resources.csv")
	if err := os.WriteFile(manifest, []byte("category,species_label,source_url,local_file\nplants,Test species,https://example.test/x.xlsx,plants-test.xlsx\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plant_resource_profiles.csv"), []byte("local_file,review_status\nplants-test.xlsx,reviewed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "audit.csv")
	records := []record{{Category: "plants", Species: "Test species", ID: "gene1", Symbol: "CYP1A1", SourceURL: "https://example.test/x.xlsx", Description: "best hit=CYP1A1; sequence column=L", Sequence: strings.Repeat("A", 400), ReviewStatus: "source-pseudogene-label"}}
	if err := writePlantRecordAudit(manifest, out, records); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("audit rows=%d", len(rows))
	}
	header := map[string]int{}
	for i, value := range rows[0] {
		header[value] = i
	}
	est, _ := strconv.ParseBool(rows[1][header["est_hint"]])
	fragment, _ := strconv.ParseBool(rows[1][header["fragment_or_pseudogene_hint"]])
	if est || !fragment {
		t.Fatalf("est=%t fragment=%t row=%v", est, fragment, rows[1])
	}
}

func TestReviewedSourceAllowsExplicitlyUnannotatedRecord(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "unannotated.csv")
	data := "category,species,symbol,id,record_key,source_url,source_note,sequence,review_status\nplants,Prunus persica,,ppa000001m,old:block-1,https://example.test/old.doc,source explicitly unannotated," + strings.Repeat("A", 400) + ",source-unannotated\n"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	rows, err := readReviewedSources(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != "ppa000001m" || rows[0].Symbol != "" || rows[0].Sequence != strings.Repeat("A", 400) {
		t.Fatalf("rows=%#v", rows)
	}
}

func TestMergeRecordsDoesNotCrossAssignSpecies(t *testing.T) {
	rows := mergeRecords([]record{{Category: "plants", Species: "Oryza sativa", Symbol: "CYP84A6"}, {Category: "plants", Species: "Arabidopsis thaliana", Symbol: "CYP84A1"}})
	if len(rows) != 2 {
		t.Fatalf("rows=%d", len(rows))
	}
	if rows[0].Species == rows[1].Species {
		t.Fatal("species relationships collapsed")
	}
}

func TestScreenshotsAndLiteratureTablesAreNotPGDSources(t *testing.T) {
	rows, err := readReviewedSources("../../sources")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if strings.Contains(strings.ToLower(r.SourceURL), "user-provided-table") || strings.Contains(strings.ToLower(r.Description), "table s2") {
			t.Fatalf("non-resource fixture entered reviewed PGD inputs: %#v", r)
		}
	}
}

func TestParseExtractedStopsAtTerminalFASTAAndSkipsAlignment(t *testing.T) {
	dir := t.TempDir()
	manifest := filepath.Join(dir, "resources.csv")
	if err := os.WriteFile(manifest, []byte("category,species_label,source_url,local_file\nplants,Test species,https://example.test/x.doc,plants-test.doc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plant_resource_profiles.csv"), []byte("local_file,review_status\nplants-test.doc,reviewed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	extracted := filepath.Join(dir, "extracted")
	if err := os.MkdirAll(extracted, 0o755); err != nil {
		t.Fatal(err)
	}
	text := ">CYP1A1 Test\n" + strings.Repeat("M", 120) + "*\nQuery: alignment text\n" + strings.Repeat("A", 800) + "\n"
	if err := os.WriteFile(filepath.Join(extracted, "plants-test.txt"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	rows, err := parseExtracted(manifest, extracted)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || len(rows[0].Sequence) != 120 {
		t.Fatalf("rows=%#v", rows)
	}
}
