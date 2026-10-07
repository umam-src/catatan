<script lang="ts">
  type User = { id: string; username: string; display_name: string };
  type Notebook = { id: string; title: string; description: string };
  type Note = { id: string; title: string; content: string; updated_at: string };
  type Source = { id: string; title: string; kind: string; content?: string; locator: string; checksum: string; metadata_json: string };
  type SearchResult = { id: string; kind: 'note' | 'source'; notebook_id: string; title: string; relevance: number };

  type Panel = 'sumber' | 'artefak';
  type View = 'catatan' | 'chat' | 'pratinjau';

  let user: User | null = null;
  let siap = false;
  let memuat = true;
  let versi = '';
  let galat = '';

  let modeMasuk: 'masuk' | 'siapkan' = 'masuk';
  let namaPengguna = '';
  let kataSandi = '';
  let konfirmasiKataSandi = '';
  let email = '';
  let namaTampilan = '';
  let menuPenggunaTerbuka = false;
  let dialogPengaturanTerbuka = false;
  let pendaftaranDiizinkan = true;
  let galatPengaturan = '';
  const kunciPendaftaran = 'catatan.pendaftaranDiizinkan';
  const kunciStatusDrawer = 'catatan.bukuTerbuka';
  const kunciStatusDrawerKonteks = 'catatan.konteksTerbuka';

  let notebooks: Notebook[] = [];
  let notebookID = '';
  let notes: Note[] = [];
  let catatanAktif: Note | null = null;

  let sources: Source[] = [];
  let sumberAktif: Source | null = null;
  let panel: Panel = 'sumber';
  let view: View = 'catatan';
  let bukuTerbuka = true;
  let konteksTerbuka = true;
  let menuKonteksAktif = '';
  let kueriPencarian = '';
  let hasilPencarian: SearchResult[] = [];
  let pencarianMemuat = false;
  let galatPencarian = '';
  let pencarianAktif = false;
  let timerPencarian: ReturnType<typeof setTimeout> | undefined;
  let pengendaliPencarian: AbortController | undefined;

  let statusSimpan: 'tersimpan' | 'menyimpan' | 'gagal' = 'tersimpan';
  let timerSimpan: ReturnType<typeof setTimeout> | undefined;
  let nomorSimpan = 0;
  let imporInput: HTMLInputElement;

  // Ponsel menampilkan satu panel pada satu waktu: daftar catatan atau editor.
  let tampilanMobile: 'daftar' | 'editor' = 'daftar';

  type DialogAksi = {
    jenis: 'teks' | 'konfirmasi';
    judul: string;
    pesan: string;
    label: string;
    bahaya: boolean;
    nilai: string;
    selesai: (hasil: string | boolean | null) => void;
  };
  let dialogAksi: DialogAksi | null = null;

  async function permintaan(path: string, init: RequestInit = {}) {
    const headers = new Headers(init.headers);
    if (init.body && !(init.body instanceof FormData)) headers.set('Content-Type', 'application/json');
    const response = await fetch(path, { ...init, headers });
    if (!response.ok) throw new Error((await response.text()) || 'Permintaan gagal.');
    return response;
  }

  function layarKecil() {
    return typeof window !== 'undefined' && window.matchMedia('(max-width: 760px)').matches;
  }

  function kembaliKeDaftar() {
    tampilanMobile = 'daftar';
    if (view === 'pratinjau') {
      view = 'catatan';
      sumberAktif = null;
    }
  }

  function tanyaTeks(judul: string, nilaiAwal = '', label = 'Simpan'): Promise<string | null> {
    return new Promise((selesai) => {
      dialogAksi = {
        jenis: 'teks',
        judul,
        pesan: '',
        label,
        bahaya: false,
        nilai: nilaiAwal,
        selesai: (hasil) => selesai(typeof hasil === 'string' ? hasil : null)
      };
    });
  }

  function konfirmasi(judul: string, pesan: string, label: string): Promise<boolean> {
    return new Promise((selesai) => {
      dialogAksi = { jenis: 'konfirmasi', judul, pesan, label, bahaya: true, nilai: '', selesai: (hasil) => selesai(hasil === true) };
    });
  }

  function tutupDialogAksi(hasil: string | boolean | null) {
    const dialog = dialogAksi;
    dialogAksi = null;
    dialog?.selesai(hasil);
  }

  function kirimDialogAksi(event: SubmitEvent) {
    event.preventDefault();
    if (!dialogAksi) return;
    tutupDialogAksi(dialogAksi.jenis === 'teks' ? dialogAksi.nilai : true);
  }

  // Membuka <dialog> sebagai modal (fokus terkunci, latar tidak aktif) dan memfokuskan elemen bertanda data-fokus.
  function bukaModal(node: HTMLDialogElement) {
    if (!node.open) node.showModal();
    const sasaran = node.querySelector<HTMLElement>('[data-fokus]');
    if (!sasaran) return;
    sasaran.focus();
    if (sasaran instanceof HTMLInputElement) sasaran.select();
  }

  function sinkronkanMenuKonteks(id: string, event: Event) {
    const details = event.currentTarget;
    if (!(details instanceof HTMLDetailsElement)) return;
    const popup = details.querySelector<HTMLDivElement>('.menu-konteks-daftar');
    if (details.open) {
      menuKonteksAktif = id;
      if (popup && !popup.matches(':popover-open')) {
        document.querySelectorAll<HTMLDivElement>('.menu-konteks-daftar:popover-open').forEach((menu) => menu.hidePopover());
        popup.showPopover();
        const summary = details.querySelector('summary');
        if (!summary) return;
        const rect = summary.getBoundingClientRect();
        const margin = 8;
        const width = popup.offsetWidth;
        const height = popup.offsetHeight;
        const left = Math.max(margin, Math.min(rect.right - width, window.innerWidth - width - margin));
        const spaceBelow = window.innerHeight - rect.bottom;
        const preferredTop = spaceBelow >= height + margin
          ? rect.bottom + 4
          : rect.top - height - 4;
        const top = Math.max(margin, Math.min(preferredTop, window.innerHeight - height - margin));
        popup.style.left = `${left}px`;
        popup.style.top = `${top}px`;
      }
      return;
    }
    if (popup?.matches(':popover-open')) popup.hidePopover();
    if (menuKonteksAktif === id) menuKonteksAktif = '';
  }

  function tutupMenuKonteksDiLuar(event: MouseEvent) {
    const target = event.target;
    if (!(target instanceof Element)) return;
    if (target.closest('.menu-konteks')) return;
    menuKonteksAktif = '';
    if (!target.closest('.menu-pengguna')) menuPenggunaTerbuka = false;
  }

  function muatPengaturanPendaftaran() {
    try {
      const nilai = localStorage.getItem(kunciPendaftaran);
      pendaftaranDiizinkan = nilai === null || nilai === 'true';
      galatPengaturan = '';
    } catch {
      galatPengaturan = 'Pengaturan pendaftaran tidak dapat dibaca dari perangkat ini.';
    }
  }

  function bukaPengaturan() {
    menuPenggunaTerbuka = false;
    muatPengaturanPendaftaran();
    dialogPengaturanTerbuka = true;
  }

  function ubahStatusDrawer(terbuka: boolean) {
    bukuTerbuka = terbuka;
    try {
      window.localStorage.setItem(kunciStatusDrawer, String(terbuka));
    } catch {
      galat = 'Status daftar buku tidak dapat disimpan pada perangkat ini.';
    }
  }

  function ubahStatusDrawerKonteks(terbuka: boolean) {
    konteksTerbuka = terbuka;
    try {
      window.localStorage.setItem(kunciStatusDrawerKonteks, String(terbuka));
    } catch {
      galat = 'Status drawer Artefak tidak dapat disimpan pada perangkat ini.';
    }
  }

  function ubahPengaturanPendaftaran(event: Event) {
    const input = event.currentTarget;
    if (!(input instanceof HTMLInputElement)) return;
    try {
      localStorage.setItem(kunciPendaftaran, String(input.checked));
      pendaftaranDiizinkan = input.checked;
      galatPengaturan = '';
    } catch {
      input.checked = pendaftaranDiizinkan;
      galatPengaturan = 'Pengaturan tidak dapat disimpan pada perangkat ini.';
    }
  }

  function tanganiTombolEscape(event: KeyboardEvent) {
    if (event.key !== 'Escape') return;
    menuPenggunaTerbuka = false;
    dialogPengaturanTerbuka = false;
  }

  async function mulai() {
    memuat = true;
    try {
      if (typeof window !== 'undefined') {
        try {
          const nilai = window.localStorage.getItem(kunciPendaftaran);
          pendaftaranDiizinkan = nilai === null || nilai === 'true';
        } catch {
          galatPengaturan = 'Pengaturan pendaftaran tidak dapat dibaca dari perangkat ini.';
        }
        try {
          const nilaiDrawer = window.localStorage.getItem(kunciStatusDrawer);
          if (nilaiDrawer === 'true' || nilaiDrawer === 'false') bukuTerbuka = nilaiDrawer === 'true';
          else if (nilaiDrawer === null && window.matchMedia('(max-width: 760px)').matches) bukuTerbuka = false;
        } catch {
          galat = 'Status daftar buku tidak dapat dibaca pada perangkat ini.';
        }
        try {
          const nilaiDrawerKonteks = window.localStorage.getItem(kunciStatusDrawerKonteks);
          if (nilaiDrawerKonteks === 'true' || nilaiDrawerKonteks === 'false') {
            konteksTerbuka = nilaiDrawerKonteks === 'true';
          } else if (nilaiDrawerKonteks === null && window.matchMedia('(max-width: 760px)').matches) {
            konteksTerbuka = false;
          }
        } catch {
          galat = 'Status drawer Artefak tidak dapat dibaca pada perangkat ini.';
        }
        if (layarKecil()) {
          bukuTerbuka = false;
          konteksTerbuka = false;
        }
      }
      const versiResponse = await fetch('/api/version');
      if (versiResponse.ok) versi = (await versiResponse.json()).version ?? '';

      const response = await fetch('/api/auth/me');
      if (!response.ok) return;
      user = (await response.json()).user;
      siap = true;
      await muatBuku();
    } catch {
      galat = 'Catatan tidak dapat dihubungi. Pastikan aplikasi lokal sedang berjalan.';
    } finally {
      memuat = false;
    }
  }

  async function kirimAutentikasi(event: SubmitEvent) {
    event.preventDefault();
    galat = '';
    if (modeMasuk === 'siapkan' && kataSandi !== konfirmasiKataSandi) {
      galat = 'Konfirmasi kata sandi tidak sama.';
      return;
    }
    try {
      const endpoint = modeMasuk === 'masuk' ? '/api/auth/login' : '/api/auth/setup';
      const body = modeMasuk === 'masuk'
        ? { username: namaPengguna.trim(), password: kataSandi }
        : { username: namaPengguna.trim(), email: email.trim(), display_name: namaTampilan.trim(), password: kataSandi };

      const response = await permintaan(endpoint, { method: 'POST', body: JSON.stringify(body) });
      user = (await response.json()).user;
      kataSandi = '';
      konfirmasiKataSandi = '';
      siap = true;
      await muatBuku();
    } catch (error) {
      galat = error instanceof Error ? error.message : 'Autentikasi gagal.';
    }
  }

  async function keluar() {
    try {
      const response = await fetch('/api/auth/logout', { method: 'POST' });
      if (!response.ok) throw new Error((await response.text()) || 'Gagal keluar.');
      menuPenggunaTerbuka = false;
      user = null;
      siap = false;
      bersihkanPencarian();
      notebooks = [];
      notes = [];
      sources = [];
      catatanAktif = null;
      sumberAktif = null;
      notebookID = '';
      menuKonteksAktif = '';
      tampilanMobile = 'daftar';
      modeMasuk = 'masuk';
      kataSandi = '';
      konfirmasiKataSandi = '';
    } catch (error) {
      galat = error instanceof Error ? error.message : 'Gagal keluar.';
    }
  }

  function bersihkanPencarian() {
    if (timerPencarian) clearTimeout(timerPencarian);
    pengendaliPencarian?.abort();
    timerPencarian = undefined;
    pengendaliPencarian = undefined;
    kueriPencarian = '';
    hasilPencarian = [];
    pencarianMemuat = false;
    galatPencarian = '';
    pencarianAktif = false;
  }

  function jadwalkanPencarian() {
    pencarianAktif = true;
    galatPencarian = '';
    if (timerPencarian) clearTimeout(timerPencarian);
    pengendaliPencarian?.abort();
    if (!kueriPencarian.trim()) {
      hasilPencarian = [];
      pencarianMemuat = false;
      pencarianAktif = false;
      return;
    }
    pencarianMemuat = true;
    timerPencarian = setTimeout(() => void cariLokal(), 250);
  }

  async function cariLokal() {
    const kueri = kueriPencarian.trim();
    if (!kueri) return;
    const pengendali = new AbortController();
    pengendaliPencarian = pengendali;
    try {
      const parameter = new URLSearchParams({ q: kueri, limit: '20' });
      if (notebookID) parameter.set('notebook_id', notebookID);
      const response = await fetch('/api/search?' + parameter.toString(), { signal: pengendali.signal });
      if (!response.ok) throw new Error((await response.text()) || 'Pencarian gagal.');
      const data = await response.json();
      if (pengendaliPencarian !== pengendali) return;
      hasilPencarian = Array.isArray(data.results) ? data.results : [];
      galatPencarian = '';
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') return;
      if (pengendaliPencarian !== pengendali) return;
      hasilPencarian = [];
      galatPencarian = error instanceof Error ? error.message : 'Pencarian gagal.';
    } finally {
      if (pengendaliPencarian === pengendali) {
        pencarianMemuat = false;
        pengendaliPencarian = undefined;
      }
    }
  }

  function pilihHasilPencarian(hasil: SearchResult) {
    if (hasil.notebook_id !== notebookID) return;
    pencarianAktif = false;
    if (hasil.kind === 'note') {
      const note = notes.find((item) => item.id === hasil.id);
      if (note) pilihCatatan(note);
      return;
    }
    const source = sources.find((item) => item.id === hasil.id);
    if (source) void pilihSumber(source);
  }

  async function muatBuku() {
    const response = await permintaan('/api/notebooks');
    notebooks = await response.json();
    if (!notebookID || !notebooks.some((item) => item.id === notebookID)) {
      notebookID = notebooks[0]?.id ?? '';
    }
    await muatIsiBuku();
  }

  async function muatIsiBuku() {
    menuKonteksAktif = '';
    if (!notebookID) {
      notes = [];
      sources = [];
      catatanAktif = null;
      sumberAktif = null;
      view = 'catatan';
      tampilanMobile = 'daftar';
      return;
    }

    const [notesResponse, sourcesResponse] = await Promise.all([
      permintaan('/api/notebooks/' + notebookID + '/notes'),
      permintaan('/api/notebooks/' + notebookID + '/sources')
    ]);

    notes = await notesResponse.json();
    sources = await sourcesResponse.json();
    catatanAktif = notes[0] ? { ...notes[0] } : null;
    sumberAktif = null;
    view = 'catatan';
    tampilanMobile = 'daftar';
    statusSimpan = 'tersimpan';
  }

  async function pilihBuku(id: string) {
    if (id === notebookID) return;
    if (timerSimpan) clearTimeout(timerSimpan);
    nomorSimpan++;
    menuKonteksAktif = '';
    notebookID = id;
    bersihkanPencarian();
    galat = '';
    try {
      await muatIsiBuku();
    } catch (error) {
      galat = error instanceof Error ? error.message : 'Gagal memuat buku.';
    }
  }

  async function buatBuku() {
    const title = await tanyaTeks('Buku baru', '', 'Buat buku');
    if (!title?.trim()) return;
    try {
      const response = await permintaan('/api/notebooks', {
        method: 'POST',
        body: JSON.stringify({ title: title.trim(), description: '' })
      });
      const notebook: Notebook = await response.json();
      notebooks = [notebook, ...notebooks];
      notebookID = notebook.id;
      await muatIsiBuku();
    } catch (error) {
      galat = error instanceof Error ? error.message : 'Gagal membuat buku.';
    }
  }

  async function ubahBuku(notebook: Notebook) {
    menuKonteksAktif = '';
    const title = await tanyaTeks('Ubah nama buku', notebook.title);
    if (!title?.trim() || title.trim() === notebook.title) return;
    try {
      const response = await permintaan('/api/notebooks/' + notebook.id, {
        method: 'PUT',
        body: JSON.stringify({ title: title.trim(), description: notebook.description })
      });
      const updated: Notebook = await response.json();
      notebooks = notebooks.map((item) => item.id === updated.id ? { ...item, ...updated } : item);
      galat = '';
    } catch (error) {
      galat = error instanceof Error ? error.message : 'Gagal mengubah buku.';
    }
  }

  async function hapusBuku(notebook: Notebook) {
    menuKonteksAktif = '';
    if (!(await konfirmasi(`Hapus buku “${notebook.title}”?`, 'Semua catatan, sumber, dan isi terkait di dalamnya juga akan dihapus.', 'Hapus buku'))) return;
    if (timerSimpan) clearTimeout(timerSimpan);
    nomorSimpan++;
    try {
      await permintaan('/api/notebooks/' + notebook.id, { method: 'DELETE' });
      const sisa = notebooks.filter((item) => item.id !== notebook.id);
      notebooks = sisa;
      notebookID = sisa[0]?.id ?? '';
      await muatIsiBuku();
      galat = '';
    } catch (error) {
      galat = error instanceof Error ? error.message : 'Gagal menghapus buku.';
    }
  }

  async function buatCatatan() {
    if (!notebookID) return;
    menuKonteksAktif = '';
    try {
      const response = await permintaan('/api/notebooks/' + notebookID + '/notes', {
        method: 'POST',
        body: JSON.stringify({ title: 'Catatan baru', content: ' ' })
      });
      const note: Note = await response.json();
      notes = [note, ...notes];
      catatanAktif = { ...note };
      view = 'catatan';
      tampilanMobile = 'editor';
      statusSimpan = 'tersimpan';
    } catch (error) {
      galat = error instanceof Error ? error.message : 'Gagal membuat catatan.';
    }
  }

  function pilihCatatan(note: Note) {
    if (timerSimpan) clearTimeout(timerSimpan);
    nomorSimpan++;
    menuKonteksAktif = '';
    catatanAktif = { ...note };
    view = 'catatan';
    tampilanMobile = 'editor';
    statusSimpan = 'tersimpan';
    galat = '';
  }

  async function ubahCatatan(note: Note) {
    menuKonteksAktif = '';
    const title = await tanyaTeks('Ubah judul catatan', note.title);
    if (!title?.trim() || title.trim() === note.title) return;
    try {
      const response = await permintaan('/api/notes/' + note.id, {
        method: 'PUT',
        body: JSON.stringify({ title: title.trim(), content: note.content.trim() || ' ' })
      });
      const data = await response.json();
      notes = notes.map((item) => item.id === note.id ? { ...item, title: title.trim(), updated_at: data.updated_at } : item);
      if (catatanAktif?.id === note.id) catatanAktif = { ...catatanAktif, title: title.trim(), updated_at: data.updated_at };
      galat = '';
    } catch (error) {
      galat = error instanceof Error ? error.message : 'Gagal mengubah catatan.';
    }
  }

  function jadwalkanSimpan() {
    if (!catatanAktif) return;
    statusSimpan = 'menyimpan';
    if (timerSimpan) clearTimeout(timerSimpan);
    const nomor = ++nomorSimpan;
    timerSimpan = setTimeout(() => void simpanCatatan(nomor), 650);
  }

  async function simpanCatatan(nomor: number) {
    if (!catatanAktif || nomor !== nomorSimpan) return;
    const note = { ...catatanAktif };
    try {
      const response = await permintaan('/api/notes/' + note.id, {
        method: 'PUT',
        body: JSON.stringify({ title: note.title, content: note.content.trim() || ' ' })
      });
      if (nomor !== nomorSimpan || catatanAktif?.id !== note.id) return;
      const data = await response.json();
      catatanAktif = { ...note, content: note.content.trim() || ' ', updated_at: data.updated_at };
      notes = notes.map((item) => item.id === note.id ? { ...item, ...catatanAktif } : item);
      statusSimpan = 'tersimpan';
      galat = '';
    } catch (error) {
      if (nomor !== nomorSimpan || catatanAktif?.id !== note.id) return;
      statusSimpan = 'gagal';
      galat = error instanceof Error ? error.message : 'Gagal menyimpan catatan.';
    }
  }

  async function hapusCatatan(note: Note) {
    menuKonteksAktif = '';
    if (!(await konfirmasi(`Hapus catatan “${note.title || 'Tanpa judul'}”?`, 'Catatan ini akan dihapus dari buku.', 'Hapus catatan'))) return;
    const id = note.id;
    nomorSimpan++;
    if (timerSimpan) clearTimeout(timerSimpan);
    try {
      await permintaan('/api/notes/' + id, { method: 'DELETE' });
      notes = notes.filter((item) => item.id !== id);
      catatanAktif = notes[0] ? { ...notes[0] } : null;
      tampilanMobile = 'daftar';
      statusSimpan = 'tersimpan';
      galat = '';
    } catch (error) {
      galat = error instanceof Error ? error.message : 'Gagal menghapus catatan.';
    }
  }

  async function pilihSumber(source: Source) {
    galat = '';
    panel = 'sumber';
    try {
      const response = await permintaan('/api/sources/' + source.id);
      sumberAktif = await response.json();
      view = 'pratinjau';
      tampilanMobile = 'editor';
      if (layarKecil()) ubahStatusDrawerKonteks(false);
    } catch (error) {
      galat = error instanceof Error ? error.message : 'Gagal membuka sumber.';
    }
  }

  async function hapusSumber(source: Source) {
    if (!(await konfirmasi(`Hapus sumber “${source.title}”?`, 'Sumber ini akan dihapus dari buku.', 'Hapus sumber'))) return;
    try {
      await permintaan('/api/sources/' + source.id, { method: 'DELETE' });
      sources = sources.filter((item) => item.id !== source.id);
      if (sumberAktif?.id === source.id) sumberAktif = null;
      view = 'catatan';
    } catch (error) {
      galat = error instanceof Error ? error.message : 'Gagal menghapus sumber.';
    }
  }

  function bukaImpor() {
    imporInput?.click();
  }

  async function imporSumber(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file || !notebookID) return;

    galat = '';
    const form = new FormData();
    form.append('file', file);
    form.append('title', file.name);

    try {
      const response = await permintaan('/api/notebooks/' + notebookID + '/sources', {
        method: 'POST',
        body: form
      });
      const source: Source = await response.json();
      sources = [source, ...sources];
      await pilihSumber(source);
    } catch (error) {
      galat = error instanceof Error ? error.message : 'Gagal mengimpor sumber.';
    } finally {
      input.value = '';
    }
  }

  $: bukuAktif = notebooks.find((item) => item.id === notebookID);
  $: inisial = (user?.display_name || user?.username || 'C').slice(0, 1).toUpperCase();

  mulai();
