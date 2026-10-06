<script lang="ts">
  type User = { id: string; username: string; display_name: string };
  type Notebook = { id: string; title: string; description: string };
  type Note = { id: string; title: string; content: string; updated_at: string };
  type Source = { id: string; title: string; kind: string; content?: string; locator: string; checksum: string; metadata_json: string };

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
  let email = '';
  let namaTampilan = '';

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

  let statusSimpan: 'tersimpan' | 'menyimpan' | 'gagal' = 'tersimpan';
  let timerSimpan: ReturnType<typeof setTimeout> | undefined;
  let nomorSimpan = 0;
  let imporInput: HTMLInputElement;

  async function permintaan(path: string, init: RequestInit = {}) {
    const headers = new Headers(init.headers);
    if (init.body && !(init.body instanceof FormData)) headers.set('Content-Type', 'application/json');
    const response = await fetch(path, { ...init, headers });
    if (!response.ok) throw new Error((await response.text()) || 'Permintaan gagal.');
    return response;
  }

  function sinkronkanMenuKonteks(id: string, event: Event) {
    const details = event.currentTarget;
    if (details instanceof HTMLDetailsElement) menuKonteksAktif = details.open ? id : '';
  }

  function tutupMenuKonteksDiLuar(event: MouseEvent) {
    const target = event.target;
    if (target instanceof Element && target.closest('.menu-konteks')) return;
    menuKonteksAktif = '';
  }

  async function mulai() {
    memuat = true;
    try {
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
    try {
      const endpoint = modeMasuk === 'masuk' ? '/api/auth/login' : '/api/auth/setup';
      const body = modeMasuk === 'masuk'
        ? { username: namaPengguna.trim(), password: kataSandi }
        : { username: namaPengguna.trim(), email: email.trim(), display_name: namaTampilan.trim(), password: kataSandi };

      const response = await permintaan(endpoint, { method: 'POST', body: JSON.stringify(body) });
      user = (await response.json()).user;
      kataSandi = '';
      siap = true;
      await muatBuku();
    } catch (error) {
      galat = error instanceof Error ? error.message : 'Autentikasi gagal.';
    }
  }

  async function keluar() {
    try {
      await fetch('/api/auth/logout', { method: 'POST' });
    } finally {
      user = null;
      siap = false;
      notebooks = [];
      notes = [];
      sources = [];
      catatanAktif = null;
      sumberAktif = null;
      notebookID = '';
      menuKonteksAktif = '';
    }
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
    statusSimpan = 'tersimpan';
  }

  async function pilihBuku(id: string) {
    if (id === notebookID) return;
    if (timerSimpan) clearTimeout(timerSimpan);
    nomorSimpan++;
    menuKonteksAktif = '';
    notebookID = id;
    galat = '';
    try {
      await muatIsiBuku();
    } catch (error) {
      galat = error instanceof Error ? error.message : 'Gagal memuat buku.';
    }
  }

  async function buatBuku() {
    const title = window.prompt('Nama buku');
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
    const title = window.prompt('Nama buku', notebook.title);
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
    if (!window.confirm(`Hapus buku “${notebook.title}”?\n\nSemua catatan, sumber, dan isi terkait di dalamnya juga akan dihapus.`)) return;
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
    statusSimpan = 'tersimpan';
    galat = '';
  }

  async function ubahCatatan(note: Note) {
    menuKonteksAktif = '';
    const title = window.prompt('Judul catatan', note.title);
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
    if (!window.confirm(`Hapus catatan “${note.title || 'Tanpa judul'}”?`)) return;
    const id = note.id;
    nomorSimpan++;
    if (timerSimpan) clearTimeout(timerSimpan);
    try {
      await permintaan('/api/notes/' + id, { method: 'DELETE' });
      notes = notes.filter((item) => item.id !== id);
      catatanAktif = notes[0] ? { ...notes[0] } : null;
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
    } catch (error) {
      galat = error instanceof Error ? error.message : 'Gagal membuka sumber.';
    }
  }

  async function hapusSumber(source: Source) {
    if (!window.confirm('Hapus sumber ini?')) return;
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

<svelte:head>
  <title>{versi ? `Catatan · ${versi}` : 'Catatan'}</title>
  <meta name="description" content="Catatan lokal yang sederhana dan tetap berada di perangkat." />
</svelte:head>

{#if memuat}
  <main class="layar-status">
    <div class="status-muat">
      <span class="logo-mark">C</span>
      <span>Memuat…</span>
    </div>
  </main>
{:else if !siap}
  <main class="layar-autentikasi">
    <section class="autentikasi">
      <div class="autentikasi-mark">C</div>
      <p class="eyebrow">CATATAN</p>
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

        <button class="tombol utama lebar" type="submit">
          {modeMasuk === 'masuk' ? 'Masuk' : 'Buat akun'}
        </button>
      </form>

      <button class="tautan" type="button" onclick={() => {
        modeMasuk = modeMasuk === 'masuk' ? 'siapkan' : 'masuk';
        galat = '';
      }}>
        {modeMasuk === 'masuk' ? 'Perangkat baru? Buat akun lokal' : 'Sudah punya akun? Masuk'}
      </button>

      {#if versi}<p class="versi">Versi {versi}</p>{/if}
    </section>
  </main>
{:else}
  <div class:tanpa-buku={!bukuTerbuka} class="aplikasi" onclick={tutupMenuKonteksDiLuar}>
    <aside class="panel-navigasi">
      <button
        class="tombol-lipat-buku"
        type="button"
        aria-label="Sembunyikan daftar buku"
        title="Sembunyikan daftar buku"
        aria-expanded={bukuTerbuka}
        aria-controls="daftar-buku"
        onclick={() => bukuTerbuka = false}
      >‹</button>
      <div class="merek">
        <span class="logo-mark kecil">C</span>
        <div>
          <strong>Catatan</strong>
          <span>Ruang kerja pribadi</span>
        </div>
      </div>

      <button class="tombol sekunder lebar" type="button" onclick={buatBuku}>+ Buku baru</button>

      <div class="judul-panel">BUKU</div>
      <nav id="daftar-buku" class="daftar-buku" aria-label="Buku">
        {#each notebooks as notebook}
          <div class:aktif={notebook.id === notebookID} class="item-buku-bar">
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
              <div class="menu-konteks-daftar">
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
        <span class="avatar">{inisial}</span>
        <div class="akun-teks">
          <strong>{user?.display_name || user?.username}</strong>
          <span>Perangkat ini</span>
        </div>
        <button class="ikon-tombol" type="button" title="Keluar" aria-label="Keluar" onclick={keluar}>↪</button>
      </div>
      {#if versi}<div class="versi-navigasi">v{versi}</div>{/if}
    </aside>

    <main class="ruang-kerja">
      <header class="bar-atas">
        {#if !bukuTerbuka}
          <button
            class="tombol-buka-buku"
            type="button"
            aria-label="Tampilkan daftar buku"
            title="Tampilkan daftar buku"
            aria-expanded={bukuTerbuka}
            aria-controls="daftar-buku"
            onclick={() => bukuTerbuka = true}
          >›</button>
        {/if}
        <div class="identitas-buku">
          <p class="eyebrow">BUKU</p>
          <h1>{bukuAktif?.title || 'Belum ada buku'}</h1>
        </div>
        <div class="aksi-atas">
          <button class="avatar" type="button" title={user?.username}>{inisial}</button>
        </div>
      </header>

      <div class="pesan-area">
        {#if galat}<div class="pesan galat" role="alert">{galat}</div>{/if}
      </div>

      {#if !notebookID}
        <div class="keadaan-kosong tanpa-buku-kosong">
          <p class="eyebrow">RUANG KERJA</p>
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
                <div class:aktif={catatanAktif?.id === note.id} class="item-catatan-bar">
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
                    <div class="menu-konteks-daftar">
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
                <p class="eyebrow">RUANG KERJA</p>
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
                onclick={() => konteksTerbuka = false}
              >›</button>
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
        {#if !konteksTerbuka}
          <button
            class="tombol-buka-konteks"
            type="button"
            aria-label="Tampilkan panel sumber dan artefak"
            title="Tampilkan panel sumber dan artefak"
            onclick={() => konteksTerbuka = true}
          >‹</button>
        {/if}
      {/if}
    </main>
  </div>
{/if}

<style>
  .identitas-buku { min-width: 0; }
  .item-buku-bar,
  .item-catatan-bar { position: relative; display: flex; align-items: stretch; min-width: 0; }
  .item-buku-bar .item-buku,
  .item-catatan-bar .item-catatan { flex: 1 1 auto; min-width: 0; }
  .item-buku-bar.aktif .item-buku,
  .item-catatan-bar.aktif .item-catatan { padding-right: 38px; }
  .menu-konteks { position: relative; flex: 0 0 auto; }
  .item-buku-bar .menu-konteks,
  .item-catatan-bar .menu-konteks { position: absolute; top: 50%; right: 6px; transform: translateY(-50%); z-index: 2; }
  .menu-konteks summary { list-style: none; width: 30px; height: 30px; display: grid; place-items: center; border-radius: 7px; color: var(--teks-2); cursor: pointer; font-size: 18px; line-height: 1; }
  .menu-konteks summary::-webkit-details-marker { display: none; }
  .menu-konteks summary:hover,
  .menu-konteks[open] summary { background: var(--permukaan-lembut); color: var(--teks); }
  .menu-konteks-daftar { position: absolute; top: 34px; right: 0; min-width: 155px; padding: 5px; border: 1px solid var(--garis); border-radius: 9px; background: var(--permukaan); box-shadow: 0 8px 24px rgba(24,25,22,.12); z-index: 10; }
  .menu-konteks-daftar button { display: block; width: 100%; padding: 8px 10px; border: 0; border-radius: 6px; background: transparent; color: var(--teks); text-align: left; cursor: pointer; font: inherit; }
  .menu-konteks-daftar button:hover:not(:disabled) { background: var(--permukaan-lembut); }
  .menu-konteks-daftar button:disabled { color: var(--teks-2); cursor: not-allowed; opacity: .65; }
  .menu-konteks-daftar .berbahaya { color: #a33a32; }
  .tab-konteks { display: flex; align-items: center; justify-content: space-between; gap: 6px; }
  .tab-konteks-pilihan { display: flex; min-width: 0; flex: 1 1 auto; }
  .tombol-lipat-konteks,
  .tombol-buka-konteks { width: 30px; height: 30px; padding: 0; border: 0; border-radius: 7px; background: transparent; color: var(--teks-2); cursor: pointer; font-size: 17px; }
  .tombol-lipat-konteks:hover,
  .tombol-buka-konteks:hover { background: var(--permukaan-lembut); color: var(--teks); }
  .workspace.konteks-terlipat { grid-template-columns: 248px minmax(0, 1fr); }
  .workspace.konteks-terlipat .panel-konteks { display: none; }
  .tombol-buka-konteks { position: fixed; top: 84px; right: 13px; z-index: 6; background: var(--permukaan-lembut); box-shadow: 0 3px 14px rgba(24,25,22,.08); }
  .tanpa-buku-kosong { min-height: calc(100vh - 145px); display: grid; place-content: center; justify-items: center; padding: 40px 24px; text-align: center; }
  .tanpa-buku-kosong h2 { margin: 4px 0 8px; }
  .tanpa-buku-kosong p:not(.eyebrow) { max-width: 420px; margin: 0 0 20px; color: var(--teks-2); }

  @media (max-width: 1120px) and (min-width: 761px) {
    .workspace.konteks-terlipat { grid-template-columns: 220px minmax(0, 1fr); }
  }

  @media (max-width: 760px) {
    .item-buku-bar .menu-konteks,
    .item-catatan-bar .menu-konteks { right: 4px; }
    .menu-konteks-daftar { right: -2px; }
    .workspace.konteks-terlipat { grid-template-columns: minmax(0, 1fr); }
    .tombol-buka-konteks { top: 76px; right: 10px; }
  }
</style>
