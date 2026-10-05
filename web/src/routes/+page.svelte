<script lang="ts">
  type Notebook = { id: string; title: string; description: string };
  type Note = { id: string; title: string; content: string; updated_at: string };
  type StatusSimpan = 'tersimpan' | 'menyimpan' | 'gagal';

  let notebooks: Notebook[] = [];
  let notes: Note[] = [];
  let selected = '';
  let activeNote: Note | null = null;
  let loading = true;
  let error = '';
  let notesRequest = 0;
  let statusSimpan: StatusSimpan = 'tersimpan';
  let timerSimpan: ReturnType<typeof setTimeout> | undefined;
  let nomorSimpan = 0;

  async function loadNotebooks() {
    loading = true; error = '';
    try {
      const response = await fetch('/api/notebooks');
      if (!response.ok) throw new Error('Gagal memuat buku.');
      notebooks = await response.json();
      if (!selected && notebooks.length) selected = notebooks[0].id;
      if (selected) await loadNotes(selected);
    } catch (err) { error = err instanceof Error ? err.message : 'Gagal memuat data.'; }
    finally { loading = false; }
  }

  async function loadNotes(notebookID = selected) {
    if (!notebookID) { notes = []; activeNote = null; return; }
    const request = ++notesRequest; error = '';
    try {
      const response = await fetch(`/api/notebooks/${notebookID}/notes`);
      if (!response.ok) throw new Error('Gagal memuat catatan.');
      const loaded: Note[] = await response.json();
      if (request !== notesRequest || notebookID !== selected) return;
      notes = loaded; activeNote = loaded[0] ? { ...loaded[0] } : null;
      statusSimpan = 'tersimpan';
    } catch (err) {
      if (request !== notesRequest) return;
      notes = []; activeNote = null; error = err instanceof Error ? err.message : 'Gagal memuat catatan.';
    }
  }

  async function selectNotebook(id: string) { if (id === selected) return; selected = id; await loadNotes(id); }

  async function createNotebook() {
    const title = prompt('Nama buku'); if (!title?.trim()) return;
    const response = await fetch('/api/notebooks', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({title, description:''}) });
    if (!response.ok) { error = 'Gagal membuat buku.'; return; }
    const notebook = await response.json(); notebooks = [notebook, ...notebooks]; selected = notebook.id; await loadNotes(notebook.id);
  }

  function jadwalkanSimpan() {
    if (!activeNote) return; statusSimpan = 'menyimpan';
    if (timerSimpan) clearTimeout(timerSimpan);
    const nomor = ++nomorSimpan; timerSimpan = setTimeout(() => void simpanCatatan(nomor), 700);
  }

  async function simpanCatatan(nomor = ++nomorSimpan) {
    if (!activeNote) return;
    const catatan = { ...activeNote };
    const response = await fetch(`/api/notes/${catatan.id}`, { method:'PUT', headers:{'Content-Type':'application/json'}, body:JSON.stringify({title:catatan.title, content:catatan.content}) });
    if (nomor !== nomorSimpan) return;
    if (!response.ok) { statusSimpan = 'gagal'; error = 'Gagal menyimpan catatan.'; return; }
    statusSimpan = 'tersimpan'; error = '';
    const data: { updated_at: string } = await response.json();
    notes = notes.map((item) => item.id === catatan.id ? { ...item, ...catatan, updated_at: data.updated_at } : item);
  }

  function pilihCatatan(catatan: Note) { if (timerSimpan) clearTimeout(timerSimpan); activeNote = { ...catatan }; statusSimpan = 'tersimpan'; error = ''; }

  async function hapusCatatan() {
    if (!activeNote || !confirm('Hapus catatan ini?')) return;
    if (timerSimpan) clearTimeout(timerSimpan);
    const id = activeNote.id; const response = await fetch(`/api/notes/${id}`, { method:'DELETE' });
    if (!response.ok) { error = 'Gagal menghapus catatan.'; return; }
    notes = notes.filter((item) => item.id !== id); activeNote = notes[0] ? { ...notes[0] } : null; statusSimpan = 'tersimpan';
  }

  async function createNote() {
    if (!selected) return;
    const response = await fetch(`/api/notebooks/${selected}/notes`, { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({title:'Catatan baru', content:''}) });
    if (!response.ok) { error = 'Gagal membuat catatan.'; return; }
    const created: Note = await response.json(); activeNote = { ...created }; notes = [created, ...notes]; statusSimpan = 'tersimpan';
  }

  loadNotebooks();
