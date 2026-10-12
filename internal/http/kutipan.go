package http

import (
	"context"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/umam-src/catatan/internal/ai"
	"github.com/umam-src/catatan/internal/db"
)

const jawabanTidakDitemukan = "Informasi tidak ditemukan dalam konteks."

var (
	penandaBarisPattern      = regexp.MustCompile(`\[L[0-9]+(?:\s*[-–—]\s*L?[0-9]+)?\]`)
	penandaKutipanPattern    = regexp.MustCompile(`\[[ \t]*[SN][0-9]+[^\]\n]*\]`)
	spasiGandaPattern        = regexp.MustCompile(`[ \t]{2,}`)
	spasiSebelumTandaPattern = regexp.MustCompile(`[ \t]+([.,;:!?])`)
)

// susunKutipan membuat kutipan dari potongan yang cocok dengan pertanyaan dan
// benar-benar dikirim ke model. Rentang baris yang bersambung digabung. Model
// tidak menulis kutipan, sehingga format jawaban model tidak memengaruhi hasil.
func susunKutipan(catatan []conversationNote, sumber []conversationSource) []Citation {
	hasil := []Citation{}
	for i, n := range catatan {
		hasil = append(hasil, kutipanDariPotongan("note", "N"+strconv.Itoa(i+1), n.ID, n.Potongan)...)
	}
	for i, src := range sumber {
		hasil = append(hasil, kutipanDariPotongan("source", "S"+strconv.Itoa(i+1), src.ID, src.Potongan)...)
	}
	return hasil
}

func kutipanDariPotongan(jenis, ref, id string, potongan []db.HasilPotongan) []Citation {
	urut := make([]db.HasilPotongan, 0, len(potongan))
	for _, p := range potongan {
		if p.Cocok {
			urut = append(urut, p)
		}
	}
	sort.SliceStable(urut, func(i, j int) bool { return urut[i].BarisAwal < urut[j].BarisAwal })
	hasil := []Citation{}
	for _, p := range urut {
		if n := len(hasil); n > 0 && p.BarisAwal <= hasil[n-1].EndLine+1 {
			if p.BarisAkhir > hasil[n-1].EndLine {
				hasil[n-1].EndLine = p.BarisAkhir
			}
			continue
		}
		hasil = append(hasil, Citation{Kind: jenis, SourceID: id, SourceRef: ref, StartLine: p.BarisAwal, EndLine: p.BarisAkhir})
	}
	return hasil
}

// tanpaBahan benar bila konteks wajib ada (cakupan Catatan, atau Sumber dipilih)
// tetapi tidak ada satu pun potongan yang dapat dikirim, sehingga model tidak
// perlu dipanggil.
func tanpaBahan(wajib bool, catatan []conversationNote, sumber []conversationSource) bool {
	if !wajib {
		return false
	}
	for _, n := range catatan {
		if len(n.Potongan) > 0 {
			return false
		}
	}
	for _, src := range sumber {
		if len(src.Potongan) > 0 {
			return false
		}
	}
	return true
}

func generateJawaban(ctx context.Context, provider ai.Provider, model string, pesan []ai.Message, kosong bool) (ai.Response, error) {
	if kosong {
		return ai.Response{Text: jawabanTidakDitemukan, Model: model}, nil
	}
	return provider.Generate(ctx, ai.Request{Model: model, Messages: pesan})
}

// bersihkanJawaban membuang penanda lama ([TIDAK_DITEMUKAN]), kutipan, dan
// nomor baris yang mungkin masih ditulis model. Jawaban tidak pernah ditolak
// karena format; jawaban kosong diganti pernyataan tidak ditemukan.
func bersihkanJawaban(teks string) string {
	teks = strings.ReplaceAll(teks, noInfoMarker, "")
	teks = penandaBarisPattern.ReplaceAllString(teks, "")
	teks = penandaKutipanPattern.ReplaceAllString(teks, "")
	teks = spasiGandaPattern.ReplaceAllString(teks, " ")
	teks = spasiSebelumTandaPattern.ReplaceAllString(teks, "$1")
	teks = strings.TrimSpace(strings.ReplaceAll(teks, " \n", "\n"))
	if teks == "" {
		return jawabanTidakDitemukan
	}
	return teks
}
