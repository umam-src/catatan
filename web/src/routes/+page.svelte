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
    if (!notebookID) {
      notes = [];
      sources = [];
      catatanAktif = null;
      sumberAktif = null;
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

  async function buatCatatan() {
    if (!notebookID) return;
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
    catatanAktif = { ...note };
    view = 'catatan';
    statusSimpan = 'tersimpan';
    galat = '';
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

  async function hapusCatatan() {
    if (!catatanAktif || !window.confirm('Hapus catatan ini?')) return;
    const id = catatanAktif.id;
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
  <div class:tanpa-buku={!bukuTerbuka} class="aplikasi">
    <aside class="panel-navigasi">
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
          <button
            class:aktif={notebook.id === notebookID}
            class="item-buku"
            type="button"
            onclick={() => pilihBuku(notebook.id)}
          >
            {notebook.title}
          </button>
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
        <div>
          <p class="eyebrow">BUKU</p>
          <h1>{bukuAktif?.title || 'Belum ada buku'}</h1>
        </div>
        <div class="aksi-atas">
          <button class="tombol sekunder" type="button" onclick={buatCatatan} disabled={!notebookID}>+ Catatan</button>
          <button class="avatar" type="button" title={user?.username}>{inisial}</button>
        </div>
      </header>

      <div class="pesan-area">
        {#if galat}<div class="pesan galat" role="alert">{galat}</div>{/if}
      </div>

      <div class="workspace">
        <section class="panel-catatan" aria-label="Daftar catatan">
          <div class="panel-header">
            <div>
              <strong>Catatan</strong>
              <span>{notes.length} catatan</span>
            </div>
            <button class="ikon-tombol" type="button" onclick={buatCatatan} disabled={!notebookID} title="Catatan baru">+</button>
          </div>

          <div class="daftar-catatan">
            {#each notes as note}
              <button
                class:aktif={catatanAktif?.id === note.id}
                class="item-catatan"
                type="button"
                onclick={() => pilihCatatan(note)}
              >
                <strong>{note.title || 'Tanpa judul'}</strong>
                <span>{note.content.trim().slice(0, 76) || 'Belum ada isi'}</span>
              </button>
            {:else}
              <div class="keadaan-kosong kecil">
                <span class="ikon-kosong">+</span>
                <strong>Belum ada catatan</strong>
                <span>Buat catatan pertama untuk buku ini.</span>
                <button class="tombol sekunder" type="button" onclick={buatCatatan} disabled={!notebookID}>Buat catatan</button>
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
              <button class="tombol-hapus" type="button" onclick={hapusCatatan}>Hapus</button>
            </div>
            <input class="judul-catatan" bind:value={catatanAktif.title} oninput={jadwalkanSimpan} aria-label="Judul catatan" />
            <textarea bind:value={catatanAktif.content} oninput={jadwalkanSimpan} aria-label="Isi catatan" placeholder="Mulai menulis…"></textarea>
          {:else}
            <div class="keadaan-kosong besar">
              <span class="ikon-kosong besar">+</span>
              <p class="eyebrow">RUANG KERJA</p>
              <h2>Pilih atau buat catatan.</h2>
              <p>Tempat menulis Anda tetap sederhana. Semua catatan tersimpan di perangkat ini.</p>
              <button class="tombol utama" type="button" onclick={buatCatatan} disabled={!notebookID}>+ Catatan baru</button>
            </div>
          {/if}
        </section>

        <aside class="panel-konteks" aria-label="Konteks">
          <div class="tab-konteks">
            <button class:aktif={panel === 'sumber'} type="button" onclick={() => panel = 'sumber'}>Sumber</button>
            <button class:aktif={panel === 'artefak'} type="button" onclick={() => panel = 'artefak'}>Artefak</button>
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
              <span class="ikon-kosong">—</span>
              <strong>Belum ada artefak.</strong>
              <span>Ruang artefak disiapkan untuk tahap berikutnya tanpa mengganggu ruang kerja sekarang.</span>
            </div>
          {/if}
        </aside>
      </div>
    </main>
  </div>
{/if}
