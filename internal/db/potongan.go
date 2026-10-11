package db

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	// batasPotonganByte adalah ukuran maksimum satu potongan (±300 token).
	batasPotonganByte = 1200
	// batasParagrafByte: potongan dapat berakhir di baris kosong setelah ukuran ini.
	batasParagrafByte = 700
	// batasTumpangTindihByte: baris terakhir potongan sebelumnya diulang bila lebih pendek dari ini.
	batasTumpangTindihByte   = 200
	batasIstilahPengambilan  = 12
	batasBatchIndeks         = 200
	batasDokumenPerTransaksi = 50
)

// BagianTeks adalah satu potongan teks beserta rentang barisnya (berbasis 1,
// dihitung dari pemisahan isi dengan "\n", sama dengan penomoran konteks chat).
type BagianTeks struct {
	BarisAwal  int
	BarisAkhir int
	Isi        string
}

// HasilPotongan adalah potongan yang terpilih untuk konteks chat.
type HasilPotongan struct {
	DokumenID  string  `json:"dokumen_id"`
	Jenis      string  `json:"jenis"`
	Judul      string  `json:"judul"`
	Urutan     int     `json:"urutan"`
	BarisAwal  int     `json:"baris_awal"`
	BarisAkhir int     `json:"baris_akhir"`
	Isi        string  `json:"isi"`
	Relevansi  float64 `json:"relevansi"`
}

// PecahTeks memecah teks menjadi potongan per heading dan paragraf dengan ukuran
// terbatas. Potongan yang terpotong karena ukuran mengulang satu baris pendek
// terakhir sebagai tumpang tindih.
func PecahTeks(isi string) []BagianTeks {
	hasil := []BagianTeks{}
	var baris []string
	awal := 0
	ukuran := 0
	punyaIsi := false

	siram := func() {
		if !punyaIsi {
			baris, awal, ukuran = nil, 0, 0
			return
		}
		a, b := 0, len(baris)-1
		for a <= b && strings.TrimSpace(baris[a]) == "" {
			a++
		}
		for b >= a && strings.TrimSpace(baris[b]) == "" {
			b--
		}
		hasil = append(hasil, BagianTeks{BarisAwal: awal + a, BarisAkhir: awal + b, Isi: strings.Join(baris[a:b+1], "\n")})
		baris, awal, ukuran, punyaIsi = nil, 0, 0, false
	}
	tambah := func(nomor int, teks string) {
		if len(baris) == 0 {
			awal = nomor
		}
		baris = append(baris, teks)
		ukuran += len(teks) + 1
		if strings.TrimSpace(teks) != "" {
			punyaIsi = true
		}
	}

	for i, teks := range strings.Split(isi, "\n") {
		nomor := i + 1
		kosong := strings.TrimSpace(teks) == ""
		if len(teks) > batasPotonganByte {
			siram()
			bagian := potongBaris(teks, batasPotonganByte)
			for _, p := range bagian {
				hasil = append(hasil, BagianTeks{BarisAwal: nomor, BarisAkhir: nomor, Isi: p})
			}
			continue
		}
		if adaHeading(teks) && punyaIsi {
			siram()
		}
		if len(baris) > 0 && ukuran+len(teks)+1 > batasPotonganByte {
			terakhir := baris[len(baris)-1]
			nomorTerakhir := awal + len(baris) - 1
			siram()
			if strings.TrimSpace(terakhir) != "" && len(terakhir) <= batasTumpangTindihByte && !adaHeading(terakhir) {
				tambah(nomorTerakhir, terakhir)
				punyaIsi = false // tumpang tindih saja belum cukup menjadi potongan
			}
		}
		tambah(nomor, teks)
		if kosong && ukuran >= batasParagrafByte {
			siram()
		}
	}
	siram()
	return hasil
}

func adaHeading(baris string) bool {
	t := strings.TrimLeft(baris, " ")
	if !strings.HasPrefix(t, "#") {
		return false
	}
	t = strings.TrimLeft(t, "#")
	return strings.HasPrefix(t, " ")
}

