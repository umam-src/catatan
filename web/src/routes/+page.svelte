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

  async function mulai() {
    loading = true;
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
      user = null; siap = false; notebooks = []; notes = []; activeNote = null; selected = '';
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
    if (!notebookID) { notes = []; activeNote = null; return; }
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
    if (!activeNote) return;
    const catatan = { ...activeNote };
    try {
      const response = await api('/api/notes/' + catatan.id, { method: 'PUT', body: JSON.stringify({ title: catatan.title, content: catatan.content }) });
      if (nomor !== nomorSimpan) return;
      statusSimpan = 'tersimpan';
      const data = await response.json();
      notes = notes.map((item) => item.id === catatan.id ? { ...item, ...catatan, updated_at: data.updated_at } : item);
      error = '';
    } catch (err) {
      if (nomor !== nomorSimpan) return;
      statusSimpan = 'gagal';
      error = err instanceof Error ? err.message : 'Gagal menyimpan catatan.';
    }
  }

  async function hapusCatatan() {
    if (!activeNote || !window.confirm('Hapus catatan ini?')) return;
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

<svelte:head><title>Buku Catatan</title></svelte:head>

{#if loading}
  <main class="center"><p>Memuat…</p></main>
{:else if !siap}
  <main class="auth">
    <section class="auth-card">
      <div class="mark">B</div>
      <h1>Buku Catatan</h1>
      <p>{mode === 'setup' ? 'Siapkan akun lokal pertama.' : 'Masuk untuk membuka catatan di perangkat ini.'}</p>
      {#if error}<div class="error" role="alert">{error}</div>{/if}
      <form onsubmit={kirimForm}>
        <label>Nama pengguna<input bind:value={username} autocomplete="username" required /></label>
        {#if mode === 'setup'}
          <label>Email <span>(opsional)</span><input bind:value={email} type="email" autocomplete="email" /></label>
          <label>Nama tampilan <span>(opsional)</span><input bind:value={displayName} autocomplete="name" /></label>
        {/if}
        <label>Kata sandi<input bind:value={password} type="password" minlength="12" autocomplete={mode === 'setup' ? 'new-password' : 'current-password'} required /></label>
        <button class="primary" type="submit">{mode === 'setup' ? 'Siapkan akun' : 'Masuk'}</button>
      </form>
      <button class="link" onclick={() => { mode = mode === 'login' ? 'setup' : 'login'; error = ''; }}> {mode === 'login' ? 'Pengaturan awal' : 'Kembali ke masuk'} </button>
    </section>
  </main>
{:else}
  <div class="app">
    <aside class="sidebar">
      <div class="brand"><span class="mark small">B</span><strong>Buku Catatan</strong></div>
      <button class="new" onclick={createNotebook}>＋ Buku baru</button>
      <div class="section-label">BUKU</div>
      {#each notebooks as notebook}
        <button class:active={selected === notebook.id} class="book" onclick={() => selectNotebook(notebook.id)}>{notebook.title}</button>
      {/each}
      <div class="account"><span>{user?.display_name || user?.username}</span><button onclick={keluar}>Keluar</button></div>
    </aside>

    <main class="workspace">
      <header><div><div class="eyebrow">RUANG CATATAN</div><h2>{notebooks.find((item) => item.id === selected)?.title ?? 'Buku Catatan'}</h2></div><button class="icon" onclick={createNote} title="Catatan baru">＋</button></header>
      <div class="body">
        <nav class="notes" aria-label="Daftar catatan">
          <div class="notes-title"><span>Catatan</span><button onclick={createNote} title="Catatan baru">＋</button></div>
          {#if !notes.length}<div class="state">Belum ada catatan.</div>{:else}{#each notes as note}<button class:chosen={activeNote?.id === note.id} class="note" onclick={() => pilihCatatan(note)}><strong>{note.title || 'Tanpa judul'}</strong><small>{note.content.slice(0, 72) || 'Belum ada isi'}</small></button>{/each}{/if}
        </nav>
        <section class="editor" aria-live="polite">
          {#if error}<div class="error" role="alert">{error}</div>{/if}
          {#if activeNote}
            <div class="editor-toolbar"><span class="save-state">{statusSimpan === 'menyimpan' ? 'Menyimpan…' : statusSimpan === 'gagal' ? 'Gagal menyimpan' : 'Tersimpan'}</span><button class="delete" onclick={hapusCatatan}>Hapus</button></div>
            <input class="title" bind:value={activeNote.title} oninput={jadwalkanSimpan} aria-label="Judul catatan" />
            <textarea bind:value={activeNote.content} oninput={jadwalkanSimpan} placeholder="Mulai menulis…" aria-label="Isi catatan"></textarea>
          {:else}
            <div class="empty"><div class="empty-mark">✦</div><h2>Mulai mencatat</h2><p>Buat catatan teks yang tetap berada di perangkat ini.</p><button class="new" onclick={createNote}>＋ Catatan baru</button></div>
          {/if}
        </section>
      </div>
    </main>
  </div>
{/if}

<style>
  :global(*) { box-sizing: border-box; }
  :global(body) { margin: 0; font-family: Inter, ui-sans-serif, system-ui, sans-serif; background: #f7f8fa; color: #202124; }
  button, input, textarea { font: inherit; }
  button:focus-visible, input:focus-visible, textarea:focus-visible { outline: 2px solid #202124; outline-offset: 2px; }
  .center, .auth { min-height: 100vh; display: grid; place-items: center; }
  .auth { background: #f7f8fa; padding: 24px; }
  .auth-card { width: min(420px, 100%); padding: 36px; background: #fff; border: 1px solid #e2e5e8; border-radius: 16px; box-shadow: 0 8px 30px #0000000d; }
  .auth-card h1 { margin: 16px 0 8px; }
  .auth-card p { color: #686c72; margin: 0 0 24px; line-height: 1.5; }
  form { display: grid; gap: 14px; }
  label { display: grid; gap: 6px; font-size: 14px; font-weight: 600; }
  label span { color: #777; font-weight: 400; }
  input { border: 1px solid #d5d9dd; border-radius: 8px; padding: 10px 12px; background: #fff; }
  .primary { border: 0; border-radius: 8px; padding: 11px 14px; background: #202124; color: #fff; cursor: pointer; }
  .link { margin-top: 16px; border: 0; background: transparent; color: #4a4f55; cursor: pointer; }
  .app { min-height: 100vh; display: grid; grid-template-columns: 250px 1fr; }
  .sidebar { padding: 24px 16px; background: #f1f3f4; border-right: 1px solid #e2e5e8; display: flex; flex-direction: column; }
  .brand { display: flex; align-items: center; gap: 10px; margin: 0 8px 24px; }
  .mark { width: 42px; height: 42px; border-radius: 12px; display: grid; place-items: center; background: #202124; color: #fff; font-weight: 700; }
  .mark.small { width: 32px; height: 32px; border-radius: 10px; }
  .new { border: 0; border-radius: 20px; padding: 10px 15px; background: #fff; box-shadow: 0 1px 4px #0001; cursor: pointer; }
  .section-label, .eyebrow { font-size: 11px; letter-spacing: .12em; color: #73777c; font-weight: 700; }
  .section-label { margin: 28px 8px 8px; }
  .book, .note, .icon, .notes-title button { width: 100%; border: 0; background: transparent; text-align: left; cursor: pointer; }
  .book { padding: 10px 12px; border-radius: 9px; }
  .book.active { background: #e2e5e9; font-weight: 600; }
  .account { margin-top: auto; padding: 16px 8px 0; display: flex; gap: 10px; justify-content: space-between; align-items: center; font-size: 13px; }
  .account button { border: 0; background: transparent; cursor: pointer; color: #8a1c13; }
  .workspace { min-width: 0; }
  header { height: 92px; padding: 20px 34px; display: flex; justify-content: space-between; align-items: center; background: #fff; border-bottom: 1px solid #e2e5e8; }
  h2 { font-size: 22px; margin: 5px 0 0; }
  .icon { width: 40px; text-align: center; font-size: 25px; }
  .body { height: calc(100vh - 92px); display: grid; grid-template-columns: 260px 1fr; }
  .notes { padding: 18px 12px; background: #fbfbfc; border-right: 1px solid #e2e5e8; overflow: auto; }
  .notes-title { display: flex; justify-content: space-between; align-items: center; padding: 6px 10px 12px; font-weight: 700; }
  .notes-title button { width: auto; font-size: 20px; }
  .note { padding: 12px; border-radius: 10px; display: block; }
  .note.chosen { background: #e8eaed; }
  .note strong, .note small { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .note small { color: #777; margin-top: 4px; }
  .state { padding: 12px 10px; color: #777; font-size: 14px; }
  .editor { padding: 52px clamp(24px, 8vw, 120px); background: #fff; overflow: auto; }
  .error { margin-bottom: 20px; padding: 10px 12px; border-radius: 8px; background: #fce8e6; color: #8a1c13; font-size: 14px; white-space: pre-wrap; }
  .editor-toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 14px; }
  .save-state { color: #686c72; font-size: 13px; }
  .delete { border: 0; background: transparent; color: #8a1c13; cursor: pointer; padding: 8px; }
  .title { width: 100%; border: 0; outline: 0; font-size: 32px; font-weight: 700; margin-bottom: 20px; background: transparent; }
  textarea { width: 100%; min-height: 65vh; resize: none; border: 0; outline: 0; line-height: 1.8; font-size: 16px; background: transparent; }
  .empty { max-width: 460px; margin: 14vh auto; text-align: center; color: #686c72; }
  .empty-mark { width: 32px; height: 32px; margin: auto; border-radius: 10px; display: grid; place-items: center; background: #202124; color: #fff; }
  .empty h2 { color: #202124; margin-bottom: 8px; }
  @media (max-width: 760px) { .app { grid-template-columns: 1fr; } .sidebar { display: none; } .body { grid-template-columns: 1fr; } .notes { display: none; } .editor { padding: 32px 20px; } header { padding: 18px 20px; } }
</style>
