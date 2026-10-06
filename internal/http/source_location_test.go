package http

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"testing"
)

func buatLokasiSumberUji(t *testing.T, handler http.Handler, sourceID string, startLine, endLine int) SourceLocation {
	t.Helper()
	res := requestUji(t, handler, http.MethodPost, "/api/sources/"+sourceID+"/locations", map[string]int{"start_line": startLine, "end_line": endLine})
	if res.Code != http.StatusCreated {
		t.Fatalf("buat lokasi: status = %d, body = %q", res.Code, res.Body.String())
	}
	var location SourceLocation
	if err := json.NewDecoder(res.Body).Decode(&location); err != nil {
		t.Fatal(err)
	}
	return location
}

func TestCreateAndListSourceLocations(t *testing.T) {
	handler := serverUji(t)
	notebook := buatBukuUjiSumber(t, handler)
	source := permintaanSumber(t, handler, notebook.ID, "bahan.txt", "Bahan", "baris satu\nbaris dua\nbaris tiga")
	if source.Code != http.StatusCreated {
		t.Fatalf("impor sumber: status = %d", source.Code)
	}
	var imported Source
	if err := json.NewDecoder(source.Body).Decode(&imported); err != nil {
		t.Fatal(err)
	}

	location := buatLokasiSumberUji(t, handler, imported.ID, 2, 3)
	if location.SourceID != imported.ID || location.StartLine != 2 || location.EndLine != 3 || location.SourceChecksum != imported.Checksum {
		t.Fatalf("lokasi = %#v", location)
	}

	res := requestUji(t, handler, http.MethodGet, "/api/sources/"+imported.ID+"/locations", nil)
	if res.Code != http.StatusOK {
		t.Fatalf("daftar lokasi: status = %d, body = %q", res.Code, res.Body.String())
	}
	var locations []SourceLocation
	if err := json.NewDecoder(res.Body).Decode(&locations); err != nil {
		t.Fatal(err)
	}
	if len(locations) != 1 || locations[0].ID != location.ID {
		t.Fatalf("daftar lokasi = %#v", locations)
	}
}

func TestSourceLocationRejectsInvalidRange(t *testing.T) {
	handler := serverUji(t)
	notebook := buatBukuUjiSumber(t, handler)
	source := imporSumberUji(t, handler, notebook.ID)

	for _, input := range []map[string]int{
		{"start_line": 0, "end_line": 1},
		{"start_line": 2, "end_line": 1},
		{"start_line": 1, "end_line": 4},
	} {
		res := requestUji(t, handler, http.MethodPost, "/api/sources/"+source.ID+"/locations", input)
		if res.Code != http.StatusBadRequest {
			t.Fatalf("rentang %#v: status = %d", input, res.Code)
		}
	}
}

func TestVerifySourceLocationRequiresCurrentSourceVersion(t *testing.T) {
	content := "baris satu\nbaris dua"
	digest := sha256.Sum256([]byte(content))
	source := Source{ID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Content: content, Checksum: hex.EncodeToString(digest[:])}
	location := SourceLocation{SourceID: source.ID, StartLine: 1, EndLine: 2, SourceChecksum: source.Checksum}
	if !verifySourceLocation(source, location) {
		t.Fatal("lokasi valid ditolak")
	}

	location.SourceChecksum = "0000000000000000000000000000000000000000000000000000000000000000"
	if verifySourceLocation(source, location) {
		t.Fatal("lokasi dengan checksum berbeda dianggap valid")
	}

	digest = sha256.Sum256([]byte("baris berubah"))
	source.Content = "baris berubah"
	source.Checksum = hex.EncodeToString(digest[:])
	if verifySourceLocation(source, location) {
		t.Fatal("lokasi versi lama dianggap valid setelah isi berubah")
	}
}

func TestListSourceLocationsRejectsUnknownSource(t *testing.T) {
	handler := serverUji(t)
	res := requestUji(t, handler, http.MethodGet, "/api/sources/00000000000000000000000000000000/locations", nil)
	if res.Code != http.StatusNotFound {
		t.Fatalf("sumber tidak ditemukan: status = %d", res.Code)
	}
}
