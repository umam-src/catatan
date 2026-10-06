<script lang="ts">
  type User = { id: string; username: string; display_name: string };
  type Notebook = { id: string; title: string; description: string };
  type Note = { id: string; title: string; content: string; updated_at: string };
  type Source = { id: string; title: string; kind: string; locator: string; checksum: string };
  type RuangMode = 'note' | 'chat' | 'split' | 'artifact' | 'preview';
  type Konteks = 'sumber' | 'artefak';
  type StatusSimpan = 'tersimpan' | 'menyimpan' | 'gagal';

  let user: User | null = null;
  let siap = false;
  let mode: 'login' | 'setup' = 'login';
  let username = '';
  let password = '';
  let email = '';
  let displayName = '';
  let error = '';
  let loading = true;
  let versi = '';
  let notebooks: Notebook[] = [];
  let notes: Note[] = [];
  let sources: Source[] = [];
  let selected = '';
  let activeNote: Note | null = null;
  let activeSource: Source | null = null;
  let ruangMode: RuangMode = 'note';
  let konteks: Konteks = 'sumber';
  let panelKonteks = true;
  let statusSimpan: StatusSimpan = 'tersimpan';
  let timerSimpan: ReturnType<typeof setTimeout> | undefined;
  let nomorSimpan = 0;

  async function api(path: string, init: RequestInit = {}) {
    const headers = new Headers(init.headers);
    if (init.body && !(init.body instanceof FormData)) headers.set('Content-Type', 'application/json');
    const response = await fetch(path, { ...init, headers });
    if (!response.ok) throw new Error(await response.text() || 'Permintaan gagal.');
    return response;
  }

  async function muatVersi() {
    try {
      const response = await fetch('/api/version');
      if (response.ok) versi = (await response.json()).version ?? '';
    } catch { versi = ''; }
  }

  async function mulai() {
    loading = true;
    await muatVersi();
    try {
      const response = await fetch('/api/auth/me');
      if (response.ok) {
        user = (await response.json()).user;
        siap = true;
        await loadNotebooks();
      }
    } catch { error = 'Layanan lokal tidak dapat dihubungi.'; }
    finally { loading = false; }
  }

  function kirimForm(event: SubmitEvent) {
    event.preventDefault();
    void (mode === 'setup' ? siapkanAkun() : masuk());
  }

  async function masuk() {
    error = '';
    try {
      const data = await (await api('/api/auth/login', { method: 'POST', body: JSON.stringify({ username, password }) })).json();
      user = data.user; password = ''; siap = true; await loadNotebooks();
    } catch (err) { error = err instanceof Error ? err.message : 'Gagal masuk.'; }
  }

  async function siapkanAkun() {
    error = '';
    try {
      const data = await (await api('/api/auth/setup', { method: 'POST', body: JSON.stringify({ username, email, display_name: displayName, password }) })).json();
      user = data.user; password = ''; siap = true; await loadNotebooks();
    } catch (err) { error = err instanceof Error ? err.message : 'Gagal menyiapkan akun.'; }
  }

  async function keluar() {
    try { await fetch('/api/auth/logout', { method: 'POST' }); }
    finally {
      user = null; siap = false; notebooks = []; notes = []; sources = [];
      activeNote = null; activeSource = null; selected = '';
    }
  }

  async function loadNotebooks() {
    try {
      notebooks = await (await api('/api/notebooks')).json();
      if (!selected && notebooks.length) selected = notebooks[0].id;
      if (selected) await muatBuku(selected);
    } catch (err) { error = err instanceof Error ? err.message : 'Gagal memuat buku.'; }
  }

  async function muatBuku(notebookID: string) {
    await Promise.all([loadNotes(notebookID), loadSources(notebookID)]);
  }

  async function loadNotes(notebookID = selected) {
    if (!notebookID) { notes = []; activeNote = null; return; }
    try {
      notes = await (await api('/api/notebooks/' + notebookID + '/notes')).json();
      activeNote = notes[0] ? { ...notes[0] } : null;
      statusSimpan = 'tersimpan';
    } catch (err) { error = err instanceof Error ? err.message : 'Gagal memuat catatan.'; }
  }

  async function loadSources(notebookID = selected) {
    if (!notebookID) { sources = []; activeSource = null; return; }
    try {
      sources = await (await api('/api/notebooks/' + notebookID + '/sources')).json();
      if (activeSource && !sources.some((source) => source.id === activeSource?.id)) activeSource = null;
    } catch (err) { error = err instanceof Error ? err.message : 'Gagal memuat sumber.'; }
  }

  async function selectNotebook(id: string) {
    if (id === selected) return;
    if (timerSimpan) clearTimeout(timerSimpan);
    selected = id; activeSource = null; ruangMode = 'note';
    await muatBuku(id);
  }

  async function createNotebook() {
    const title = window.prompt('Nama buku');
    if (!title?.trim()) return;
    try {
      const notebook = await (await api('/api/notebooks', { method: 'POST', body: JSON.stringify({ title: title.trim(), description: '' }) })).json();
      notebooks = [notebook, ...notebooks]; selected = notebook.id; await muatBuku(notebook.id);
    } catch (err) { error = err instanceof Error ? err.message : 'Gagal membuat buku.'; }
  }

  async function createNote() {
    if (!selected) return;
    try {
      const created: Note = await (await api('/api/notebooks/' + selected + '/notes', { method: 'POST', body: JSON.stringify({ title: 'Catatan baru', content: '' }) })).json();
      activeNote = { ...created }; notes = [created, ...notes]; statusSimpan = 'tersimpan'; ruangMode = 'note';
    } catch (err) { error = err instanceof Error ? err.message : 'Gagal membuat catatan.'; }
  }

  function pilihCatatan(note: Note) {
    if (timerSimpan) clearTimeout(timerSimpan);
    activeNote = { ...note }; activeSource = null; statusSimpan = 'tersimpan'; error = ''; ruangMode = 'note';
  }

  async function pilihSumber(source: Source) {
    try {
      activeSource = await (await api('/api/sources/' + source.id)).json();
      ruangMode = 'preview';
      error = '';
    } catch (err) { error = err instanceof Error ? err.message : 'Gagal membuka sumber.'; }
  }

  async function hapusSumber(source: Source) {
    if (!window.confirm('Hapus sumber ini?')) return;
    try {
      await api('/api/sources/' + source.id, { method: 'DELETE' });
      sources = sources.filter((item) => item.id !== source.id);
      if (activeSource?.id === source.id) activeSource = null;
    } catch (err) { error = err instanceof Error ? err.message : 'Gagal menghapus sumber.'; }
  }

  async function imporSumber(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file || !selected) return;
    const form = new FormData();
    form.append('file', file);
    try {
      await api('/api/notebooks/' + selected + '/sources', { method: 'POST', body: form });
      await loadSources(selected);
      konteks = 'sumber';
    } catch (err) { error = err instanceof Error ? err.message : 'Gagal mengimpor sumber.'; }
    finally { input.value = ''; }
  }

  function jadwalkanSimpan() {
    if (!activeNote) return;
    statusSimpan = 'menyimpan';
    if (timerSimpan) clearTimeout(timerSimpan);
    const nomor = ++nomorSimpan;
    timerSimpan = setTimeout(() => void simpanCatatan(nomor), 700);
  }

  async function simpanCatatan(nomor: number) {
    if (!activeNote || nomor !== nomorSimpan) return;
    const catatan = { ...activeNote };
    try {
      const data = await (await api('/api/notes/' + catatan.id, { method: 'PUT', body: JSON.stringify({ title: catatan.title, content: catatan.content }) })).json();
      if (nomor !== nomorSimpan || activeNote?.id !== catatan.id) return;
      statusSimpan = 'tersimpan';
      notes = notes.map((item) => item.id === catatan.id ? { ...item, ...catatan, updated_at: data.updated_at } : item);
      error = '';
    } catch (err) {
      if (nomor !== nomorSimpan || activeNote?.id !== catatan.id) return;
      statusSimpan = 'gagal'; error = err instanceof Error ? err.message : 'Gagal menyimpan catatan.';
    }
  }

  async function hapusCatatan() {
    if (!activeNote || !window.confirm('Hapus catatan ini?')) return;
    nomorSimpan++;
    if (timerSimpan) clearTimeout(timerSimpan);
    try {
      await api('/api/notes/' + activeNote.id, { method: 'DELETE' });
      notes = notes.filter((item) => item.id !== activeNote?.id);
      activeNote = notes[0] ? { ...notes[0] } : null;
      statusSimpan = 'tersimpan';
    } catch (err) { error = err instanceof Error ? err.message : 'Gagal menghapus catatan.'; }
  }

  mulai();
