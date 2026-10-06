<script lang="ts">
  type User = { id: string; username: string; display_name: string };
  type Notebook = { id: string; title: string; description: string };
  type Note = { id: string; title: string; content: string; updated_at: string };
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
  let selected = '';
  let activeNote: Note | null = null;
  let statusSimpan: StatusSimpan = 'tersimpan';
  let timerSimpan: ReturnType<typeof setTimeout> | undefined;
  let nomorSimpan = 0;

  async function api(path: string, init: RequestInit = {}) {
    const headers = new Headers(init.headers);
    if (init.body) headers.set('Content-Type', 'application/json');
    const response = await fetch(path, { ...init, headers });
    if (!response.ok) throw new Error(await response.text() || 'Permintaan gagal.');
    return response;
  }

  async function muatVersi() {
    try {
      const response = await fetch('/api/version');
      if (response.ok) {
        const data = await response.json();
        versi = data.version ?? '';
      }
    } catch {
      versi = '';
    }
  }

  async function mulai() {
    loading = true;
    await muatVersi();
    try {
      const response = await fetch('/api/auth/me');
      if (response.ok) {
        const data = await response.json();
        user = data.user;
        siap = true;
        await loadNotebooks();
      }
    } catch {
      error = 'Layanan lokal tidak dapat dihubungi.';
    } finally {
      loading = false;
    }
  }

  function kirimForm(event: SubmitEvent) {
    event.preventDefault();
    void (mode === 'setup' ? siapkanAkun() : masuk());
  }

  async function masuk() {
    error = '';
    try {
      const response = await api('/api/auth/login', { method: 'POST', body: JSON.stringify({ username, password }) });
      const data = await response.json();
      user = data.user;
      password = '';
      siap = true;
      await loadNotebooks();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Gagal masuk.';
    }
  }

  async function siapkanAkun() {
    error = '';
    try {
      const response = await api('/api/auth/setup', { method: 'POST', body: JSON.stringify({ username, email, display_name: displayName, password }) });
      const data = await response.json();
      user = data.user;
      password = '';
      siap = true;
      await loadNotebooks();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Gagal menyiapkan akun.';
    }
  }

  async function keluar() {
    try { await fetch('/api/auth/logout', { method: 'POST' }); } finally {
      user = null;
      siap = false;
      notebooks = [];
      notes = [];
      activeNote = null;
      selected = '';
    }
  }

  async function loadNotebooks() {
    try {
      const response = await api('/api/notebooks');
      notebooks = await response.json();
      if (!selected && notebooks.length) selected = notebooks[0].id;
      if (selected) await loadNotes(selected);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Gagal memuat buku.';
    }
  }

  async function loadNotes(notebookID = selected) {
    if (!notebookID) {
      notes = [];
      activeNote = null;
      return;
    }
    try {
      const response = await api('/api/notebooks/' + notebookID + '/notes');
      const loaded: Note[] = await response.json();
      notes = loaded;
      activeNote = loaded[0] ? { ...loaded[0] } : null;
      statusSimpan = 'tersimpan';
    } catch (err) {
      error = err instanceof Error ? err.message : 'Gagal memuat catatan.';
    }
  }

  async function selectNotebook(id: string) {
    if (id === selected) return;
    if (timerSimpan) clearTimeout(timerSimpan);
    selected = id;
    await loadNotes(id);
  }

  async function createNotebook() {
    const title = window.prompt('Nama buku');
    if (!title?.trim()) return;
    try {
      const response = await api('/api/notebooks', { method: 'POST', body: JSON.stringify({ title: title.trim(), description: '' }) });
      const notebook = await response.json();
      notebooks = [notebook, ...notebooks];
      selected = notebook.id;
      await loadNotes(notebook.id);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Gagal membuat buku.';
    }
  }

  async function createNote() {
    if (!selected) return;
    try {
      const response = await api('/api/notebooks/' + selected + '/notes', { method: 'POST', body: JSON.stringify({ title: 'Catatan baru', content: '' }) });
      const created: Note = await response.json();
      activeNote = { ...created };
      notes = [created, ...notes];
      statusSimpan = 'tersimpan';
    } catch (err) {
      error = err instanceof Error ? err.message : 'Gagal membuat catatan.';
    }
  }

  function pilihCatatan(note: Note) {
    if (timerSimpan) clearTimeout(timerSimpan);
    activeNote = { ...note };
    statusSimpan = 'tersimpan';
    error = '';
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
      const response = await api('/api/notes/' + catatan.id, { method: 'PUT', body: JSON.stringify({ title: catatan.title, content: catatan.content }) });
      if (nomor !== nomorSimpan || activeNote?.id !== catatan.id) return;
      statusSimpan = 'tersimpan';
      const data = await response.json();
      notes = notes.map((item) => item.id === catatan.id ? { ...item, ...catatan, updated_at: data.updated_at } : item);
      error = '';
    } catch (err) {
      if (nomor !== nomorSimpan || activeNote?.id !== catatan.id) return;
      statusSimpan = 'gagal';
      error = err instanceof Error ? err.message : 'Gagal menyimpan catatan.';
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
    } catch (err) {
      error = err instanceof Error ? err.message : 'Gagal menghapus catatan.';
    }
  }

  mulai();
