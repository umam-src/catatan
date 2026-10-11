package http

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/umam-src/catatan/internal/db"
)

// pilihPotongan mengisi Potongan pada Catatan dan Sumber terpilih dengan potongan
// yang paling relevan terhadap pertanyaan, dibatasi jumlah dan anggaran byte.
func (s *Server) pilihPotongan(r *http.Request, percakapanID, pertanyaan string, catatan *conversationNote, sumber []conversationSource) error {
	if catatan == nil && len(sumber) == 0 {
		return nil
	}
	var bukuID string
	if err := s.db.QueryRowContext(r.Context(), `SELECT notebook_id FROM conversations WHERE id = ?`, percakapanID).Scan(&bukuID); err != nil {
		return err
	}
	dokumenIDs := make([]string, 0, len(sumber)+1)
	if catatan != nil {
		dokumenIDs = append(dokumenIDs, "note:"+catatan.ID)
	}
	for _, src := range sumber {
		dokumenIDs = append(dokumenIDs, "source:"+src.ID)
	}
	hasil, err := s.db.CariPotongan(r.Context(), userID(r), bukuID, dokumenIDs, pertanyaan, batasPotonganKonteks, anggaranKonteksByte)
	if err != nil {
		return err
	}
	perDokumen := map[string][]db.HasilPotongan{}
	for _, h := range hasil {
		perDokumen[h.DokumenID] = append(perDokumen[h.DokumenID], h)
	}
	urut := func(id string) []db.HasilPotongan {
		p := perDokumen[id]
		sort.SliceStable(p, func(i, j int) bool { return p[i].Urutan < p[j].Urutan })
		return p
	}
	if catatan != nil {
		catatan.Potongan, catatan.DariPotongan = urut("note:"+catatan.ID), true
	}
	for i := range sumber {
		sumber[i].Potongan, sumber[i].DariPotongan = urut("source:"+sumber[i].ID), true
	}
	return nil
}

// tulisIsiKonteks menulis isi konteks dengan nomor baris asli. Bila dariPotongan,
// hanya potongan terpilih yang ditulis; celah di antaranya ditandai "...".
func tulisIsiKonteks(b *strings.Builder, isi string, dariPotongan bool, potongan []db.HasilPotongan) {
	if !dariPotongan {
		for n, baris := range strings.Split(isi, "\n") {
			fmt.Fprintf(b, "[L%d] %s\n", n+1, baris)
		}
		return
	}
	if len(potongan) == 0 {
		b.WriteString("(tidak ada bagian yang relevan dengan pertanyaan)\n")
		return
	}
	akhir := 0
	for i, p := range potongan {
		if i > 0 && p.BarisAwal > akhir+1 {
			b.WriteString("...\n")
		}
		for n, baris := range strings.Split(p.Isi, "\n") {
			nomor := p.BarisAwal + n
			if nomor <= akhir && p.BarisAwal != p.BarisAkhir {
				continue // baris tumpang tindih sudah ditulis
			}
			fmt.Fprintf(b, "[L%d] %s\n", nomor, baris)
		}
		if p.BarisAkhir > akhir {
			akhir = p.BarisAkhir
		}
	}
}