</script>

<svelte:window onkeydown={tanganiTombolEscape} />

<svelte:head>
  <title>{versi ? `Catatan · ${versi}` : 'Catatan'}</title>
  <meta name="description" content="Catatan lokal yang sederhana dan tetap berada di perangkat." />
</svelte:head>

{#if memuat}
  <main class="layar-status">
    <div class="status-muat">
      <span class="logo-mark">
        <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 6.3C9.8 5 7.2 4.7 4.5 5.4v13c2.7-.7 5.3-.4 7.5.9m0-13c2.2-1.3 4.8-1.6 7.5-.9v13c-2.7-.7-5.3-.4-7.5.9m0-13v13" />
        </svg>
      </span>
      <span>Memuat…</span>
    </div>
  </main>
{:else if !siap}
  <main class="layar-autentikasi">
    <section class="autentikasi">
      <div class="autentikasi-mark">
        <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 6.3C9.8 5 7.2 4.7 4.5 5.4v13c2.7-.7 5.3-.4 7.5.9m0-13c2.2-1.3 4.8-1.6 7.5-.9v13c-2.7-.7-5.3-.4-7.5.9m0-13v13" />
        </svg>
      </div>
      <h1>{modeMasuk === 'masuk' ? 'Selamat datang kembali.' : 'Mulai di perangkat ini.'}</h1>
      <p class="pengantar">
        {modeMasuk === 'masuk'
          ? 'Masuk untuk melanjutkan catatan yang tersimpan secara lokal.'
          : 'Buat akun lokal. Data Anda tetap berada di perangkat ini.'}
      </p>

      {#if galat}
        <div class="pesan galat" role="alert">{galat}</div>
      {/if}

      <form onsubmit={kirimAutentikasi}>
        <label>
          Nama pengguna
          <input bind:value={namaPengguna} autocomplete="username" required />
        </label>

        {#if modeMasuk === 'siapkan'}
          <label>
            Nama tampilan <span class="opsional">opsional</span>
            <input bind:value={namaTampilan} autocomplete="name" />
          </label>
          <label>
            Email <span class="opsional">opsional</span>
            <input bind:value={email} type="email" autocomplete="email" />
          </label>
        {/if}

        <label>
          Kata sandi
          <input
            bind:value={kataSandi}
            type="password"
            minlength="12"
            autocomplete={modeMasuk === 'siapkan' ? 'new-password' : 'current-password'}
            required
          />
        </label>

        {#if modeMasuk === 'siapkan'}
          <label>
            Konfirmasi kata sandi
            <input
              bind:value={konfirmasiKataSandi}
              type="password"
              minlength="12"
              autocomplete="new-password"
              required
            />
          </label>
        {/if}

        <button class="tombol utama lebar" type="submit">
          {modeMasuk === 'masuk' ? 'Masuk' : 'Buat akun'}
        </button>
      </form>

      {#if modeMasuk === 'siapkan'}
        <button class="tautan" type="button" onclick={() => {
          modeMasuk = 'masuk';
          kataSandi = '';
          konfirmasiKataSandi = '';
          galat = '';
        }}>Sudah punya akun? Masuk</button>
      {:else if pendaftaranDiizinkan}
        <button class="tautan" type="button" onclick={() => {
          modeMasuk = 'siapkan';
          kataSandi = '';
          konfirmasiKataSandi = '';
          galat = '';
        }}>Perangkat baru? Buat akun lokal</button>
      {/if}

      {#if versi}<p class="versi">Versi {versi}</p>{/if}
    </section>
  </main>
{:else}
  <div class:tanpa-buku={!bukuTerbuka} class="aplikasi" onclick={tutupMenuKonteksDiLuar}>
    <aside class:tersembunyi={!bukuTerbuka} class="panel-navigasi">
      <button
        class="tombol-lipat-buku"
        type="button"
        aria-label="Sembunyikan daftar buku"
        title="Sembunyikan daftar buku"
        aria-expanded={bukuTerbuka}
        aria-controls="daftar-buku"
        onclick={() => ubahStatusDrawer(false)}
      >
        <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M15 6l-6 6 6 6" /></svg>
      </button>
      <div class="merek">
        <span class="logo-mark kecil">
          <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 6.3C9.8 5 7.2 4.7 4.5 5.4v13c2.7-.7 5.3-.4 7.5.9m0-13c2.2-1.3 4.8-1.6 7.5-.9v13c-2.7-.7-5.3-.4-7.5.9m0-13v13" />
          </svg>
        </span>
        <div>
          <strong>Catatan</strong>
          <span>Ruang kerja pribadi</span>
        </div>
      </div>

      <button class="tombol sekunder lebar" type="button" onclick={buatBuku}>+ Buku baru</button>

      <div class="judul-panel">Buku</div>
      <nav id="daftar-buku" class="daftar-buku" aria-label="Buku">
        {#each notebooks as notebook}
          <div
            class:aktif={notebook.id === notebookID}
            class:menu-terbuka={menuKonteksAktif === `buku:${notebook.id}`}
            class="item-buku-bar"
          >
            <button class="item-buku" type="button" onclick={() => pilihBuku(notebook.id)}>
              {notebook.title}
            </button>
            <details
              class="menu-konteks"
              open={menuKonteksAktif === `buku:${notebook.id}`}
              ontoggle={(event) => sinkronkanMenuKonteks(`buku:${notebook.id}`, event)}
              onclick={(event) => event.stopPropagation()}
            >
              <summary aria-label={"Menu " + notebook.title} title="Menu buku">⋯</summary>
              <div class="menu-konteks-daftar" popover="manual">
                <button type="button" onclick={() => ubahBuku(notebook)}>Ubah nama</button>
                <button type="button" disabled title="Akan tersedia pada pengaturan izin">Izin akses</button>
                <button class="berbahaya" type="button" onclick={() => hapusBuku(notebook)}>Hapus buku</button>
              </div>
            </details>
          </div>
        {:else}
          <p class="teks-kosong">Belum ada buku.</p>
        {/each}
      </nav>

      <div class="akun">
        <button class="tombol-pengaturan-samping" type="button" onclick={bukaPengaturan}>
          <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <line x1="4" y1="6" x2="20" y2="6" />
            <circle cx="9" cy="6" r="2" fill="var(--permukaan-lembut)" />
            <line x1="4" y1="12" x2="20" y2="12" />
            <circle cx="15" cy="12" r="2" fill="var(--permukaan-lembut)" />
            <line x1="4" y1="18" x2="20" y2="18" />
            <circle cx="11" cy="18" r="2" fill="var(--permukaan-lembut)" />
          </svg>
          <span>Pengaturan</span>
        </button>
      </div>
      {#if versi}<div class="versi-navigasi">v{versi}</div>{/if}
    </aside>
    {#if bukuTerbuka}
      <button
        class="lapisan-drawer-mobile"
        type="button"
        aria-label="Tutup daftar buku"
        onclick={() => ubahStatusDrawer(false)}
      ></button>
    {/if}

    <main class:layar-editor={tampilanMobile === 'editor'} class="ruang-kerja">
      <header class="bar-atas">
        {#if !bukuTerbuka}
          <button
            class="tombol-buka-buku"
            type="button"
            aria-label="Tampilkan daftar buku"
            title="Tampilkan daftar buku"
            aria-expanded={bukuTerbuka}
            aria-controls="daftar-buku"
            onclick={() => ubahStatusDrawer(true)}
          >
            <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M4 7h16M4 12h16M4 17h16" /></svg>
          </button>
        {/if}
        <button class="tombol-kembali-mobile" type="button" onclick={kembaliKeDaftar}>
          <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 6l-6 6 6 6" /></svg>
          Catatan
        </button>
        <div class="identitas-buku">
          <h1>{bukuAktif?.title || 'Belum ada buku'}</h1>
        </div>
        <div class="pencarian" class:aktif={pencarianAktif}>
          <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="6.5" /><path d="m16 16 4.5 4.5" /></svg>
          <input
            value={kueriPencarian}
            oninput={(event) => {
              const input = event.currentTarget as HTMLInputElement;
              kueriPencarian = input.value;
              jadwalkanPencarian();
            }}
            onfocus={() => { if (kueriPencarian.trim()) pencarianAktif = true; }}
            aria-label="Cari catatan dan sumber"
            placeholder="Cari catatan dan sumber…"
            autocomplete="off"
          />
          {#if kueriPencarian}
            <button type="button" aria-label="Bersihkan pencarian" title="Bersihkan pencarian" onclick={bersihkanPencarian}>×</button>
          {/if}
          {#if pencarianAktif}
            <div class="hasil-pencarian" role="status" aria-live="polite">
              {#if pencarianMemuat}
                <div class="status-pencarian">Mencari…</div>
              {:else if galatPencarian}
                <div class="status-pencarian galat-pencarian" role="alert">{galatPencarian}</div>
              {:else if !kueriPencarian.trim()}
                <div class="status-pencarian">Ketik kata untuk mencari.</div>
              {:else if !hasilPencarian.length}
                <div class="status-pencarian">Tidak ada hasil di buku ini.</div>
              {:else}
                {#each hasilPencarian as hasil}
                  <button class="hasil-pencarian-item" type="button" onclick={() => pilihHasilPencarian(hasil)}>
                    <span class="hasil-pencarian-ikon">{hasil.kind === 'note' ? 'C' : 'S'}</span>
                    <span class="hasil-pencarian-teks">
                      <strong>{hasil.title || (hasil.kind === 'note' ? 'Tanpa judul' : 'Tanpa nama')}</strong>
                      <small>{hasil.kind === 'note' ? 'Catatan' : 'Sumber'} · {bukuAktif?.title || 'Buku aktif'}</small>
                    </span>
                  </button>
                {/each}
              {/if}
            </div>
          {/if}
        </div>
        <div class="aksi-atas">
          <div class="menu-pengguna">
            <button
              class="pemicu-pengguna"
              type="button"
              aria-label="Menu pengguna"
              aria-haspopup="menu"
              aria-expanded={menuPenggunaTerbuka}
              onclick={() => menuPenggunaTerbuka = !menuPenggunaTerbuka}
            >
              <span class="avatar">{inisial}</span>
            </button>
            {#if menuPenggunaTerbuka}
              <div class="daftar-menu-pengguna" role="menu" aria-label="Menu pengguna">
                <div class="identitas-menu-pengguna">
                  <strong>{user?.display_name || user?.username}</strong>
                  <span>@{user?.username}</span>
                </div>
                <button type="button" role="menuitem" disabled>Profil <span>Segera</span></button>
                <button type="button" role="menuitem" onclick={keluar}>Keluar</button>
              </div>
            {/if}
          </div>
        </div>
      </header>

      <div class="pesan-area">
        {#if galat}<div class="pesan galat" role="alert">{galat}</div>{/if}
      </div>

      {#if !notebookID}
        <div class="keadaan-kosong tanpa-buku-kosong">
          <h2>Belum ada buku</h2>
          <p>Buat buku untuk mulai menyimpan catatan.</p>
          <button class="tombol utama" type="button" onclick={buatBuku}>+ Buku baru</button>
        </div>
      {:else}
        <div class:konteks-terlipat={!konteksTerbuka} class="workspace">
          <section class="panel-catatan" aria-label="Daftar catatan">
            <div class="panel-header">
              <div>
                <strong>Catatan</strong>
                <span>{notes.length} catatan</span>
              </div>
              <button class="ikon-tombol" type="button" onclick={buatCatatan} disabled={!notebookID} title="Catatan baru" aria-label="Catatan baru">+</button>
            </div>

            <div class="daftar-catatan">
              {#each notes as note}
                <div
                  class:aktif={catatanAktif?.id === note.id}
                  class:menu-terbuka={menuKonteksAktif === `catatan:${note.id}`}
                  class="item-catatan-bar"
                >
                  <button class="item-catatan" type="button" onclick={() => pilihCatatan(note)}>
                    <strong>{note.title || 'Tanpa judul'}</strong>
                    <span>{note.content.trim().slice(0, 76) || 'Belum ada isi'}</span>
                  </button>
                  <details
                    class="menu-konteks"
                    open={menuKonteksAktif === `catatan:${note.id}`}
                    ontoggle={(event) => sinkronkanMenuKonteks(`catatan:${note.id}`, event)}
                    onclick={(event) => event.stopPropagation()}
                  >
                    <summary aria-label={"Menu " + (note.title || 'catatan')} title="Menu catatan">⋯</summary>
                    <div class="menu-konteks-daftar" popover="manual">
                      <button type="button" onclick={() => ubahCatatan(note)}>Ubah nama</button>
                      <button type="button" disabled title="Akan tersedia pada pengaturan izin">Izin akses</button>
                      <button class="berbahaya" type="button" onclick={() => hapusCatatan(note)}>Hapus catatan</button>
                    </div>
                  </details>
                </div>
              {:else}
                <div class="keadaan-kosong kecil">
                  <strong>Belum ada catatan</strong>
                  <span>Mulai dengan catatan pertama untuk buku ini.</span>
                  <button class="tombol sekunder" type="button" onclick={buatCatatan}>Buat catatan</button>
                </div>
              {/each}
            </div>
          </section>

          <section class="editor" aria-label="Ruang kerja">
            {#if view === 'pratinjau' && sumberAktif}
              <div class="editor-atas">
                <button class="tautan-kembali" type="button" onclick={() => { view = 'catatan'; sumberAktif = null; }}>← Kembali ke catatan</button>
                <span class="status">Sumber asli</span>
              </div>
              <div class="pratinjau">
                <p class="eyebrow">{sumberAktif.kind}</p>
                <h2>{sumberAktif.title}</h2>
                <p class="meta-sumber">{sumberAktif.locator || 'Sumber lokal'}</p>
                <pre>{sumberAktif.content || 'Isi sumber kosong.'}</pre>
              </div>
            {:else if catatanAktif}
              <div class="editor-atas">
                <span class="status {statusSimpan}">
                  <i></i>
                  {statusSimpan === 'menyimpan' ? 'Menyimpan…' : statusSimpan === 'gagal' ? 'Belum tersimpan' : 'Tersimpan'}
                </span>
              </div>
              <input class="judul-catatan" bind:value={catatanAktif.title} oninput={jadwalkanSimpan} aria-label="Judul catatan" />
              <textarea bind:value={catatanAktif.content} oninput={jadwalkanSimpan} aria-label="Isi catatan" placeholder="Mulai menulis…"></textarea>
            {:else}
              <div class="keadaan-kosong besar">
                <h2>Belum ada catatan</h2>
                <p>Buat catatan pertama untuk mulai menulis.</p>
                <button class="tombol utama" type="button" onclick={buatCatatan}>Buat catatan</button>
              </div>
            {/if}
          </section>

          <aside class="panel-konteks" aria-label="Konteks">
            <div class="tab-konteks">
              <div class="tab-konteks-pilihan">
                <button class:aktif={panel === 'sumber'} type="button" onclick={() => panel = 'sumber'}>Sumber</button>
                <button class:aktif={panel === 'artefak'} type="button" onclick={() => panel = 'artefak'}>Artefak</button>
              </div>
              <button
                class="tombol-lipat-konteks"
                type="button"
                aria-label="Sembunyikan panel sumber dan artefak"
                title="Sembunyikan panel sumber dan artefak"
                onclick={() => ubahStatusDrawerKonteks(false)}
              >
                <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M9 6l6 6-6 6" /></svg>
              </button>
            </div>

            {#if panel === 'sumber'}
              <div class="konteks-header">
                <div>
                  <strong>Sumber</strong>
                  <span>{sources.length} sumber</span>
                </div>
                <button class="ikon-tombol" type="button" onclick={bukaImpor} disabled={!notebookID} title="Tambah sumber">+</button>
              </div>
              <input bind:this={imporInput} class="tersembunyi" type="file" accept=".txt,.md,.csv,.json,.html,.xml,.log,text/*" onchange={imporSumber} />
              <div class="daftar-sumber">
                {#each sources as source}
                  <article class:aktif={sumberAktif?.id === source.id} class="item-sumber">
                    <button type="button" onclick={() => pilihSumber(source)}>
                      <strong>{source.title}</strong>
                      <span>{source.kind}</span>
                    </button>
                    <button class="hapus-kecil" type="button" title="Hapus sumber" aria-label={"Hapus " + source.title} onclick={() => hapusSumber(source)}>×</button>
                  </article>
                {:else}
                  <div class="konteks-kosong">
                    <strong>Belum ada sumber.</strong>
                    <span>Tambahkan berkas teks untuk digunakan sebagai referensi.</span>
                    <button class="tombol sekunder" type="button" onclick={bukaImpor} disabled={!notebookID}>+ Tambah sumber</button>
                  </div>
                {/each}
              </div>
            {:else}
              <div class="konteks-header">
                <div>
                  <strong>Artefak</strong>
                  <span>Hasil kerja</span>
                </div>
              </div>
              <div class="konteks-kosong">
                <strong>Belum ada artefak.</strong>
                <span>Ruang artefak disiapkan untuk tahap berikutnya tanpa mengganggu ruang kerja sekarang.</span>
              </div>
            {/if}
          </aside>
        </div>
        {#if konteksTerbuka}
          <button
            class="lapisan-konteks-mobile"
            type="button"
            aria-label="Tutup panel sumber dan artefak"
            onclick={() => ubahStatusDrawerKonteks(false)}
          ></button>
        {/if}
        {#if !konteksTerbuka}
          <button
            class="tombol-buka-konteks"
            type="button"
            aria-label="Tampilkan panel sumber dan artefak"
            title="Tampilkan panel sumber dan artefak"
            onclick={() => ubahStatusDrawerKonteks(true)}
          >
            <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M15 6l-6 6 6 6" /></svg>
          </button>
        {/if}
        <nav class="bar-bawah-mobile" aria-label="Navigasi ponsel">
          <button class:aktif={!konteksTerbuka} type="button" onclick={() => { ubahStatusDrawerKonteks(false); kembaliKeDaftar(); }}>
            <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M6 3h9l4 4v14H6z" /><path d="M9 12h7M9 16h7" /></svg>
            <span>Catatan</span>
          </button>
          <button class:aktif={konteksTerbuka && panel === 'sumber'} type="button" onclick={() => { panel = 'sumber'; ubahStatusDrawerKonteks(true); }}>
            <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M5 4h9l5 5v11H5z" /><path d="M14 4v5h5" /></svg>
            <span>Sumber</span>
          </button>
          <button class:aktif={konteksTerbuka && panel === 'artefak'} type="button" onclick={() => { panel = 'artefak'; ubahStatusDrawerKonteks(true); }}>
            <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3l8 4.5v9L12 21l-8-4.5v-9z" /><path d="M12 12l8-4.5M12 12v9M12 12L4 7.5" /></svg>
            <span>Artefak</span>
          </button>
        </nav>
      {/if}
    </main>
    {#if dialogPengaturanTerbuka}
      <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
      <dialog
        class="dialog dialog-pengaturan"
        aria-labelledby="judul-pengaturan"
        use:bukaModal
        onclose={() => (dialogPengaturanTerbuka = false)}
        onclick={(event) => {
          if (event.target === event.currentTarget) dialogPengaturanTerbuka = false;
        }}
      >
        <header>
          <div>
            <h2 id="judul-pengaturan">Pengaturan</h2>
            <p>Berlaku untuk perangkat ini dan tetap tersimpan tanpa jaringan.</p>
          </div>
          <button class="tombol-tutup-dialog" type="button" aria-label="Tutup pengaturan" onclick={() => (dialogPengaturanTerbuka = false)}>×</button>
        </header>
        <div class="isi-pengaturan">
          <label class="item-pengaturan">
            <span>
              <strong>Pengguna boleh mendaftar</strong>
              <small>Tampilkan opsi pembuatan akun pada halaman masuk di perangkat ini.</small>
            </span>
            <input type="checkbox" checked={pendaftaranDiizinkan} onchange={ubahPengaturanPendaftaran} />
          </label>
          {#if galatPengaturan}
            <p class="pesan-pengaturan" role="alert">{galatPengaturan}</p>
          {/if}
        </div>
        <footer>
          <button data-fokus class="tombol utama" type="button" onclick={() => (dialogPengaturanTerbuka = false)}>Selesai</button>
        </footer>
      </dialog>
    {/if}
    {#if dialogAksi}
      <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
      <dialog
        class="dialog dialog-aksi"
        aria-labelledby="judul-dialog-aksi"
        use:bukaModal
        onclose={() => tutupDialogAksi(null)}
        onclick={(event) => {
          if (event.target === event.currentTarget) tutupDialogAksi(null);
        }}
      >
        <form onsubmit={kirimDialogAksi}>
          <header>
            <div>
              <h2 id="judul-dialog-aksi">{dialogAksi.judul}</h2>
              {#if dialogAksi.pesan}<p>{dialogAksi.pesan}</p>{/if}
            </div>
          </header>
          {#if dialogAksi.jenis === 'teks'}
            <div class="isi-dialog">
              <input data-fokus bind:value={dialogAksi.nilai} aria-label={dialogAksi.judul} autocomplete="off" required />
            </div>
          {/if}
          <footer>
            <button data-fokus class="tombol sekunder" type="button" onclick={() => tutupDialogAksi(null)}>Batal</button>
            <button class:bahaya={dialogAksi.bahaya} class="tombol utama" type="submit">{dialogAksi.label}</button>
          </footer>
        </form>
      </dialog>
    {/if}
  </div>
{/if}

<style>
  .pencarian { position: relative; flex: 1 1 360px; max-width: 520px; min-width: 180px; display: flex; align-items: center; gap: 8px; padding: 0 10px; border: 1px solid var(--garis); border-radius: var(--radius); background: var(--permukaan); color: var(--teks-2); }
  .pencarian:focus-within, .pencarian.aktif { border-color: var(--teks-3); }
  .pencarian > svg { width: 18px; height: 18px; flex: 0 0 auto; }
  .pencarian input { width: 100%; min-width: 0; height: 42px; padding: 0; border: 0; outline: 0; background: transparent; color: var(--teks); }
  .pencarian input::placeholder { color: var(--teks-3); }
  .pencarian > button { width: 30px; height: 30px; flex: 0 0 auto; border: 0; border-radius: 50%; background: transparent; color: var(--teks-2); cursor: pointer; font-size: 1.15rem; }
  .pencarian > button:hover { background: var(--permukaan-hover); color: var(--teks); }
  .hasil-pencarian { position: absolute; top: calc(100% + 8px); left: 0; right: 0; z-index: 50; overflow: hidden; border: 1px solid var(--garis); border-radius: var(--radius); background: var(--permukaan); box-shadow: var(--bayangan-dialog); }
  .status-pencarian { padding: 14px 16px; color: var(--teks-2); font-size: .9rem; }
  .galat-pencarian { color: var(--bahaya); }
  .hasil-pencarian-item { display: flex; width: 100%; gap: 10px; align-items: center; padding: 11px 12px; border: 0; background: transparent; color: var(--teks); text-align: left; cursor: pointer; }
  .hasil-pencarian-item:hover, .hasil-pencarian-item:focus-visible { background: var(--permukaan-hover); }
  .hasil-pencarian-ikon { display: grid; width: 30px; height: 30px; flex: 0 0 auto; place-items: center; border-radius: 8px; background: var(--permukaan-lembut); color: var(--teks-2); font-size: .75rem; font-weight: 700; }
  .hasil-pencarian-teks { min-width: 0; display: grid; gap: 2px; }
  .hasil-pencarian-teks strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .hasil-pencarian-teks small { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--teks-2); }
  @media (max-width: 760px) {
    .pencarian { order: 3; flex-basis: 100%; max-width: none; }
    .hasil-pencarian { position: fixed; top: calc(var(--tinggi-bar) + 8px); left: 12px; right: 12px; }
  }

  .item-buku-bar,
  .item-catatan-bar { position: relative; display: flex; align-items: stretch; min-width: 0; }
  .item-buku-bar.menu-terbuka,
  .item-catatan-bar.menu-terbuka { z-index: 20; }
  .item-buku-bar .item-buku,
  .item-catatan-bar .item-catatan { flex: 1 1 auto; min-width: 0; }
  .menu-konteks { position: absolute; top: 50%; right: 4px; transform: translateY(-50%); z-index: 2; }
  .menu-konteks summary { list-style: none; width: 40px; height: 40px; display: grid; place-items: center; border-radius: var(--radius-kecil); color: var(--teks-2); cursor: pointer; font-size: 1.25rem; line-height: 1; }
  .menu-konteks summary::-webkit-details-marker { display: none; }
  .menu-konteks summary:hover,
  .menu-konteks[open] summary { background: var(--permukaan-hover); color: var(--teks); }
  .menu-konteks-daftar { position: absolute; top: 42px; right: 0; min-width: 190px; padding: 6px; border: 1px solid var(--garis); border-radius: var(--radius); background: var(--permukaan); color: var(--teks); box-shadow: var(--bayangan-dialog); z-index: 10; }
  .menu-konteks-daftar:popover-open { position: fixed; inset: auto; max-width: calc(100vw - 16px); max-height: calc(100dvh - 16px); margin: 0; overflow: auto; z-index: 1000; }
  .menu-konteks-daftar button { display: block; width: 100%; min-height: 44px; padding: 8px 12px; border: 0; border-radius: var(--radius-kecil); background: transparent; color: var(--teks); text-align: left; cursor: pointer; }
  .menu-konteks-daftar button:hover:not(:disabled) { background: var(--permukaan-hover); }
  .menu-konteks-daftar button:disabled { color: var(--teks-3); cursor: not-allowed; }
  .menu-konteks-daftar .berbahaya { color: var(--bahaya); }

  /* Pada perangkat dengan penunjuk, menu ⋯ muncul saat butir disorot atau terpilih. */
  @media (hover: hover) {
    .menu-konteks summary { opacity: 0; }
    .item-buku-bar:hover summary,
    .item-catatan-bar:hover summary,
    .item-buku-bar:focus-within summary,
    .item-catatan-bar:focus-within summary,
    .item-buku-bar.aktif summary,
    .item-catatan-bar.aktif summary,
    .menu-konteks[open] summary { opacity: 1; }
  }

  .workspace.konteks-terlipat { grid-template-columns: var(--lebar-catatan) minmax(0, 1fr); }
  .workspace.konteks-terlipat .panel-konteks { display: none; }
  .tombol-buka-konteks { position: fixed; top: calc(var(--tinggi-bar) + 12px); right: 12px; z-index: 6; border: 1px solid var(--garis); background: var(--permukaan); box-shadow: var(--bayangan-kecil); }
  .tanpa-buku-kosong { flex: 1 1 auto; display: grid; place-content: center; justify-items: center; padding: 40px 24px; text-align: center; }
  .tanpa-buku-kosong h2 { margin: 4px 0 8px; }
  .tanpa-buku-kosong p { max-width: 420px; margin: 0 0 20px; color: var(--teks-2); }
</style>