</script>

<svelte:head>
  <title>{versi ? `Catatan ${versi}` : 'Catatan'}</title>
  <meta name="description" content="Catatan lokal yang tetap berada di perangkat ini." />
</svelte:head>

{#if loading}
  <main class="layar-tengah">
    <div class="muat"><span class="logo kecil">C</span><span>Memuat Catatan…</span></div>
  </main>
{:else if !siap}
  <main class="layar-masuk">
    <section class="kartu-masuk">
      <div class="logo">C</div>
      <div class="label-kecil">CATATAN LOKAL</div>
      <h1>{mode === 'setup' ? 'Mulai dengan Catatan' : 'Selamat datang kembali'}</h1>
      <p>{mode === 'setup' ? 'Buat akun lokal untuk menyimpan catatan di perangkat ini.' : 'Masuk untuk melanjutkan ke catatan Anda.'}</p>

      {#if error}<div class="galat" role="alert">{error}</div>{/if}

      <form onsubmit={kirimForm}>
        <label>Nama pengguna
          <input bind:value={username} autocomplete="username" required />
        </label>
        {#if mode === 'setup'}
          <label>Email <span>opsional</span>
            <input bind:value={email} type="email" autocomplete="email" />
          </label>
          <label>Nama tampilan <span>opsional</span>
            <input bind:value={displayName} autocomplete="name" />
          </label>
        {/if}
        <label>Kata sandi
          <input bind:value={password} type="password" minlength="12" autocomplete={mode === 'setup' ? 'new-password' : 'current-password'} required />
        </label>
        <button class="tombol utama" type="submit">{mode === 'setup' ? 'Buat akun lokal' : 'Masuk'}</button>
      </form>

      <button class="tombol-teks" onclick={() => { mode = mode === 'login' ? 'setup' : 'login'; error = ''; }}>
        {mode === 'login' ? 'Ini perangkat baru? Siapkan akun' : 'Sudah punya akun? Masuk'}
      </button>
      {#if versi}<div class="versi">v{versi}</div>{/if}
    </section>
  </main>
{:else}
  <div class="aplikasi">
    <aside class="navigasi">
      <div class="identitas">
        <span class="logo kecil">C</span>
        <div><strong>Catatan</strong><small>Lokal, sederhana, milik Anda</small></div>
      </div>

      <button class="tombol buat" onclick={createNotebook}>＋ Buku baru</button>

      <div class="judul-bagian">BUKU</div>
      <nav aria-label="Daftar buku">
        {#each notebooks as notebook}
          <button class:aktif={selected === notebook.id} class="buku" onclick={() => selectNotebook(notebook.id)}>
            <span>{notebook.title}</span>
          </button>
        {:else}
          <div class="kosong-nav">Belum ada buku.</div>
        {/each}
      </nav>

      <div class="akun">
        <div class="avatar">{(user?.display_name || user?.username || 'C').slice(0, 1).toUpperCase()}</div>
        <div class="akun-nama"><strong>{user?.display_name || user?.username}</strong><small>Perangkat ini</small></div>
        <button aria-label="Keluar" title="Keluar" onclick={keluar}>↪</button>
      </div>
      {#if versi}<div class="versi-nav">v{versi}</div>{/if}
    </aside>

    <main class="ruang">
      <header class="kepala">
        <div>
          <div class="label-kecil">RUANG CATATAN</div>
          <h1>{notebooks.find((item) => item.id === selected)?.title ?? 'Catatan'}</h1>
        </div>
        <button class="tombol ikon" onclick={createNote} disabled={!selected} title="Catatan baru" aria-label="Catatan baru">＋</button>
      </header>

      <div class="isi">
        <nav class="daftar" aria-label="Daftar catatan">
          <div class="kepala-daftar">
            <div><strong>Catatan</strong><span>{notes.length} catatan</span></div>
            <button class="ikon kecil-ikon" onclick={createNote} disabled={!selected} title="Catatan baru" aria-label="Catatan baru">＋</button>
          </div>
          {#if !notes.length}
            <div class="kosong">
              <span class="simbol">✦</span>
              <strong>Belum ada catatan</strong>
              <p>Mulai dengan membuat catatan pertama di buku ini.</p>
              <button class="tombol sekunder" onclick={createNote} disabled={!selected}>＋ Catatan baru</button>
            </div>
          {:else}
            {#each notes as note}
              <button class:dipilih={activeNote?.id === note.id} class="item-catatan" onclick={() => pilihCatatan(note)}>
                <strong>{note.title || 'Tanpa judul'}</strong>
                <span>{note.content.slice(0, 90) || 'Belum ada isi'}</span>
              </button>
            {/each}
          {/if}
        </nav>

        <section class="editor" aria-live="polite">
          {#if error}<div class="galat" role="alert">{error}</div>{/if}
          {#if activeNote}
            <div class="bilah-editor">
              <span class="status {statusSimpan}">
                <i></i>{statusSimpan === 'menyimpan' ? 'Menyimpan…' : statusSimpan === 'gagal' ? 'Gagal menyimpan' : 'Tersimpan'}
              </span>
              <button class="hapus" onclick={hapusCatatan}>Hapus</button>
            </div>
            <input class="judul-editor" bind:value={activeNote.title} oninput={jadwalkanSimpan} aria-label="Judul catatan" />
            <textarea bind:value={activeNote.content} oninput={jadwalkanSimpan} placeholder="Mulai menulis…" aria-label="Isi catatan"></textarea>
          {:else}
            <div class="editor-kosong">
              <span class="simbol besar">✦</span>
              <div class="label-kecil">RUANG MENULIS</div>
              <h2>Pilih atau buat catatan</h2>
              <p>Catatan disimpan di perangkat ini dan tetap dapat digunakan tanpa jaringan.</p>
              <button class="tombol utama" onclick={createNote} disabled={!selected}>＋ Catatan baru</button>
            </div>
          {/if}
        </section>
      </div>
    </main>
  </div>
{/if}

<style>
  :global(*) { box-sizing: border-box; }
  :global(html) { color-scheme: light; background: #f7f8fa; }
  :global(body) { margin: 0; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #f7f8fa; color: #202124; }
  button, input, textarea { font: inherit; }
  button:focus-visible, input:focus-visible, textarea:focus-visible { outline: 2px solid #202124; outline-offset: 2px; }
  button:disabled { opacity: .45; cursor: not-allowed; }

  .layar-tengah, .layar-masuk { min-height: 100vh; display: grid; place-items: center; }
  .layar-tengah { background: #f7f8fa; }
  .muat { display: flex; align-items: center; gap: 12px; color: #686c72; font-size: 14px; }
  .layar-masuk { padding: 24px; background: radial-gradient(circle at top, #fff, #f7f8fa 60%); }
  .kartu-masuk { width: min(430px, 100%); padding: 42px; background: #fff; border: 1px solid #e2e5e8; border-radius: 20px; box-shadow: 0 16px 50px #0000000d; }
  .logo { width: 48px; height: 48px; display: grid; place-items: center; border-radius: 14px; background: #202124; color: #fff; font-weight: 800; font-size: 20px; }
  .logo.kecil { width: 34px; height: 34px; border-radius: 10px; font-size: 15px; flex: 0 0 auto; }
  .label-kecil { margin-top: 22px; color: #777b80; font-size: 10px; font-weight: 800; letter-spacing: .14em; }
  .kartu-masuk h1 { margin: 8px 0; font-size: 28px; letter-spacing: -.03em; }
  .kartu-masuk > p { margin: 0 0 26px; color: #686c72; line-height: 1.55; }
  form { display: grid; gap: 15px; }
  label { display: grid; gap: 7px; color: #303236; font-size: 13px; font-weight: 700; }
  label span { color: #8a8e93; font-weight: 400; }
  input { min-height: 42px; border: 1px solid #d5d9dd; border-radius: 9px; padding: 9px 12px; color: #202124; background: #fff; }
  input:focus { border-color: #686c72; }
  .tombol { border: 0; border-radius: 9px; min-height: 42px; padding: 9px 14px; cursor: pointer; }
  .tombol.utama { background: #202124; color: #fff; font-weight: 700; }
  .tombol.sekunder { background: #fff; border: 1px solid #d5d9dd; color: #303236; }
  .tombol-teks { display: block; margin: 17px auto 0; border: 0; background: transparent; color: #4a4f55; cursor: pointer; font-size: 13px; }
  .versi { margin-top: 25px; text-align: center; color: #a0a4a9; font-size: 11px; }
  .galat { margin-bottom: 18px; padding: 10px 12px; border: 1px solid #f1c5c0; border-radius: 9px; background: #fce8e6; color: #8a1c13; font-size: 13px; white-space: pre-wrap; }

  .aplikasi { min-height: 100vh; display: grid; grid-template-columns: 268px minmax(0, 1fr); }
  .navigasi { min-width: 0; padding: 22px 15px 15px; background: #f1f3f4; border-right: 1px solid #e2e5e8; display: flex; flex-direction: column; }
  .identitas { display: flex; align-items: center; gap: 10px; padding: 0 8px 22px; }
  .identitas strong, .identitas small { display: block; }
  .identitas strong { font-size: 15px; }
  .identitas small { margin-top: 2px; color: #777b80; font-size: 10px; }
  .buat { width: 100%; background: #fff; border: 1px solid #e2e5e8; box-shadow: 0 2px 8px #00000008; text-align: left; }
  .judul-bagian { margin: 28px 9px 8px; color: #777b80; font-size: 10px; font-weight: 800; letter-spacing: .14em; }
  .buku { width: 100%; min-height: 39px; padding: 9px 12px; border: 0; border-radius: 9px; background: transparent; color: #303236; text-align: left; cursor: pointer; }
  .buku:hover { background: #f7f8fa; }
  .buku.aktif { background: #e2e5e9; font-weight: 700; }
  .kosong-nav { padding: 9px; color: #8a8e93; font-size: 12px; }
  .akun { margin-top: auto; padding: 14px 6px 0; border-top: 1px solid #e2e5e8; display: flex; align-items: center; gap: 9px; }
  .avatar { width: 30px; height: 30px; display: grid; place-items: center; border-radius: 50%; background: #202124; color: #fff; font-size: 12px; font-weight: 700; }
  .akun-nama { min-width: 0; flex: 1; }
  .akun-nama strong, .akun-nama small { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .akun-nama strong { font-size: 12px; }
  .akun-nama small { margin-top: 2px; color: #777b80; font-size: 10px; }
  .akun > button { width: 32px; height: 32px; border: 0; border-radius: 8px; background: transparent; color: #686c72; cursor: pointer; }
  .versi-nav { margin: 10px 9px 0; color: #a0a4a9; font-size: 10px; }

  .ruang { min-width: 0; background: #fff; }
  .kepala { height: 82px; padding: 17px 30px; display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #e2e5e8; }
  .kepala .label-kecil { margin: 0; }
  .kepala h1 { margin: 5px 0 0; font-size: 21px; letter-spacing: -.02em; }
  .ikon { width: 40px; min-height: 40px; padding: 0; background: #202124; color: #fff; border-radius: 10px; font-size: 22px; }
  .isi { height: calc(100vh - 82px); display: grid; grid-template-columns: 280px minmax(0, 1fr); }
  .daftar { min-width: 0; padding: 17px 12px; overflow: auto; background: #fbfbfc; border-right: 1px solid #e2e5e8; }
  .kepala-daftar { padding: 3px 8px 12px; display: flex; align-items: center; justify-content: space-between; }
  .kepala-daftar strong, .kepala-daftar span { display: block; }
  .kepala-daftar strong { font-size: 13px; }
  .kepala-daftar span { margin-top: 2px; color: #8a8e93; font-size: 10px; }
  .kecil-ikon { width: 30px; min-height: 30px; background: transparent; color: #202124; font-size: 20px; }
  .item-catatan { width: 100%; padding: 12px; border: 0; border-radius: 10px; background: transparent; text-align: left; cursor: pointer; }
  .item-catatan:hover { background: #f1f3f4; }
  .item-catatan.dipilih { background: #e8eaed; }
  .item-catatan strong, .item-catatan span { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .item-catatan strong { font-size: 13px; }
  .item-catatan span { margin-top: 4px; color: #777b80; font-size: 11px; }
  .kosong { margin: 38px 8px; color: #777b80; text-align: center; }
  .kosong .simbol { margin: 0 auto 10px; }
  .kosong strong { display: block; color: #4a4f55; font-size: 12px; }
  .kosong p { font-size: 11px; line-height: 1.5; }
  .simbol { width: 28px; height: 28px; display: grid; place-items: center; border-radius: 9px; background: #e8eaed; color: #4a4f55; }
  .editor { min-width: 0; padding: 43px clamp(22px, 7vw, 110px); overflow: auto; }
  .bilah-editor { min-height: 28px; margin-bottom: 10px; display: flex; justify-content: space-between; align-items: center; }
  .status { display: flex; align-items: center; gap: 6px; color: #777b80; font-size: 11px; }
  .status i { width: 6px; height: 6px; border-radius: 50%; background: #9ca0a5; }
  .status.menyimpan i { background: #686c72; }
  .status.gagal i { background: #8a1c13; }
  .hapus { border: 0; background: transparent; color: #8a1c13; cursor: pointer; font-size: 12px; }
  .judul-editor { width: 100%; padding: 0; border: 0; outline: 0; font-size: clamp(27px, 3vw, 38px); font-weight: 750; letter-spacing: -.035em; background: transparent; }
  textarea { width: 100%; min-height: 65vh; margin-top: 17px; padding: 0; border: 0; outline: 0; resize: none; background: transparent; color: #303236; font-size: 16px; line-height: 1.8; }
  .editor-kosong { max-width: 480px; margin: 14vh auto; text-align: center; color: #686c72; }
  .editor-kosong .besar { width: 42px; height: 42px; margin: 0 auto 18px; }
  .editor-kosong .label-kecil { margin: 0; }
  .editor-kosong h2 { margin: 7px 0; color: #202124; font-size: 25px; }
  .editor-kosong p { margin: 0 auto 20px; max-width: 380px; font-size: 13px; line-height: 1.6; }

  @media (max-width: 900px) {
    .aplikasi { grid-template-columns: 220px minmax(0, 1fr); }
    .isi { grid-template-columns: 240px minmax(0, 1fr); }
    .editor { padding: 36px 30px; }
  }
  @media (max-width: 700px) {
    .aplikasi { display: block; }
    .navigasi { min-height: auto; padding: 12px 14px; border-right: 0; border-bottom: 1px solid #e2e5e8; }
    .identitas { padding-bottom: 10px; }
    .navigasi .buat { width: auto; align-self: flex-start; }
    .judul-bagian, .navigasi nav { display: none; }
    .akun { margin-top: 10px; padding-top: 10px; }
    .versi-nav { display: none; }
    .kepala { height: 72px; padding: 14px 18px; }
    .kepala h1 { font-size: 18px; }
    .isi { height: auto; min-height: calc(100vh - 72px); display: block; }
    .daftar { max-height: 180px; border-right: 0; border-bottom: 1px solid #e2e5e8; }
    .editor { min-height: calc(100vh - 300px); padding: 28px 18px 40px; }
    .editor-kosong { margin: 10vh auto; }
  }
  @media (prefers-reduced-motion: reduce) {
    *, *::before, *::after { scroll-behavior: auto !important; transition-duration: .01ms !important; animation-duration: .01ms !important; }
  }
</style>