// potongBaris memotong satu baris panjang di batas kata bila memungkinkan dan
// selalu di batas rune.
func potongBaris(baris string, batas int) []string {
	out := []string{}
	for len(baris) > batas {
		titik := batas
		for titik > 0 && !utf8.RuneStart(baris[titik]) {
			titik--
		}
		if titik == 0 {
			titik = batas
		}
		if spasi := strings.LastIndexByte(baris[:titik], ' '); spasi > batas/2 {
			titik = spasi
		}
		if p := strings.TrimSpace(baris[:titik]); p != "" {
			out = append(out, p)
		}
		baris = strings.TrimLeft(baris[titik:], " ")
	}
	if p := strings.TrimSpace(baris); p != "" {
		out = append(out, p)
	}
	return out
}

var kataTanya = map[string]struct{}{}

func init() {
	for _, k := range strings.Fields(`yang dan di ke dari untuk pada dengan adalah ini itu apa apakah siapa kapan
		dimana mana bagaimana mengapa kenapa berapa atau saya aku kamu anda kami kita dalam akan ada tidak
		bisa dapat sebuah oleh juga sudah belum the is of an what who when how why`) {
		kataTanya[k] = struct{}{}
	}
}

// ekspresiPengambilan mengubah pertanyaan menjadi ekspresi FTS5 aman dengan
// istilah digabung OR; bm25 menempatkan potongan dengan istilah terbanyak di atas.
func ekspresiPengambilan(pertanyaan string) string {
	istilah := strings.FieldsFunc(strings.ToLower(pertanyaan), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	dipakai := map[string]struct{}{}
	hasil := make([]string, 0, batasIstilahPengambilan)
	for _, t := range istilah {
		if utf8.RuneCountInString(t) < 2 {
			continue
		}
		if _, tanya := kataTanya[t]; tanya {
			continue
		}
		if _, ada := dipakai[t]; ada {
			continue
		}
		dipakai[t] = struct{}{}
		hasil = append(hasil, `"`+t+`"`)
		if len(hasil) == batasIstilahPengambilan {
			break
		}
	}
	return strings.Join(hasil, " OR ")
}

// IndeksPotonganTertunda memecah ulang dokumen buku yang berubah sejak pemecahan terakhir.
func (d *DB) IndeksPotonganTertunda(ctx context.Context, bukuID string) error {
	if _, err := d.ExecContext(ctx, `DELETE FROM potongan_tertunda WHERE dokumen_id NOT IN (SELECT id FROM search_documents)`); err != nil {
		return err
	}
	for {
		rows, err := d.QueryContext(ctx, `SELECT pt.dokumen_id FROM potongan_tertunda pt
			JOIN search_documents sd ON sd.id = pt.dokumen_id
			WHERE sd.notebook_id = ? LIMIT ?`, bukuID, batasBatchIndeks)
		if err != nil {
			return err
		}
		ids := []string{}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			ids = append(ids, id)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
		if len(ids) == 0 {
			return nil
		}
		for len(ids) > 0 {
			n := min(len(ids), batasDokumenPerTransaksi)
			if err := d.pecahDokumen(ctx, bukuID, ids[:n]); err != nil {
				return err
			}
			ids = ids[n:]
		}
	}
}

// pecahDokumen memecah beberapa dokumen dalam satu transaksi (mengurangi biaya commit).
// Isi dibaca di dalam transaksi setelah mengambil penanda tertunda, agar suntingan
// yang terjadi di antaranya tidak menghasilkan potongan usang.
func (d *DB) pecahDokumen(ctx context.Context, bukuID string, ids []string) error {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, id := range ids {
		res, err := tx.ExecContext(ctx, `DELETE FROM potongan_tertunda WHERE dokumen_id = ?`, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			continue // sudah ditangani permintaan lain
		}
		var isi string
		err = tx.QueryRowContext(ctx, `SELECT content FROM search_documents WHERE id = ? AND notebook_id = ?`, id, bukuID).Scan(&isi)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM potongan WHERE dokumen_id = ?`, id); err != nil {
			return err
		}
		for urutan, bagian := range PecahTeks(isi) {
			if _, err := tx.ExecContext(ctx, `INSERT INTO potongan(dokumen_id, buku_id, urutan, baris_awal, baris_akhir, isi) VALUES (?, ?, ?, ?, ?, ?)`,
				id, bukuID, urutan, bagian.BarisAwal, bagian.BarisAkhir, bagian.Isi); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

// CariPotongan mengambil potongan paling relevan terhadap pertanyaan pada satu buku,
// dibatasi oleh jumlah dan anggaran byte. dokumenIDs (berbentuk "note:ID"/"source:ID")
// membatasi cakupan; bila diisi dan tidak ada potongan yang cocok, potongan awal
// dokumen dipakai agar permintaan seperti "ringkas catatan ini" tetap berfungsi.
func (d *DB) CariPotongan(ctx context.Context, pemilikID, bukuID string, dokumenIDs []string, pertanyaan string, batasJumlah, anggaranByte int) ([]HasilPotongan, error) {
	if pemilikID == "" || bukuID == "" {
		return nil, errors.New("pemilik dan buku pengambilan wajib diisi")
	}
	if batasJumlah <= 0 || anggaranByte <= 0 {
		return []HasilPotongan{}, nil
	}
	if err := d.IndeksPotonganTertunda(ctx, bukuID); err != nil {
		return nil, err
	}
	hasil := []HasilPotongan{}
	if ekspresi := ekspresiPengambilan(pertanyaan); ekspresi != "" {
		args := []any{ekspresi, pemilikID, bukuID}
		saring := ""
		if len(dokumenIDs) > 0 {
			saring = " AND p.dokumen_id IN (?" + strings.Repeat(",?", len(dokumenIDs)-1) + ")"
			for _, id := range dokumenIDs {
				args = append(args, id)
			}
		}
		args = append(args, batasJumlah)
		rows, err := d.QueryContext(ctx, `SELECT p.dokumen_id, sd.kind, sd.title, p.urutan, p.baris_awal, p.baris_akhir, p.isi,
			bm25(indeks_potongan) AS relevansi
			FROM indeks_potongan
			JOIN potongan p ON p.id = indeks_potongan.rowid
			JOIN search_documents sd ON sd.id = p.dokumen_id
			JOIN notebooks n ON n.id = p.buku_id
			WHERE indeks_potongan MATCH ? AND n.owner_id = ? AND p.buku_id = ?`+saring+`
			ORDER BY relevansi ASC, p.id ASC
			LIMIT ?`, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		total := 0
		for rows.Next() {
			var h HasilPotongan
			if err := rows.Scan(&h.DokumenID, &h.Jenis, &h.Judul, &h.Urutan, &h.BarisAwal, &h.BarisAkhir, &h.Isi, &h.Relevansi); err != nil {
				return nil, err
			}
			if len(hasil) > 0 && total+len(h.Isi) > anggaranByte {
				break
			}
			total += len(h.Isi)
			hasil = append(hasil, h)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}
	if len(hasil) > 0 || len(dokumenIDs) == 0 {
		return hasil, nil
	}
	return d.potonganAwal(ctx, pemilikID, bukuID, dokumenIDs, batasJumlah, anggaranByte)
}

// potonganAwal mengambil potongan pertama tiap dokumen, bergiliran antar dokumen.
func (d *DB) potonganAwal(ctx context.Context, pemilikID, bukuID string, dokumenIDs []string, batasJumlah, anggaranByte int) ([]HasilPotongan, error) {
	hasil := []HasilPotongan{}
	total := 0
	for urutan := 0; len(hasil) < batasJumlah; urutan++ {
		ditambah := false
		for _, id := range dokumenIDs {
			var h HasilPotongan
			err := d.QueryRowContext(ctx, `SELECT p.dokumen_id, sd.kind, sd.title, p.urutan, p.baris_awal, p.baris_akhir, p.isi
				FROM potongan p
				JOIN search_documents sd ON sd.id = p.dokumen_id
				JOIN notebooks n ON n.id = p.buku_id
				WHERE p.dokumen_id = ? AND p.buku_id = ? AND n.owner_id = ? AND p.urutan = ?`, id, bukuID, pemilikID, urutan).
				Scan(&h.DokumenID, &h.Jenis, &h.Judul, &h.Urutan, &h.BarisAwal, &h.BarisAkhir, &h.Isi)
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			if err != nil {
				return nil, err
			}
			if len(hasil) > 0 && total+len(h.Isi) > anggaranByte {
				return hasil, nil
			}
			total += len(h.Isi)
			hasil = append(hasil, h)
			ditambah = true
			if len(hasil) >= batasJumlah {
				return hasil, nil
			}
		}
		if !ditambah {
			break
		}
	}
	return hasil, nil
}