</script>
<svelte:head><title>Buku Catatan</title></svelte:head>

<div class="app">
  <aside class="sidebar">
    <div class="brand"><span class="mark">B</span><span>Buku Catatan</span></div>
    <button class="new" onclick={createNotebook}>＋ Buku baru</button>
    <div class="section-label">BUKU</div>
    {#each notebooks as notebook}<button class:active={selected === notebook.id} class="book" onclick={() => selectNotebook(notebook.id)}>{notebook.title}</button>{/each}
  </aside>
  <main class="workspace">
    <header><div><div class="eyebrow">RUANG CATATAN</div><h1>{notebooks.find(n => n.id === selected)?.title ?? 'Buku Catatan'}</h1></div><button class="icon" onclick={createNote} title="Catatan baru">＋</button></header>
    <div class="body">
      <nav class="notes" aria-label="Daftar catatan">
        <div class="notes-title"><span>Catatan</span><button onclick={createNote} title="Catatan baru">＋</button></div>
        {#if loading}<div class="state">Memuat…</div>{:else if !notes.length}<div class="state">Belum ada catatan.</div>{:else}{#each notes as note}<button class:chosen={activeNote?.id === note.id} class="note" onclick={() => pilihCatatan(note)}><strong>{note.title || 'Tanpa judul'}</strong><small>{note.content.slice(0, 72) || 'Belum ada isi'}</small></button>{/each}{/if}
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



<svelte:head><title>Buku Catatan</title></svelte:head>

<div class="app">
  <aside class="sidebar">
    <div class="brand"><span class="mark">B</span><span>Buku Catatan</span></div>
    <button class="new" onclick={createNotebook}>＋ Buku baru</button>
    <div class="section-label">BUKU</div>
    {#each notebooks as notebook}
      <button class:active={selected === notebook.id} class="book" onclick={() => selectNotebook(notebook.id)}>{notebook.title}</button>
    {/each}
  </aside>

  <main class="workspace">
    <header><div><div class="eyebrow">RUANG CATATAN</div><h1>{notebooks.find(n => n.id === selected)?.title ?? 'Buku Catatan'}</h1></div><button class="icon" onclick={createNote} title="Catatan baru">＋</button></header>
    <div class="body">
      <nav class="notes" aria-label="Daftar catatan">
        <div class="notes-title"><span>Catatan</span><button onclick={createNote} title="Catatan baru">＋</button></div>
        {#if loading}
          <div class="state">Memuat…</div>
        {:else if !notes.length}
          <div class="state">Belum ada catatan.</div>
        {:else}
          {#each notes as note}
            <button class:chosen={activeNote?.id === note.id} class="note" onclick={() => activeNote = note}><strong>{note.title || 'Tanpa judul'}</strong><small>{note.content.slice(0, 72) || 'Belum ada isi'}</small></button>
          {/each}
        {/if}
      </nav>
      <section class="editor" aria-live="polite">
        {#if error}
          <div class="error" role="alert">{error}</div>
        {/if}
        {#if activeNote}
          <input class="title" bind:value={activeNote.title} onblur={saveNote} aria-label="Judul catatan" />
          <textarea bind:value={activeNote.content} onblur={saveNote} placeholder="Mulai menulis…" aria-label="Isi catatan"></textarea>
        {:else}
          <div class="empty"><div class="empty-mark">✦</div><h2>Mulai mencatat</h2><p>Buat catatan teks yang tetap berada di perangkat ini.</p><button class="new" onclick={createNote}>＋ Catatan baru</button></div>
        {/if}
      </section>
    </div>
  </main>
</div>

<style>
  :global(*) { box-sizing: border-box; }
  :global(body) { margin:0; font-family: Inter, ui-sans-serif, system-ui, sans-serif; background:#f7f8fa; color:#202124; }
  button,input,textarea { font:inherit; }
  button:focus-visible,input:focus-visible,textarea:focus-visible { outline:2px solid #202124; outline-offset:2px; }
  .app { min-height:100vh; display:grid; grid-template-columns:250px 1fr; }
  .sidebar { padding:24px 16px; background:#f1f3f4; border-right:1px solid #e2e5e8; }
  .brand { display:flex; align-items:center; gap:10px; font-weight:700; margin:0 8px 24px; }
  .mark,.empty-mark { width:32px;height:32px;border-radius:10px;display:grid;place-items:center;background:#202124;color:white; }
  .new { border:0; border-radius:20px; padding:10px 15px; background:#fff; box-shadow:0 1px 4px #0001; cursor:pointer; }
  .section-label,.eyebrow { font-size:11px; letter-spacing:.12em; color:#73777c; font-weight:700; }
  .section-label { margin:28px 8px 8px; }
  .book,.note,.icon,.notes-title button { width:100%; border:0; background:transparent; text-align:left; cursor:pointer; }
  .book { padding:10px 12px; border-radius:9px; }
  .book.active { background:#e2e5e9; font-weight:600; }
  .workspace { min-width:0; }
  header { height:92px; padding:20px 34px; display:flex; justify-content:space-between; align-items:center; background:#fff; border-bottom:1px solid #e2e5e8; }
  h1 { font-size:22px; margin:5px 0 0; }
  .icon { width:40px; text-align:center; font-size:25px; }
  .body { height:calc(100vh - 92px); display:grid; grid-template-columns:260px 1fr; }
  .notes { padding:18px 12px; background:#fbfbfc; border-right:1px solid #e2e5e8; overflow:auto; }
  .notes-title { display:flex; justify-content:space-between; align-items:center; padding:6px 10px 12px; font-weight:700; }
  .notes-title button { width:auto; font-size:20px; }
  .note { padding:12px; border-radius:10px; display:block; }
  .note.chosen { background:#e8eaed; }
  .note strong,.note small { display:block; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
  .note small { color:#777; margin-top:4px; }
  .state { padding:12px 10px; color:#777; font-size:14px; }
  .editor { padding:52px clamp(24px,8vw,120px); background:#fff; overflow:auto; }
  .error { margin-bottom:20px; padding:10px 12px; border-radius:8px; background:#fce8e6; color:#8a1c13; font-size:14px; }
  .editor-toolbar { display:flex; justify-content:space-between; align-items:center; margin-bottom:14px; }
  .save-state { color:#686c72; font-size:13px; }
  .delete { border:0; background:transparent; color:#8a1c13; cursor:pointer; padding:8px; }
  .title { width:100%; border:0; outline:0; font-size:32px; font-weight:700; margin-bottom:20px; background:transparent; }
  textarea { width:100%; min-height:65vh; resize:none; border:0; outline:0; line-height:1.8; font-size:16px; background:transparent; }
  .empty { max-width:460px; margin:14vh auto; text-align:center; color:#686c72; }
  .empty-mark { margin:auto; }
  .empty h2 { color:#202124; margin-bottom:8px; }
  @media (max-width:760px) { .app{grid-template-columns:1fr}.sidebar{display:none}.body{grid-template-columns:1fr}.notes{display:none}.editor{padding:32px 20px}header{padding:18px 20px} }
</style>