</script>

<svelte:head>
  <title>{versi ? `Catatan ${versi}` : 'Catatan'}</title>
  <meta name="description" content="Catatan lokal yang tetap berada di perangkat ini." />
</svelte:head>

{#if loading}
  <main class="layar-tengah"><div class="muat"><span class="logo kecil">C</span><span>Memuat Catatan…</span></div></main>
{:else if !siap}
  <main class="layar-masuk">
    <section class="kartu-masuk">
      <div class="logo">C</div><div class="label-kecil">CATATAN LOKAL</div>
      <h1>{mode === 'setup' ? 'Mulai dengan Catatan' : 'Selamat datang kembali'}</h1>
      <p>{mode === 'setup' ? 'Buat akun lokal untuk menyimpan catatan di perangkat ini.' : 'Masuk untuk melanjutkan ke catatan Anda.'}</p>
      {#if error}<div class="galat" role="alert">{error}</div>{/if}
      <form onsubmit={kirimForm}>
        <label>Nama pengguna<input bind:value={username} autocomplete="username" required /></label>
        {#if mode === 'setup'}
          <label>Email <span>opsional</span><input bind:value={email} type="email" autocomplete="email" /></label>
          <label>Nama tampilan <span>opsional</span><input bind:value={displayName} autocomplete="name" /></label>
        {/if}
        <label>Kata sandi<input bind:value={password} type="password" minlength="12" autocomplete={mode === 'setup' ? 'new-password' : 'current-password'} required /></label>
        <button class="tombol utama" type="submit">{mode === 'setup' ? 'Buat akun lokal' : 'Masuk'}</button>
      </form>
      <button class="tombol-teks" onclick={() => { mode = mode === 'login' ? 'setup' : 'login'; error = ''; }}>{mode === 'login' ? 'Ini perangkat baru? Siapkan akun' : 'Sudah punya akun? Masuk'}</button>
      {#if versi}<div class="versi">v{versi}</div>{/if}
    </section>
  </main>
{:else}
  <div class="aplikasi">
    <aside class="navigasi" class:tersembunyi={!panelKonteks}>
      <div class="identitas"><span class="logo kecil">C</span><div><strong>Catatan</strong><small>Lokal, sederhana, milik Anda</small></div></div>
      <button class="tombol buat" onclick={createNotebook}>＋ Buku baru</button>
      <div class="judul-bagian">CATATAN</div>
      <nav aria-label="Daftar catatan" class="daftar-kiri">
        {#each notes as note}
          <button class:dipilih={activeNote?.id === note.id} class="item-kiri" onclick={() => pilihCatatan(note)}>
            <strong>{note.title || 'Tanpa judul'}</strong><span>{note.content.slice(0, 70) || 'Belum ada isi'}</span>
          </button>
        {:else}<div class="kosong-kecil">Belum ada catatan.</div>{/each}
      </nav>
      <button class="tombol sekunder tambah-catatan" onclick={createNote} disabled={!selected}>＋ Catatan</button>
      <div class="buku-kecil">
        <div class="judul-bagian">BUKU</div>
        {#each notebooks as notebook}
          <button class:aktif={selected === notebook.id} class="buku" onclick={() => selectNotebook(notebook.id)}>{notebook.title}</button>
        {/each}
      </div>
      <div class="akun"><div class="avatar">{(user?.display_name || user?.username || 'C').slice(0, 1).toUpperCase()}</div><div class="akun-nama"><strong>{user?.display_name || user?.username}</strong><small>Perangkat ini</small></div><button aria-label="Keluar" onclick={keluar}>↪</button></div>
      {#if versi}<div class="versi-nav">v{versi}</div>{/if}
    </aside>

    <main class="ruang">
      <header class="kepala">
        <div><div class="label-kecil">BUKU</div><h1>{notebooks.find((item) => item.id === selected)?.title ?? 'Catatan'}</h1></div>
        <div class="aksi-kepala">
          <button class:aktif={ruangMode === 'note'} class="mode" onclick={() => ruangMode = 'note'}>Note</button>
          <button class:aktif={ruangMode === 'chat'} class="mode" onclick={() => ruangMode = 'chat'}>Chat</button>
          <button class:aktif={ruangMode === 'split'} class="mode" onclick={() => ruangMode = 'split'}>Split</button>
          <button class="tombol ikon" onclick={() => panelKonteks = !panelKonteks} aria-label="Tampilkan konteks" title="Tampilkan konteks">☰</button>
        </div>
      </header>

      <div class="isi-utama">
        <section class="ruang-kerja" aria-live="polite">
          {#if error}<div class="galat" role="alert">{error}</div>{/if}

          {#if ruangMode === 'chat'}
            <div class="ruang-kosong"><span class="simbol besar">C</span><div class="label-kecil">RUANG KERJA</div><h2>Chat</h2><p>Ruang Chat menggunakan sumber yang dipilih sebagai konteks. Mesin Chat akan ditambahkan tanpa mengubah susunan ruang kerja ini.</p></div>
          {:else if ruangMode === 'split'}
            <div class="split">
              <section class="sub-ruang note-ruang">
                {#if activeNote}
                  <div class="bilah-editor"><span class="status {statusSimpan}"><i></i>{statusSimpan === 'menyimpan' ? 'Menyimpan…' : statusSimpan === 'gagal' ? 'Gagal menyimpan' : 'Tersimpan'}</span><button class="hapus" onclick={hapusCatatan}>Hapus</button></div>
                  <input class="judul-editor" bind:value={activeNote.title} oninput={jadwalkanSimpan} aria-label="Judul catatan" />
                  <textarea bind:value={activeNote.content} oninput={jadwalkanSimpan} placeholder="Mulai menulis…" aria-label="Isi catatan"></textarea>
                {:else}<div class="ruang-kosong kecil"><h2>Pilih atau buat catatan</h2><button class="tombol utama" onclick={createNote}>＋ Catatan baru</button></div>{/if}
              </section>
              <section class="sub-ruang chat-ruang"><div class="sub-kepala">Chat</div><div class="ruang-kosong kecil"><p>Tanya tentang catatan atau sumber terpilih di sini.</p><span class="status">Mesin Chat belum diaktifkan.</span></div></section>
            </div>
          {:else if ruangMode === 'preview' && activeSource}
            <div class="preview"><div class="bilah-editor"><span class="status">Pratinjau sumber</span><button class="tombol sekunder" onclick={() => ruangMode = 'note'}>Kembali</button></div><h2>{activeSource.title}</h2><p class="metadata">{activeSource.kind} · {activeSource.checksum.slice(0, 12)}…</p><pre>{activeSource.content}</pre></div>
          {:else}
            {#if activeNote}
              <div class="bilah-editor"><span class="status {statusSimpan}"><i></i>{statusSimpan === 'menyimpan' ? 'Menyimpan…' : statusSimpan === 'gagal' ? 'Gagal menyimpan' : 'Tersimpan'}</span><button class="hapus" onclick={hapusCatatan}>Hapus</button></div>
              <input class="judul-editor" bind:value={activeNote.title} oninput={jadwalkanSimpan} aria-label="Judul catatan" />
              <textarea bind:value={activeNote.content} oninput={jadwalkanSimpan} placeholder="Mulai menulis…" aria-label="Isi catatan"></textarea>
            {:else}<div class="ruang-kosong"><span class="simbol besar">✦</span><div class="label-kecil">RUANG KERJA</div><h2>Pilih atau buat catatan</h2><p>Catatan disimpan di perangkat ini dan tetap dapat digunakan tanpa jaringan.</p><button class="tombol utama" onclick={createNote} disabled={!selected}>＋ Catatan baru</button></div>{/if}
          {/if}
        </section>

        {#if panelKonteks}
          <aside class="konteks">
            <div class="tab-konteks"><button class:aktif={konteks === 'sumber'} onclick={() => konteks = 'sumber'}>Sumber</button><button class:aktif={konteks === 'artefak'} onclick={() => konteks = 'artefak'}>Artefak</button></div>
            {#if konteks === 'sumber'}
              <div class="konteks-isi">
                <div class="judul-panel"><strong>Sumber</strong><label class="tombol sekunder tambah-sumber">＋ Tambah sumber<input type="file" accept=".txt,.md,.csv,.json,.xml,.html,text/plain,text/markdown,text/csv,application/json,text/xml,text/html" onchange={imporSumber} /></label></div>
                {#each sources as source}
                  <div class:aktif={activeSource?.id === source.id} class="sumber-item">
                    <button onclick={() => pilihSumber(source)}><strong>{source.title}</strong><span>{source.kind}</span></button>
                    <button class="hapus-sumber" onclick={() => hapusSumber(source)} aria-label="Hapus sumber" title="Hapus sumber">×</button>
                  </div>
                {:else}<div class="kosong-panel">Belum ada sumber. Tambahkan dokumen untuk digunakan sebagai referensi.</div>{/each}
                {#if activeSource && ruangMode !== 'preview'}<div class="pratinjau"><strong>Pratinjau</strong><p>{activeSource.content?.slice(0, 500) || 'Isi sumber tidak tersedia.'}</p><button class="tombol sekunder" onclick={() => ruangMode = 'preview'}>Buka</button></div>{/if}
              </div>
            {:else}
              <div class="konteks-isi"><div class="judul-panel"><strong>Artefak</strong><button class="tombol sekunder" disabled>＋ Buat artefak</button></div><div class="kosong-panel">Belum ada artefak. Hasil kerja dari Chat dapat ditambahkan di sini ketika kemampuan Artefak diaktifkan.</div></div>
            {/if}
          </aside>
        {/if}
      </div>
    </main>
  </div>
{/if}

<style>
  :global(*){box-sizing:border-box}:global(html){color-scheme:light;background:#f7f8fa}:global(body){margin:0;font-family:Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;background:#f7f8fa;color:#202124}button,input,textarea{font:inherit}button:focus-visible,input:focus-visible,textarea:focus-visible{outline:2px solid #202124;outline-offset:2px}button:disabled{opacity:.45;cursor:not-allowed}
  .layar-tengah,.layar-masuk{min-height:100vh;display:grid;place-items:center}.layar-tengah{background:#f7f8fa}.muat{display:flex;align-items:center;gap:12px;color:#686c72;font-size:14px}.layar-masuk{padding:24px;background:#f7f8fa}.kartu-masuk{width:min(430px,100%);padding:42px;background:#fff;border:1px solid #e2e5e8;border-radius:20px;box-shadow:0 16px 50px #0000000d}.logo{width:48px;height:48px;display:grid;place-items:center;border-radius:14px;background:#202124;color:#fff;font-weight:800;font-size:20px}.logo.kecil{width:34px;height:34px;border-radius:10px;font-size:15px;flex:0 0 auto}.label-kecil{color:#777b80;font-size:10px;font-weight:800;letter-spacing:.14em}.kartu-masuk .label-kecil{margin-top:22px}.kartu-masuk h1{margin:8px 0;font-size:28px;letter-spacing:-.03em}.kartu-masuk>p{margin:0 0 26px;color:#686c72;line-height:1.55}form{display:grid;gap:15px}label{display:grid;gap:7px;color:#303236;font-size:13px;font-weight:700}label span{color:#8a8e93;font-weight:400}input{min-height:42px;border:1px solid #d5d9dd;border-radius:9px;padding:9px 12px;color:#202124;background:#fff}.tombol{border:0;border-radius:9px;min-height:42px;padding:9px 14px;cursor:pointer}.tombol.utama{background:#202124;color:#fff;font-weight:700}.tombol.sekunder{background:#fff;border:1px solid #d5d9dd;color:#303236}.tombol-teks{display:block;margin:17px auto 0;border:0;background:transparent;color:#4a4f55;cursor:pointer;font-size:13px}.versi{margin-top:25px;text-align:center;color:#a0a4a9;font-size:11px}.galat{margin-bottom:18px;padding:10px 12px;border:1px solid #f1c5c0;border-radius:9px;background:#fce8e6;color:#8a1c13;font-size:13px;white-space:pre-wrap}
  .aplikasi{min-height:100vh;display:grid;grid-template-columns:250px minmax(0,1fr)}.navigasi{min-width:0;padding:20px 14px 14px;background:#f1f3f4;border-right:1px solid #e2e5e8;display:flex;flex-direction:column}.identitas{display:flex;align-items:center;gap:10px;padding:0 8px 20px}.identitas strong,.identitas small{display:block}.identitas strong{font-size:15px}.identitas small{margin-top:2px;color:#777b80;font-size:10px}.buat{width:100%;background:#fff;border:1px solid #e2e5e8}.judul-bagian{margin:22px 8px 8px;color:#777b80;font-size:10px;font-weight:800;letter-spacing:.14em}.daftar-kiri{min-height:0;overflow:auto}.item-kiri{width:100%;padding:10px;border:0;border-radius:9px;background:transparent;text-align:left;cursor:pointer}.item-kiri:hover,.item-kiri.dipilih{background:#e2e5e9}.item-kiri strong,.item-kiri span{display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.item-kiri strong{font-size:12px}.item-kiri span{margin-top:3px;color:#777b80;font-size:10px}.kosong-kecil{padding:10px;color:#8a8e93;font-size:11px}.tambah-catatan{margin-top:10px}.buku-kecil{border-top:1px solid #e2e5e8;margin-top:14px}.buku{width:100%;border:0;border-radius:8px;background:transparent;padding:8px;text-align:left;color:#303236;cursor:pointer;font-size:11px}.buku.aktif{background:#e2e5e9;font-weight:700}.akun{margin-top:auto;padding:12px 5px 0;border-top:1px solid #e2e5e8;display:flex;align-items:center;gap:8px}.avatar{width:30px;height:30px;display:grid;place-items:center;border-radius:50%;background:#202124;color:#fff;font-size:12px;font-weight:700}.akun-nama{min-width:0;flex:1}.akun-nama strong,.akun-nama small{display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.akun-nama strong{font-size:12px}.akun-nama small{color:#777b80;font-size:10px}.akun>button{border:0;background:transparent;color:#686c72;cursor:pointer}.versi-nav{margin:9px 8px 0;color:#a0a4a9;font-size:10px}
  .ruang{min-width:0;background:#fff}.kepala{height:78px;padding:14px 24px;display:flex;justify-content:space-between;align-items:center;border-bottom:1px solid #e2e5e8}.kepala h1{margin:5px 0 0;font-size:20px}.aksi-kepala{display:flex;align-items:center;gap:5px}.mode{border:0;background:transparent;color:#777b80;border-radius:8px;padding:8px 10px;cursor:pointer}.mode.aktif{background:#e8eaed;color:#202124;font-weight:700}.ikon{width:38px;min-height:38px;padding:0;background:#202124;color:#fff;font-size:17px}.isi-utama{min-height:calc(100vh - 78px);display:grid;grid-template-columns:minmax(0,1fr) 300px}.ruang-kerja{min-width:0;padding:34px clamp(20px,6vw,90px);overflow:auto}.konteks{min-width:0;border-left:1px solid #e2e5e8;background:#fbfbfc;display:flex;flex-direction:column}.tab-konteks{display:grid;grid-template-columns:1fr 1fr;border-bottom:1px solid #e2e5e8}.tab-konteks button{border:0;background:transparent;padding:14px 8px;color:#777b80;cursor:pointer}.tab-konteks button.aktif{color:#202124;font-weight:700;border-bottom:2px solid #202124}.konteks-isi{padding:14px;overflow:auto}.judul-panel{display:flex;justify-content:space-between;align-items:center;gap:8px;margin-bottom:12px}.judul-panel strong{font-size:13px}.tambah-sumber{position:relative;overflow:hidden;font-size:11px;padding:7px 9px;min-height:32px}.tambah-sumber input{position:absolute;inset:0;opacity:0;cursor:pointer;min-height:0}.sumber-item{display:flex;align-items:center;border:1px solid #e2e5e8;border-radius:9px;background:#fff;margin-bottom:7px}.sumber-item.aktif{border-color:#9ca0a5}.sumber-item>button:first-child{flex:1;min-width:0;border:0;background:transparent;text-align:left;padding:10px;cursor:pointer}.sumber-item strong,.sumber-item span{display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.sumber-item strong{font-size:11px}.sumber-item span{margin-top:3px;color:#8a8e93;font-size:10px}.hapus-sumber{border:0;background:transparent;color:#8a1c13;padding:8px;cursor:pointer}.kosong-panel{padding:25px 8px;color:#8a8e93;font-size:11px;line-height:1.5;text-align:center}.pratinjau{margin-top:12px;padding:12px;border:1px solid #e2e5e8;border-radius:10px;background:#fff}.pratinjau strong{font-size:11px}.pratinjau p{max-height:130px;overflow:hidden;color:#686c72;font-size:10px;line-height:1.5;white-space:pre-wrap}.bilah-editor{min-height:28px;margin-bottom:12px;display:flex;justify-content:space-between;align-items:center}.status{display:flex;align-items:center;gap:6px;color:#777b80;font-size:11px}.status i{width:6px;height:6px;border-radius:50%;background:#9ca0a5}.status.gagal i{background:#8a1c13}.hapus{border:0;background:transparent;color:#8a1c13;cursor:pointer;font-size:12px}.judul-editor{width:100%;padding:0;border:0;outline:0;font-size:clamp(27px,3vw,38px);font-weight:750;letter-spacing:-.035em;background:transparent}textarea{width:100%;min-height:65vh;margin-top:17px;padding:0;border:0;outline:0;resize:none;background:transparent;color:#303236;font-size:16px;line-height:1.8}.ruang-kosong{max-width:500px;margin:15vh auto;text-align:center;color:#686c72}.ruang-kosong h2{margin:8px 0;color:#202124;font-size:25px}.ruang-kosong p{line-height:1.6;font-size:13px}.ruang-kosong.kecil{margin:10vh auto}.simbol{width:42px;height:42px;margin:0 auto 18px;display:grid;place-items:center;border-radius:10px;background:#e8eaed;color:#4a4f55}.split{display:grid;grid-template-columns:minmax(0,1fr) minmax(260px,.8fr);min-height:calc(100vh - 150px);border:1px solid #e2e5e8;border-radius:12px;overflow:hidden}.sub-ruang{min-width:0;padding:24px;overflow:auto}.chat-ruang{border-left:1px solid #e2e5e8;background:#fbfbfc}.sub-kepala{font-weight:700;font-size:12px;border-bottom:1px solid #e2e5e8;padding-bottom:12px}.preview h2{margin:10px 0 4px;font-size:26px}.metadata{color:#8a8e93;font-size:11px}.preview pre{white-space:pre-wrap;word-break:break-word;font:inherit;line-height:1.7;color:#303236}
  @media(max-width:1050px){.aplikasi{grid-template-columns:220px minmax(0,1fr)}.isi-utama{grid-template-columns:minmax(0,1fr) 270px}.split{grid-template-columns:1fr}.chat-ruang{border-left:0;border-top:1px solid #e2e5e8}.ruang-kerja{padding:28px}.konteks{position:fixed;right:0;top:78px;bottom:0;width:min(340px,90vw);z-index:3;box-shadow:-8px 0 25px #00000012}}
  @media(max-width:760px){.aplikasi{display:block}.navigasi{min-height:auto;padding:12px 14px;border-right:0;border-bottom:1px solid #e2e5e8}.identitas{padding-bottom:10px}.daftar-kiri,.buku-kecil,.judul-bagian{display:none}.tambah-catatan{width:auto;align-self:flex-start}.akun{margin-top:10px;padding-top:10px}.kepala{height:72px;padding:12px 14px}.kepala h1{font-size:17px}.mode{display:none}.isi-utama{display:block;min-height:calc(100vh - 72px)}.ruang-kerja{min-height:calc(100vh - 150px);padding:24px 16px}.konteks{top:72px;width:100%;box-shadow:none}.split{min-height:calc(100vh - 170px)}textarea{min-height:55vh}.preview pre{font-size:13px}}
  @media(prefers-reduced-motion:reduce){*,*::before,*::after{scroll-behavior:auto!important;transition-duration:.01ms!important;animation-duration:.01ms!important}}
</style>
