<script lang="ts">
  type Notebook = { id: string; title: string; description: string };
  type Note = { id: string; title: string; content: string };

  let notebooks: Notebook[] = [];
  let notes: Note[] = [];
  let selected = '';
  let activeNote: Note | null = null;
  let loading = true;

  async function loadNotebooks() {
    notebooks = await (await fetch('/api/notebooks')).json();
    if (!selected && notebooks.length) selected = notebooks[0].id;
    if (selected) await loadNotes();
    loading = false;
  }

  async function loadNotes() {
    notes = await (await fetch(`/api/notebooks/${selected}/notes`)).json();
    activeNote = notes[0] ?? null;
  }

  async function createNotebook() {
    const title = prompt('Nama buku');
    if (!title?.trim()) return;
    const response = await fetch('/api/notebooks', { method: 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify({title, description:''}) });
    const notebook = await response.json();
    notebooks = [notebook, ...notebooks]; selected = notebook.id; await loadNotes();
  }

  async function createNote() {
    if (!selected) return;
    const response = await fetch(`/api/notebooks/${selected}/notes`, { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({title:'Catatan baru', content:''}) });
    activeNote = await response.json(); notes = [activeNote, ...notes];
  }

  async function saveNote() {
    if (!activeNote) return;
    await fetch(`/api/notes/${activeNote.id}`, { method:'PUT', headers:{'Content-Type':'application/json'}, body:JSON.stringify(activeNote) });
  }

  $: if (selected && !loading) loadNotes();
  loadNotebooks();
</script>

<svelte:head><title>Buku Catatan</title></svelte:head>

<div class="app">
  <aside class="sidebar">
    <div class="brand"><span class="mark">B</span><span>Buku Catatan</span></div>
    <button class="new" onclick={createNotebook}>＋ Buku baru</button>
    <div class="section-label">BUKU</div>
    {#each notebooks as notebook}
      <button class:active={selected === notebook.id} class="book" onclick={() => selected = notebook.id}>{notebook.title}</button>
    {/each}
  </aside>

  <main class="workspace">
    <header><div><div class="eyebrow">RUANG CATATAN</div><h1>{notebooks.find(n => n.id === selected)?.title ?? 'Buku Catatan'}</h1></div><button class="icon" onclick={createNote} title="Catatan baru">＋</button></header>
    <div class="body">
      <nav class="notes">
        <div class="notes-title"><span>Catatan</span><button onclick={createNote}>＋</button></div>
        {#each notes as note}
          <button class:chosen={activeNote?.id === note.id} class="note" onclick={() => activeNote = note}><strong>{note.title || 'Tanpa judul'}</strong><small>{note.content.slice(0, 72) || 'Belum ada isi'}</small></button>
        {/each}
      </nav>
      <section class="editor">
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
  .editor { padding:52px clamp(24px,8vw,120px); background:#fff; overflow:auto; }
  .title { width:100%; border:0; outline:0; font-size:32px; font-weight:700; margin-bottom:20px; background:transparent; }
  textarea { width:100%; min-height:65vh; resize:none; border:0; outline:0; line-height:1.8; font-size:16px; background:transparent; }
  .empty { max-width:460px; margin:14vh auto; text-align:center; color:#686c72; }
  .empty-mark { margin:auto; }
  .empty h2 { color:#202124; margin-bottom:8px; }
  @media (max-width:760px) { .app{grid-template-columns:1fr}.sidebar{display:none}.body{grid-template-columns:1fr}.notes{display:none}.editor{padding:32px 20px}header{padding:18px 20px} }
</style>
